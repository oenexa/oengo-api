const { restaurants, orders } = require('../data/store');
const { OEN_EUR_EXCHANGE_RATE } = require('../config/constants');

function estimateDelivery({ restaurantId = 'user_restaurant', address = 'Piazza del Plebiscito 1, Napoli' }) {
    const rest = restaurants[restaurantId] || restaurants['user_restaurant'];

    const baseFee = rest.deliveryFeeEUR || 2.50;
    const distanceKm = 1.8;
    const etaMins = (rest.prepEtaMinutes || 15) + 12;

    return {
        success: true,
        restaurantId: rest.id,
        restaurantName: rest.name,
        customerAddress: address,
        distanceKm,
        estimatedDurationMins: etaMins,
        deliveryFeeEUR: baseFee,
        deliveryFeeOEN: (baseFee / OEN_EUR_EXCHANGE_RATE).toFixed(2),
        freeDeliveryThresholdEUR: 35.00
    };
}

function getTrackingTelemetry(orderId) {
    const order = orders[orderId];
    if (!order) {
        const error = new Error('Order not found');
        error.statusCode = 404;
        throw error;
    }

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

    const restCoords = { lat: 40.8401, lng: 14.2497, name: rest.name, address: rest.address };
    const customerCoords = { lat: 40.8359, lng: 14.2488, address: 'Piazza del Plebiscito 1, Napoli' };
    
    let courierCoords = null;
    if (order.status === 'IN_TRANSIT') {
        courierCoords = { lat: 40.8380, lng: 14.2492, heading: 'South', riderName: 'Bob Rider', vehicle: 'Electric Bicycle 🚲' };
    } else if (order.status === 'READY_FOR_PICKUP' || order.status === 'PREPARING') {
        courierCoords = { lat: 40.8403, lng: 14.2499, heading: 'At Restaurant Counter', riderName: 'Bob Rider', vehicle: 'Electric Bicycle 🚲' };
    }

    return {
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
    };
}

module.exports = {
    estimateDelivery,
    getTrackingTelemetry
};
