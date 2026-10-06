# ⚙️ OENGO-API — Backend Architecture & Engineering Guidelines
> **Repository Mandate: Pure Backend API, Event Mesh & Financial Smart Contracts**  
> **Target Service: REST API, WebSockets, gRPC, PostgreSQL Double-Entry Ledger & OENEXA L1 Escrow**  
> **Strict Isolation Boundary: ZERO UI rendering, ZERO JSX/HTML templates, ZERO frontend bundlers**

---

## 1. Architectural Persona & Scope

`oengo-api` is the **Pure Backend Engine** for the OENGO on-demand delivery ecosystem. It serves as the single source of truth for all business logic, financial ledgers, transactional state machines, event streaming, and blockchain escrow interactions.

### Strict Boundary Rules
* **No Frontend Code**: Do NOT place React components, Next.js routes, CSS styles, or browser-specific bundles inside this repository.
* **Pure API Contracts**: All communication with clients (`oengo-web`, mobile apps) occurs strictly over versioned JSON REST APIs (`/api/v1/...`) and WebSocket streams.
* **Data Integrity & Invariants**: All business rules (stock deductions, balance checks, commission splits, proof-of-delivery verifications) must be strictly enforced on the backend, never trusting client payloads.

---

## 2. Directory Structure & Architecture Standards

The backend follows **Clean Architecture & Domain-Driven Design (DDD)**:

```
oengo-api/
├── contracts/
│   └── escrow/                     # Go WASM Smart Contracts (L1 Blockchain trust layer)
│       ├── main.go                 # State machine: createOrder, confirmPickup, confirmDelivery
│       └── main_test.go            # 100% test coverage for escrow lifecycle
├── src/
│   ├── config/                     # Environment, constants, commission bounds (0%-30%)
│   ├── data/                       # In-memory store / database connection pooling
│   ├── routes/                     # HTTP Router layer (Gin / Express)
│   │   ├── adminRoutes.js          # Platform management, commission overrides
│   │   ├── deliveryRoutes.js       # Courier telemetry, online/offline, proximity
│   │   ├── orderRoutes.js          # Order placement, state transitions, barcode verification
│   │   ├── paymentRoutes.js        # Card intents, tokenized checkout, Instant Pay
│   │   ├── restaurantRoutes.js     # Menus, operating hours, KDS ticket queue
│   │   ├── rpcRoutes.js            # L1 Blockchain relayer proxy
│   │   └── walletRoutes.js         # Double-entry ledger, deposits, coin rewards
│   ├── services/                   # Core business usecases & domain logic
│   │   ├── commissionService.js    # Transparent 5% default commission engine
│   │   ├── deliveryService.js      # Redis GEO proximity matching & telemetry
│   │   ├── escrowService.js        # Go WASM & L1 escrow contract bridge
│   │   ├── orderService.js         # Order state machine & PIN/Barcode generation
│   │   ├── paymentService.js       # PCI-DSS tokenized card & multi-rail intents
│   │   ├── restaurantService.js    # Menu catalog & KDS operational controls
│   │   └── walletService.js        # Banking-grade ledger & balance invariants
│   ├── utils/                      # Hashes, HMAC-SHA256 barcode salting, brand detection
│   └── app.js                      # Express / Gin HTTP application bootstrapping
├── test/                           # Automated API test suites
├── Dockerfile                      # Production container definition
├── package.json
└── server.js                       # Server entrypoint (Port 3001)
```

---

## 3. Core Subsystems & Responsibilities

### A. Banking-Grade Wallet & Double-Entry Ledger
* All monetary movements must adhere to the fundamental accounting equation:
  $$\sum \text{Debits} = \sum \text{Credits}$$
* Zero negative balances allowed (`balance >= 0`).
* Accounts:
  * `1001`: Payment Gateway Clearing (Asset)
  * `1002`: Customer Available Wallet (Liability)
  * `1003`: Customer Locked Order Escrow (Liability)
  * `2001`: Restaurant Payable (Liability)
  * `2002`: Courier Payable (Liability)
  * `3001`: Platform Commission Revenue (Revenue)
  * `3002`: Loyalty Coin Expense Reserve (Expense)

### B. Order State Machine & Commission Split
* Default Platform Commission: **5%** (retaining 95% for merchant). Configurable from 0% to 30%.
* State Machine Flow:
  $$\text{CREATED} \rightarrow \text{ACCEPTED} \rightarrow \text{PREPARING} \rightarrow \text{READY\_FOR\_PICKUP} \rightarrow \text{IN\_TRANSIT} \rightarrow \text{DELIVERED}$$
* Upon delivery verification (scanned Package Barcode + Customer PIN):
  $$\text{Restaurant Share} = \text{Food Subtotal} \times (1 - \text{CommissionPct})$$
  $$\text{Courier Share} = (\text{Food Subtotal} \times \text{CommissionPct}) + \text{Delivery Fee} + \text{Tip}$$
  $$\text{Customer Rewards} = 5\% \text{ Food Subtotal in OENGO Coins}$$

### C. Cryptographic Proof-of-Delivery
* Delivery Barcodes must be dynamically salted: $\text{HMAC-SHA256}(\text{orderId} \parallel \text{secret} \parallel \text{timestamp})$.
* 4-digit PIN fallback with rate-limiting (max 3 failed attempts).

### D. Security & Compliance
* **PCI-DSS Level 1**: Zero Primary Account Numbers (PAN) or CVVs stored in database or logs. Use tokenized card intents (`card_intent_xxx`).
* **Idempotency**: All payment and escrow release endpoints require `Idempotency-Key` headers stored in Redis with 24h TTL.

---

## 4. Instructions for Future AI Agents & Developers

1. **Maintain Pure Backend Isolation**: Never import React, DOM APIs, or client-side UI libraries into `oengo-api`.
2. **Mandatory Test Verification**:
   * Always run `npm test` before pushing to `main`.
   * Always run `go test -v ./...` in `contracts/escrow`.
   * Pushing failing or unverified code is strictly prohibited.
3. **Database & Schema Changes**:
   * All schema changes must use UUID primary keys (`gen_random_uuid()`), check constraints, and audit logs.
4. **API Versioning**:
   * All new public routes must be prefixed with `/api/v1/`. Existing backward-compatible routes must be preserved.
