const app = require('./src/app');
const { PORT, OENEXA_RPC_URL } = require('./src/config/constants');

const server = app.listen(PORT, () => {
    console.log(`🍔 Oengo Food Delivery Commerce Gateway running on http://localhost:${PORT}`);
    console.log(`🔗 Connected to pure OENEXA node at ${OENEXA_RPC_URL}`);
});

module.exports = { app, server };
