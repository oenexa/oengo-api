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

// ── Restaurant Partner Profiles ─────────────────────────────────────────────
const restaurants = {
    'user_restaurant': {
        id: 'user_restaurant',
        name: 'Napoli Woodfire Pizza',
        tagline: 'Authentic Neapolitan Pizza & Artisan Italian Delicacies',
        cuisine: 'Italian, Pizza, Artisan',
        rating: 4.9,
        reviewCount: 342,
        address: 'Via Toledo 42, Napoli / Historic District',
        icon: '🍕',
        isOpen: true,
        prepEtaMinutes: 15,
        deliveryFeeEUR: 2.50,
        minOrderEUR: 15.00,
        deliveryRadiusKm: 5.5,
        cryptoWalletAddress: '0xNapoli_Restaurant_MLDSA65',
        fiatBalanceEUR: 1250.00
    },
    'rest_burger': {
        id: 'rest_burger',
        name: 'Smash & Co. Gourmet Burgers',
        tagline: 'Double Smashed Dry-Aged Angus & Loaded Brioche',
        cuisine: 'American, Burgers, Craft',
        rating: 4.8,
        reviewCount: 218,
        address: 'Corso Umberto I 118, Napoli',
        icon: '🍔',
        isOpen: true,
        prepEtaMinutes: 20,
        deliveryFeeEUR: 2.99,
        minOrderEUR: 12.00,
        deliveryRadiusKm: 6.0,
        cryptoWalletAddress: '0xSmashBurger_MLDSA65',
        fiatBalanceEUR: 890.00
    },
    'rest_sushi': {
        id: 'rest_sushi',
        name: 'Tokyo Bloom Omakase & Bento',
        tagline: 'Sustainably Sourced Sashimi, Nigiri & Crispy Gyoza',
        cuisine: 'Japanese, Sushi, Asian',
        rating: 4.9,
        reviewCount: 451,
        address: 'Via Chiaia 84, Napoli',
        icon: '🍣',
        isOpen: true,
        prepEtaMinutes: 25,
        deliveryFeeEUR: 3.50,
        minOrderEUR: 20.00,
        deliveryRadiusKm: 7.0,
        cryptoWalletAddress: '0xTokyoBloom_MLDSA65',
        fiatBalanceEUR: 2140.00
    },
    'rest_green': {
        id: 'rest_green',
        name: 'Verde Organic Bowls & Cold Press',
        tagline: 'Superfood Macro Bowls, Wild Greens & Fresh Smoothies',
        cuisine: 'Healthy, Vegan, Organic',
        rating: 4.7,
        reviewCount: 129,
        address: 'Piazza dei Martiri 16, Napoli',
        icon: '🥗',
        isOpen: true,
        prepEtaMinutes: 12,
        deliveryFeeEUR: 1.99,
        minOrderEUR: 10.00,
        deliveryRadiusKm: 4.5,
        cryptoWalletAddress: '0xVerdeOrganic_MLDSA65',
        fiatBalanceEUR: 620.00
    }
};

// ── Live Menu & Catalog Datastore ───────────────────────────────────────────
const menus = {
    'user_restaurant': [
        {
            id: 'menu_margherita',
            category: 'Pizza & Mains',
            name: 'Artisanal Margherita Pizza',
            description: 'San Marzano D.O.P. tomatoes, fresh buffalo mozzarella, fragrant basil, extra virgin olive oil.',
            priceEUR: 16.50,
            priceOEN: '1.21',
            inStock: true,
            prepMinutes: 12,
            badge: 'Bestseller'
        },
        {
            id: 'menu_diavola',
            category: 'Pizza & Mains',
            name: 'Spicy Diavola Pizza',
            description: 'Spianata Calabrese spicy salami, smoked provolone, chili flakes, organic tomato reduction.',
            priceEUR: 18.00,
            priceOEN: '1.32',
            inStock: true,
            prepMinutes: 14,
            badge: 'Spicy'
        },
        {
            id: 'menu_arancini',
            category: 'Starters',
            name: 'Truffle & Porcini Arancini',
            description: 'Crispy golden saffron risotto balls stuffed with black truffle cream and melted fontina cheese.',
            priceEUR: 12.00,
            priceOEN: '0.88',
            inStock: true,
            prepMinutes: 8,
            badge: 'Vegetarian'
        },
        {
            id: 'menu_burrata',
            category: 'Starters',
            name: 'Pugliese Burrata & Heirloom Salad',
            description: 'Creamy artisanal burrata, heirloom cherry tomatoes, aged balsamic glaze, toasted pine nuts.',
            priceEUR: 14.00,
            priceOEN: '1.03',
            inStock: true,
            prepMinutes: 6,
            badge: 'Chef Special'
        },
        {
            id: 'menu_tiramisu',
            category: 'Desserts',
            name: 'Classic Espresso Tiramisù',
            description: 'Layered savoiardi soaked in single-origin espresso and Marsala, mascarpone cream, dark cocoa.',
            priceEUR: 8.50,
            priceOEN: '0.62',
            inStock: true,
            prepMinutes: 5,
            badge: 'Homemade'
        },
        {
            id: 'menu_sanpellegrino',
            category: 'Beverages',
            name: 'San Pellegrino Blood Orange',
            description: 'Sparkling Italian citrus beverage crafted with sun-ripened Sicilian blood oranges.',
            priceEUR: 4.50,
            priceOEN: '0.33',
            inStock: true,
            prepMinutes: 2,
            badge: 'Cold'
        }
    ],
    'rest_burger': [
        {
            id: 'menu_smash_truffle',
            category: 'Burgers',
            name: 'Double Truffle Smash Burger',
            description: 'Two 100g dry-aged Angus patties, double American cheese, black truffle aioli, grilled onions, brioche.',
            priceEUR: 15.50,
            priceOEN: '1.14',
            inStock: true,
            prepMinutes: 12,
            badge: 'Bestseller'
        },
        {
            id: 'menu_bacon_bbq',
            category: 'Burgers',
            name: 'Smoked Bacon BBQ Burger',
            description: 'Crispy applewood smoked bacon, aged cheddar, bourbon BBQ glaze, house pickles.',
            priceEUR: 14.50,
            priceOEN: '1.07',
            inStock: true,
            prepMinutes: 10,
            badge: 'Popular'
        },
        {
            id: 'menu_parm_fries',
            category: 'Sides',
            name: 'Parmesan & Rosemary Fries',
            description: 'Triple-cooked rustic skin-on potatoes dusted with 24-month Parmigiano Reggiano and fresh rosemary.',
            priceEUR: 6.00,
            priceOEN: '0.44',
            inStock: true,
            prepMinutes: 6,
            badge: 'Crispy'
        }
    ],
    'rest_sushi': [
        {
            id: 'menu_omakase_set',
            category: 'Nigiri & Sets',
            name: 'Chef Omakase 12-Piece Set',
            description: 'Otoro fatty tuna, Scottish salmon, yellowtail hamachi, Hokkaido scallop, sweet unagi eel.',
            priceEUR: 28.00,
            priceOEN: '2.06',
            inStock: true,
            prepMinutes: 18,
            badge: 'Chef Special'
        },
        {
            id: 'menu_dragon_roll',
            category: 'Special Rolls',
            name: 'Crunchy Tiger Dragon Roll',
            description: 'Tempura black tiger prawn, Hass avocado, cucumber, unagi reduction, flying fish tobiko.',
            priceEUR: 16.00,
            priceOEN: '1.18',
            inStock: true,
            prepMinutes: 14,
            badge: 'Bestseller'
        },
        {
            id: 'menu_gyoza',
            category: 'Starters',
            name: 'Crispy Duck & Ponzu Gyoza',
            description: 'Pan-seared handmade dumplings filled with confit duck and ginger scallion, citrus ponzu dip.',
            priceEUR: 9.50,
            priceOEN: '0.70',
            inStock: true,
            prepMinutes: 8,
            badge: 'Hot'
        }
    ],
    'rest_green': [
        {
            id: 'menu_bali_bowl',
            category: 'Superfood Bowls',
            name: 'Bali Sunset Macro Bowl',
            description: 'Tri-color organic quinoa, Hass avocado, edamame, roasted sweet potato, toasted nori, tahini miso.',
            priceEUR: 13.50,
            priceOEN: '0.99',
            inStock: true,
            prepMinutes: 8,
            badge: 'Organic'
        },
        {
            id: 'menu_green_cleanse',
            category: 'Cold Press',
            name: 'Immunity Cold Press Elixir',
            description: 'Cold-pressed Tuscan kale, Granny Smith apple, fresh ginger root, Meyer lemon, organic celery.',
            priceEUR: 6.50,
            priceOEN: '0.48',
            inStock: true,
            prepMinutes: 3,
            badge: 'Detox'
        }
    ]
};

// Seed an initial demo order for the kitchen KDS
const seedOrder = {
    id: 'ord_seed_101',
    buyerId: 'user_customer',
    buyerName: 'Alice Customer',
    buyerAddress: '0xAlice_Customer_MLDSA65',
    restaurantId: 'user_restaurant',
    restaurantAddress: '0xNapoli_Restaurant_MLDSA65',
    courierId: null,
    courierAddress: null,
    items: [
        { name: 'Artisanal Margherita Pizza', qty: 1, price: 16.50 },
        { name: 'Truffle & Porcini Arancini', qty: 1, price: 12.00 }
    ],
    amount: 28.50,
    deliveryFee: 3.50,
    tip: 2.00,
    total: 34.00,
    commissionPct: 5,
    paymentMethod: 'CREDIT_CARD',
    cardPayment: { brand: 'Visa', last4: '4242', transactionId: 'txn_seed_4242' },
    paymentStatus: 'PAID',
    status: 'AWAITING_RESTAURANT',
    pickupBarcode: 'PKG-DEMO01',
    pickupBarcodeHash: crypto.createHash('sha256').update('PKG-DEMO01').digest('hex'),
    deliveryPin: '4821',
    deliveryPinHash: crypto.createHash('sha256').update('4821').digest('hex'),
    createdAt: new Date(Date.now() - 4 * 60000).toISOString(),
    escrowLocked: true
};
orders[seedOrder.id] = seedOrder;

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
    order.acceptedAt = new Date().toISOString();
    res.json({ success: true, message: 'Order accepted, kitchen preparing', order });
});

// POST /api/orders/:id/ready - Kitchen marks order ready for courier counter pickup
app.post('/api/orders/:id/ready', (req, res) => {
    const order = orders[req.params.id];
    if (!order) return res.status(404).json({ error: 'Order not found' });
    if (order.status !== 'PREPARING') {
        return res.status(400).json({ error: `Order cannot be marked ready from status ${order.status}` });
    }

    order.status = 'READY_FOR_PICKUP';
    order.readyAt = new Date().toISOString();
    res.json({
        success: true,
        message: 'Order is packed and ready for courier counter pickup',
        pickupBarcode: order.pickupBarcode,
        order
    });
});

// POST /api/orders/:id/decline - Restaurant declines order & triggers automatic escrow refund
app.post('/api/orders/:id/decline', (req, res) => {
    const order = orders[req.params.id];
    if (!order) return res.status(404).json({ error: 'Order not found' });
    if (order.status !== 'AWAITING_RESTAURANT' && order.status !== 'PREPARING') {
        return res.status(400).json({ error: `Order cannot be declined from status ${order.status}` });
    }

    const { reason = 'Kitchen at peak capacity' } = req.body;
    order.status = 'CANCELLED_BY_RESTAURANT';
    order.cancelReason = reason;
    order.escrowLocked = false;
    order.refundStatus = 'REFUNDED_100_PERCENT';
    order.cancelledAt = new Date().toISOString();

    // If paid via digital wallet, refund immediately to user's wallet
    if (order.paymentMethod === 'DIGITAL_WALLET') {
        const buyer = users[order.buyerId];
        if (buyer) buyer.digitalWallet.fiatBalanceEUR += order.total;
    }

    res.json({
        success: true,
        message: `Order declined by kitchen (${reason}). Escrow 100% refunded to customer.`,
        order
    });
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

    if (order.status !== 'PREPARING' && order.status !== 'READY_FOR_PICKUP') {
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

// ── 3. RESTAURANT PARTNER PORTAL & MENU CATALOG APIS ─────────────────────────

// GET /api/restaurants - List all partner restaurants (supports ?cuisine=... and ?search=...)
app.get('/api/restaurants', (req, res) => {
    const { cuisine, search } = req.query;
    let list = Object.values(restaurants);

    if (cuisine && cuisine !== 'ALL') {
        const cLower = cuisine.toLowerCase();
        list = list.filter(r => r.cuisine.toLowerCase().includes(cLower));
    }

    if (search) {
        const sLower = search.toLowerCase();
        list = list.filter(r => 
            r.name.toLowerCase().includes(sLower) || 
            r.cuisine.toLowerCase().includes(sLower) ||
            r.tagline.toLowerCase().includes(sLower)
        );
    }

    res.json({
        success: true,
        count: list.length,
        restaurants: list
    });
});

// GET /api/restaurants/:id - Restaurant profile & live analytics
app.get('/api/restaurants/:id', (req, res) => {
    const restaurant = restaurants[req.params.id];
    if (!restaurant) return res.status(404).json({ error: 'Restaurant not found' });

    // Compute live operational stats
    const restaurantOrders = Object.values(orders).filter(o => o.restaurantId === restaurant.id);
    const activeOrders = restaurantOrders.filter(o => 
        ['AWAITING_RESTAURANT', 'PREPARING', 'READY_FOR_PICKUP', 'IN_TRANSIT'].includes(o.status)
    );
    const deliveredOrders = restaurantOrders.filter(o => o.status === 'DELIVERED');
    
    // Revenue calculations
    const grossDeliveredRevenueEUR = deliveredOrders.reduce((sum, o) => sum + (o.restaurantPayout || (o.amount * 0.95)), 0);
    const activePipelineRevenueEUR = activeOrders.reduce((sum, o) => sum + (o.amount * ((100 - (o.commissionPct || defaultCommissionPct)) / 100)), 0);
    const totalRevenueTodayEUR = parseFloat((grossDeliveredRevenueEUR + activePipelineRevenueEUR).toFixed(2));
    
    // Commission comparison: Oengo (5%) vs Legacy Delivery Apps (30%)
    const commissionRetainedPct = 100 - defaultCommissionPct;
    const legacyAppFeePct = 30;
    const legacyLostRevenueEUR = parseFloat((totalRevenueTodayEUR * (legacyAppFeePct - defaultCommissionPct) / 100).toFixed(2));

    res.json({
        success: true,
        restaurant: {
            ...restaurant,
            stats: {
                totalOrdersCount: restaurantOrders.length,
                activeOrdersCount: activeOrders.length,
                deliveredOrdersCount: deliveredOrders.length,
                grossRevenueTodayEUR: totalRevenueTodayEUR,
                fiatBalanceEUR: restaurant.fiatBalanceEUR,
                commissionRetainedPct,
                platformCommissionPct: defaultCommissionPct,
                legacyLostRevenueEUR,
                avgPrepTimeMinutes: restaurant.prepEtaMinutes || 15
            }
        }
    });
});

// PUT /api/restaurants/:id - Update restaurant operating status & prep times
app.put('/api/restaurants/:id', (req, res) => {
    const restaurant = restaurants[req.params.id];
    if (!restaurant) return res.status(404).json({ error: 'Restaurant not found' });

    const { isOpen, prepEtaMinutes, name, tagline, deliveryRadiusKm } = req.body;
    if (typeof isOpen === 'boolean') restaurant.isOpen = isOpen;
    if (typeof prepEtaMinutes === 'number') restaurant.prepEtaMinutes = prepEtaMinutes;
    if (name) restaurant.name = name;
    if (tagline) restaurant.tagline = tagline;
    if (typeof deliveryRadiusKm === 'number') restaurant.deliveryRadiusKm = deliveryRadiusKm;

    res.json({
        success: true,
        message: 'Restaurant profile updated successfully',
        restaurant
    });
});

// GET /api/restaurants/:id/menu - List all menu dishes & availability
app.get('/api/restaurants/:id/menu', (req, res) => {
    const menuList = menus[req.params.id] || [];
    res.json({
        success: true,
        restaurantId: req.params.id,
        count: menuList.length,
        menu: menuList
    });
});

// POST /api/restaurants/:id/menu - Add a new dish to the restaurant catalog
app.post('/api/restaurants/:id/menu', (req, res) => {
    const { name, category = 'Mains', description = '', priceEUR, prepMinutes = 15, badge = 'New' } = req.body;
    if (!name || !priceEUR) {
        return res.status(400).json({ error: 'Dish name and price (EUR) are required' });
    }

    if (!menus[req.params.id]) menus[req.params.id] = [];
    const priceNum = parseFloat(priceEUR);
    const newItem = {
        id: `menu_${Date.now()}`,
        category,
        name: name.trim(),
        description: description.trim(),
        priceEUR: priceNum,
        priceOEN: (priceNum / 13.60).toFixed(2), // 1 OEN ~= 13.60 EUR
        inStock: true,
        prepMinutes: parseInt(prepMinutes, 10) || 15,
        badge
    };

    menus[req.params.id].push(newItem);
    res.status(201).json({
        success: true,
        message: `Dish "${newItem.name}" added to catalog`,
        item: newItem
    });
});

// PUT /api/restaurants/:id/menu/:itemId - Update dish (availability toggle, price, details)
app.put('/api/restaurants/:id/menu/:itemId', (req, res) => {
    const menuList = menus[req.params.id];
    if (!menuList) return res.status(404).json({ error: 'Restaurant menu not found' });

    const item = menuList.find(i => i.id === req.params.itemId);
    if (!item) return res.status(404).json({ error: 'Menu item not found' });

    const { inStock, priceEUR, name, description, category, prepMinutes, badge } = req.body;
    if (typeof inStock === 'boolean') item.inStock = inStock;
    if (priceEUR !== undefined) {
        const p = parseFloat(priceEUR);
        item.priceEUR = p;
        item.priceOEN = (p / 13.60).toFixed(2);
    }
    if (name) item.name = name.trim();
    if (description !== undefined) item.description = description.trim();
    if (category) item.category = category;
    if (prepMinutes) item.prepMinutes = parseInt(prepMinutes, 10);
    if (badge) item.badge = badge;

    res.json({
        success: true,
        message: `Dish "${item.name}" updated`,
        item
    });
});

// DELETE /api/restaurants/:id/menu/:itemId - Remove dish from catalog
app.delete('/api/restaurants/:id/menu/:itemId', (req, res) => {
    const menuList = menus[req.params.id];
    if (!menuList) return res.status(404).json({ error: 'Restaurant menu not found' });

    const index = menuList.findIndex(i => i.id === req.params.itemId);
    if (index === -1) return res.status(404).json({ error: 'Menu item not found' });

    const removed = menuList.splice(index, 1)[0];
    res.json({
        success: true,
        message: `Dish "${removed.name}" deleted from menu`,
        deletedItemId: removed.id
    });
});

// GET /api/restaurants/:id/orders - List live orders for kitchen KDS
app.get('/api/restaurants/:id/orders', (req, res) => {
    const { status } = req.query;
    let list = Object.values(orders).filter(o => o.restaurantId === req.params.id);

    if (status) {
        const statuses = status.split(',').map(s => s.trim());
        list = list.filter(o => statuses.includes(o.status));
    }

    // Sort newest first
    list.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));

    res.json({
        success: true,
        count: list.length,
        orders: list
    });
});

// ── 4. CUSTOMER STOREFRONT & TRACKING APIS ──────────────────────────

// POST /api/delivery/estimate - Calculate delivery fee and ETA based on address
app.post('/api/delivery/estimate', (req, res) => {
    const { restaurantId = 'user_restaurant', address = 'Piazza del Plebiscito 1, Napoli' } = req.body;
    const rest = restaurants[restaurantId] || restaurants['user_restaurant'];

    const baseFee = rest.deliveryFeeEUR || 2.50;
    const distanceKm = 1.8;
    const etaMins = (rest.prepEtaMinutes || 15) + 12;

    res.json({
        success: true,
        restaurantId: rest.id,
        restaurantName: rest.name,
        customerAddress: address,
        distanceKm,
        estimatedDurationMins: etaMins,
        deliveryFeeEUR: baseFee,
        deliveryFeeOEN: (baseFee / 13.60).toFixed(2),
        freeDeliveryThresholdEUR: 35.00
    });
});

// GET /api/orders/:id/track - Live courier GPS route & visual tracking telemetry
app.get('/api/orders/:id/track', (req, res) => {
    const order = orders[req.params.id];
    if (!order) return res.status(404).json({ error: 'Order not found' });

    const rest = restaurants[order.restaurantId] || restaurants['user_restaurant'];
    const stages = {
        'AWAITING_RESTAURANT': { percent: 15, step: 1, title: 'Order Received & Escrow Secured', etaMinutes: 28 },
        'PREPARING': { percent: 45, step: 2, title: 'Kitchen Preparing Your Meal', etaMinutes: order.prepEtaMinutes || 18 },
        'READY_FOR_PICKUP': { percent: 65, step: 3, title: 'Food Packed, Courier Arrived at Counter', etaMinutes: 12 },
        'IN_TRANSIT': { percent: 85, step: 4, title: 'Courier on the Way to Your Doorstep', etaMinutes: 6 },
        'DELIVERED': { percent: 100, step: 5, title: 'Delivered & Escrow Settled', etaMinutes: 0 },
        'CANCELLED_BY_RESTAURANT': { percent: 0, step: 0, title: 'Order Cancelled & 100% Refunded', etaMinutes: 0 }
    };

    const stageInfo = stages[order.status] || stages['AWAITING_RESTAURANT'];

    // Simulated high-precision coordinates for map visualization in Naples
    const restCoords = { lat: 40.8401, lng: 14.2497, name: rest.name, address: rest.address };
    const customerCoords = { lat: 40.8359, lng: 14.2488, address: 'Piazza del Plebiscito 1, Napoli' };
    
    let courierCoords = null;
    if (order.status === 'IN_TRANSIT') {
        courierCoords = { lat: 40.8380, lng: 14.2492, heading: 'South', riderName: 'Bob Rider', vehicle: 'Electric Bicycle 🚲' };
    } else if (order.status === 'READY_FOR_PICKUP' || order.status === 'PREPARING') {
        courierCoords = { lat: 40.8403, lng: 14.2499, heading: 'At Restaurant Counter', riderName: 'Bob Rider', vehicle: 'Electric Bicycle 🚲' };
    }

    res.json({
        success: true,
        orderId: order.id,
        status: order.status,
        progressPercent: stageInfo.percent,
        currentStep: stageInfo.step,
        statusTitle: stageInfo.title,
        etaMinutes: stageInfo.etaMinutes,
        deliveryPin: order.deliveryPin,
        pickupBarcode: order.pickupBarcode,
        restaurant: restCoords,
        customer: customerCoords,
        courier: courierCoords,
        escrowLocked: order.escrowLocked,
        paymentMethod: order.paymentMethod,
        cardPayment: order.cardPayment,
        total: order.total,
        items: order.items
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
