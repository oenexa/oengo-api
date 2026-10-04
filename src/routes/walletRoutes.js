const express = require('express');
const router = express.Router();
const walletService = require('../services/walletService');

router.get('/:userId', async (req, res) => {
    try {
        const result = await walletService.getWalletBalance(req.params.userId);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 500).json({ error: err.message });
    }
});

router.post('/topup', (req, res) => {
    try {
        const result = walletService.topupWallet(req.body);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 500).json({ error: err.message });
    }
});

module.exports = router;
