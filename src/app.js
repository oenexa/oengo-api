const express = require('express');
const cors = require('cors');

const adminRoutes = require('./routes/adminRoutes');
const paymentRoutes = require('./routes/paymentRoutes');
const walletRoutes = require('./routes/walletRoutes');
const orderRoutes = require('./routes/orderRoutes');
const restaurantRoutes = require('./routes/restaurantRoutes');
const deliveryRoutes = require('./routes/deliveryRoutes');
const rpcRoutes = require('./routes/rpcRoutes');
const deliveryService = require('./services/deliveryService');

const app = express();

app.use(cors());
app.use(express.json());

// Direct route for /api/orders/:id/track
app.get('/api/orders/:id/track', (req, res) => {
    try {
        const result = deliveryService.getTrackingTelemetry(req.params.id);
        res.json(result);
    } catch (err) {
        res.status(err.statusCode || 404).json({ error: err.message });
    }
});

// Mount modular sub-routers
app.use('/api/admin', adminRoutes);
app.use('/api/payments', paymentRoutes);
app.use('/api/wallet', walletRoutes);
app.use('/api/orders', orderRoutes);
app.use('/api/restaurants', restaurantRoutes);
app.use('/api/delivery', deliveryRoutes);
app.use('/api/rpc', rpcRoutes);

// Health check endpoint
app.get('/health', (req, res) => {
    res.json({ status: 'UP', service: 'oengo-api', timestamp: new Date().toISOString() });
});

module.exports = app;
