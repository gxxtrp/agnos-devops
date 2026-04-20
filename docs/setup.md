# Setup Instructions

## Prerequisites

- [minikube](https://minikube.sigs.k8s.io/docs/start/) >= 1.32
- [podman](https://podman.io/getting-started/installation) >= 4.x
- [kubectl](https://kubernetes.io/docs/tasks/tools/) >= 1.29
- [helm](https://helm.sh/docs/intro/install/) >= 3.14
- [helmfile](https://github.com/helmfile/helmfile/releases) >= 0.162

---

## 1. Start Minikube with Podman driver

```bash
minikube start \
  --driver=podman \
  --container-runtime=containerd \
  --cpus=4 \
  --memory=8192 \
  --kubernetes-version=v1.29.0
```

Enable required addons:

```bash
minikube addons enable metrics-server
minikube addons enable storage-provisioner
```

## 2. Create Namespaces

```bash
kubectl apply -f k8s/namespaces.yaml
```

## 3. Install Envoy Gateway

```bash
helm install eg oci://docker.io/envoyproxy/gateway-helm \
  --version v1.0.0 \
  --namespace envoy-gateway-system \
  --create-namespace

kubectl wait --timeout=5m \
  -n envoy-gateway-system \
  deployment/envoy-gateway \
  --for=condition=Available
```

## 4. Install ArgoCD

```bash
helmfile apply -f k8s/argocd/helmfile.yaml

kubectl wait --timeout=5m \
  -n argocd \
  deployment/argocd-server \
  --for=condition=Available
```

Get the initial admin password:

```bash
kubectl get secret argocd-initial-admin-secret -n argocd \
  -o jsonpath="{.data.password}" | base64 -d
```

## 5. Create GHCR Pull Secret

Replace placeholders with your values:

```bash
for ns in dev uat prod; do
  kubectl create secret docker-registry ghcr-secret \
    --docker-server=ghcr.io \
    --docker-username=<YOUR_GITHUB_USERNAME> \
    --docker-password=<YOUR_GITHUB_PAT> \
    --docker-email=<YOUR_EMAIL> \
    -n $ns
done
```

## 6. Configure MySQL Secret

Edit `k8s/mysql/secret.yaml` and replace the base64 placeholder values:

```bash
echo -n 'your-root-password' | base64
echo -n 'your-db-user' | base64
echo -n 'your-db-password' | base64
```

Then apply:

```bash
kubectl apply -f k8s/mysql/secret.yaml -n prod
kubectl apply -f k8s/mysql/secret.yaml -n uat
kubectl apply -f k8s/mysql/secret.yaml -n dev
```

## 7. Apply Base Infrastructure

```bash
# RBAC
kubectl apply -f k8s/rbac/

# MySQL
kubectl apply -f k8s/mysql/ -n prod

# Envoy Gateway routes
kubectl apply -f k8s/envoy-gateway/

# Monitoring stack
kubectl apply -f k8s/monitoring/

# ArgoCD AppProject + all Applications
kubectl apply -f k8s/argocd/project.yaml
kubectl apply -f k8s/argocd/applications/workloads/dev/
kubectl apply -f k8s/argocd/applications/workloads/uat/
kubectl apply -f k8s/argocd/applications/workloads/prod/
kubectl apply -f k8s/argocd/applications/systems/
kubectl apply -f k8s/argocd/applications/monitoring/
```

ArgoCD will take over syncing all resources from this point. See [argocd.md](argocd.md) for sync policies.
