# Oengo API & Contracts Developer & AI Agent Strict Enforcement Policy 🛡️

> **MANDATORY POLICY FOR ALL HUMAN DEVELOPERS AND AI AGENTS (Antigravity, Cursor, Copilot, Claude, etc.)**  
> **EFFECTIVE DATE: IMMEDIATE — ZERO TOLERANCE FOR UNVERIFIED CODE**

This document defines the non-negotiable rules and quality gates for contributing to **`oengo-api`**. Every developer and AI assistant operating in this codebase is strictly bound by these rules.

---

## 1. The Zero-Bug & 100% Verification Mandate

1. **No Code Without Verification**: Code must be cleanly designed, defensively written, and verified locally before submission.
2. **100% Test Passing Rate for Smart Contracts**:
   - Zero test failures are tolerated in WASM smart contracts (`contracts/escrow/`).
   - Every contract method must have automated unit tests covering order creation, payment escrow, delivery confirmation, and refund pathways.
3. **API Reliability**:
   - Express API routes must handle invalid request bodies, missing fields, and RPC node timeouts with clean error responses (never unhandled exceptions or crash loops).

---

## 2. Mandatory Security & Escrow Protection

Before modifying or adding code that handles escrow funds, payments, or transaction broadcasting:

1. **Escrow Invariants**: Order funds must be guaranteed to be locked during transit and only released to the vendor upon valid user delivery confirmation or refund timeout.
2. **Web3 RPC Safety**: All calls to `oen_sendRawTransaction` must validate hex string signatures before broadcasting to the Layer-1 node.
3. **Secret Protection**: Never commit private keys, keystores, or database credentials to git. Use `.env` with `.env.example` templates.

---

## 3. Strict Main Branch Gate (Pre-Push Enforcement)

**NO CODE MAY BE PUSHED TO THE `main` BRANCH WITHOUT PASSING THE VERIFICATION CHECKLIST:**

### Pre-Push Verification Commands:
```bash
# 1. Test WASM Smart Contracts (Must be 100% PASS)
cd contracts/escrow
go test -v ./...

# 2. Verify API Gateway dependencies & syntax
npm test # or node -c server.js
```

**AI Agent Rule**: If you are an AI assistant executing commands or proposing commits, you **MUST** run the verification suite in your sandbox/terminal and verify a `0` exit code before executing `git push origin main`. Pushing failing code or skipping tests is a critical protocol violation.
