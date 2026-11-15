# Deployment Guide: oengo-api

## GitOps Architecture
All deployments to Staging and Production are managed declaratively by **ArgoCD**.
- **Staging Application:** \oenexa-staging-oengo-api\
- **Production Application:** \oenexa-prod-oengo-api\

## Continuous Delivery Process
1. PR merged to \main\ triggers CI pipeline (lint, test, security scan, docker build).
2. Signed container image pushed to internal container registry with Cosign.
3. ArgoCD detects updated tag in Git and reconciles cluster state.
4. Health checks verified before traffic shift.