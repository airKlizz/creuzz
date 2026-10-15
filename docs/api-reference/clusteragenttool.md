# ClusterAgentTool

`ClusterAgentTool` defines a CLI tool as a lightweight OCI image, mounted as a read-only volume in agent pods.

- **API Group**: `creuzz.io`
- **Version**: `v1alpha1`
- **Kind**: `ClusterAgentTool`
- **Scope**: Cluster

## Description

Tools are OCI images that contain CLI utilities (terraform, node, python, etc.). When referenced by an `OpenCodeAgent`, the tool image is mounted as a Kubernetes `image` volume (requires K8s 1.31+ or containerd 2.x).

The `binPath` is automatically prepended to the agent's `PATH` environment variable.

## Specification

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `image` | string | Yes | OCI image reference (e.g., `"hashicorp/terraform:1.5.7"`) |
| `mountPath` | string | Yes | Path where the image is mounted (e.g., `"/opt/tools/terraform"`) |
| `binPath` | string | No | Directory to add to PATH (e.g., `"/opt/tools/terraform/bin"`) |

## Examples

### Terraform

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

### Node.js

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentTool
metadata:
  name: node-20
spec:
  image: "node:20-alpine"
  mountPath: "/opt/tools/node"
  binPath: "/opt/tools/node/bin"
```

### Python

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentTool
metadata:
  name: python-3-12
spec:
  image: "python:3.12-slim"
  mountPath: "/opt/tools/python"
  binPath: "/usr/local/bin"
```

### Custom Internal Tool

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentTool
metadata:
  name: internal-cli
spec:
  image: "registry.internal/tools/cli:v2.1.0"
  mountPath: "/opt/tools/cli"
  binPath: "/opt/tools/cli/bin"
```

## Usage in OpenCodeAgent

```yaml
apiVersion: creuzz.io/v1alpha1
kind: OpenCodeAgent
metadata:
  name: my-agent
  namespace: team-backend
spec:
  # ...
  clusterToolRefs:
    - name: "terraform-1-5"
    - name: "node-20"
```

## Resulting Pod Configuration

```yaml
volumes:
- name: tool-terraform-1-5
  image:
    reference: hashicorp/terraform:1.5.7
    pullPolicy: IfNotPresent
- name: tool-node-20
  image:
    reference: node:20-alpine
    pullPolicy: IfNotPresent

containers:
- name: opencode
  env:
  - name: PATH
    value: "/opt/tools/terraform/bin:/opt/tools/node/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
  volumeMounts:
  - name: tool-terraform-1-5
    mountPath: /opt/tools/terraform
    readOnly: true
  - name: tool-node-20
    mountPath: /opt/tools/node
    readOnly: true
```

## Best Practices

1. **Use specific tags** — Avoid `latest` tag for reproducibility
2. **Keep images small** — Use Alpine or Distroless bases
3. **Set binPath correctly** — Ensure it matches the actual binary location in the image
4. **Version your tools** — Use semantic versioning in the resource name (e.g., `terraform-1-5`)

## Troubleshooting

| Issue | Solution |
|-------|----------|
| `image volume` not supported | Upgrade to K8s 1.31+ or containerd 2.x |
| Tool not found in PATH | Verify `binPath` matches the image's binary directory |
| Image pull fails | Check image reference and pull secrets |

## See Also

- [OpenCodeAgent](opencodeagent.md) — Use tools in agents
- [Managing Tools](../guides/managing-tools.md) — Guide for tool management
