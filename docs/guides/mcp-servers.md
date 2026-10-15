# Setting Up MCP Servers

This guide covers how to configure Model Context Protocol (MCP) servers for your agents.

## Overview

MCP servers provide additional context and tools to OpenCode agents, such as:
- **Databases**: PostgreSQL, MySQL, MongoDB
- **APIs**: GitHub, GitLab, Jira, Slack
- **File Systems**: Local or cloud storage
- **Search**: Elasticsearch, Algolia
- **Custom**: Your internal services

## Architecture

```
┌─────────────────────────────────────────┐
│           Agent Pod                     │
│                                         │
│  ┌─────────────┐    ┌─────────────┐    │
│  │   OpenCode  │◄──►│ MCP Server  │    │
│  │   (main)    │    │  (sidecar)  │    │
│  │             │    │             │    │
│  │  Reads:     │    │  Connects:  │    │
│  │  opencode   │    │  • Database │    │
│  │  .json      │    │  • API      │    │
│  │             │    │  • Files    │    │
│  └─────────────┘    └─────────────┘    │
│         │                  │            │
│         └──────────────────┘            │
│              localhost:SSE              │
└─────────────────────────────────────────┘
```

## Creating an MCP Server

### PostgreSQL

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

Required secret:
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
  POSTGRES_SSLMODE: "require"
```

### MySQL

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentMCPServer
metadata:
  name: mysql-prod
spec:
  image: "mcp/mysql:1.0.0"
  port: 8002
  args: ["--port", "8002"]
  envFromSecret: "mysql-creds"
```

### GitHub

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentMCPServer
metadata:
  name: github-mcp
spec:
  image: "mcp/github:latest"
  port: 8003
  args: ["--port", "8003", "--read-only"]
  envFromSecret: "github-token"
```

Secret:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: github-token
  namespace: team-backend
type: Opaque
stringData:
  GITHUB_TOKEN: "ghp_xxxxxxxxxxxx"
```

### Filesystem (Workspace)

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentMCPServer
metadata:
  name: filesystem
spec:
  image: "mcp/filesystem:1.0.0"
  port: 8004
  args: ["--port", "8004", "--root", "/workspace"]
```

### Custom Internal API

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentMCPServer
metadata:
  name: internal-api
spec:
  image: "registry.internal/mcp/internal-api:2.0.0"
  port: 8005
  args: ["--port", "8005", "--base-url", "https://api.internal"]
  envFromSecret: "internal-api-creds"
```

## Multiple MCP Servers

Agents can use multiple MCP servers:

```yaml
apiVersion: creuzz.io/v1alpha1
kind: OpenCodeAgent
metadata:
  name: full-stack-agent
  namespace: team-backend
spec:
  # ...
  clusterMcpRefs:
    - name: "postgres-staging"
    - name: "github-mcp"
    - name: "filesystem"
    - name: "internal-api"
```

Each server gets a unique port. The `opencode.json` is generated with all servers:

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
        "url": "http://localhost:8003"
      },
      "filesystem": {
        "type": "sse",
        "url": "http://localhost:8004"
      },
      "internal-api": {
        "type": "sse",
        "url": "http://localhost:8005"
      }
    }
  }
}
```

## Port Allocation

Each MCP server needs a unique port. Plan your port allocation:

| Port Range | Purpose |
|------------|---------|
| 8001-8010 | Database servers |
| 8011-8020 | API integrations |
| 8021-8030 | File/Search servers |
| 8031-8040 | Custom servers |

## Resource Limits

MCP servers can be resource-intensive. Consider adding resource limits by extending the operator or using LimitRanges in the namespace.

## Health Checks

Ensure your MCP server images support health checks:

```yaml
# In the MCP server image, expose a health endpoint
# GET /health should return 200 OK
```

## Debugging

### Check MCP Server Logs

```bash
# List containers in the pod
kubectl get pod <agent-pod> -n <namespace> -o jsonpath='{.spec.containers[*].name}'

# View MCP server logs
kubectl logs <agent-pod> -c mcp-postgres-staging -n <namespace>
```

### Test MCP Connection

```bash
# Exec into the main container
kubectl exec -it <agent-pod> -n <namespace> -- sh

# Test SSE connection to MCP server
curl -N http://localhost:8001/sse
```

### Verify opencode.json

```bash
kubectl get configmap <agent-name>-config -n <namespace> -o jsonpath='{.data.opencode\.json}' | jq
```

## Security

### Network Isolation

MCP servers are only accessible from within the Pod (localhost). They are not exposed externally.

### Secret Handling

- Credentials are injected via `envFromSecret`
- Secrets are never logged or exposed in pod specs
- Use namespace-scoped secrets

### Read-Only Mode

Where possible, configure MCP servers in read-only mode:

```yaml
args: ["--port", "8001", "--read-only"]
```

## Best Practices

1. **One server per purpose** — Don't combine database and API in one server
2. **Use specific image tags** — Avoid `latest`
3. **Set resource limits** — Prevent runaway containers
4. **Enable read-only mode** — When write access isn't needed
5. **Monitor logs** — Set up log aggregation for MCP servers
6. **Test connectivity** — Verify servers work before deploying agents

## Troubleshooting

| Issue | Solution |
|-------|----------|
| MCP server not starting | Check logs: `kubectl logs <pod> -c mcp-<name>` |
| Connection refused | Verify port matches in spec and server config |
| Secret not found | Create secret in agent's namespace |
| Agent can't connect | Check `opencode.json` in ConfigMap |
| Port conflict | Use unique ports for each server |
| High memory usage | Add resource limits or use smaller images |

## See Also

- [ClusterAgentMCPServer API Reference](../api-reference/clusteragentmcpserver.md)
- [Environment Variables](environment-variables.md)
- [Network Policies](network-policies.md)
