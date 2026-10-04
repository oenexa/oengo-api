# OenGo API: Backend Modular Design System & SRP Standards ⚙️

> **Repository**: `oengo-api`  
> **Status**: Production Architecture Standard  
> **Target Audience**: Human Engineers & Autonomous AI Coding Agents  
> **Framework**: Node.js 24 LTS / Express.js / REST & Web3 RPC Proxy  
> **Last Updated**: 2026-10-04  

---

## 1. Architectural Philosophy & Principles

### 1.1 Single Responsibility Principle (SRP)
Every file and function in `oengo-api` must have **one, and only one**, reason to change:
- **Routes (`src/routes/**`)**: Pure HTTP request/response routing. Validate input parameters and delegate immediately to services. No database mutations or business algorithms in route handlers.
- **Services (`src/services/**`)**: Pure business logic and domain operations (escrow calculations, commission policies, payment authorizations, kitchen state transitions).
- **Data Store (`src/data/**`)**: State persistence abstraction and in-memory store models.
- **Utilities (`src/utils/**`)**: Pure, side-effect-free helper functions (Luhn validation, card brand detection, SHA-256 hashing).
- **Configuration (`src/config/**`)**: Environment variables, default rates, and platform thresholds.

### 1.2 Separation of Concerns: 3-Tier Layering
```text
HTTP Request  ──►  Router Layer (src/routes/)
                         │
                         ▼
                   Service Layer (src/services/)
                         │
        ┌────────────────┴────────────────┐
        ▼                                 ▼
Data Store (src/data/)         L1 Node RPC (Web3 Proxy)
```

---

## 2. Directory Structure & File Taxonomy (`oengo-api`)

```text
oengo-api/
├── server.js                              # Lean application entry point (< 20 lines)
├── ARCHITECTURE_AND_MODULAR_DESIGN.md     # Backend architecture & AI coding standards
├── DEVELOPER_AND_AI_RULES.md              # Quality gate policy
├── package.json                           # Dependencies and test runner script
│
├── src/
│   ├── app.js                             # Express application assembly, CORS & middleware
│   │
│   ├── config/                            # Environment & configuration constants
│   │   └── constants.js                   # RPC URL, commission bounds, exchange rates
│   │
│   ├── utils/                             # Pure stateless utility functions
│   │   ├── cardBrand.js                   # Regex card detection & card length/CVC validation
│   │   └── hash.js                        # SHA-256 cryptographic hashing helper
│   │
│   ├── data/                              # State persistence & datastores
│   │   └── store.js                       # Users, restaurants, menus, orders, savedCards
│   │
│   ├── services/                          # Business logic & domain services
│   │   ├── commissionService.js           # Query & update platform fee (0% to 30%)
│   │   ├── paymentService.js              # Card intent creation, payment authorization
│   │   ├── walletService.js               # Dual-wallet balances (Fiat EUR + On-Chain OEN)
│   │   ├── orderService.js                # Order creation, order listing, courier assignment
│   │   ├── escrowService.js               # Kitchen acceptance, pickup, PIN settlement & refunds
│   │   ├── restaurantService.js           # Catalog, cuisine filters, dish CRUD, financial analytics
│   │   └── deliveryService.js             # Delivery fee, distance ETA, live tracking telemetry
│   │
│   └── routes/                            # Modular Express router endpoints
│       ├── adminRoutes.js                 # /api/admin/commission
│       ├── paymentRoutes.js               # /api/payments/* (card-intent, confirm-card, methods)
│       ├── walletRoutes.js                # /api/wallet/* (balance inquiry, topup)
│       ├── orderRoutes.js                 # /api/orders/* (create, get, accept, ready, decline, verify)
│       ├── restaurantRoutes.js            # /api/restaurants/* (listing, profile, menu CRUD, KDS)
│       ├── deliveryRoutes.js              # /api/delivery/* (estimate, tracking)
│       └── rpcRoutes.js                   # /api/rpc/broadcast (L1 node gateway proxy)
│
└── test/                                  # Automated Node.js native test runner suites
    ├── payments.test.js                   # Card brand & Luhn tests
    ├── restaurants.test.js                # Financial retention & KDS transition tests
    ├── storefront.test.js                 # Catalog & search tests
    └── modular_services.test.js           # Direct unit tests for all SRP service modules
```

---

## 3. Service Specifications & Domain Boundaries

### 3.1 `commissionService.js`
- **Responsibility**: Manages the dynamic platform commission policy.
- **Invariants**: Commission rate must stay within bounded limits `[0%, 30%]`. Defaults to `5%`. Computes merchant retained share (`100 - commissionPct`).

### 3.2 `paymentService.js`
- **Responsibility**: Card intent generation, client secrets, instant card authorization, card saving, and payment method enumeration.
- **Invariants**: Never stores raw CVV. Validates length between 13 and 19 digits.

### 3.3 `walletService.js`
- **Responsibility**: Aggregates customer/merchant off-chain fiat credits and queries on-chain non-custodial balance from Oenexa L1 Node (`oen_getBalance`).
- **Invariants**: Gracefully falls back to simulated balance if the L1 node is offline or initializing.

### 3.4 `orderService.js`
- **Responsibility**: Generates secure order records with scannable pickup barcode (`PKG-XXXXXX`) and a 4-digit secret delivery PIN.
- **Invariants**: Stores cryptographic SHA-256 hashes of barcodes and PINs for zero-knowledge verification.

### 3.5 `escrowService.js`
- **Responsibility**: State transitions for smart contract escrow settlement:
  1. `acceptOrder`: Kitchen begins preparation.
  2. `markOrderReady`: Counter packaging complete.
  3. `declineOrder`: Cancels order and triggers **100% immediate customer refund**.
  4. `confirmPickup`: Verifies courier scanned package barcode hash.
  5. `confirmDelivery`: Verifies customer delivery PIN hash, computes autonomous settlement (platform fee, restaurant 95% payout, courier payout), and credits 5% loyalty cashback.

### 3.6 `restaurantService.js`
- **Responsibility**: Partner catalog management, cuisine filtering, dish availability toggles (86-ing items), and financial retention analytics.

### 3.7 `deliveryService.js`
- **Responsibility**: Fee calculation, distance estimation, and live GPS coordinate simulation for tracking.

---

## 4. Backend Developer & AI Agent Guidelines

Future engineers and autonomous AI agents working in `oengo-api` **must** strictly adhere to the following rules:

### Rule 1: No Monolithic Files
- **Maximum file size limit**: 200 lines for routes; 250 lines for services.
- Never write business logic inside route controllers. Always delegate to a dedicated service in `src/services/`.

### Rule 2: Pure Utilities
- Utilities in `src/utils/` must be pure functions with zero database side-effects and zero external network calls.

### Rule 3: Graceful Error Handling
- Service methods must throw typed errors with `statusCode` properties (e.g. 400 for validation errors, 404 for missing resources).
- Route handlers must catch these errors and respond with appropriate HTTP status codes.

### Rule 4: Zero-Bug Quality Gate
Before submitting any pull request or committing code:
- Execute `npm test` — all test suites in `test/*.test.js` must pass with 100% success rate (currently 22/22 passing).
