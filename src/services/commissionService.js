const { MIN_COMMISSION_PCT, MAX_COMMISSION_PCT } = require('../config/constants');
const { commissionState } = require('../data/store');

function getCommissionConfig() {
    return {
        success: true,
        commissionPct: commissionState.defaultCommissionPct,
        minPct: MIN_COMMISSION_PCT,
        maxPct: MAX_COMMISSION_PCT,
        restaurantSharePct: 100 - commissionState.defaultCommissionPct
    };
}

function updateCommissionConfig(ratePct) {
    const num = Number(ratePct);
    if (isNaN(num) || num < MIN_COMMISSION_PCT || num > MAX_COMMISSION_PCT) {
        throw new Error(`Invalid commission rate. Must be a number between ${MIN_COMMISSION_PCT}% and ${MAX_COMMISSION_PCT}%.`);
    }
    commissionState.defaultCommissionPct = num;
    return {
        success: true,
        message: `Platform commission rate updated to ${commissionState.defaultCommissionPct}%`,
        commissionPct: commissionState.defaultCommissionPct,
        restaurantSharePct: 100 - commissionState.defaultCommissionPct
    };
}

module.exports = {
    getCommissionConfig,
    updateCommissionConfig
};
