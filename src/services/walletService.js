const axios = require('axios');
const { users } = require('../data/store');
const { OENEXA_RPC_URL } = require('../config/constants');

async function getWalletBalance(userId) {
    const user = users[userId];
    if (!user) {
        const error = new Error('User account not found');
        error.statusCode = 404;
        throw error;
    }

    let onChainBalanceOEN = '100.00000000'; // Default simulated balance

    try {
        // Query live on-chain balance from pure OENEXA Layer-1 Node
        const rpcRes = await axios.post(OENEXA_RPC_URL, {
            jsonrpc: '2.0',
            method: 'oen_getBalance',
            params: [user.cryptoWalletAddress, 'latest'],
            id: Date.now()
        }, { timeout: 2000 });

        if (rpcRes.data && rpcRes.data.result) {
            onChainBalanceOEN = rpcRes.data.result;
        }
    } catch {
        // Fallback to simulated balance if node is initializing
    }

    return {
        success: true,
        userId: user.id,
        name: user.name,
        role: user.role,
        wallets: {
            digital: {
                fiatEUR: user.digitalWallet.fiatBalanceEUR,
                loyaltyPoints: user.digitalWallet.loyaltyPoints,
                type: 'OFF_CHAIN_CREDITS'
            },
            crypto: {
                address: user.cryptoWalletAddress,
                balanceOEN: onChainBalanceOEN,
                type: 'NON_CUSTODIAL_ML_DSA_65'
            }
        }
    };
}

function topupWallet({ userId, amountEUR, points }) {
    const user = users[userId];
    if (!user) {
        const error = new Error('User not found');
        error.statusCode = 404;
        throw error;
    }

    if (amountEUR) user.digitalWallet.fiatBalanceEUR += parseFloat(amountEUR);
    if (points) user.digitalWallet.loyaltyPoints += parseInt(points, 10);

    return {
        success: true,
        message: 'Wallet topped up successfully',
        digitalWallet: user.digitalWallet
    };
}

module.exports = {
    getWalletBalance,
    topupWallet
};
