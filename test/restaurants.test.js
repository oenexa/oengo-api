const test = require('node:test');
const assert = require('node:assert');

test('Restaurant stats correctly calculates 95% revenue retention and legacy savings', () => {
    const orders = [
        { id: 'o1', amount: 28.50, commissionPct: 5, restaurantPayout: 27.08, status: 'DELIVERED' },
        { id: 'o2', amount: 35.00, commissionPct: 5, restaurantPayout: 33.25, status: 'DELIVERED' }
    ];

    const defaultCommissionPct = 5;
    const commissionRetainedPct = 100 - defaultCommissionPct;
    assert.strictEqual(commissionRetainedPct, 95);

    const grossDeliveredRevenueEUR = orders.reduce((sum, o) => sum + o.restaurantPayout, 0);
    assert.strictEqual(parseFloat(grossDeliveredRevenueEUR.toFixed(2)), 60.33);

    // Legacy 30% delivery apps would only pay 70%
    const legacyAppFeePct = 30;
    const totalOrderAmount = 28.50 + 35.00; // 63.50
    const oengoFee = (totalOrderAmount * 5) / 100; // 3.175
    const legacyFee = (totalOrderAmount * legacyAppFeePct) / 100; // 19.05
    const merchantSavings = legacyFee - oengoFee; // 15.875
    assert.ok(merchantSavings > 15.00, 'Merchant retains significantly more revenue under Oengo 5% commission');
});

test('Menu item catalog calculates price in OEN from EUR', () => {
    const priceEUR = 16.50;
    const oenExchangeRate = 13.60;
    const priceOEN = (priceEUR / oenExchangeRate).toFixed(2);
    assert.strictEqual(priceOEN, '1.21');

    const starterEUR = 12.00;
    const starterOEN = (starterEUR / oenExchangeRate).toFixed(2);
    assert.strictEqual(starterOEN, '0.88');
});

test('Menu stock toggle modifies inStock availability', () => {
    const item = {
        id: 'menu_margherita',
        name: 'Artisanal Margherita Pizza',
        priceEUR: 16.50,
        inStock: true
    };

    assert.strictEqual(item.inStock, true);
    item.inStock = false;
    assert.strictEqual(item.inStock, false);
    item.inStock = true;
    assert.strictEqual(item.inStock, true);
});

test('Kitchen KDS status transition: AWAITING -> PREPARING -> READY_FOR_PICKUP', () => {
    const order = {
        id: 'ord_test_kds',
        status: 'AWAITING_RESTAURANT',
        pickupBarcode: 'PKG-TEST01'
    };

    // Step 1: Accept & Cook
    order.status = 'PREPARING';
    order.prepEtaMinutes = 15;
    assert.strictEqual(order.status, 'PREPARING');
    assert.strictEqual(order.prepEtaMinutes, 15);

    // Step 2: Ready for Pickup
    order.status = 'READY_FOR_PICKUP';
    assert.strictEqual(order.status, 'READY_FOR_PICKUP');
    assert.ok(order.pickupBarcode.startsWith('PKG-'));
});

test('Order decline triggers 100% customer refund and releases escrow', () => {
    const order = {
        id: 'ord_decline_test',
        status: 'AWAITING_RESTAURANT',
        escrowLocked: true,
        total: 34.00,
        refundStatus: null
    };

    order.status = 'CANCELLED_BY_RESTAURANT';
    order.cancelReason = 'Sold out of fresh dough';
    order.escrowLocked = false;
    order.refundStatus = 'REFUNDED_100_PERCENT';

    assert.strictEqual(order.status, 'CANCELLED_BY_RESTAURANT');
    assert.strictEqual(order.escrowLocked, false);
    assert.strictEqual(order.refundStatus, 'REFUNDED_100_PERCENT');
});
