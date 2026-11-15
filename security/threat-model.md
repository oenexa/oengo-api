# Threat Model: oengo-api

## STRIDE Analysis
- **Spoofing:** Mitigated via mTLS between internal services and JWT validation from \id.oenexa.com\.
- **Tampering:** In-transit TLS 1.3 encryption; database encryption at rest (AES-256-GCM).
- **Repudiation:** Cryptographically signed immutable audit logs.
- **Information Disclosure:** Strict PII redaction and zero-log policy for sensitive keys.
- **Denial of Service:** Ingress rate limiting and HPA autoscaling.
- **Elevation of Privilege:** Strict RBAC and zero-trust service account permissions.