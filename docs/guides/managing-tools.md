# Managing Tools

This guide covers how to create, update, and manage `ClusterAgentTool` resources.

## Overview

Tools are OCI images containing CLI utilities. They are mounted as read-only volumes in agent pods and added to the agent's `PATH`.

## Creating a Tool

### Basic Tool

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentTool
metadata:
  name: terraform-1-5
spec:
  image: "hashicorp/terraform:1.5.7"
  mountPath: "/opt/tools/terraform"
  binPath: "/opt/tools/terraform/bin"
```

### Tool with Custom Entrypoint

Some images need specific binaries. Ensure `binPath` points to the correct directory:

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentTool
metadata:
  name: kubectl-1-28
spec:
  image: "bitnami/kubectl:1.28"
  mountPath: "/opt/tools/kubectl"
  binPath: "/opt/bitnami/kubectl/bin"  # Note: bitnami uses different path
```

### Multi-Tool Image

If an image contains multiple tools, you can reference the same image with different names:

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentTool
metadata:
  name: aws-cli
spec:
  image: "amazon/aws-cli:2.15.0"
  mountPath: "/opt/tools/aws"
  binPath: "/usr/local/bin"
---
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentTool
metadata:
  name: aws-cli-v2
spec:
  image: "amazon/aws-cli:2.15.0"
  mountPath: "/opt/tools/aws-v2"
  binPath: "/usr/local/bin"
```

## Versioning Strategy

### Semantic Versioning in Names

Use version numbers in resource names for clarity:

| Name | Image | Use Case |
|------|-------|----------|
| `terraform-1-5` | `hashicorp/terraform:1.5.7` | Legacy projects |
| `terraform-1-6` | `hashicorp/terraform:1.6.0` | Current projects |
| `node-18` | `node:18-alpine` | Legacy Node.js |
| `node-20` | `node:20-alpine` | Current Node.js |
| `node-22` | `node:22-alpine` | Latest Node.js |

### Updating a Tool

To update a tool, create a new version and update agent references:

```bash
# 1. Create new version
kubectl apply -f terraform-1-6.yaml

# 2. Update agents to use new version
kubectl patch opencodeagent my-agent -n team-backend --type='json' \
  -p='[{"op": "replace", "path": "/spec/clusterToolRefs/0/name", "value": "terraform-1-6"}]'

# 3. Verify rollout
kubectl get pod my-agent-agent -n team-backend -w

# 4. Clean up old version (optional)
kubectl delete clusteragenttool terraform-1-5
```

## Internal Tools

For internal tools, use your private registry:

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentTool
metadata:
  name: internal-cli
spec:
  image: "registry.internal.company.com/tools/cli:v2.1.0"
  mountPath: "/opt/tools/internal-cli"
  binPath: "/opt/tools/internal-cli/bin"
```

Ensure image pull secrets are configured for the agent namespace.

## Building Custom Tool Images

### Dockerfile Example

```dockerfile
FROM alpine:3.19

# Install tools
RUN apk add --no-cache curl jq bash

# Add custom scripts
COPY scripts/ /opt/tools/bin/
RUN chmod +x /opt/tools/bin/*

# Set PATH
ENV PATH="/opt/tools/bin:${PATH}"
```

### Build and Push

```bash
# Build
docker build -t registry.internal/tools/my-tools:1.0.0 .

# Push
docker push registry.internal/tools/my-tools:1.0.0

# Create ClusterAgentTool
cat << EOF | kubectl apply -f -
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentTool
metadata:
  name: my-tools
spec:
  image: "registry.internal/tools/my-tools:1.0.0"
  mountPath: "/opt/tools/my-tools"
  binPath: "/opt/tools/bin"
EOF
```

## PATH Construction

The agent's `PATH` is constructed by concatenating all `binPath` values:

```
PATH = binPath(tool1):binPath(tool2):...:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
```

Example with two tools:
```yaml
# Agent references:
clusterToolRefs:
  - name: "terraform-1-5"   # binPath: /opt/tools/terraform/bin
  - name: "node-20"         # binPath: /opt/tools/node/bin

# Resulting PATH:
PATH="/opt/tools/terraform/bin:/opt/tools/node/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
```

## Verification

Check tool availability in a running agent:

```bash
# Exec into the agent pod
kubectl exec -it <agent-pod> -n <namespace> -- sh

# Check PATH
echo $PATH

# Verify tools are available
terraform version
node --version
npm --version
```

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Tool not found | Verify `binPath` matches binary location in image |
| Permission denied | Ensure binaries are executable in the image |
| Image pull fails | Check image reference and pull secrets |
| Wrong version | Verify image tag matches expected version |
| PATH not updated | Pod needs to be recreated (tools are immutable) |

## Best Practices

1. **Pin versions** — Always use specific tags, never `latest`
2. **Keep images small** — Use Alpine or Distroless bases
3. **Test images** — Verify tools work before adding to catalog
4. **Document tools** — Add descriptions in resource annotations
5. **Use semantic names** — `terraform-1-5`, not `tool1`

## See Also

- [ClusterAgentTool API Reference](../api-reference/clusteragenttool.md)
- [Creating Your First Agent](creating-your-first-agent.md)
