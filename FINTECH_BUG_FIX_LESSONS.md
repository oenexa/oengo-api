# 🛠️ FinTech Bug Fixes: Developer Learning Guide

This document explains the critical financial bugs we found in the early Ledger design and exactly how we fixed them. Use this as a reference when building future financial features to avoid common pitfalls!

---

### 1. The "Lost Update" Bug (Race Conditions)
**What it was:**
If two payments arrived at the exact same millisecond, the code would load the wallet balance (e.g., €50), add the money in Go memory, and save it back. If both processes loaded €50 at the same time, they would both save €60, instead of the correct €70!

**How we fixed it:**
We stopped doing the math in Go memory. We switched to an **Atomic SQL Update**. By using `gorm.Expr`, we tell the database to do the math directly in the database engine, locking the row securely.

*Bad Code (Go Memory Math):*
```go
wallet.Balance = wallet.Balance + amount
tx.Save(&wallet) 
```

*Good Code (Atomic SQL Math):*
```go
tx.Model(&wallet).UpdateColumn("balance", gorm.Expr("balance + ?", amount))
```

---

### 2. The "Double Billing" Bug (Network Retries)
**What it was:**
If a user clicked "Pay", but their WiFi dropped, the frontend might retry the payment. The server would process both requests, charging the user twice for one order.

**How we fixed it:**
We added an **Idempotency Key** to the `LedgerTransaction` table with a `UNIQUE` database index. Now, every checkout request sends a unique string (like an order ID + timestamp). If a retry happens, the database instantly blocks the duplicate key.

---

### 3. The "Rounding Error" Bug (Floating Point Math)
**What it was:**
Computers struggle to store decimals as `float64`. Calculating `0.1 + 0.2` in standard code actually equals `0.30000000000000004`. In a banking app, these micro-cents add up and destroy the accounting balance.

**How we fixed it:**
We replaced all `float64` types with the `decimal.Decimal` library in Go, and used `NUMERIC(28,8)` in PostgreSQL. This ensures exact math with zero precision loss, even up to 8 decimal places (perfect for crypto/tokens).

---

### 4. The "Rogue Edit" Bug (Mutable Data)
**What it was:**
A ledger is supposed to be a permanent receipt. But originally, there was nothing stopping another developer's code (or a database admin) from updating or deleting a past transaction.

**How we fixed it:**
1. **App Level:** Added GORM Hooks (`BeforeUpdate` and `BeforeDelete`) that automatically return a hard error (`"immutable record"`) if anyone tries to save over an existing ledger entry.
2. **Data Level:** Added a `SHA-256` Cryptographic Hash chain. Every new transaction hashes itself together with the `PreviousHash`. If anyone secretly edits a row in the database, the hashes won't match, and the tamper is instantly detected!
