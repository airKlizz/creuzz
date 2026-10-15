# Installation

This guide covers installing the Creuzz operator on a Kubernetes cluster.

## Prerequisites

### Kubernetes Cluster

- **Kubernetes 1.31+** — Required for OCI image volume support (`image` volume type)
- **containerd 2.x** — Alternative to K8s 1.31+ for image volume support
- **Network Plugin** — Any CNI that supports NetworkPolicy (Calico, Cilium, Weave, etc.)

Verify image volume support:

```bash
kubectl get nodes -o jsonpath='{.items[0].status.nodeInfo.containerRuntimeVersion}'
# Should show: containerd://2.x.x or Kubernetes 1.31+
```

### Tools

- `kubectl` — Kubernetes CLI
- `docker` — For building the operator image
- `make` — For running build commands
- `go` 1.26+ — For development builds

### Cluster Access

Ensure `kubectl` is configured with admin access to your cluster:

```bash
kubectl cluster-info
kubectl auth can-i create clusterroles --all-namespaces
# Should return: yes
```

## Installation Methods

### Method 1: Deploy from Source (Recommended for Development)

```bash
# Clone the repository
git clone https://github.com/airklizz/creuzz.git
cd creuzz

# Install CRDs
make install

# Build and deploy the operator
make docker-build IMG=creuzz-operator:dev
make deploy IMG=creuzz-operator:dev
```

### Method 2: Deploy with Pre-built Image

```bash
# Install CRDs
kubectl apply -f https://raw.githubusercontent.com/airklizz/creuzz/main/config/crd/bases/

# Deploy the operator
kubectl apply -k config/default
```

### Method 3: Using Kustomize

```bash
# Build and apply all manifests
make build-installer IMG=creuzz-operator:latest
kubectl apply -f dist/install.yaml
```

## Verification

Check that the operator is running:

```bash
# Check CRDs are installed
kubectl get crds | grep creuzz.io

# Expected output:
# clusteragentegressrules.creuzz.io
# clusteragentenvs.creuzz.io
# clusteragentmcpservers.creuzz.io
# clusteragentskills.creuzz.io
# clusteragenttools.creuzz.io
# opencodeagents.creuzz.io

# Check operator deployment
kubectl get deployment -n creuzz-system creuzz-controller-manager

# Check operator logs
kubectl logs -n creuzz-system deployment/creuzz-controller-manager -f
```

Expected log output:
```
INFO  Starting manager
INFO  Starting Controller  {"controller": "opencodeagent", "controllerGroup": "creuzz.io"}
INFO  Starting workers     {"controller": "opencodeagent", "worker count": 1}
```

## Configuration

### Leader Election

By default, leader election is enabled. To disable (single replica):

```bash
kubectl patch deployment creuzz-controller-manager -n creuzz-system \
  --type='json' -p='[{"op": "remove", "path": "/spec/template/spec/containers/0/args/0"}]'
```

### Metrics

Metrics are exposed on port 8443 (HTTPS) by default. To access:

```bash
kubectl port-forward -n creuzz-system deployment/creuzz-controller-manager 8443:8443
curl -k https://localhost:8443/metrics
```

### Resource Limits

Default resource limits for the operator:

| Resource | Request | Limit |
|----------|---------|-------|
| CPU | 10m | 500m |
| Memory | 64Mi | 128Mi |

To adjust, edit `config/manager/manager.yaml`:

```yaml
resources:
  limits:
    cpu: 1000m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

## Upgrading

```bash
# Pull latest changes
git pull

# Regenerate manifests
make manifests generate

# Rebuild and redeploy
make docker-build IMG=creuzz-operator:new-tag
make deploy IMG=creuzz-operator:new-tag
```

## Uninstallation

```bash
# Delete the operator
make undeploy

# Delete CRDs (this will delete all Creuzz resources!)
make uninstall

# Or manually
kubectl delete -k config/default
kubectl delete -f config/crd/bases/
```

## Next Steps

- [Getting Started](getting-started.md) — Create your first agent
- [Architecture](architecture.md) — Understand the design
