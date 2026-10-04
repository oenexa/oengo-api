const { users, orders, commissionState } = require('../data/store');
const { MIN_COMMISSION_PCT, MAX_COMMISSION_PCT } = require('../config/constants');
const { sha256 } = require('../utils/hash');

function createOrder({
    buyerId = 'user_customer',
    restaurantId = 'user_restaurant',
    items = [],
    amount = 25.00,
    deliveryFee = 3.50,
    tip = 2.00,
    paymentMethod = 'CREDIT_CARD',
    cardPayment = null,
    commissionPct
}) {
    const buyer = users[buyerId];
    const restaurant = users[restaurantId];
    if (!buyer || !restaurant) {
        const error = new Error('Invalid buyer or restaurant ID');
        error.statusCode = 400;
        throw error;
    }

    let activeCommissionPct = commissionState.defaultCommissionPct;
    if (commissionPct !== undefined && commissionPct !== null) {
        const parsed = Number(commissionPct);
        if (!isNaN(parsed) && parsed >= MIN_COMMISSION_PCT && parsed <= MAX_COMMISSION_PCT) {
            activeCommissionPct = parsed;
        }
    }

    const orderId = `ord_${Date.now()}`;
    const pickupBarcode = `PKG-${orderId.slice(-6).toUpperCase()}`;
    const deliveryPin = `${Math.floor(1000 + Math.random() * 9000)}`;

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
        status: 'AWAITING_RESTAURANT',
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

    return {
        success: true,
        message: `Order created via ${methodLabel} and funds locked in smart contract escrow`,
        order: newOrder
    };
}

function getAllOrders() {
    const list = Object.values(orders);
    return { success: true, count: list.length, orders: list };
}

function getOrderById(id) {
    const order = orders[id];
    if (!order) {
        const error = new Error('Order not found');
        error.statusCode = 404;
        throw error;
    }
    return { success: true, order };
}

function assignCourier(orderId, courierId = 'user_courier') {
    const order = orders[orderId];
    if (!order) {
        const error = new Error('Order not found');
        error.statusCode = 404;
        throw error;
    }

    const courier = users[courierId];
    if (!courier) {
        const error = new Error('Courier not found');
        error.statusCode = 400;
        throw error;
    }

    order.courierId = courierId;
    order.courierAddress = courier.cryptoWalletAddress;
    return { success: true, message: `Courier ${courier.name} assigned`, order };
}

module.exports = {
    createOrder,
    getAllOrders,
    getOrderById,
    assignCourier
};
