require('dotenv').config();
const express = require('express');
const axios = require('axios');
const cors = require('cors');

const app = express();
app.use(cors());
app.use(express.json());

const PORT = process.env.PORT || 3001;
const OENEXA_RPC_URL = process.env.OENEXA_RPC_URL || 'http://localhost:8545';

// Mock local database
const orders = {};

// 1. Off-chain Shopping: Create an order in the local database
app.post('/api/orders', (req, res) => {
    const { buyer, seller, amount } = req.body;
    const orderId = `ord_${Date.now()}`;
    
    orders[orderId] = {
        id: orderId,
        buyer,
        seller,
        amount,
        status: 'AWAITING_PAYMENT'
    };
    
    res.json({ success: true, order: orders[orderId] });
});

// 2. Web3 Gateway: Proxy a raw transaction to the pure OENEXA node
app.post('/api/rpc/broadcast', async (req, res) => {
    try {
        const { signedTxHex } = req.body;
        
        // Broadcast the transaction to the pure OENEXA Layer 1 node
        const response = await axios.post(OENEXA_RPC_URL, {
            jsonrpc: '2.0',
            method: 'oen_sendRawTransaction',
            params: [signedTxHex],
            id: Date.now()
        });
        
        res.json(response.data);
    } catch (error) {
        console.error('RPC Error:', error.message);
        res.status(500).json({ error: 'Failed to broadcast transaction to OENEXA network' });
    }
});

// 3. Confirm Delivery: Poll the contract state (stubbed)
app.get('/api/orders/:id', async (req, res) => {
    const order = orders[req.params.id];
    if (!order) return res.status(404).json({ error: 'Order not found' });
    
    res.json({ success: true, order });
});

app.listen(PORT, () => {
    console.log(`🍔 Oengo App Food Delivery Backend running on http://localhost:${PORT}`);
    console.log(`🔗 Connected to pure OENEXA node at ${OENEXA_RPC_URL}`);
});
