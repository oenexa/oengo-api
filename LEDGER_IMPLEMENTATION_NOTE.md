# 🏦 OENGO LEDGER IMPLEMENTATION CHEATSHEET

Here is the symbolic roadmap for what we must implement next for the Double-Entry Ledger System.

### ⚙️ 1. Core Ledger Engine (The Brain)
*   **⚖️ `PostTransaction(tx *gorm.DB, entries []LedgerEntry)`**
    *   *Rule*: `SUM(Debits) == SUM(Credits)`. Reject if not zero.
    *   *Rule*: Execute purely inside an Atomic SQL Transaction (`tx.Begin()`).
    *   *Rule*: Append-only. Never update/delete past entries.

### 🛒 2. Order Lifecycle (The Flow)
*   **🔒 Step A: Order Placed (Escrow Lock)**
    *   🔴 `DEBIT`: Customer Wallet Account (Asset)
    *   🟢 `CREDIT`: Platform Escrow Account (Liability)
*   **🚀 Step B: Order Delivered (Settlement)**
    *   🔴 `DEBIT`: Platform Escrow Account (Liability)
    *   🟢 `CREDIT`: Restaurant Wallet (Revenue - 95%)
    *   🟢 `CREDIT`: Courier Wallet (Delivery Fee + Tip)
    *   🟢 `CREDIT`: Platform Wallet (Commission - 5%)
*   **⏪ Step C: Order Cancelled (Refund)**
    *   🔴 `DEBIT`: Platform Escrow Account (Liability)
    *   🟢 `CREDIT`: Customer Wallet Account (Asset)

### 💰 3. Wallet API Services (The Edge)
*   **💵 `GetBalance(userID)`**
    *   Read cached `Wallet.Balance`.
    *   *Periodic Sync*: `SELECT SUM(amount) WHERE direction='CREDIT' - SUM(amount) WHERE direction='DEBIT'`.
*   **🏦 `DepositFunds(userID, amount)`**
    *   🔴 `DEBIT`: External Bank/Stripe (Asset)
    *   🟢 `CREDIT`: Customer Wallet (Asset)
*   **🏧 `WithdrawFunds(userID, amount)`**
    *   🔴 `DEBIT`: User Wallet (Asset)
    *   🟢 `CREDIT`: External Bank Payout (Asset)

### 🛡️ 4. Future-Proofing
*   **🧾 Idempotency**: Use `Idempotency-Key` headers on all `/deposit` or `/checkout` routes to prevent accidental double-billing if the network drops.
*   **🕵️‍♂️ Reconciliation Cron**: A nightly script to verify `Wallet.Balance == Ledger SUM()`.
