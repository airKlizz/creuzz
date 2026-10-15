# Troubleshooting

Common issues and solutions for the Creuzz operator.

## Installation Issues

### CRDs Not Installing

**Symptom**: `make install` fails or CRDs don't appear

```bash
# Check if CRDs exist
kubectl get crds | grep creuzz.io

# Manual installation
kubectl apply -f config/crd/bases/

# Check for errors
kubectl apply -f config/crd/bases/ --dry-run=client
```

**Common causes**:
- Insufficient RBAC permissions
- Invalid CRD schema
- Kustomize not installed

### Operator Pod Not Starting

**Symptom**: `creuzz-controller-manager` pod in `CrashLoopBackOff` or `Pending`

```bash
# Check pod status
kubectl get pods -n creuzz-system

# View logs
kubectl logs -n creuzz-system deployment/creuzz-controller-manager

# Describe pod for events
kubectl describe pod -n creuzz-system -l control-plane=controller-manager
```

**Common causes**:
- Image pull errors (check `imagePullSecrets`)
- Resource limits too low
- Missing RBAC permissions

## Agent Issues

### Agent Stuck in Pending

**Symptom**: `OpenCodeAgent` phase is `Pending` for a long time

```bash
# Check agent status
kubectl get opencodeagent <name> -n <namespace> -o yaml

# Check events
kubectl get events -n <namespace> --sort-by='.lastTimestamp' | grep <name>
```

**Common causes**:
- Missing cluster-scoped references (check `status.conditions`)
- Insufficient cluster resources
- PVC not binding (check StorageClass)

### Agent in Failed State

**Symptom**: `OpenCodeAgent` phase is `Failed`

```bash
# Check failure reason
kubectl get opencodeagent <name> -n <namespace> -o jsonpath='{.status.conditions[0].message}'
```

**Common causes**:

| Error Message | Solution |
|--------------|----------|
| `ClusterAgentTool "x" not found` | Create the missing `ClusterAgentTool` |
| `ClusterAgentSkill "x" not found` | Create the missing `ClusterAgentSkill` |
| `ClusterAgentMCPServer "x" not found` | Create the missing `ClusterAgentMCPServer` |
| `ClusterAgentEnv "x" not found` | Create the missing `ClusterAgentEnv` |
| `ClusterAgentEgressRule "x" not found` | Create the missing `ClusterAgentEgressRule` |

### Pod Not Starting

**Symptom**: Agent is `Ready` but Pod is not running

```bash
# Check pod status
kubectl get pod <agent-name>-agent -n <namespace>

# Describe pod
kubectl describe pod <agent-name>-agent -n <namespace>

# Check init container logs
kubectl logs <agent-name>-agent -n <namespace> -c git-clone

# Check main container logs
kubectl logs <agent-name>-agent -n <namespace> -c opencode
```

**Common causes**:

| Issue | Solution |
|-------|----------|
| `ImagePullBackOff` | Check image reference and pull secrets |
| `CrashLoopBackOff` | Check container logs for errors |
| `Init:0/1` stuck | Git clone failing (check SSH secret, repo URL) |
| `CreateContainerConfigError` | Missing secret or ConfigMap |

### Git Clone Failing

**Symptom**: Init container `git-clone` fails

```bash
# Check init container logs
kubectl logs <agent-name>-agent -n <namespace> -c git-clone
```

**Common causes**:

| Error | Solution |
|-------|----------|
| `Permission denied (publickey)` | Check SSH secret has `id_ed25519` key |
| `Host key verification failed` | Add host to `known_hosts` in secret |
| `Repository not found` | Verify `repoUrl` and credentials |
| `Could not resolve hostname` | Check egress rules allow git host |

**Verify secret**:
```bash
kubectl get secret github-ssh-creds -n <namespace> -o yaml
# Should contain: id_ed25519, known_hosts
```

## Network Issues

### Agent Can't Reach External API

**Symptom**: Agent can't connect to allowed external service

```bash
# Test connectivity from pod
kubectl exec <agent-pod> -n <namespace> -- curl -v https://api.example.com

# Check NetworkPolicy
kubectl get networkpolicy <agent-name>-netpol -n <namespace> -o yaml

# Check DNS
kubectl exec <agent-pod> -n <namespace> -- nslookup api.example.com
```

**Solutions**:
1. Add `ClusterAgentEgressRule` for the domain
2. Reference it in the agent's `clusterEgressRefs`
3. Wait for NetworkPolicy update (automatic)

### DNS Resolution Failing

**Symptom**: `nslookup` fails inside agent pod

```bash
# Check kube-dns is running
kubectl get pods -n kube-system -l k8s-app=kube-dns

# Test DNS from pod
kubectl exec <agent-pod> -n <namespace> -- nslookup kubernetes.default
```

**Solutions**:
- Verify kube-dns/CoreDNS is running
- Check NetworkPolicy allows DNS (port 53 to kube-system)
- Check CNI plugin is functioning

### MCP Server Not Connecting

**Symptom**: OpenCode can't connect to MCP server

```bash
# Check MCP sidecar is running
kubectl get pod <agent-pod> -n <namespace> -o jsonpath='{.spec.containers[*].name}'

# Check MCP server logs
kubectl logs <agent-pod> -n <namespace> -c mcp-<server-name>

# Test localhost connection
kubectl exec <agent-pod> -n <namespace> -c opencode -- curl http://localhost:8001/health
```

**Solutions**:
- Verify port matches in `ClusterAgentMCPServer` spec
- Check MCP server image is correct
- Verify secret exists in agent's namespace

## Resource Issues

### PVC Not Binding

**Symptom**: `PersistentVolumeClaim` stuck in `Pending`

```bash
kubectl get pvc <agent-name>-workspace -n <namespace>
kubectl describe pvc <agent-name>-workspace -n <namespace>
```

**Solutions**:
- Check StorageClass exists: `kubectl get storageclass`
- Verify cluster has dynamic provisioning
- Check resource quotas in namespace

### Out of Resources

**Symptom**: Pod stuck in `Pending` with `Insufficient cpu/memory`

```bash
kubectl describe pod <agent-pod> -n <namespace> | grep -A 5 "Events:"
```

**Solutions**:
- Reduce `spec.runtime.cpu` and `spec.runtime.memory`
- Add more nodes to cluster
- Check cluster autoscaler

## Debugging Commands

### Collect Debug Information

```bash
#!/bin/bash
# debug-agent.sh <namespace> <agent-name>

NS=$1
NAME=$2

echo "=== Agent Status ==="
kubectl get opencodeagent $NAME -n $NS -o yaml

echo "=== Pod Status ==="
kubectl get pod $NAME-agent -n $NS -o yaml

echo "=== Pod Events ==="
kubectl describe pod $NAME-agent -n $NS

echo "=== Pod Logs (init) ==="
kubectl logs $NAME-agent -n $NS -c git-clone --tail=50

echo "=== Pod Logs (main) ==="
kubectl logs $NAME-agent -n $NS -c opencode --tail=50

echo "=== ConfigMap ==="
kubectl get configmap $NAME-config -n $NS -o yaml

echo "=== NetworkPolicy ==="
kubectl get networkpolicy $NAME-netpol -n $NS -o yaml

echo "=== Service ==="
kubectl get service $NAME-svc -n $NS -o yaml

echo "=== Operator Logs ==="
kubectl logs -n creuzz-system deployment/creuzz-controller-manager --tail=100
```

### Enable Debug Logging

```bash
# Patch operator to enable debug logging
kubectl patch deployment creuzz-controller-manager -n creuzz-system --type='json' \
  -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--zap-log-level=debug"}]'
```

## Getting Help

If you can't resolve the issue:

1. Collect debug information (see above)
2. Check [GitHub Issues](https://github.com/airklizz/creuzz/issues)
3. Open a new issue with:
   - Creuzz version
   - Kubernetes version
   - Error messages
   - Debug output
   - Steps to reproduce

## See Also

- [Installation](installation.md)
- [Architecture](architecture.md)
- [Guides](guides/)
