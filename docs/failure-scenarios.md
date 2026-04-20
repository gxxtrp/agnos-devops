# Failure Scenario Handling

## a. API crashes during peak hours

The system is designed to self-heal:

- HPA automatically scales up replicas (up to 10) when CPU exceeds 70%
- PodDisruptionBudget (`minAvailable: 1`) ensures at least one pod stays running during voluntary disruptions
- livenessProbe on `/health` restarts crashed pods automatically (after `initialDelaySeconds: 15`)
- readinessProbe removes unhealthy pods from the load balancer before they receive traffic
- Pod anti-affinity spreads replicas across nodes to avoid single-node failures

Manual rollback if needed:

```bash
kubectl rollout undo deployment/api -n prod
kubectl rollout status deployment/api -n prod
```

---

## b. Worker fails and infinitely retries

- livenessProbe (`exec: cat /tmp/worker-alive`) kills the container if the alive file is not updated, breaking the retry loop
- Kubernetes restartPolicy with exponential backoff prevents tight crash loops
- The `WorkerStalled` alert fires if no job runs for 5 minutes
- The `CrashLooping` alert fires if restarts exceed 5 in 15 minutes

To inspect and recover:

```bash
kubectl describe pod -l app=worker -n prod
kubectl logs -l app=worker -n prod --previous
kubectl delete pod -l app=worker -n prod  # force reschedule
```

---

## c. Bad deployment released

Every image is tagged with the git SHA, making every release fully traceable and reversible.

Immediate rollback:

```bash
# Roll back to previous revision
kubectl rollout undo deployment/api -n prod

# Roll back to a specific revision
kubectl rollout history deployment/api -n prod
kubectl rollout undo deployment/api --to-revision=<N> -n prod
```

Prevention layers:
- CI gates (lint → test → security scan → build) must all pass before CD runs
- CD gates: dev → uat (manual approval) → prod (manual approval)
- Images are immutable and tagged by SHA — no `latest` overwrites in prod

---

## d. Kubernetes node down

- Pod anti-affinity (`preferredDuringSchedulingIgnoredDuringExecution`) spreads API replicas across nodes so a single node failure doesn't take down all pods
- HPA detects reduced capacity and reschedules pods on healthy nodes to maintain replica count
- PodDisruptionBudget ensures Kubernetes doesn't evict too many pods at once during node drain
- StatefulSet for MySQL automatically reschedules the pod on another node; the PVC is reattached (requires `ReadWriteOnce` storage that supports node migration, or use a distributed storage class)

For MySQL high availability in production, consider:

```bash
# Use a storage class that supports multi-node access
# or deploy MySQL with a replication operator (e.g., mysql-operator)
helm install mysql-operator mysql-operator/mysql-operator -n prod
```
