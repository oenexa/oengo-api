const test = require('node:test');
const assert = require('node:assert');

const { detectCardBrand, validateCard } = require('../src/utils/cardBrand');
const { sha256 } = require('../src/utils/hash');
const commissionService = require('../src/services/commissionService');
const paymentService = require('../src/services/paymentService');
const restaurantService = require('../src/services/restaurantService');
const orderService = require('../src/services/orderService');
const escrowService = require('../src/services/escrowService');

test('cardBrand utility detects Visa, Mastercard, Amex, and Discover via module import', () => {
    assert.strictEqual(detectCardBrand('4111111111111111'), 'Visa');
    assert.strictEqual(detectCardBrand('5105105105105100'), 'Mastercard');
    assert.strictEqual(detectCardBrand('378282246310005'), 'American Express');
    assert.strictEqual(detectCardBrand('6011000000000000'), 'Discover');
});

test('validateCard helper rejects invalid lengths and missing holders', () => {
    const res1 = validateCard('123', 'Alice', '123');
    assert.strictEqual(res1.valid, false);

    const res2 = validateCard('4111111111111111', '', '123');
    assert.strictEqual(res2.valid, false);

    const res3 = validateCard('4111111111111111', 'Alice', '12');
    assert.strictEqual(res3.valid, false);

    const res4 = validateCard('4111111111111111', 'Alice Customer', '123');
    assert.strictEqual(res4.valid, true);
    assert.strictEqual(res4.brand, 'Visa');
});

test('sha256 utility produces deterministic cryptographic hashes', () => {
    const h1 = sha256('PKG-TEST01');
    const h2 = sha256('PKG-TEST01');
    assert.strictEqual(h1, h2);
    assert.strictEqual(h1.length, 64);
});

test('commissionService queries and bounds manual commission updates', () => {
    const cfg = commissionService.getCommissionConfig();
    assert.strictEqual(cfg.commissionPct, 5);
    assert.strictEqual(cfg.restaurantSharePct, 95);

    const updated = commissionService.updateCommissionConfig(3);
    assert.strictEqual(updated.commissionPct, 3);
    assert.strictEqual(updated.restaurantSharePct, 97);

    // Out of bounds test
    assert.throws(() => commissionService.updateCommissionConfig(40), /Invalid commission rate/);

    // Reset back to 5
    commissionService.updateCommissionConfig(5);
});

test('paymentService creates card intent and confirms instant payment', () => {
    const intent = paymentService.createCardIntent({ amount: 35.50 });
    assert.strictEqual(intent.success, true);
    assert.strictEqual(intent.status, 'REQUIRES_PAYMENT_METHOD');
    assert.ok(intent.paymentIntentId.startsWith('pi_'));

    const confirmation = paymentService.confirmCardPayment({
        paymentIntentId: intent.paymentIntentId,
        cardNumber: '4242 4242 4242 4242',
        cardHolderName: 'Alice Customer',
        cardExpMonth: '12',
        cardExpYear: '2028',
        cardCvc: '123',
        saveCard: true,
        customerId: 'user_customer'
    });
    assert.strictEqual(confirmation.success, true);
    assert.strictEqual(confirmation.brand, 'Visa');
    assert.strictEqual(confirmation.last4, '4242');
    assert.strictEqual(confirmation.status, 'SUCCEEDED');
});

test('restaurantService lists filtered restaurants and adds/updates dishes', () => {
    const all = restaurantService.listRestaurants({});
    assert.ok(all.count >= 4);

    const italian = restaurantService.listRestaurants({ cuisine: 'Italian' });
    assert.ok(italian.restaurants.some(r => r.name.includes('Napoli')));

    // Add dish
    const newDish = restaurantService.addDish('user_restaurant', {
        name: 'Gourmet Calzone',
        priceEUR: 15.00,
        category: 'Pizza & Mains'
    });
    assert.strictEqual(newDish.success, true);
    assert.strictEqual(newDish.item.name, 'Gourmet Calzone');

    // Update dish
    const updated = restaurantService.updateDish('user_restaurant', newDish.item.id, {
        inStock: false
    });
    assert.strictEqual(updated.item.inStock, false);

    // Clean up / delete dish
    const deleted = restaurantService.deleteDish('user_restaurant', newDish.item.id);
    assert.strictEqual(deleted.success, true);
});

test('orderService and escrowService complete order lifecycle with PIN settlement', () => {
    const orderRes = orderService.createOrder({
        buyerId: 'user_customer',
        restaurantId: 'user_restaurant',
        items: [{ name: 'Artisanal Margherita Pizza', qty: 1, price: 16.50 }],
        amount: 16.50,
        deliveryFee: 2.50,
        tip: 1.00,
        paymentMethod: 'CREDIT_CARD',
        cardPayment: { brand: 'Visa', last4: '4242' },
        commissionPct: 5
    });

    const orderId = orderRes.order.id;
    assert.strictEqual(orderRes.order.status, 'AWAITING_RESTAURANT');

    // Kitchen accepts
    const acceptRes = escrowService.acceptOrder(orderId, 20);
    assert.strictEqual(acceptRes.order.status, 'PREPARING');

    // Kitchen marks ready
    const readyRes = escrowService.markOrderReady(orderId);
    assert.strictEqual(readyRes.order.status, 'READY_FOR_PICKUP');

    // Assign courier
    orderService.assignCourier(orderId, 'user_courier');

    // Courier confirms pickup with barcode
    const pickupRes = escrowService.confirmPickup(orderId, readyRes.pickupBarcode);
    assert.strictEqual(pickupRes.order.status, 'IN_TRANSIT');

    // Courier confirms delivery with customer secret PIN
    const deliveryRes = escrowService.confirmDelivery(orderId, orderRes.order.deliveryPin);
    assert.strictEqual(deliveryRes.order.status, 'DELIVERED');
    assert.strictEqual(deliveryRes.settlement.commissionPct, 5);
    assert.ok(deliveryRes.settlement.restaurantPayoutEUR > 15);
});
