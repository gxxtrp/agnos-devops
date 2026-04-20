# GitHub Actions Runner for Local Cluster

The CD pipeline's ArgoCD sync jobs run on a `self-hosted` runner because GitHub-hosted runners cannot reach a local minikube cluster. This doc covers setup and all available methods for exposing ArgoCD to the pipeline.

---

## Method 1: Self-Hosted Runner (Recommended for local dev)

The runner runs on your local machine alongside minikube, so it can reach `localhost` directly.

### Install the runner

1. Go to your GitHub repo → Settings → Actions → Runners → New self-hosted runner
2. Select Linux → copy and run the provided commands, e.g.:

```bash
mkdir actions-runner && cd actions-runner

# Download (replace VERSION and HASH with values from GitHub UI)
curl -o actions-runner-linux-x64.tar.gz -L \
  https://github.com/actions/runner/releases/download/v2.x.x/actions-runner-linux-x64-2.x.x.tar.gz

tar xzf actions-runner-linux-x64.tar.gz

# Configure (use the token shown in GitHub UI)
./config.sh --url https://github.com/gxxtrp/agnos-devops --token <YOUR_TOKEN>

# Start
./run.sh
```

### Run as a systemd service (so it survives reboots)

```bash
sudo ./svc.sh install
sudo ./svc.sh start
sudo systemctl status actions.runner.*
```

### Set ARGOCD_SERVER secret

Since the runner is local, set:

```
ARGOCD_SERVER = localhost:8080
```

Port-forward ArgoCD before running the pipeline (or keep it running):

```bash
kubectl port-forward svc/argocd-server 8080:80 -n argocd
```

---

## Method 2: Expose ArgoCD via ngrok (Quick tunnel)

Useful for temporary access without a self-hosted runner.

```bash
# Install ngrok: https://ngrok.com/download
ngrok http 8080
```

ngrok gives you a public URL like `https://abc123.ngrok.io`. Set:

```
ARGOCD_SERVER = abc123.ngrok.io:443
```

Remove `--insecure` from the argocd login command in `cd.yaml` if using HTTPS.

Note: free ngrok URLs change on every restart.

---

## Method 3: Expose ArgoCD via cloudflared (Persistent tunnel)

More stable than ngrok, free tier available.

```bash
# Install cloudflared: https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/
cloudflared tunnel --url http://localhost:8080
```

Set the tunnel URL as `ARGOCD_SERVER`.

---

## Method 4: NodePort on minikube

Expose ArgoCD directly via NodePort and use the minikube IP.

```bash
kubectl patch svc argocd-server -n argocd \
  -p '{"spec": {"type": "NodePort"}}'

# Get the NodePort
kubectl get svc argocd-server -n argocd

# Get minikube IP
minikube ip
```

Set:

```
ARGOCD_SERVER = <minikube-ip>:<nodeport>
```

This only works if the GitHub Actions runner can reach the minikube network (i.e. self-hosted runner or same machine).

---

## Generate ArgoCD API Token

Required for `ARGOCD_TOKEN` secret regardless of method:

```bash
# Port-forward
kubectl port-forward svc/argocd-server 8080:80 -n argocd

# Login
argocd login localhost:8080 --username admin --password <initial-password> --insecure

# Generate token
argocd account generate-token --account admin
```

Copy the output into GitHub secret `ARGOCD_TOKEN`.

---

## Summary of GitHub Secrets Required

| Secret | Value |
|--------|-------|
| `GIT_TOKEN` | GitHub PAT with `repo` write scope |
| `ARGOCD_SERVER` | ArgoCD address reachable from the runner |
| `ARGOCD_TOKEN` | ArgoCD API token from `argocd account generate-token` |

---

## Adding Secrets to GitHub

Go to your repo → Settings → Secrets and variables → Actions → New repository secret

Direct URL:
```
https://github.com/gxxtrp/agnos-devops/settings/secrets/actions
```

### GIT_TOKEN

1. Go to `https://github.com/settings/tokens` (personal account settings, not repo settings)
2. Click "Generate new token (classic)"
3. Set expiration, check `repo` scope (full control of private repositories)
4. Copy the token — it's only shown once
5. Add as secret named `GIT_TOKEN`

### ARGOCD_SERVER

Value depends on your chosen method:

| Method | Value |
|--------|-------|
| Self-hosted runner | `localhost:8080` |
| ngrok | `abc123.ngrok.io:443` |
| cloudflared | your tunnel URL |
| NodePort | `<minikube-ip>:<nodeport>` |

Add as secret named `ARGOCD_SERVER`.

### ARGOCD_TOKEN

```bash
# Make sure ArgoCD is port-forwarded
kubectl port-forward svc/argocd-server 8080:80 -n argocd

# Login
argocd login localhost:8080 --username admin --password <initial-password> --insecure

# Generate token
argocd account generate-token --account admin
```

Copy the output and add as secret named `ARGOCD_TOKEN`.
