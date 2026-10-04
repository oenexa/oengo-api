require('dotenv').config();
const express = require('express');
const axios = require('axios');
const cors = require('cors');
const crypto = require('crypto');

const app = express();
app.use(cors());
app.use(express.json());

const PORT = process.env.PORT || 3001;
const OENEXA_RPC_URL = process.env.OENEXA_RPC_URL || 'http://localhost:8545';

// ── In-Memory Datastores (Production uses PostgreSQL / Redis) ───────────────
const users = {
    'user_customer': {
        id: 'user_customer',
        name: 'Alice Customer',
        role: 'CUSTOMER',
        digitalWallet: {
            fiatBalanceEUR: 50.00,
            loyaltyPoints: 120,
            cashbackRatePct: 5
        },
        cryptoWalletAddress: '0xAlice_Customer_MLDSA65'
    },
    'user_restaurant': {
        id: 'user_restaurant',
        name: 'Napoli Woodfire Pizza',
        role: 'RESTAURANT',
        digitalWallet: {
            fiatBalanceEUR: 1250.00,
            loyaltyPoints: 0
        },
        cryptoWalletAddress: '0xNapoli_Restaurant_MLDSA65'
    },
    'user_courier': {
        id: 'user_courier',
        name: 'Bob Delivery Rider',
        role: 'COURIER',
        digitalWallet: {
            fiatBalanceEUR: 85.50,
            loyaltyPoints: 40
        },
        cryptoWalletAddress: '0xBob_Rider_MLDSA65'
    }
};

const orders = {};

// ── Commission Configuration (Manually Settable, Default: 5%) ────────────────
let defaultCommissionPct = 5;
const MIN_COMMISSION_PCT = 0;
const MAX_COMMISSION_PCT = 30;

// In-Memory Saved Cards for Instant 1-Tap Checkout
const savedCards = {
    'user_customer': [
        {
            id: 'card_default_visa',
            brand: 'Visa',
            last4: '4242',
            expMonth: 12,
            expYear: 2028,
            cardHolder: 'Alice Customer',
            isDefault: true
        }
    ]
};

// Helper: Card Brand Detector
function detectCardBrand(number) {
    const cleaned = (number || '').replace(/\D/g, '');
    if (/^4/.test(cleaned)) return 'Visa';
    if (/^5[1-5]/.test(cleaned) || /^2[2-7]/.test(cleaned)) return 'Mastercard';
    if (/^3[47]/.test(cleaned)) return 'American Express';
    if (/^6(?:011|5)/.test(cleaned)) return 'Discover';
    return 'Credit Card';
}

// Helper: SHA256
function sha256(str) {
    return crypto.createHash('sha256').update(str).digest('hex');
}

// ── ADMIN COMMISSION CONFIGURATION APIS ──────────────────────────────────────

// GET /api/admin/commission - Query current platform commission rate
app.get('/api/admin/commission', (req, res) => {
    res.json({
        success: true,
        commissionPct: defaultCommissionPct,
        minPct: MIN_COMMISSION_PCT,
        maxPct: MAX_COMMISSION_PCT,
        restaurantSharePct: 100 - defaultCommissionPct
    });
});

// POST /api/admin/commission - Manually update global platform commission
app.post('/api/admin/commission', (req, res) => {
    const { ratePct } = req.body;
    const num = Number(ratePct);
    if (isNaN(num) || num < MIN_COMMISSION_PCT || num > MAX_COMMISSION_PCT) {
        return res.status(400).json({
            error: `Invalid commission rate. Must be a number between ${MIN_COMMISSION_PCT}% and ${MAX_COMMISSION_PCT}%.`
        });
    }
    defaultCommissionPct = num;
    res.json({
        success: true,
        message: `Platform commission rate updated to ${defaultCommissionPct}%`,
        commissionPct: defaultCommissionPct,
        restaurantSharePct: 100 - defaultCommissionPct
    });
});

// ── PAYMENT GATEWAY: DIRECT CREDIT CARD & MULTI-RAIL APIS ───────────────────

// POST /api/payments/card-intent - Initialize card payment intent
app.post('/api/payments/card-intent', (req, res) => {
    const { amount, currency = 'EUR', customerId = 'user_customer' } = req.body;
    const intentId = `pi_${Date.now()}_${crypto.randomBytes(4).toString('hex')}`;
    const clientSecret = `sec_${crypto.randomBytes(16).toString('hex')}`;
    res.json({
        success: true,
        paymentIntentId: intentId,
        clientSecret,
        amount: parseFloat(amount || 0),
        currency,
        status: 'REQUIRES_PAYMENT_METHOD'
    });
});

// POST /api/payments/confirm-card - Authorize & capture instant credit card payment
app.post('/api/payments/confirm-card', (req, res) => {
    const {
        paymentIntentId,
        cardNumber = '',
        cardHolderName = '',
        cardExpMonth = '12',
        cardExpYear = '2028',
        cardCvc = '',
        saveCard = false,
        customerId = 'user_customer'
    } = req.body;

    const cleanedNumber = cardNumber.replace(/\D/g, '');
    if (cleanedNumber.length < 13 || cleanedNumber.length > 19) {
        return res.status(400).json({ error: 'Invalid card number length (must be 13-19 digits)' });
    }
    if (!cardHolderName || !cardHolderName.trim()) {
        return res.status(400).json({ error: 'Cardholder name is required' });
    }
    if (!cardCvc || cardCvc.length < 3 || cardCvc.length > 4) {
        return res.status(400).json({ error: 'Invalid CVV/CVC security code' });
    }

    const brand = detectCardBrand(cleanedNumber);
    const last4 = cleanedNumber.slice(-4);
    const transactionId = `txn_${Date.now()}_${crypto.randomBytes(4).toString('hex')}`;

    const cardInfo = {
        id: `card_${Date.now()}`,
        brand,
        last4,
        expMonth: parseInt(cardExpMonth, 10),
        expYear: parseInt(cardExpYear, 10),
        cardHolder: cardHolderName.trim()
    };

    if (saveCard) {
        if (!savedCards[customerId]) savedCards[customerId] = [];
        savedCards[customerId].push(cardInfo);
    }

    res.json({
        success: true,
        message: `Credit card payment authorized and captured instantly via ${brand}`,
        transactionId,
        paymentIntentId: paymentIntentId || `pi_${Date.now()}`,
        brand,
        last4,
        status: 'SUCCEEDED'
    });
});

// GET /api/payments/methods/:userId - List saved customer payment methods
app.get('/api/payments/methods/:userId', (req, res) => {
    const methods = savedCards[req.params.userId] || [
        { id: 'card_demo_visa', brand: 'Visa', last4: '4242', expMonth: 12, expYear: 2028, cardHolder: 'Alice Customer', isDefault: true }
    ];
    res.json({ success: true, methods });
});

// ── 1. DUAL-WALLET APIS ───────────────────────────────────────────────────────

// GET /api/wallet/:userId - Unified Digital & Crypto balance inquiry
app.get('/api/wallet/:userId', async (req, res) => {
    const user = users[req.params.userId];
    if (!user) {
        return res.status(404).json({ error: 'User account not found' });
    }

    let onChainBalanceOEN = '100.00000000'; // Default simulated balance

    try {
        // Query live on-chain balance from pure OENEXA Layer-1 Node
        const rpcRes = await axios.post(OENEXA_RPC_URL, {
            jsonrpc: '2.0',
            method: 'oen_getBalance',
            params: [user.cryptoWalletAddress, 'latest'],
            id: Date.now()
        }, { timeout: 2000 });

        if (rpcRes.data && rpcRes.data.result) {
            onChainBalanceOEN = rpcRes.data.result;
        }
    } catch {
        // Fallback to simulated balance if node is initializing
    }

    res.json({
        success: true,
        userId: user.id,
        name: user.name,
        role: user.role,
        wallets: {
            digital: {
                fiatEUR: user.digitalWallet.fiatBalanceEUR,
                loyaltyPoints: user.digitalWallet.loyaltyPoints,
                type: 'OFF_CHAIN_CREDITS'
            },
            crypto: {
                address: user.cryptoWalletAddress,
                balanceOEN: onChainBalanceOEN,
                type: 'NON_CUSTODIAL_ML_DSA_65'
            }
        }
    });
});

// POST /api/wallet/topup - Top up digital fiat balance or loyalty points
app.post('/api/wallet/topup', (req, res) => {
    const { userId, amountEUR, points } = req.body;
    const user = users[userId];
    if (!user) return res.status(404).json({ error: 'User not found' });

    if (amountEUR) user.digitalWallet.fiatBalanceEUR += parseFloat(amountEUR);
    if (points) user.digitalWallet.loyaltyPoints += parseInt(points, 10);

    res.json({
        success: true,
        message: 'Wallet topped up successfully',
        digitalWallet: user.digitalWallet
    });
});

// ── 2. ORDER & ESCROW LIFECYCLE APIS ─────────────────────────────────────────

// POST /api/orders - Customer places order (Funds locked in Escrow)
app.post('/api/orders', async (req, res) => {
    try {
        const {
            buyerId = 'user_customer',
            restaurantId = 'user_restaurant',
            items = [],
            amount = 25.00,
            deliveryFee = 3.50,
            tip = 2.00,
            paymentMethod = 'CREDIT_CARD', // 'CREDIT_CARD' | 'DIGITAL_WALLET' | 'CRYPTO_OEN'
            cardPayment = null,
            commissionPct
        } = req.body;

        const buyer = users[buyerId];
        const restaurant = users[restaurantId];
        if (!buyer || !restaurant) {
            return res.status(400).json({ error: 'Invalid buyer or restaurant ID' });
        }

        // Determine active commission percentage (default: 5%, bounded between 0% and 30%)
        let activeCommissionPct = defaultCommissionPct;
        if (commissionPct !== undefined && commissionPct !== null) {
            const parsed = Number(commissionPct);
            if (!isNaN(parsed) && parsed >= MIN_COMMISSION_PCT && parsed <= MAX_COMMISSION_PCT) {
                activeCommissionPct = parsed;
            }
        }

        const orderId = `ord_${Date.now()}`;
        const pickupBarcode = `PKG-${orderId.slice(-6).toUpperCase()}`;
        const deliveryPin = `${Math.floor(1000 + Math.random() * 9000)}`; // 4-digit secret PIN

        const newOrder = {
            id: orderId,
            buyerId,
            buyerAddress: buyer.cryptoWalletAddress,
            restaurantId,
            restaurantAddress: restaurant.cryptoWalletAddress,
            courierId: null,
            courierAddress: null,
            items,
            amount: parseFloat(amount),
            deliveryFee: parseFloat(deliveryFee),
            tip: parseFloat(tip),
            total: parseFloat(amount) + parseFloat(deliveryFee) + parseFloat(tip),
            commissionPct: activeCommissionPct,
            paymentMethod,
            cardPayment: paymentMethod === 'CREDIT_CARD' ? cardPayment : null,
            paymentStatus: 'PAID',
            status: 'AWAITING_RESTAURANT', // AWAITING_RESTAURANT -> PREPARING -> IN_TRANSIT -> DELIVERED
            pickupBarcode,
            pickupBarcodeHash: sha256(pickupBarcode),
            deliveryPin,
            deliveryPinHash: sha256(deliveryPin),
            createdAt: new Date().toISOString(),
            escrowLocked: true
        };

        orders[orderId] = newOrder;

        const methodLabel = paymentMethod === 'CREDIT_CARD' && cardPayment
            ? `Credit Card (${cardPayment.brand} •••• ${cardPayment.last4})`
            : paymentMethod;

        res.status(201).json({
            success: true,
            message: `Order created via ${methodLabel} and funds locked in smart contract escrow`,
            order: newOrder
        });
    } catch (err) {
        res.status(500).json({ error: err.message });
    }
});

// GET /api/orders - List all orders (Filtered by role if requested)
app.get('/api/orders', (req, res) => {
    const list = Object.values(orders);
    res.json({ success: true, count: list.length, orders: list });
});

// GET /api/orders/:id - Order details
app.get('/api/orders/:id', (req, res) => {
    const order = orders[req.params.id];
    if (!order) return res.status(404).json({ error: 'Order not found' });
    res.json({ success: true, order });
});

// POST /api/orders/:id/accept - Restaurant accepts order & starts cooking
app.post('/api/orders/:id/accept', (req, res) => {
    const order = orders[req.params.id];
    if (!order) return res.status(404).json({ error: 'Order not found' });
    if (order.status !== 'AWAITING_RESTAURANT') {
        return res.status(400).json({ error: `Cannot accept order with status ${order.status}` });
    }

    order.status = 'PREPARING';
    order.prepEtaMinutes = req.body.prepEtaMinutes || 15;
    res.json({ success: true, message: 'Order accepted, kitchen preparing', order });
});

// POST /api/orders/:id/assign-courier - Proximity dispatch assigns courier
app.post('/api/orders/:id/assign-courier', (req, res) => {
    const order = orders[req.params.id];
    if (!order) return res.status(404).json({ error: 'Order not found' });

    const courierId = req.body.courierId || 'user_courier';
    const courier = users[courierId];
    if (!courier) return res.status(400).json({ error: 'Courier not found' });

    order.courierId = courierId;
    order.courierAddress = courier.cryptoWalletAddress;
    res.json({ success: true, message: `Courier ${courier.name} assigned`, order });
});

// POST /api/orders/:id/confirm-pickup - Courier scans package barcode at restaurant
app.post('/api/orders/:id/confirm-pickup', (req, res) => {
    const { scannedBarcode } = req.body;
    const order = orders[req.params.id];
    if (!order) return res.status(404).json({ error: 'Order not found' });

    if (order.status !== 'PREPARING') {
        return res.status(400).json({ error: `Order not ready for pickup (current: ${order.status})` });
    }

    if (sha256(scannedBarcode) !== order.pickupBarcodeHash) {
        return res.status(400).json({ error: 'Invalid pickup barcode verification failed' });
    }

    order.status = 'IN_TRANSIT';
    order.pickedUpAt = new Date().toISOString();
    res.json({ success: true, message: 'Pickup confirmed! Order is now IN_TRANSIT', order });
});

// POST /api/orders/:id/confirm-delivery - Courier scans delivery barcode or enters PIN
app.post('/api/orders/:id/confirm-delivery', async (req, res) => {
    const { proofCode } = req.body; // Can be 4-digit PIN or scanned customer delivery barcode
    const order = orders[req.params.id];
    if (!order) return res.status(404).json({ error: 'Order not found' });

    if (order.status !== 'IN_TRANSIT') {
        return res.status(400).json({ error: `Order must be IN_TRANSIT to confirm delivery (current: ${order.status})` });
    }

    // Verify proof
    if (sha256(proofCode) !== order.deliveryPinHash && proofCode !== order.deliveryPin) {
        return res.status(400).json({ error: 'Proof-of-delivery failed: Barcode or PIN incorrect' });
    }

    // Autonomous Settlement Math based on order's locked commission percentage:
    const commPct = typeof order.commissionPct === 'number' ? order.commissionPct : defaultCommissionPct;
    const platformCut = parseFloat(((order.amount * commPct) / 100).toFixed(2));
    const restaurantPayout = parseFloat((order.amount - platformCut).toFixed(2));
    const courierPayout = parseFloat((platformCut + order.deliveryFee + order.tip).toFixed(2));

    order.status = 'DELIVERED';
    order.settledAt = new Date().toISOString();
    order.commissionPct = commPct;
    order.platformCut = platformCut;
    order.restaurantPayout = restaurantPayout;
    order.courierPayout = courierPayout;

    // Credit Customer Loyalty Cashback into Digital Wallet (5% of order subtotal)
    const cashback = parseFloat((order.amount * 0.05).toFixed(2));
    const customer = users[order.buyerId];
    if (customer) {
        customer.digitalWallet.fiatBalanceEUR += cashback;
        customer.digitalWallet.loyaltyPoints += Math.floor(order.amount);
    }

    // Disburse to restaurant & courier digital wallets if applicable
    const restaurant = users[order.restaurantId];
    if (restaurant) restaurant.digitalWallet.fiatBalanceEUR += restaurantPayout;
    const courier = users[order.courierId];
    if (courier) courier.digitalWallet.fiatBalanceEUR += courierPayout;

    res.json({
        success: true,
        message: `Order delivered & escrow autonomously settled on-chain! [Commission: ${commPct}%]`,
        settlement: {
            commissionPct: commPct,
            platformFeeEUR: platformCut.toFixed(2),
            restaurantPayoutEUR: restaurantPayout.toFixed(2),
            courierPayoutEUR: courierPayout.toFixed(2),
            customerCashbackEUR: cashback.toFixed(2),
            customerLoyaltyPointsAwarded: Math.floor(order.amount)
        },
        order
    });
});

// POST /api/rpc/broadcast - Web3 Gateway Proxy to pure OENEXA Layer-1 Node
app.post('/api/rpc/broadcast', async (req, res) => {
    try {
        const { signedTxHex } = req.body;
        const response = await axios.post(OENEXA_RPC_URL, {
            jsonrpc: '2.0',
            method: 'oen_sendRawTransaction',
            params: [signedTxHex],
            id: Date.now()
        }, { timeout: 5000 });
        res.json(response.data);
    } catch (error) {
        res.status(500).json({ error: 'Failed to broadcast transaction to OENEXA network', detail: error.message });
    }
});

// Server boot
app.listen(PORT, () => {
    console.log(`🍔 Oengo App Food Delivery Backend running on http://localhost:${PORT}`);
    console.log(`🔗 Connected to pure OENEXA node at ${OENEXA_RPC_URL}`);
});
