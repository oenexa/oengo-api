# Operations Guide: oengo-api

## Service Level Objectives (SLOs)
- **Availability:** 99.99% uptime.
- **Latency:** P95 < 50ms, P99 < 150ms.
- **Error Budget:** <= 0.01% 5xx responses per 30-day window.

## Health Endpoints
- **Liveness:** \GET /api/health/live\ -> 200 OK
- **Readiness:** \GET /api/health/ready\ -> 200 OK
- **Metrics:** \GET /metrics\ -> Prometheus format