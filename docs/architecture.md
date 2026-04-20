# Architecture Overview

```
Internet
   │
   ▼
Envoy Gateway (envoy-gateway-system namespace)
   │
   ▼
api (Deployment, HPA, PDB)  ──►  mysql (StatefulSet)
   │
worker (Deployment)  ──────────►  mysql (StatefulSet)
   │
   ▼
OpenTelemetry Collector (monitoring namespace)
   │
   ├──► Prometheus ──► Grafana
   └──► Logging

GitHub Actions (CI)
   │  build + push image → GHCR
   │  commit image tag → git
   ▼
ArgoCD (argocd namespace)
   │  watches k8s/services/*/dev|uat|prod
   ├──► dev  (auto-sync)
   ├──► uat  (manual sync)
   └──► prod (manual sync)
```

## Components

| Component | Description |
|-----------|-------------|
| api | Go HTTP server exposing `/health` and `/metrics`. Sends traces via OTLP. |
| worker | Go background job running every 60s. Sends traces via OTLP. |
| mysql | MySQL 8.0 StatefulSet with persistent storage. |
| Envoy Gateway | Kubernetes Gateway API implementation for ingress. |
| OTel Collector | Receives OTLP traces/metrics, exports to Prometheus and logs. |
| Prometheus | Scrapes metrics from OTel Collector and Kubernetes pods. |
| Grafana | Dashboards and alerting UI backed by Prometheus. |
| ArgoCD | GitOps continuous delivery — syncs cluster state from git. |

## Namespaces

| Namespace | Purpose |
|-----------|---------|
| `dev` | Development environment |
| `uat` | User acceptance testing |
| `prod` | Production workloads |
| `monitoring` | Observability stack (OTel, Prometheus, Grafana) |
| `argocd` | ArgoCD GitOps controller |
| `envoy-gateway-system` | Envoy Gateway ingress controller |

## Repository Structure

```
.
├── services/
│   ├── api/               # Go API service (main.go, Dockerfile, go.mod)
│   └── worker/            # Go worker service (main.go, Dockerfile, go.mod)
├── k8s/
│   ├── namespaces.yaml
│   ├── argocd/
│   │   ├── helmfile.yaml  # ArgoCD helm install
│   │   ├── values.yaml    # ArgoCD helm values
│   │   ├── project.yaml   # AppProject
│   │   └── applications/
│   │       └── workloads/
│   │           ├── dev/   # Application manifests for dev
│   │           ├── uat/   # Application manifests for uat
│   │           └── prod/  # Application manifests for prod
│   ├── services/
│   │   ├── api/
│   │   │   ├── base/      # Deployment, Service, HPA, PDB (namespace-agnostic)
│   │   │   ├── dev/       # ConfigMap + resource/hpa patches for dev
│   │   │   ├── uat/       # ConfigMap + resource/hpa patches for uat
│   │   │   └── prod/      # ConfigMap + resource/hpa patches for prod
│   │   └── worker/
│   │       ├── base/      # Deployment, Service (namespace-agnostic)
│   │       ├── dev/       # ConfigMap + resource patch for dev
│   │       ├── uat/       # ConfigMap + resource patch for uat
│   │       └── prod/      # ConfigMap + resource patch for prod
│   ├── mysql/             # StatefulSet, Service, PVC, Secret
│   ├── rbac/              # ServiceAccounts, Roles, RoleBindings
│   ├── envoy-gateway/     # GatewayClass, Gateway, HTTPRoute
│   └── monitoring/        # OTel Collector, Prometheus, Grafana, Alerts
├── .github/workflows/
│   ├── ci.yaml            # Lint, test, scan, build, push
│   └── cd.yaml            # Commit image tag + ArgoCD sync per env
└── docs/
```
