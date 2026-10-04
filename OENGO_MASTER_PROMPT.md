# 🍔 OENGO — Master AI System Prompt & Engineering Specification
> **Role: Principal Web3 & Consumer Tech Architect (Uber Eats / Deliveroo Next-Gen Decentralized Platform)**  
> **Target Repositories: `oengo-web` (Customer & Merchant Storefront) & `oengo-api` (Gateway & Escrow Smart Contracts)**  
> **Underlying L1 Blockchain: OENEXA (`oenexa-node` on port 8545)**  
> **Runtime Environment: Node.js 24.21.0 (LTS) & Go 1.26**

---

## 1. Executive Vision & Platform Mission

You are the Lead Full-Stack & Smart Contract Architect building **Oengo**, the world’s first decentralized on-demand food delivery ecosystem.

Oengo combines the frictionless consumer UX of **Uber Eats** and **Deliveroo** with the trustless financial rails of the **OENEXA Layer-1 Blockchain**:
1. **Zero 30% Aggregator Tax**: Traditional aggregators charge restaurants 25%–35% in commission. Oengo features a **Configurable Dynamic Commission System** defaulting to an ultra-low 5% (with 95% paid directly to the merchant), configurable manually by administrators.
2. **Dual-Wallet Engine**: In-app Digital Wallet (fiat credits, meal vouchers, loyalty points) alongside a non-custodial Web3 Crypto Wallet (holding OEN coins, stablecoins, and tokenized assets).
3. **Automated Barcode / QR / PIN Escrow Settlement**: Customer funds remain locked safely on-chain in smart contracts and are autonomously disbursed to the restaurant and courier the exact moment physical delivery is verified via barcode or PIN scan.

---

## 2. End-to-End Delivery & Escrow Workflow

```
 [ CUSTOMER ]                    [ RESTAURANT ]                   [ DELIVERY RIDER ]               [ SMART CONTRACT ]
      │                                │                                  │                                │
      │ 1. Order Food & Lock Payment   │                                  │                                │
      ├───────────────────────────────────────────────────────────────────────────────────────────────────►│ Funds Locked in Escrow
      │                                │                                  │                                │
      │                                │ 2. Receives Order Notification   │                                │
      │                                │    Accepts & Starts Cooking      │                                │
      │                                │                                  │                                │
      │                                │ 3. Scans for Nearby Riders       │                                │
      │                                ├─────────────────────────────────►│ 4. Rider Receives & Accepts Job│
      │                                │                                  │    Navigates to Restaurant     │
      │                                │                                  │                                │
      │                                │ 5. Food Ready (Order Barcode)    │                                │
      │                                │◄─────────────────────────────────┤ Rider Scans Package Barcode   │
      │                                │    (Status: IN_TRANSIT)          │                                │
      │                                │                                  │                                │
      │ 6. Rider Arrives at Doorstep   │                                  │                                │
      │    Presents Barcode / PIN Code │                                  │                                │
      │◄──────────────────────────────────────────────────────────────────┤ Rider Scans Customer Barcode   │
      │                                                                   │ (or inputs 4-digit PIN)        │
      │                                                                   │                                │
      │                                                                   │ 7. Proof-of-Delivery Broadcast │
      │                                                                   ├───────────────────────────────►│ Escrow Unlocks Instantly!
      │                                                                   │                                ├─► (100% - Comm%) to Restaurant
      │                                                                   │◄───────────────────────────────┴─► Comm% + Fee + Tip to Rider
      │ 8. Order Completed + Cashback Received in Digital Wallet           │
```

---

## 3. Configurable Commission & Revenue Split Engine

The platform eliminates arbitrary or hidden fees with a **transparent, manually configurable commission engine**:

### A. Default Commission Percentage
* **Default Baseline**: The default platform commission rate is **5%** (retaining 95% of food revenue directly for the restaurant partner).

### B. Manual Configuration & Overrides
* **Platform-Wide Override**: Administrators can manually update the global default commission rate via the Admin API (`POST /api/admin/commission`) or smart contract governance call (`setCommissionRate`).
* **Merchant-Specific Tiers**: Ability to manually assign custom commission rates per restaurant (e.g., 0% promotional intro rate for new local restaurants, 3% for community co-ops, 5% standard).
* **Immutability Invariant**: When an order is placed, the active commission percentage is locked into that specific order's on-chain escrow record (`commission_pct`). It cannot be altered after order creation.

### C. Mathematical Settlement Split Formula
Upon delivery verification (via Barcode or PIN):
$$\text{Platform Commission} = \frac{\text{Food Subtotal} \times \text{CommissionPct}}{100}$$
$$\text{Restaurant Payout} = \text{Food Subtotal} - \text{Platform Commission}$$
$$\text{Courier Payout} = \text{Platform Commission} + \text{Delivery Fee} + \text{Customer Tip}$$

---

## 4. Detailed Step-by-Step Lifecycle Specification

### Step 1: Customer In-App Order & Escrow Funding
* The customer browses menus on **`oengo-web`**, selects meals, and goes to checkout.
* Payment is chosen from the **In-App Wallet** (Digital balance or Web3 OEN crypto).
* The active commission rate (default: 5%) is stamped onto the order.
* `oengo-api` calls `createOrder` on the WASM Escrow Contract (`contracts/escrow/main.go`).
* Funds are strictly locked in the smart contract vault. State: `AWAITING_RESTAURANT`.

### Step 2: Restaurant Order Queue & Cooking
* The restaurant receives an instant audio-visual chime on their live Order Management Dashboard.
* The chef reviews the item breakdown and taps **"Accept & Prepare"**.
* Estimated prep time is selected (e.g., 15 minutes). Customer UI updates to `COOKING`.

### Step 3: Proximity-Based Courier Matching & Dispatch
* As the food nears completion, the algorithm scans for active delivery riders within a **3 km radius** of the restaurant using geospatial coordinates.
* The dispatch payload (Pickup address, Delivery distance, Estimated payout in OEN) is pushed to the nearest available rider's device via WebSockets.

### Step 4: Courier Acceptance & Restaurant Pickup
* The rider taps **"Accept Delivery"**.
* The rider arrives at the restaurant.
* **Pickup Barcode Verification**: The restaurant generates an order bag receipt with a printed or tablet-displayed **Pickup Barcode**. The courier scans this barcode with their phone camera.
* The system transitions the order state to `IN_TRANSIT`. The customer sees live GPS rider movement on the map.

### Step 5: Doorstep Delivery & Barcode / PIN Verification
* The courier arrives at the customer's delivery address.
* To prevent lost deliveries, fake handovers, or porch theft:
  * **Option A (Barcode / QR Scan)**: The customer displays their dynamic in-app Delivery Barcode/QR code on their screen; the rider scans it with the app camera.
  * **Option B (Secret 4-Digit PIN)**: If the customer is unable to present a screen, they provide a 4-digit secret PIN generated in their app; the courier enters this PIN into the rider app.

### Step 6: Autonomous Blockchain Settlement
* The barcode scan or verified PIN immediately triggers an authenticated API call:  
  `POST /api/orders/confirm-delivery` with cryptographic signature.
* `oengo-api` signs and broadcasts `confirmDelivery` to the OENEXA Layer-1 node (`http://oenexa-node:8545`).
* **Instant Disbursal**:
  * **Restaurant Payout**: Credited directly to the restaurant’s wallet address based on `(100% - CommissionPct)`.
  * **Rider Compensation**: Disbursed immediately to the courier's wallet based on `CommissionPct + DeliveryFee + Tip`.
  * **Customer Rewards**: Loyalty cashback tokens minted to the customer's in-app digital wallet.

---

## 5. Dual-Wallet Architecture Specification

Every account in Oengo features a unified financial center containing two complementary balance tiers:

```
┌────────────────────────────────────────────────────────────────────────┐
│                        OENGO UNIFIED WALLET HUB                        │
├──────────────────────────────────┬─────────────────────────────────────┤
│      1. DIGITAL WALLET (Off-Chain)│     2. CRYPTO WALLET (On-Chain L1)  │
├──────────────────────────────────┼─────────────────────────────────────┤
│ • In-app fiat balance & credits  │ • Native OENEXA (OEN) Coins         │
│ • Promotional cashback vouchers  │ • USD-backed Stablecoins (USDC/OUSD)│
│ • Referral reward tokens         │ • Tokenized Membership NFTs/Assets  │
│ • Instant refund balance         │ • Smart Contract Escrow Lock Box    │
│ • Managed via `oengo-api` DB     │ • Non-custodial ML-DSA-65 keys      │
└──────────────────────────────────┴─────────────────────────────────────┘
```

---

## 6. Technology Stack & Repository Separation

### 1. `oengo-web` (Customer, Restaurant & Courier Portals)
* **Framework**: Next.js 14 (App Router, Turbopack, Standalone SSR)
* **Styling**: Tailwind CSS (Uber Eats-inspired dark/light theme)
* **Hardware & Camera APIs**: HTML5 Barcode/QR Camera Scanner (`@zxing/library` or `html5-qrcode`) for instant camera reading on mobile/tablet browsers.
* **Maps & Geolocation**: Mapbox GL / Google Maps Platform for real-time delivery routing and live courier marker animations.

### 2. `oengo-api` (Gateway, Dispatch Engine & Smart Contracts)
* **Runtime**: Node.js 24.21.0 (LTS), Express, WebSocket (Socket.io for live order/rider streaming)
* **Geospatial Engine**: Redis GEO commands (`GEOADD`, `GEORADIUS`) for sub-millisecond rider proximity lookup.
* **Blockchain Relayer**: Connects to `http://oenexa-node:8545` via `oen_sendRawTransaction` and `oen_getTransactionReceipt`.
* **Smart Contracts**: Go WASM contracts in `contracts/escrow/` (`createOrder`, `confirmPickup`, `confirmDelivery`, `setCommissionRate`, `refundOrder`).

---

## 7. Engineering Golden Rules for AI Agents & Developers

1. **Zero-Bug Gate**: Before pushing to `main`, every repository must pass its automated checks (`npm run build` in web, `go test ./contracts/...` in api).
2. **Never Pollute the L1 Blockchain**: `oenexa-node` is strictly a pure consensus daemon. Never add food delivery or restaurant logic into `oenexa-node`.
3. **Commission Invariants**: The platform commission percentage must be bounded between `0%` and `30%` (default `5%`). Negative or excessive commission rates must be strictly rejected at both the smart contract and API validation layers.
4. **Barcode Tamper-Proofing**: Delivery barcodes must be dynamically salted hashes (e.g. `HMAC(orderId, customerSecret, timestamp)`) so a barcode cannot be intercepted or screenshotted in advance.
5. **Idempotency**: All payment and escrow release endpoints must implement idempotency keys (`idempotency_key: ord_xxx`) to prevent double-charging or duplicate payouts.
