# 🍔 OENGO — Master Engineering & Development Plan
> **Decentralized On-Demand Food Delivery Platform (Uber Eats / Deliveroo Standard)**  
> **Payment Rails: Direct Credit/Debit Card (Visa/Mastercard/Amex/Apple Pay) • Digital Wallet • On-Chain WASM Escrow (OENEXA L1)**  
> **Underlying Blockchain: OENEXA Layer-1 (`oenexa-node` on port 8545)**  
> **Runtime Environment: Node.js 24.21.0 (LTS) & Go 1.26**

---

## 1. Executive Roadmap Overview

```
┌──────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                           OENGO DEVELOPMENT ROADMAP                                              │
├──────────────────┬──────────────────┬──────────────────┬──────────────────┬──────────────────┬───────────────────┤
│     PHASE 1      │     PHASE 2      │     PHASE 3      │     PHASE 4      │     PHASE 5      │      PHASE 6      │
│  Dual-Wallet &   │ Instant Credit   │   Restaurant     │   Customer Web   │ Proximity Rider  │ Proof-of-Delivery │
│ Escrow Contract  │  Card Gateway &  │ Partner Portal   │  Storefront &    │ Dispatch Engine  │  (Barcode/PIN) &  │
│  Foundation (✅) │ Multi-Rail (✅)  │ & KDS (🚀 NEXT)  │  Live Tracking   │ & Barcode Scan   │ Multi-Disbursal   │
└──────────────────┴──────────────────┴──────────────────┴──────────────────┴──────────────────┴───────────────────┘
```

---

## 2. Phased Development Breakdown

### Phase 1: Dual-Wallet & Escrow Smart Contract Foundation (✅ COMPLETED)
*Goal: Establish the underlying financial engine, database schemas, and on-chain trust layer.*

* **Task 1.1: Database Schema & Entity Design (`oengo-api`)** ✅
  * Defined core data models: `User`, `Wallet`, `Restaurant`, `MenuItem`, `Order`.
* **Task 1.2: Dual-Wallet Ledger Service (`oengo-api`)** ✅
  * Digital Wallet balance inquiry, top-ups, and cashback rewards ledger.
  * Web3 non-custodial crypto wallet address querying via Layer-1 RPC.
* **Task 1.3: Advanced WASM Escrow Contract with Configurable Commission (`oengo-api/contracts/escrow`)** ✅
  * Implemented Go WASM contract state machine:
    * `setCommissionRate(ratePct)` ➡️ Updates global baseline commission (bounded 0% to 30%).
    * `getCommissionRate()` ➡️ Returns active platform commission parameters.
    * `createOrder(...)` ➡️ Locks funds and order's active commission percentage.
    * `assignCourier(...)` ➡️ Binds delivery courier.
    * `confirmPickup(...)` ➡️ Validates restaurant pickup via barcode hash.
    * `confirmDelivery(...)` ➡️ Autonomously releases `(100% - commPct)` to restaurant & `commPct + fee + tip` to courier.
    * `refundOrder(...)` ➡️ Returns 100% of funds upon cancellation.
  * 100% unit test coverage in `main_test.go` (`TestEscrowCompleteLifecycle`, `TestEscrowRefundFlow`, `TestEscrowConfigurableCommission`).

---

### Phase 2: Instant Credit Card Payment Gateway & Multi-Rail Checkout (✅ COMPLETED)
*Goal: Enable mainstream customers to pay immediately with credit/debit cards (Visa, Mastercard, Amex, Apple Pay) while seamlessly locking funds in escrow.*

* **Task 2.1: Payment Gateway Service & Intent Lifecycle (`oengo-api`)** ✅
  * `POST /api/payments/card-intent`: Creates encrypted payment intent with order subtotal, delivery fee, tip, and currency.
  * `POST /api/payments/confirm-card`: Authorizes and captures credit card transaction, validates brand detection (Visa, Mastercard, Amex, Discover), CVV, and cardholder details.
  * `GET /api/payments/methods/:userId`: Returns saved tokenized payment methods for 1-click checkout.
  * Tokenized card receipts (`txn_xxx`, `brand`, `last4`) adhering to PCI-DSS Level 1 isolation (zero raw card numbers stored in application database).
* **Task 2.2: Frontend Multi-Rail Payment Selector & Card Form (`oengo-web`)** ✅
  * Tri-Rail Payment Selector:
    * 💳 **Credit / Debit Card (Instant)**: Interactive card form with real-time brand badges (Visa, Mastercard, Amex), Cardholder Name, 16-digit Card Number, MM/YY Expiration, CVV, test card quick-fill presets, and "Save Card for 1-Tap Checkout".
    * 💶 **In-App Digital Wallet**: 1-click debit from user's preloaded fiat balance or accumulated vouchers.
    * ⚡ **Web3 Crypto (OEN)**: Non-custodial Layer-1 blockchain transaction signed with post-quantum ML-DSA-65 keys.
  * Active order displays the authenticated payment method badge (`💳 Paid with Visa •••• 4242`).
* **Task 2.3: Credit Card to Escrow Bridge (`oengo-api`)** ✅
  * Automatically bridges confirmed card payments directly into the order escrow vault (`Status: AWAITING_RESTAURANT`, `escrowLocked: true`).
  * Locks active commission percentage onto order (default: 5%).
  * Seamlessly preserves the physical delivery pipeline: Restaurant Acceptance ➡️ Proximity Courier Match ➡️ Barcode Counter Pickup ➡️ Doorstep PIN Verification ➡️ Multi-Disbursal.
* **Task 2.4: Automated Test Suite & Verification Gate** ✅
  * Automated unit tests in `oengo-api/test/payments.test.js` (6/6 tests passing in 63ms).
  * Go smart contract tests (`go test -count=1 -v ./...`) passing 100% in `contracts/escrow`.
  * Next.js 16.3.8 Turbopack build (`npm run build`) passing with zero TypeScript and ESLint errors.
  * Pushed to GitHub repositories `oenexa/oengo-api` and `oenexa/oengo-web` on branch `main`.

---

### Phase 3: Restaurant Partner Portal (`oengo-web/merchant`) (🚀 NEXT MILESTONE)
*Goal: Empower food vendors to manage storefronts, menus, and incoming orders in real time.*

* **Task 3.1: Merchant Dashboard & Onboarding**
  * Restaurant profile setup (Logo, cuisine tags, delivery radius, preparation time).
  * Web3 wallet & bank payout binding for instant daily revenue settlements.
* **Task 3.2: Live Menu & Catalog Management**
  * Category grouping (Starters, Mains, Drinks, Desserts).
  * Real-time item availability toggles (In Stock / Sold Out).
  * Dual-currency pricing (fiat display with live OEN conversion).
* **Task 3.3: Live Kitchen Display System (KDS)**
  * Real-time incoming order stream via WebSockets.
  * Audio-visual chime for incoming card & crypto orders.
  * One-tap actions: "Accept & Prep" (with 10/15/20 min ETA slider) or "Decline & Auto-Refund".
  * Printable/displayable **Order Pickup Barcode** for the delivery courier.

---

### Phase 4: Customer Web Storefront & Live Tracking (`oengo-web`)
*Goal: High-conversion, frictionless food ordering experience inspired by Uber Eats and Deliveroo.*

* **Task 4.1: Restaurant Discovery & Geolocation**
  * Address entry with automatic geolocation and delivery fee estimation.
  * Filtering by cuisine, dietary needs, ratings, and preparation speed.
* **Task 4.2: Interactive Menu & Cart Experience**
  * Meal customization modal (sizes, spice levels, optional add-ons).
  * Sticky cart summary with subtotal, dynamic delivery fee, and carbon-neutral delivery option.
* **Task 4.3: Real-Time Order Tracking & Map**
  * Visual stepper: `Order Placed` ➡️ `Cooking` ➡️ `On the Way` ➡️ `Arrived`.
  * Mapbox / Leaflet animated delivery courier route.

---

### Phase 5: Proximity Dispatch & Courier Workflow (`oengo-web/rider` & `oengo-api`)
*Goal: Intelligent, low-latency courier matching and restaurant pickup verification.*

* **Task 5.1: Redis Geospatial Courier Matching**
  * Courier GPS pinging stored via Redis `GEOADD`.
  * When food enters the final 5 minutes of preparation, `GEORADIUS` finds couriers within 3 km.
  * WebSocket dispatch push with estimated payout (EUR / OEN) and delivery distance.
* **Task 5.2: Courier Job Acceptance & Navigation**
  * Turn-by-turn routing to restaurant pickup location.
* **Task 5.3: Camera-Based Pickup Barcode Scanner**
  * Integrated HTML5 camera scanner (`@zxing/browser` or `html5-qrcode`).
  * Courier scans the restaurant's printed or tablet-displayed Order Barcode.
  * Order status transitions to `IN_TRANSIT`.

---

### Phase 6: Doorstep Delivery Handover & Autonomous Multi-Disbursal
*Goal: Zero-theft, verified physical delivery with instant multi-party escrow release.*

* **Task 6.1: Doorstep Delivery Handover**
  * Courier arrives at customer's location.
  * **Option A**: Courier camera scans customer's dynamic Delivery Barcode.
  * **Option B**: Customer provides the 4-digit secret PIN; courier enters it into rider app.
* **Task 6.2: Cryptographic Proof-of-Delivery Relay**
  * `oengo-api` verifies the barcode HMAC / PIN validity.
  * Signs and broadcasts `confirmDelivery` to the OENEXA Layer-1 node (`http://oenexa-node:8545`).
* **Task 6.3: Automated Multi-Party Disbursal with Dynamic Commission Split**
  * Escrow unlocks funds:
    * `(100% - CommissionPct)` transferred directly to restaurant (95% default).
    * `CommissionPct + DeliveryFee + Tip` transferred directly to courier (5% + fee + tip).
    * Customer in-app Digital Wallet credited with 5% loyalty cashback tokens.

---

### Phase 7: Multi-Container Testnet Verification & Production Hardening
*Goal: End-to-end integration and security assurance across all 4 Docker containers.*

* **Task 7.1: 4-Node Testnet Integration**
  * Run full ordering cycle against the live local PBFT network (`http://localhost:8545`).
  * Verify ML-DSA-65 post-quantum signature verification on all escrow transactions.
* **Task 7.2: Single-Command Ecosystem Verification**
  * Launch all 4 services via `docker compose up -d --build` or `.\start-ecosystem-local.ps1`.
  * Validate healthchecks, DNS resolution, and latency metrics.

---

## 3. Technology Stack Matrix

| Subsystem | Technology / Runtime | Key Libraries / Modules |
| :--- | :--- | :--- |
| **Backend API** | Node.js 24.21.0 (LTS) | Express, WebSocket, CORS, Node Crypto |
| **Smart Contracts** | Go 1.26 (WASM-targetable) | `crypto/sha256`, state machine engine |
| **Web Frontend** | Next.js 16.3.8 (Turbopack) | React 19, Tailwind CSS v4, Lucide Icons |
| **Database & Cache** | Redis 7 + PostgreSQL 16 | Geospatial indexing (`GEOADD`/`GEORADIUS`) |
| **Blockchain L1** | OENEXA Layer-1 Node (Go 1.26) | ML-DSA-65 (FIPS 204), PBFT BFT-Consensus |
| **Containerization** | Docker & Compose | Alpine Linux minimal multi-stage images |
