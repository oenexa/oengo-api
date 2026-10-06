# 🍔 OENGO-API — Enterprise Backend & Smart Contract Service

The official, pure backend engine, REST API, WebSocket hub, and Go WASM smart contract suite for the **OENGO** on-demand food delivery and super-app ecosystem.

---

## 🏛️ Architecture & Separation of Concerns

* **Pure Backend Mandate**: Provides REST APIs (`/api/v1`), WebSocket streams, and Layer-1 blockchain escrow integration. Zero UI or client-side rendering.
* **Client Companion**: The frontend UI is housed separately in the [`oengo-web`](../oengo-web) repository.
* **Underlying Blockchain**: Connects to the OENEXA Layer-1 node (`http://oenexa-node:8545`) via JSON-RPC.

For in-depth architectural patterns, double-entry ledger equations, and directory guidelines, see **[BACKEND_ARCHITECTURE.md](./BACKEND_ARCHITECTURE.md)**.

---

## 🚀 Quick Start

### Prerequisites
- Node.js 24.21.0+ (LTS)
- Go 1.26+ (for smart contract tests)
- Docker & Docker Compose

### Local Development
```bash
# Install dependencies
npm install

# Run backend development server (Port 3001)
npm run dev

# Run automated API test suite
npm test

# Run Go WASM escrow smart contract test suite
cd contracts/escrow
go test -v ./...
```

---

## 📡 Core API Endpoints

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/health` | Service health status |
| `GET` | `/api/restaurants` | List partner restaurants with metadata |
| `GET` | `/api/restaurants/:id/menu` | Categorized menu catalog with custom modifiers |
| `POST` | `/api/orders` | Place order, calculate commission, and lock escrow |
| `GET` | `/api/orders/:id/track` | Real-time order progress & courier telemetry |
| `POST` | `/api/payments/card-intent` | Tokenized card intent creation (PCI-DSS compliant) |
| `POST` | `/api/payments/confirm-card`| Authorize card and lock payment into escrow vault |
| `GET` | `/api/wallet/:userId` | Double-entry digital balance & coin rewards |
| `POST` | `/api/orders/confirm-pickup`| Courier scans package barcode at restaurant counter |
| `POST` | `/api/orders/confirm-delivery`| Courier scans customer barcode or enters PIN |
