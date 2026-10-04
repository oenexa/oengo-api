# OenGo & Oenexa: Architecture & Modular Design System

> **Document Status**: Production Architecture Standard  
> **Target Audience**: Human Software Engineers & Autonomous AI Coding Agents  
> **Last Updated**: 2026-10-04  
> **Scope**: `oengo-web` (Frontend), `oengo-api` (Commerce Gateway), and `oenexa-node` (L1 Blockchain)

---

## 1. Architectural Philosophy & Principles

### 1.1 Single Responsibility Principle (SRP)
Every module, file, and React component must have **one, and only one**, reason to change. 
- **Anti-Pattern**: Putting 800+ lines of state management, API calls, SVG maps, forms, modals, and checkout math into a single `page.tsx`.
- **Approved Pattern**: Modular components with focused scopes (<150 lines average), strictly typed input props, and isolated responsibilities.

### 1.2 Container / Presentational Separation
- **Page Orchestrators (`src/app/**/page.tsx`)**: Act strictly as coordinators. They fetch initial data, hold high-level screen states (active tabs, selected restaurant, cart state), and compose presentational subcomponents.
- **Presentational Components (`src/components/**`)**: Receive props and emit typed callbacks. They do not initiate arbitrary out-of-band state mutations or hardcode API URLs.

### 1.3 Decoupled API & Network Layer
- **Central API Client (`src/lib/api.ts`)**: All HTTP communication, error parsing, and fallback behaviors are centralized.
- Components never execute raw `fetch()` calls or invent ad-hoc endpoints.

### 1.4 Strict Domain Modeling
- **TypeScript Contracts (`src/types/`)**: Domain models (Restaurant, MenuItem, Order, Payment, Wallet) are defined once in dedicated type files and re-exported via `src/types/index.ts`. No duplicate interfaces across files.

---

## 2. Directory Structure & File Taxonomy (`oengo-web`)

```text
oengo-web/
├── src/
│   ├── app/                               # Next.js App Router (Entrypoint Orchestrators)
│   │   ├── layout.tsx                     # Root layout with fonts and metadata
│   │   ├── page.tsx                       # Customer Portal Container (Directory, Menu, Checkout, Live Tracking)
│   │   └── merchant/
│   │       └── page.tsx                   # Kitchen / Merchant Portal Container (KDS, Menu, Vault)
│   │
│   ├── components/                        # Modular React Components (Categorized by Domain)
│   │   ├── common/                        # Shared cross-domain UI components
│   │   │   ├── Header.tsx                 # Navigation header, location address, portal links
│   │   │   └── WalletBadge.tsx            # Digital wallet & Web3 crypto balance indicator
│   │   │
│   │   ├── customer/                      # Customer discovery, browsing & cart components
│   │   │   ├── CuisineFilter.tsx          # Category filters and live restaurant search
│   │   │   ├── RestaurantCard.tsx         # Individual restaurant card (rating, ETA, badge)
│   │   │   ├── RestaurantDirectory.tsx    # Responsive grid of partner storefronts
│   │   │   ├── StorefrontBanner.tsx       # Restaurant hero banner & 95% retention guarantee
│   │   │   ├── DishCard.tsx               # Dish presentation with dual EUR/OEN pricing
│   │   │   ├── MenuCatalog.tsx            # Categorized dish grid with quick-add actions
│   │   │   └── CartDrawer.tsx             # Sticky slide-out cart, tips, eco-delivery & checkout CTA
│   │   │
│   │   ├── checkout/                      # Tri-Rail checkout & payment components
│   │   │   ├── CommissionSelector.tsx     # Dynamic platform commission buttons (0%, 3%, 5%, 10%)
│   │   │   ├── CardPaymentForm.tsx        # Instant credit/debit card form with brand badge
│   │   │   └── TriRailPaymentSelector.tsx # Payment method switcher (Card, In-App Wallet, Crypto)
│   │   │
│   │   ├── tracking/                      # Real-time order fulfillment & simulation components
│   │   │   ├── OrderTrackingStepper.tsx   # 5-stage visual progress stepper
│   │   │   ├── LiveVectorMap.tsx          # Vector street map (Naples) with moving courier polyline
│   │   │   ├── DoorstepCredentials.tsx    # Scannable barcode & customer verification PIN
│   │   │   ├── LifecycleSimulator.tsx     # Multi-party action triggers (Kitchen, Courier, Customer)
│   │   │   └── ActiveOrderTrackingView.tsx# Unified tracking layout
│   │   │
│   │   └── merchant/                      # Merchant operations & KDS portal components
│   │       ├── MerchantHeader.tsx         # Store status toggle, audio chime switch, navigation
│   │       ├── FinancialRibbon.tsx        # Real-time revenue retention metrics & legacy savings
│   │       ├── KdsTicketCard.tsx          # Real-time kitchen display ticket with prep ETA & actions
│   │       ├── MenuManager.tsx            # 86-item availability toggles & item removal
│   │       ├── AddDishModal.tsx           # New menu item creation modal
│   │       ├── CounterBarcodeModal.tsx    # Enlarged counter pickup verification barcode
│   │       └── VaultAnalytics.tsx         # Smart contract escrow settlement history & payout logs
│   │
│   ├── lib/                               # Core libraries, utilities & clients
│   │   ├── api.ts                         # Typed API client for all backend endpoints
│   │   ├── constants.ts                   # Environment defaults, exchange rates, commission configs
│   │   └── utils.ts                       # Currency formatters, card brand detection, helpers
│   │
│   └── types/                             # Domain type definitions
│       ├── index.ts                       # Barrel re-export for clean imports
│       ├── menu.ts                        # MenuItem, CartItem interfaces
│       ├── order.ts                       # OrderData, OrderItem, OrderStatus, TrackingStep
│       ├── payment.ts                     # PaymentMethodOption, CardPaymentData, CardIntentResponse
│       ├── restaurant.ts                  # Restaurant, RestaurantStats, RestaurantProfile
│       └── wallet.ts                      # WalletData, DigitalWallet, CryptoWallet
```

---

## 3. Component Specification & Boundaries

### 3.1 Common Layer (`src/components/common/`)
- **`Header.tsx`**: Renders brand identity, current delivery address modal/trigger, customer/merchant switcher, and embedded `WalletBadge`.
- **`WalletBadge.tsx`**: Shows connected wallet status (Digital Wallet balance in EUR, and Oenexa Crypto balance in OEN).

### 3.2 Customer Layer (`src/components/customer/`)
- **`CuisineFilter.tsx`**: Manages search query string and selected cuisine category pills.
- **`RestaurantCard.tsx`**: Clean presentation card displaying restaurant badge, rating, delivery time, and minimum order.
- **`RestaurantDirectory.tsx`**: Maps over filtered restaurants and renders grid layout.
- **`StorefrontBanner.tsx`**: Visual header for selected restaurant with back button, chef bio, and 95% retention badge.
- **`DishCard.tsx`**: Card for menu item with dual-pricing calculation (`formatEUR` & `formatOEN`), image, and add-to-cart trigger.
- **`MenuCatalog.tsx`**: Categorizes dishes into Starters, Mains, Desserts, and Beverages.
- **`CartDrawer.tsx`**: Slide-out cart displaying item quantities, tip selectors (0%, 5%, 10%, 15%), carbon-neutral delivery toggle, and embedded checkout drawer.

### 3.3 Checkout Layer (`src/components/checkout/`)
- **`CommissionSelector.tsx`**: Visual platform fee selection (0%, 3%, 5%, 10%). Calculates dynamic merchant payout and displays platform subsidy note for 0% fee.
- **`CardPaymentForm.tsx`**: Credit card inputs with live Luhn validation, real-time card brand detection (Visa, Mastercard, Amex, Discover), CVV masking, and 1-tap test presets.
- **`TriRailPaymentSelector.tsx`**: Tabbed selection between:
  1. Credit / Debit Card (Instant clearing via payment rails).
  2. In-App Digital Wallet (Pre-funded EUR balance).
  3. Web3 Crypto Wallet (Direct OEN transaction on Oenexa L1).

### 3.4 Tracking Layer (`src/components/tracking/`)
- **`OrderTrackingStepper.tsx`**: High-contrast 5-stage visual stepper (`CONFIRMED` → `PREPARING` → `READY_FOR_PICKUP` → `IN_TRANSIT` → `DELIVERED`).
- **`LiveVectorMap.tsx`**: Stylized SVG vector map depicting merchant location, animated courier route, and delivery destination.
- **`DoorstepCredentials.tsx`**: Generates high-density visual barcode for counter handoff and a 4-digit verification PIN for doorstep delivery.
- **`LifecycleSimulator.tsx`**: Sandbox controls enabling instant simulation of kitchen readiness, courier pickup, and doorstep PIN delivery without external webhooks.
- **`ActiveOrderTrackingView.tsx`**: Unified layout combining the map, credentials, stepper, and order item recap.

### 3.5 Merchant / KDS Layer (`src/components/merchant/`)
- **`MerchantHeader.tsx`**: Kitchen portal navigation with store open/closed switch and audio chime alerts for incoming orders.
- **`FinancialRibbon.tsx`**: Displays Gross Sales, 95% Merchant Retention, and Legacy Platform Savings ($ saved vs 30% aggregator fees).
- **`KdsTicketCard.tsx`**: Real-time kitchen display ticket showing order items, customer notes, prep time selector (15m, 25m, 40m), and "Mark Ready" action.
- **`MenuManager.tsx`**: Live catalog manager with in-stock/sold-out switches (86-item toggle) and dish removal actions.
- **`AddDishModal.tsx`**: Modal for creating new items with name, price in EUR, category, and image URL.
- **`CounterBarcodeModal.tsx`**: Full-screen modal presenting high-density pickup barcode for courier scanning.
- **`VaultAnalytics.tsx`**: Escrow settlement transparency table showing on-chain/instant payout events with tx hashes.

---

## 4. Developer & AI Agent Guidelines

Future engineers and autonomous AI agents working in this repository **must** strictly adhere to the following rules:

### Rule 1: No Monolithic Files
- **Maximum file size limit**: 300 lines of code for components; 400 lines for page orchestrators.
- If a file exceeds this limit, extract logical subcomponents into the appropriate `src/components/<category>/` directory.

### Rule 2: Strict Typing & Shared Contracts
- Never use `any` in component props or API calls.
- All new models must be added to `src/types/<category>.ts` and re-exported in `src/types/index.ts`.
- Ensure frontend types match backend response schemas in `oengo-api/server.js`.

### Rule 3: Network Centralization
- Never invoke `fetch('http://localhost:3001/...')` directly inside a component.
- Add typed methods to `src/lib/api.ts` with error handling and fallback defaults.

### Rule 4: Zero-Bug Quality Gate
Before submitting any pull request or committing code:
1. `oengo-web`: Execute `npm run build` — must finish with exit code 0 and zero TypeScript errors.
2. `oengo-api`: Execute `npm test` — all test suites must pass (15/15 passing).
3. `oenexa-node`: Execute `go test ./...` — must finish with exit code 0.
