const express = require('express');
const router = express.Router();
const commissionService = require('../services/commissionService');

router.get('/commission', (req, res) => {
    try {
        const config = commissionService.getCommissionConfig();
        res.json(config);
    } catch (err) {
        res.status(500).json({ error: err.message });
    }
});

router.post('/commission', (req, res) => {
    try {
        const result = commissionService.updateCommissionConfig(req.body.ratePct);
        res.json(result);
    } catch (err) {
        res.status(400).json({ error: err.message });
    }
});

module.exports = router;
