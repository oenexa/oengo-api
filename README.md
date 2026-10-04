# Oengo API & Smart Contract Services 🍔

The official backend gateway, relayer API, and WASM smart contract suite for the **Oengo** decentralized food delivery platform.

## Architecture
- **API Relayer (`server.js`)**: Express server providing off-chain order management and relaying Web3 signed transactions to `oenexa-node` via `oen_sendRawTransaction`.
- **Escrow Smart Contracts (`contracts/escrow/`)**: WASM smart contracts enforcing escrow locking and funds release upon delivery confirmation.

## Quick Start
```bash
# Install dependencies
npm install

# Run development server
npm run dev

# Run contract tests
cd contracts/escrow
go test -v ./...
```
