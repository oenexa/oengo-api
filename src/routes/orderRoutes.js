const express = require('express');
const router = express.Router();
const orderService = require('../services/orderService');
const escrowService = require('../services/escrowService');

// POST /api/orders - Customer places order (Funds locked in Escrow)
router.post('/', async (req, res) => {
    try {
        const result = orderService.createOrder(req.body);
        res.status(201).json(result);
    } catch (err) {
        res.status(err.statusCode || 500).json({ error: err.message });
    }
});

// GET /api/orders - List all orders
router.get('/', (req, res) => {
    try {
        const result = orderService.getAllOrders();
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 500).json({ error: err.message });
    }
});

// GET /api/orders/:id - Order details
router.get('/:id', (req, res) => {
    try {
        const result = orderService.getOrderById(req.params.id);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 404).json({ error: err.message });
    }
});

// POST /api/orders/:id/accept - Kitchen accepts order & starts cooking
router.post('/:id/accept', (req, res) => {
    try {
        const result = escrowService.acceptOrder(req.params.id, req.body.prepEtaMinutes);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 400).json({ error: err.message });
    }
});

// POST /api/orders/:id/ready - Kitchen marks order ready for courier pickup
router.post('/:id/ready', (req, res) => {
    try {
        const result = escrowService.markOrderReady(req.params.id);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 400).json({ error: err.message });
    }
});

// POST /api/orders/:id/decline - Kitchen declines order & refunds 100%
router.post('/:id/decline', (req, res) => {
    try {
        const result = escrowService.declineOrder(req.params.id, req.body.reason);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 400).json({ error: err.message });
    }
});

// POST /api/orders/:id/assign-courier - Proximity dispatch assigns courier
router.post('/:id/assign-courier', (req, res) => {
    try {
        const result = orderService.assignCourier(req.params.id, req.body.courierId);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 400).json({ error: err.message });
    }
});

// POST /api/orders/:id/confirm-pickup - Courier scans package barcode at counter
router.post('/:id/confirm-pickup', (req, res) => {
    try {
        const result = escrowService.confirmPickup(req.params.id, req.body.scannedBarcode);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 400).json({ error: err.message });
    }
});

// POST /api/orders/:id/confirm-delivery - Courier scans delivery barcode or enters PIN
router.post('/:id/confirm-delivery', (req, res) => {
    try {
        const result = escrowService.confirmDelivery(req.params.id, req.body.proofCode);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 400).json({ error: err.message });
    }
});

module.exports = router;
