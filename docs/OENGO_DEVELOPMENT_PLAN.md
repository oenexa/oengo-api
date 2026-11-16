# 🍔 OENGO — Master Engineering & Development Plan
> **Decentralized On-Demand Food Delivery Platform (Uber Eats / Deliveroo Standard)**  
> **Architecture: Dual-Wallet (Digital + Web3 Crypto) • On-Chain WASM Escrow • Proximity Dispatch • Barcode/PIN Proof-of-Delivery**  
> **Underlying Blockchain: OENEXA Layer-1 (`oenexa-node` on port 8545)**  
> **Runtime Environment: Node.js 24.21.0 (LTS) & Go 1.26**

---

## 1. Executive Roadmap Overview

```
 ┌────────────────────────────────────────────────────────────────────────────────────────┐
 │                              OENGO DEVELOPMENT ROADMAP                                 │
 ├─────────────────┬─────────────────┬─────────────────┬─────────────────┬────────────────┤
 │     PHASE 1     │     PHASE 2     │     PHASE 3     │     PHASE 4     │    PHASE 5     │
 │  Dual-Wallet &  │   Restaurant    │  Customer Web   │ Proximity Rider │ Proof-of-      │
 │  Escrow Contract│ Partner Portal  │   Storefront    │ Dispatch Engine │ Delivery &     │
 │   Foundation    │ & Live Orders   │ & Live Tracking │ & Barcode Scan  │ L1 Settlement  │
 └─────────────────┴─────────────────┴─────────────────┴─────────────────┴────────────────┘
```

---

## 2. Phased Development Breakdown

### Phase 1: Dual-Wallet & Escrow Smart Contract Foundation
*Goal: Establish the underlying financial engine, database schemas, and on-chain trust layer.*

* **Task 1.1: Database Schema Design (`oengo-api`)**
  * Implement PostgreSQL database schema using Prisma ORM:
    * `User` (Customer, Merchant, Courier roles)
    * `Wallet` (Digital Balance, Loyalty Cashback Points, Web3 Public Address)
    * `Restaurant` (Name, Geo-Coordinates [Lat/Long], Cuisine, Operating Hours)
    * `MenuItem` (Title, Description, Price in EUR & ONXC, Category, Photo)
    * `Order` (Status, CartItems, Subtotal, DeliveryFee, Tip, EscrowTxHash, PIN, BarcodeHash)
* **Task 1.2: Dual-Wallet Ledger Service (`oengo-api`)**
  * **Digital Wallet**: Ledger endpoints for credits, promotional vouchers, and cashback rewards.
  * **Web3 Crypto Wallet**: Balance querying via `onxc_getBalance` and raw transaction broadcasting via `onxc_sendRawTransaction`.
* **Task 1.3: Advanced WASM Escrow Contract (`oengo-api/contracts/escrow`)**
  * Extend Go WASM contract with complete lifecycle state machine:
    * `createOrder(orderId, buyer, restaurant, courier, amount, fee)` ➡️ Locks funds in contract.
    * `assignCourier(orderId, courier)` ➡️ Binds delivery courier.
    * `confirmPickup(orderId, pickupHash)` ➡️ Validates restaurant pickup.
    * `confirmDelivery(orderId, proofHash)` ➡️ Autonomously releases 95% to restaurant & 5%+tips to courier.
    * `refundOrder(orderId)` ➡️ Automated return of funds on merchant decline or cancellation timeout.
  * 100% unit test coverage in `main_test.go`.

---

### Phase 2: Restaurant Partner Portal (`oengo-web/merchant`)
*Goal: Empower food vendors to manage storefronts, menus, and incoming orders in real time.*

* **Task 2.1: Merchant Dashboard & Onboarding**
  * Restaurant profile setup (Logo, cuisine tags, delivery radius, preparation time).
  * Web3 wallet binding for instant daily revenue settlements.
* **Task 2.2: Live Menu & Catalog Management**
  * Category grouping (Starters, Mains, Drinks, Desserts).
  * Real-time item availability toggles (In Stock / Sold Out).
  * Dual-currency pricing (fiat display with live ONXC conversion).
* **Task 2.3: Live Kitchen Display System (KDS)**
  * Real-time incoming order stream via WebSockets.
  * Audio-visual chime for new orders.
  * One-tap actions: "Accept & Prep" (with 10/15/20 min ETA slider) or "Decline & Auto-Refund".
  * Printable/displayable **Order Pickup Barcode** for the delivery courier.

---

### Phase 3: Customer Web Storefront (`oengo-web`)
*Goal: High-conversion, frictionless food ordering experience inspired by Uber Eats and Deliveroo.*

* **Task 3.1: Restaurant Discovery & Geolocation**
  * Address entry with automatic geolocation and delivery fee estimation.
  * Filtering by cuisine, dietary needs, ratings, and preparation speed.
* **Task 3.2: Interactive Menu & Cart Experience**
  * Meal customization modal (sizes, spice levels, optional add-ons).
  * Sticky cart summary with subtotal, dynamic delivery fee, and carbon-neutral delivery option.
* **Task 3.3: 1-Click Dual-Wallet Checkout**
  * Payment selector:
    * **Option A**: Pay instantly with in-app Digital Wallet balance.
    * **Option B**: Sign transaction using OENEXA Web3 Crypto Wallet.
  * Automatic generation of the dynamic **Customer Delivery Barcode** and secret **4-digit PIN**.
* **Task 3.4: Real-Time Order Tracking & Map**
  * Visual stepper: `Order Placed` ➡️ `Cooking` ➡️ `On the Way` ➡️ `Arrived`.
  * Mapbox / Google Maps animated delivery courier route.

---

### Phase 4: Proximity Dispatch & Courier Workflow (`oengo-web/rider` & `oengo-api`)
*Goal: Intelligent, low-latency courier matching and restaurant pickup verification.*

* **Task 4.1: Redis Geospatial Courier Matching**
  * Courier GPS pinging stored via Redis `GEOADD`.
  * When food enters the final 5 minutes of preparation, `GEORADIUS` finds couriers within 3 km.
  * WebSocket dispatch push with estimated payout (ONXC) and delivery distance.
* **Task 4.2: Courier Job Acceptance & Navigation**
  * Turn-by-turn routing to restaurant pickup location.
* **Task 4.3: Camera-Based Pickup Barcode Scanner**
  * Integrated HTML5 camera scanner (`@zxing/browser` or `html5-qrcode`).
  * Courier scans the restaurant's printed or tablet-displayed Order Barcode.
  * Order status transitions to `IN_TRANSIT`.

---

### Phase 5: Doorstep Delivery Handover & Autonomous L1 Settlement
*Goal: Zero-theft, verified physical delivery with instant on-chain escrow disbursal.*

* **Task 5.1: Doorstep Delivery Handover**
  * Courier arrives at customer's location.
  * **Option A**: Courier camera scans customer's dynamic Delivery Barcode.
  * **Option B**: Customer provides the 4-digit secret PIN; courier enters it into rider app.
* **Task 5.2: Cryptographic Proof-of-Delivery Relay**
  * `oengo-api` verifies the barcode HMAC / PIN validity.
  * Signs and broadcasts `confirmDelivery` to the OENEXA Layer-1 node (`http://oenexa-node:8545`).
* **Task 5.3: Automated Multi-Party Disbursal**
  * WASM contract unlocks escrow:
    * **95%** transferred directly to the restaurant's wallet.
    * **5% base + 100% tip** transferred directly to the courier's wallet.
    * Customer in-app Digital Wallet credited with loyalty cashback tokens.

---

### Phase 6: Multi-Container Testnet Verification & Production Hardening
*Goal: End-to-end integration and security assurance across all 4 Docker containers.*

* **Task 6.1: 4-Node Testnet Integration**
  * Run full ordering cycle against the live local PBFT network (`http://localhost:8545`).
  * Verify ML-DSA-65 post-quantum signature verification on all escrow transactions.
* **Task 6.2: Single-Command Ecosystem Verification**
  * Launch all 4 services via `docker compose up -d --build`.
  * Validate healthchecks, DNS resolution, and latency metrics.

---

## 3. Immediate Next Implementation Step

We are ready to begin **Phase 1: Foundation & Dual-Wallet Infrastructure**:
1. Implement the Prisma database schema in `oengo-api` for Users, Wallets, Restaurants, and Orders.
2. Upgrade `contracts/escrow/main.go` to support the multi-party lifecycle (`confirmPickup`, `confirmDelivery` with barcode/PIN verification).
3. Connect the dual-wallet balance ledger.
