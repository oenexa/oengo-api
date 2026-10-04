function detectCardBrand(number) {
    const cleaned = (number || '').replace(/\D/g, '');
    if (/^4/.test(cleaned)) return 'Visa';
    if (/^5[1-5]/.test(cleaned) || /^2[2-7]/.test(cleaned)) return 'Mastercard';
    if (/^3[47]/.test(cleaned)) return 'American Express';
    if (/^6(?:011|5)/.test(cleaned)) return 'Discover';
    return 'Credit Card';
}

function validateCard(cardNumber, cardHolderName, cardCvc) {
    const cleanedNumber = (cardNumber || '').replace(/\D/g, '');
    if (cleanedNumber.length < 13 || cleanedNumber.length > 19) {
        return { valid: false, error: 'Invalid card number length (must be 13-19 digits)' };
    }
    if (!cardHolderName || !cardHolderName.trim()) {
        return { valid: false, error: 'Cardholder name is required' };
    }
    if (!cardCvc || cardCvc.length < 3 || cardCvc.length > 4) {
        return { valid: false, error: 'Invalid CVV/CVC security code' };
    }
    return { valid: true, cleanedNumber, brand: detectCardBrand(cleanedNumber) };
}

module.exports = {
    detectCardBrand,
    validateCard
};
