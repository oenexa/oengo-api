const express = require('express');
const axios = require('axios');
const router = express.Router();
const { OENEXA_RPC_URL } = require('../config/constants');

// POST /api/rpc/broadcast - Web3 Gateway Proxy to pure OENEXA Layer-1 Node
router.post('/broadcast', async (req, res) => {
    try {
        const { signedTxHex } = req.body;
        const response = await axios.post(OENEXA_RPC_URL, {
            jsonrpc: '2.0',
            method: 'oen_sendRawTransaction',
            params: [signedTxHex],
            id: Date.now()
        }, { timeout: 5000 });
        res.json(response.data);
    } catch (error) {
        res.status(500).json({ error: 'Failed to broadcast transaction to OENEXA network', detail: error.message });
    }
});

module.exports = router;
