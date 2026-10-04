const crypto = require('crypto');
const { savedCards } = require('../data/store');
const { validateCard } = require('../utils/cardBrand');

function createCardIntent({ amount = 0, currency = 'EUR', customerId = 'user_customer' }) {
    const intentId = `pi_${Date.now()}_${crypto.randomBytes(4).toString('hex')}`;
    const clientSecret = `sec_${crypto.randomBytes(16).toString('hex')}`;
    return {
        success: true,
        paymentIntentId: intentId,
        clientSecret,
        amount: parseFloat(amount || 0),
        currency,
        status: 'REQUIRES_PAYMENT_METHOD'
    };
}

function confirmCardPayment({
    paymentIntentId,
    cardNumber = '',
    cardHolderName = '',
    cardExpMonth = '12',
    cardExpYear = '2028',
    cardCvc = '',
    saveCard = false,
    customerId = 'user_customer'
}) {
    const validation = validateCard(cardNumber, cardHolderName, cardCvc);
    if (!validation.valid) {
        const error = new Error(validation.error);
        error.statusCode = 400;
        throw error;
    }

    const { brand, cleanedNumber } = validation;
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

    return {
        success: true,
        message: `Credit card payment authorized and captured instantly via ${brand}`,
        transactionId,
        paymentIntentId: paymentIntentId || `pi_${Date.now()}`,
        brand,
        last4,
        status: 'SUCCEEDED'
    };
}

function getSavedPaymentMethods(userId) {
    const methods = savedCards[userId] || [
        { id: 'card_demo_visa', brand: 'Visa', last4: '4242', expMonth: 12, expYear: 2028, cardHolder: 'Alice Customer', isDefault: true }
    ];
    return { success: true, methods };
}

module.exports = {
    createCardIntent,
    confirmCardPayment,
    getSavedPaymentMethods
};
