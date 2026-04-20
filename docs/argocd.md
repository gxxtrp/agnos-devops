# ArgoCD

ArgoCD handles all deployments via GitOps. GitHub Actions CI builds and pushes images, then CD commits the new image tag to git — ArgoCD detects the change and syncs the cluster.

## Flow

```
CI: build + push image to GHCR
        │
        ▼
CD: commit new image tag to k8s/services/*/base/deployment.yaml
        │
        ▼
ArgoCD detects git change
        │
   ┌────┴──────────────────────┐
   ▼                            ▼
api-dev / worker-dev        (auto-sync)
   │
   ▼ (manual approval)
api-uat / worker-uat
   │
   ▼ (manual approval)
api-prod / worker-prod
```

## Install ArgoCD on Minikube

```bash
# Install helmfile if not already installed
# https://github.com/helmfile/helmfile/releases

helmfile apply -f k8s/argocd/helmfile.yaml

# Wait for it to be ready
kubectl wait --timeout=5m -n argocd deployment/argocd-server --for=condition=Available

# Get initial admin password
kubectl get secret argocd-initial-admin-secret -n argocd \
  -o jsonpath="{.data.password}" | base64 -d

# Port-forward the UI
kubectl port-forward svc/argocd-server 8080:80 -n argocd
# Open http://localhost:8080 — user: admin
```

## Apply ArgoCD Resources

```bash
# AppProject
kubectl apply -f k8s/argocd/project.yaml

# Workloads (api + worker per env)
kubectl apply -f k8s/argocd/applications/workloads/dev/
kubectl apply -f k8s/argocd/applications/workloads/uat/
kubectl apply -f k8s/argocd/applications/workloads/prod/

# Systems (mysql, rbac, envoy-gateway)
kubectl apply -f k8s/argocd/applications/systems/

# Monitoring (otel, prometheus, grafana, alerts)
kubectl apply -f k8s/argocd/applications/monitoring/
```

## Sync Policies

| App | Sync | Notes |
|-----|------|-------|
| api-dev | Auto | Syncs on git change, self-heals drift |
| worker-dev | Auto | Syncs on git change, self-heals drift |
| api-uat | Manual | Triggered by CD pipeline after GitHub approval |
| worker-uat | Manual | Triggered by CD pipeline after GitHub approval |
| api-prod | Manual | Triggered by CD pipeline after GitHub approval |
| worker-prod | Manual | Triggered by CD pipeline after GitHub approval |

## Required GitHub Secrets

| Secret | Description |
|--------|-------------|
| `GIT_TOKEN` | GitHub PAT with repo write access (for image tag commits) |
| `ARGOCD_SERVER` | ArgoCD server address (e.g. `localhost:8080`) |
| `ARGOCD_TOKEN` | ArgoCD API token — generate via UI: Settings → Accounts → Tokens |

## Manual Sync via CLI

```bash
argocd login $ARGOCD_SERVER --auth-token $ARGOCD_TOKEN --grpc-web --insecure

argocd app sync api-prod
argocd app wait api-prod --health --timeout 180
```

## Rollback

```bash
# View history
argocd app history api-prod

# Roll back to a previous revision
argocd app rollback api-prod <revision-id>
```
