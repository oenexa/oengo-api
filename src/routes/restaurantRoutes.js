const express = require('express');
const router = express.Router();
const restaurantService = require('../services/restaurantService');

// GET /api/restaurants - List all restaurants (supports cuisine & search filters)
router.get('/', (req, res) => {
    try {
        const result = restaurantService.listRestaurants(req.query);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 500).json({ error: err.message });
    }
});

// GET /api/restaurants/:id - Restaurant profile & live analytics
router.get('/:id', (req, res) => {
    try {
        const result = restaurantService.getRestaurantProfile(req.params.id);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 404).json({ error: err.message });
    }
});

// PUT /api/restaurants/:id - Update restaurant profile
router.put('/:id', (req, res) => {
    try {
        const result = restaurantService.updateRestaurantProfile(req.params.id, req.body);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 400).json({ error: err.message });
    }
});

// GET /api/restaurants/:id/menu - List all menu dishes
router.get('/:id/menu', (req, res) => {
    try {
        const result = restaurantService.getMenu(req.params.id);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 500).json({ error: err.message });
    }
});

// POST /api/restaurants/:id/menu - Add a new dish
router.post('/:id/menu', (req, res) => {
    try {
        const result = restaurantService.addDish(req.params.id, req.body);
        res.status(201).json(result);
    } catch (err) {
        res.status(err.statusCode || 400).json({ error: err.message });
    }
});

// PUT /api/restaurants/:id/menu/:itemId - Update dish (availability toggle, price, details)
router.put('/:id/menu/:itemId', (req, res) => {
    try {
        const result = restaurantService.updateDish(req.params.id, req.params.itemId, req.body);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 400).json({ error: err.message });
    }
});

// DELETE /api/restaurants/:id/menu/:itemId - Remove dish from catalog
router.delete('/:id/menu/:itemId', (req, res) => {
    try {
        const result = restaurantService.deleteDish(req.params.id, req.params.itemId);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 400).json({ error: err.message });
    }
});

// GET /api/restaurants/:id/orders - List live orders for kitchen KDS
router.get('/:id/orders', (req, res) => {
    try {
        const result = restaurantService.getRestaurantOrders(req.params.id, req.query.status);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 500).json({ error: err.message });
    }
});

module.exports = router;
