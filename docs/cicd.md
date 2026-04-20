# CI/CD Pipeline

## CI (`ci.yaml`) — triggers on push/PR to `main` and `develop`

1. lint — golangci-lint for `services/api` and `services/worker`
2. test — `go test ./...` with race detection
3. security-scan — Trivy filesystem scan on source + config scan on Dockerfiles
4. build — builds and pushes Docker images to GHCR tagged with the commit SHA

Images are pushed to:
- `ghcr.io/OWNER/agnos-api:<sha>`
- `ghcr.io/OWNER/agnos-worker:<sha>`

## CD (`cd.yaml`) — triggers on push to `main`

CD no longer runs `kubectl apply` directly. Instead it:

1. update-image-tags — commits the new image SHA into `k8s/services/*/base/deployment.yaml` and pushes to git
2. deploy-dev — triggers `argocd app sync` for `api-dev` and `worker-dev`, waits for healthy
3. deploy-uat — requires manual approval (GitHub environment gate), then syncs `api-uat` and `worker-uat`
4. deploy-prod — requires manual approval, then syncs `api-prod` and `worker-prod`

ArgoCD handles the actual apply to the cluster from git. The CD pipeline is just the trigger.

## Promotion Flow

```
push to main
     │
     ▼
CI: lint → test → scan → build → push to GHCR
     │
     ▼
CD: commit image tag to git
     │
     ▼
ArgoCD detects change → auto-sync dev
     │
     ▼ (manual approval in GitHub)
ArgoCD sync uat
     │
     ▼ (manual approval in GitHub)
ArgoCD sync prod
```

## Required GitHub Secrets

| Secret | Description |
|--------|-------------|
| `GIT_TOKEN` | GitHub PAT with repo write access (for image tag commits) |
| `ARGOCD_SERVER` | ArgoCD server address (e.g. `localhost:8080`) |
| `ARGOCD_TOKEN` | ArgoCD API token — generate via UI: Settings → Accounts → Tokens |

## Required GitHub Environments

Create these in repo Settings → Environments with required reviewers:

- `dev` — no approval needed (auto-deploys)
- `uat` — add required reviewers
- `prod` — add required reviewers
