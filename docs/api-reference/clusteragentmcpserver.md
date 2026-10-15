# ClusterAgentMCPServer

`ClusterAgentMCPServer` defines a Model Context Protocol (MCP) server to run as a sidecar container alongside OpenCode.

- **API Group**: `creuzz.io`
- **Version**: `v1alpha1`
- **Kind**: `ClusterAgentMCPServer`
- **Scope**: Cluster

## Description

MCP servers provide additional context and tools to the OpenCode agent (databases, APIs, file systems, etc.). Each `ClusterAgentMCPServer` runs as a sidecar container in the agent Pod, sharing the network namespace (localhost access).

The agent's `opencode.json` is automatically configured to connect to the MCP server via SSE (Server-Sent Events).

## Specification

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `image` | string | Yes | OCI image reference for the MCP server |
| `port` | integer | Yes | Port the MCP server listens on (1-65535) |
| `args` | []string | No | Command-line arguments for the container |
| `envFromSecret` | string | No | Name of Secret in target namespace to inject as env vars |

## Examples

### PostgreSQL MCP Server

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentMCPServer
metadata:
  name: postgres-staging
spec:
  image: "mcp/postgres:1.2.0"
  port: 8001
  args: ["--port", "8001"]
  envFromSecret: "staging-db-creds"
```

Required secret (`staging-db-creds`):
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: staging-db-creds
  namespace: team-backend
type: Opaque
stringData:
  POSTGRES_HOST: "postgres.staging.internal"
  POSTGRES_PORT: "5432"
  POSTGRES_USER: "agent"
  POSTGRES_PASSWORD: "secret"
  POSTGRES_DATABASE: "billing"
```

### GitHub MCP Server

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentMCPServer
metadata:
  name: github-mcp
spec:
  image: "mcp/github:latest"
  port: 8002
  args: ["--port", "8002", "--read-only"]
  envFromSecret: "github-token"
```

### Filesystem MCP Server

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentMCPServer
metadata:
  name: filesystem
spec:
  image: "mcp/filesystem:1.0.0"
  port: 8003
  args: ["--port", "8003", "--root", "/workspace"]
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
  clusterMcpRefs:
    - name: "postgres-staging"
    - name: "github-mcp"
```

## Generated Configuration

### opencode.json

```json
{
  "mcp": {
    "servers": {
      "postgres-staging": {
        "type": "sse",
        "url": "http://localhost:8001"
      },
      "github-mcp": {
        "type": "sse",
        "url": "http://localhost:8002"
      }
    }
  }
}
```

### Sidecar Container

```yaml
containers:
- name: mcp-postgres-staging
  image: mcp/postgres:1.2.0
  args: ["--port", "8001"]
  ports:
  - containerPort: 8001
    protocol: TCP
  envFrom:
  - secretRef:
      name: staging-db-creds
```

## Network Architecture

```
┌─────────────────────────────────────────┐
│           Agent Pod                     │
│  ┌─────────────┐    ┌─────────────┐    │
│  │   OpenCode  │◄──►│ MCP Sidecar │    │
│  │   (main)    │    │  (postgres) │    │
│  │  Port 4096  │    │  Port 8001  │    │
│  └─────────────┘    └──────┬──────┘    │
│                            │            │
│                     localhost:SSE       │
│                            │            │
│                     ┌──────▼──────┐    │
│                     │  PostgreSQL │    │
│                     │   (external)│    │
│                     └─────────────┘    │
│                                         │
│  ┌─────────────┐    ┌─────────────┐    │
│  │ MCP Sidecar │    │ MCP Sidecar │    │
│  │  (github)   │    │(filesystem) │    │
│  │  Port 8002  │    │  Port 8003  │    │
│  └─────────────┘    └─────────────┘    │
└─────────────────────────────────────────┘
```

## Best Practices

1. **Use specific image tags** — Avoid `latest`
2. **Set resource limits** — MCP servers can be memory-intensive
3. **Use secrets for credentials** — Never hardcode in the spec
4. **One server per port** — Each MCP server needs a unique port
5. **Health checks** — Ensure the MCP server image supports health checks

## Troubleshooting

| Issue | Solution |
|-------|----------|
| MCP server not connecting | Verify port matches in spec and server config |
| Secret not found | Create the secret in the agent's namespace |
| Sidecar crashes | Check logs: `kubectl logs <pod> -c mcp-<name>` |
| Port conflict | Use unique ports for each MCP server |

## See Also

- [OpenCodeAgent](opencodeagent.md) — Use MCP servers in agents
- [Setting Up MCP Servers](../guides/mcp-servers.md) — Detailed guide
