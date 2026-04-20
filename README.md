# Agnos DevOps

Production-ready Kubernetes deployment for the Agnos platform — API service, background worker, and MySQL, with full observability, GitOps via ArgoCD, and Envoy Gateway ingress.

## Docs

- [Architecture Overview](docs/architecture.md)
- [Setup Instructions](docs/setup.md)
- [Usage](docs/usage.md)
- [CI/CD Pipeline](docs/cicd.md)
- [ArgoCD](docs/argocd.md)
- [Failure Scenario Handling](docs/failure-scenarios.md)

## Quick Start

```bash
# 1. Start minikube
minikube start --driver=docker --cpus=4 --memory=8192

# 2. Create namespaces
kubectl apply -f k8s/namespaces.yaml

# 3. Install Envoy Gateway
helm install eg oci://docker.io/envoyproxy/gateway-helm \
  --version v1.0.0 --namespace envoy-gateway-system --create-namespace

# 4. Install ArgoCD
helmfile apply -f k8s/argocd/helmfile.yamla

# 5. Apply base infrastructure
kubectl apply -f k8s/rbac/
kubectl apply -f k8s/mysql/ -n prod
kubectl apply -f k8s/envoy-gateway/
kubectl apply -f k8s/monitoring/

# 6. Register ArgoCD apps
kubectl apply -f k8s/argocd/project.yaml
kubectl apply -f k8s/argocd/applications/workloads/dev/
kubectl apply -f k8s/argocd/applications/workloads/uat/
kubectl apply -f k8s/argocd/applications/workloads/prod/
kubectl apply -f k8s/argocd/applications/systems/
kubectl apply -f k8s/argocd/applications/monitoring/
```

See [docs/setup.md](docs/setup.md) for the full guide including secrets and GHCR pull secret configuration.
