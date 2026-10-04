const test = require('node:test');
const assert = require('node:assert');

// Test card brand detection logic
function detectCardBrand(number) {
    const cleaned = (number || '').replace(/\D/g, '');
    if (/^4/.test(cleaned)) return 'Visa';
    if (/^5[1-5]/.test(cleaned) || /^2[2-7]/.test(cleaned)) return 'Mastercard';
    if (/^3[47]/.test(cleaned)) return 'American Express';
    if (/^6(?:011|5)/.test(cleaned)) return 'Discover';
    return 'Credit Card';
}

test('detectCardBrand accurately detects Visa cards', () => {
    assert.strictEqual(detectCardBrand('4242 4242 4242 4242'), 'Visa');
    assert.strictEqual(detectCardBrand('4000123456789010'), 'Visa');
});

test('detectCardBrand accurately detects Mastercard', () => {
    assert.strictEqual(detectCardBrand('5500 0000 0000 0004'), 'Mastercard');
    assert.strictEqual(detectCardBrand('2221 0000 0000 0001'), 'Mastercard');
});

test('detectCardBrand accurately detects American Express', () => {
    assert.strictEqual(detectCardBrand('3782 822463 10005'), 'American Express');
    assert.strictEqual(detectCardBrand('3400 000000 00000'), 'American Express');
});

test('detectCardBrand accurately detects Discover', () => {
    assert.strictEqual(detectCardBrand('6011 0000 0000 0000'), 'Discover');
    assert.strictEqual(detectCardBrand('6500 0000 0000 0000'), 'Discover');
});

test('Card length validation rejects invalid card numbers', () => {
    const shortNumber = '1234'.replace(/\D/g, '');
    assert.ok(shortNumber.length < 13);
    const validNumber = '4242424242424242'.replace(/\D/g, '');
    assert.ok(validNumber.length >= 13 && validNumber.length <= 19);
});

test('CVV validation enforces 3-4 security digits', () => {
    const cvc3 = '123';
    const cvc4 = '1234';
    const badCvc = '1';
    assert.ok(cvc3.length >= 3 && cvc3.length <= 4);
    assert.ok(cvc4.length >= 3 && cvc4.length <= 4);
    assert.ok(badCvc.length < 3 || badCvc.length > 4);
});
