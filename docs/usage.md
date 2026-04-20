# Usage Instructions

## Access the API

```bash
kubectl port-forward svc/api 8080:80 -n prod

# Health check
curl http://localhost:8080/health

# Metrics
curl http://localhost:8080/metrics
```

## Access Grafana

```bash
kubectl port-forward svc/grafana 3000:3000 -n monitoring
# Open http://localhost:3000 — default user: admin
```

## Access Prometheus

```bash
kubectl port-forward svc/prometheus 9090:9090 -n monitoring
# Open http://localhost:9090
```

## Access ArgoCD UI

```bash
kubectl port-forward svc/argocd-server 8080:80 -n argocd
# Open http://localhost:8080 — user: admin
```

## Check Logs

```bash
# API logs
kubectl logs -l app=api -n prod -f

# Worker logs
kubectl logs -l app=worker -n prod -f
```

## Trigger a Deployment

Push to `main` — CI builds and pushes the image, CD commits the tag, ArgoCD syncs dev automatically. UAT and prod require manual approval in GitHub Actions.

To manually sync via ArgoCD CLI:

```bash
argocd app sync api-prod
argocd app sync worker-prod
```

## Rollback

Via ArgoCD (recommended):

```bash
argocd app history api-prod
argocd app rollback api-prod <revision-id>
```

Via kubectl (emergency):

```bash
kubectl rollout undo deployment/api -n prod
kubectl rollout undo deployment/worker -n prod
```
