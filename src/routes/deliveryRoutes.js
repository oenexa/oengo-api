const express = require('express');
const router = express.Router();
const deliveryService = require('../services/deliveryService');

// POST /api/delivery/estimate - Calculate delivery fee and ETA
router.post('/estimate', (req, res) => {
    try {
        const result = deliveryService.estimateDelivery(req.body);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 500).json({ error: err.message });
    }
});

// GET /api/orders/:id/track - Live courier GPS route & visual telemetry
// Also mountable under /api/delivery/track/:id or mapped in app.js
router.get('/track/:id', (req, res) => {
    try {
        const result = deliveryService.getTrackingTelemetry(req.params.id);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 404).json({ error: err.message });
    }
});

module.exports = router;
