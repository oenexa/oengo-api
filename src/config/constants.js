require('dotenv').config();

module.exports = {
    PORT: process.env.PORT || 3001,
    OENEXA_RPC_URL: process.env.OENEXA_RPC_URL || 'http://localhost:8545',
    DEFAULT_COMMISSION_PCT: 5,
    MIN_COMMISSION_PCT: 0,
    MAX_COMMISSION_PCT: 30,
    OEN_EUR_EXCHANGE_RATE: 13.60, // 1 OEN ~= 13.60 EUR
    LEGACY_APP_FEE_PCT: 30,
    LOYALTY_CASHBACK_PCT: 0.05
};
