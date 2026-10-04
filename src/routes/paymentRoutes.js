const express = require('express');
const router = express.Router();
const paymentService = require('../services/paymentService');

router.post('/card-intent', (req, res) => {
    try {
        const result = paymentService.createCardIntent(req.body);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 500).json({ error: err.message });
    }
});

router.post('/confirm-card', (req, res) => {
    try {
        const result = paymentService.confirmCardPayment(req.body);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 500).json({ error: err.message });
    }
});

router.get('/methods/:userId', (req, res) => {
    try {
        const result = paymentService.getSavedPaymentMethods(req.params.userId);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 500).json({ error: err.message });
    }
});

module.exports = router;
