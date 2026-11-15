# RB-001: Mitigating Service Degradation on oengo-api

## Symptoms
- Alert \High5xxRate\ or \HighLatency\ firing.
- Influx of HTTP 502/503 errors on the Ingress Gateway.

## Immediate Triage Steps
1. **Verify Pod Status:** \kubectl get pods -n oenexa-production -l app=oengo-api\
2. **Inspect Memory/CPU:** \kubectl top pods -n oenexa-production -l app=oengo-api\
3. **Check Downstream Connectivity:** Verify database connectivity and Redis cache response time.
4. **Trigger Pod Restart (if stuck):** \kubectl rollout restart deployment/oengo-api -n oenexa-production\
5. **Rollback (if post-deployment):** \rgocd app rollback oenexa-prod-oengo-api\
"@

# 7. GitOps K8s Manifests
Write-IfMissing "deploy/base/deployment.yaml" @"
apiVersion: apps/v1
kind: Deployment
metadata:
  name: oengo-api
  labels:
    app.kubernetes.io/name: oengo-api
    app.kubernetes.io/part-of: oenexa
spec:
  replicas: 2
  selector:
    matchLabels:
      app: oengo-api
  template:
    metadata:
      labels:
        app: oengo-api
    spec:
      containers:
        - name: oengo-api
          image: oenexa/
          ports:
            - containerPort: 8080
          readinessProbe:
            httpGet:
              path: /api/health/ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /api/health/live
              port: 8080
            initialDelaySeconds: 15
            periodSeconds: 20
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 512Mi