# OpenCodeAgent

`OpenCodeAgent` is the namespaced resource that instantiates an OpenCode v2 agent environment.

- **API Group**: `creuzz.io`
- **Version**: `v1alpha1`
- **Kind**: `OpenCodeAgent`
- **Scope**: Namespaced

## Description

An `OpenCodeAgent` declares an agent to be deployed for a specific project. The controller reconciles the agent by generating:

- A **Pod** with init container (git clone), MCP sidecars, and main OpenCode container
- A **Service** exposing the OpenCode API on port 4096
- A **ConfigMap** with skills and MCP configuration
- A **NetworkPolicy** with zero-trust egress rules
- A **PersistentVolumeClaim** for the workspace

## Specification

### `spec.project`

Defines the Git repository to clone into the agent workspace.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `repoUrl` | string | Yes | Git repository URL (SSH or HTTPS) |
| `branch` | string | Yes | Git branch or ref to checkout |
| `secretRef` | string | Yes | Name of Secret containing `id_ed25519` and `known_hosts` |

**Example:**
```yaml
project:
  repoUrl: "git@github.com:orga/billing-service.git"
  branch: "refactor/v2"
  secretRef: "github-ssh-creds"
```

### `spec.runtime`

Defines compute resources and container image.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `image` | string | Yes | OpenCode v2 base container image |
| `cpu` | string | Yes | CPU request/limit (e.g., `"2"`, `"500m"`) |
| `memory` | string | Yes | Memory request/limit (e.g., `"4Gi"`, `"512Mi"`) |
| `storage` | string | Yes | Workspace PVC size (e.g., `"10Gi"`) |

**Example:**
```yaml
runtime:
  image: "registry.internal/opencode-base:v2.0.0"
  cpu: "2"
  memory: "4Gi"
  storage: "10Gi"
```

### `spec.clusterToolRefs`

List of `ClusterAgentTool` names to mount as OCI image volumes.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Name of the `ClusterAgentTool` |

**Example:**
```yaml
clusterToolRefs:
  - name: "terraform-1-5"
  - name: "node-20"
```

### `spec.clusterSkillRefs`

List of `ClusterAgentSkill` names to inject as configuration files.

**Example:**
```yaml
clusterSkillRefs:
  - name: "clean-architecture"
  - name: "security-best-practices"
```

### `spec.clusterMcpRefs`

List of `ClusterAgentMCPServer` names to run as sidecar containers.

**Example:**
```yaml
clusterMcpRefs:
  - name: "postgres-staging"
  - name: "github-mcp"
```

### `spec.clusterEnvRefs`

List of `ClusterAgentEnv` names to inject environment variables.

**Example:**
```yaml
clusterEnvRefs:
  - name: "aws-dev-profile"
  - name: "llm-config"
```

### `spec.clusterEgressRefs`

List of `ClusterAgentEgressRule` names to allow network egress.

**Example:**
```yaml
clusterEgressRefs:
  - name: "allow-anthropic"
  - name: "allow-github"
```

### `spec.ttlAfterIdle`

Optional duration of inactivity after which the agent scales to zero.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `ttlAfterIdle` | string | No | Duration (e.g., `"2h"`, `"30m"`) |

**Example:**
```yaml
ttlAfterIdle: "2h"
```

## Status

| Field | Type | Description |
|-------|------|-------------|
| `phase` | string | `Pending`, `Ready`, `Failed`, or `Suspended` |
| `url` | string | HTTPS URL to access the OpenCode API |
| `podName` | string | Name of the generated Pod |
| `serviceName` | string | Name of the generated Service |
| `pairingTokenSecret` | string | Name of Secret containing pairing token |
| `conditions` | []Condition | Standard Kubernetes conditions |

**Example status:**
```yaml
status:
  phase: Ready
  url: https://agent-team-backend-billing-migration.internal.local
  podName: billing-migration-agent
  serviceName: billing-migration-svc
  conditions:
  - type: Ready
    status: "True"
    reason: Reconciliation
    message: Agent is ready
```

## Complete Example

```yaml
apiVersion: creuzz.io/v1alpha1
kind: OpenCodeAgent
metadata:
  name: billing-migration
  namespace: team-backend
spec:
  project:
    repoUrl: "git@github.com:orga/billing-service.git"
    branch: "refactor/v2"
    secretRef: "github-ssh-creds"
  runtime:
    image: "registry.internal/opencode-base:v2.0.0"
    cpu: "2"
    memory: "4Gi"
    storage: "10Gi"
  clusterToolRefs:
    - name: "terraform-1-5"
    - name: "node-20"
  clusterSkillRefs:
    - name: "clean-architecture"
  clusterMcpRefs:
    - name: "postgres-staging"
  clusterEnvRefs:
    - name: "aws-dev-profile"
  clusterEgressRefs:
    - name: "allow-anthropic"
    - name: "allow-github"
  ttlAfterIdle: "2h"
```

## Generated Resources

| Resource | Name Pattern | Description |
|----------|--------------|-------------|
| ConfigMap | `<name>-config` | Skills + `opencode.json` |
| NetworkPolicy | `<name>-netpol` | Zero-trust network rules |
| PVC | `<name>-workspace` | Workspace volume |
| Pod | `<name>-agent` | Agent runtime |
| Service | `<name>-svc` | ClusterIP on port 4096 |

## See Also

- [ClusterAgentTool](clusteragenttool.md) — Define tools
- [ClusterAgentSkill](clusteragentskill.md) — Define skills
- [ClusterAgentMCPServer](clusteragentmcpserver.md) — Define MCP servers
- [ClusterAgentEnv](clusteragentenv.md) — Define environment profiles
- [ClusterAgentEgressRule](clusteragentegressrule.md) — Define egress rules
