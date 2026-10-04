const { users, orders, commissionState } = require('../data/store');
const { sha256 } = require('../utils/hash');

function acceptOrder(orderId, prepEtaMinutes = 15) {
    const order = orders[orderId];
    if (!order) {
        const error = new Error('Order not found');
        error.statusCode = 404;
        throw error;
    }
    if (order.status !== 'AWAITING_RESTAURANT') {
        const error = new Error(`Cannot accept order with status ${order.status}`);
        error.statusCode = 400;
        throw error;
    }

    order.status = 'PREPARING';
    order.prepEtaMinutes = prepEtaMinutes || 15;
    order.acceptedAt = new Date().toISOString();
    return { success: true, message: 'Order accepted, kitchen preparing', order };
}

function markOrderReady(orderId) {
    const order = orders[orderId];
    if (!order) {
        const error = new Error('Order not found');
        error.statusCode = 404;
        throw error;
    }
    if (order.status !== 'PREPARING') {
        const error = new Error(`Order cannot be marked ready from status ${order.status}`);
        error.statusCode = 400;
        throw error;
    }

    order.status = 'READY_FOR_PICKUP';
    order.readyAt = new Date().toISOString();
    return {
        success: true,
        message: 'Order is packed and ready for courier counter pickup',
        pickupBarcode: order.pickupBarcode,
        order
    };
}

function declineOrder(orderId, reason = 'Kitchen at peak capacity') {
    const order = orders[orderId];
    if (!order) {
        const error = new Error('Order not found');
        error.statusCode = 404;
        throw error;
    }
    if (order.status !== 'AWAITING_RESTAURANT' && order.status !== 'PREPARING') {
        const error = new Error(`Order cannot be declined from status ${order.status}`);
        error.statusCode = 400;
        throw error;
    }

    order.status = 'CANCELLED_BY_RESTAURANT';
    order.cancelReason = reason;
    order.escrowLocked = false;
    order.refundStatus = 'REFUNDED_100_PERCENT';
    order.cancelledAt = new Date().toISOString();

    if (order.paymentMethod === 'DIGITAL_WALLET') {
        const buyer = users[order.buyerId];
        if (buyer) buyer.digitalWallet.fiatBalanceEUR += order.total;
    }

    return {
        success: true,
        message: `Order declined by kitchen (${reason}). Escrow 100% refunded to customer.`,
        order
    };
}

function confirmPickup(orderId, scannedBarcode) {
    const order = orders[orderId];
    if (!order) {
        const error = new Error('Order not found');
        error.statusCode = 404;
        throw error;
    }

    if (order.status !== 'PREPARING' && order.status !== 'READY_FOR_PICKUP') {
        const error = new Error(`Order not ready for pickup (current: ${order.status})`);
        error.statusCode = 400;
        throw error;
    }

    if (sha256(scannedBarcode) !== order.pickupBarcodeHash) {
        const error = new Error('Invalid pickup barcode verification failed');
        error.statusCode = 400;
        throw error;
    }

    order.status = 'IN_TRANSIT';
    order.pickedUpAt = new Date().toISOString();
    return { success: true, message: 'Pickup confirmed! Order is now IN_TRANSIT', order };
}

function confirmDelivery(orderId, proofCode) {
    const order = orders[orderId];
    if (!order) {
        const error = new Error('Order not found');
        error.statusCode = 404;
        throw error;
    }

    if (order.status !== 'IN_TRANSIT') {
        const error = new Error(`Order must be IN_TRANSIT to confirm delivery (current: ${order.status})`);
        error.statusCode = 400;
        throw error;
    }

    if (sha256(proofCode) !== order.deliveryPinHash && proofCode !== order.deliveryPin) {
        const error = new Error('Proof-of-delivery failed: Barcode or PIN incorrect');
        error.statusCode = 400;
        throw error;
    }

    const commPct = typeof order.commissionPct === 'number' ? order.commissionPct : commissionState.defaultCommissionPct;
    const platformCut = parseFloat(((order.amount * commPct) / 100).toFixed(2));
    const restaurantPayout = parseFloat((order.amount - platformCut).toFixed(2));
    const courierPayout = parseFloat((platformCut + order.deliveryFee + order.tip).toFixed(2));

    order.status = 'DELIVERED';
    order.settledAt = new Date().toISOString();
    order.commissionPct = commPct;
    order.platformCut = platformCut;
    order.restaurantPayout = restaurantPayout;
    order.courierPayout = courierPayout;

    // Loyalty cashback (5%)
    const cashback = parseFloat((order.amount * 0.05).toFixed(2));
    const customer = users[order.buyerId];
    if (customer) {
        customer.digitalWallet.fiatBalanceEUR += cashback;
        customer.digitalWallet.loyaltyPoints += Math.floor(order.amount);
    }

    // Disburse to restaurant & courier digital wallets
    const restaurant = users[order.restaurantId];
    if (restaurant) restaurant.digitalWallet.fiatBalanceEUR += restaurantPayout;
    const courier = users[order.courierId];
    if (courier) courier.digitalWallet.fiatBalanceEUR += courierPayout;

    return {
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
    };
}

module.exports = {
    acceptOrder,
    markOrderReady,
    declineOrder,
    confirmPickup,
    confirmDelivery
};
