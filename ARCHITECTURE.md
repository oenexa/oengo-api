# Architecture Specification: oengo-api

> **Domain:** Commerce/Backend  
> **Status:** APPROVED & ACTIVE  
> **Classification:** Enterprise Architecture Baseline  

## 1. Domain Purpose & Responsibilities
Unified API service for OENGO marketplace order matching, driver routing, and merchant catalog dispatch

## 2. Architectural Boundaries & Invariants
- **Single Responsibility:** oengo-api manages exclusively its bounded context.
- **Data Sovereignty:** Exclusively owns its datastore (db_oengo_api). Zero cross-service database access.
- **Inter-service Communication:** Strictly via REST/gRPC synchronous APIs or CloudEvents asynchronous message bus.
- **Secret Management:** Dynamically fetched via Vault; zero static credentials.

## 3. Component Architecture & Data Flow
\\\mermaid
flowchart LR
    Client["Client / API Gateway"] -->|"REST / gRPC"| Service["oengo-api"]
    Service -->|"Read/Write"| DB[("Private Datastore")]
    Service -->|"CloudEvents"| Kafka["Kafka / NATS Event Bus"]
\\\
"@

Write-IfMissing "ROADMAP.md" @"
# Roadmap: oengo-api

## Q1 2026 - Q2 2026: Foundation & High Availability
- [x] Initial service scaffolding and domain interface definition.
- [x] GitOps pipeline integration via ArgoCD.
- [x] Enterprise observability (Prometheus/Grafana) integration.

## Q3 2026 - Q4 2026: Scale & Optimization
- [ ] Automated chaos engineering resilience validation.
- [ ] Post-quantum cryptography readiness audit.
- [ ] Multi-region active-active read replica support.