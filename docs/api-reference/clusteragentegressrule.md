# ClusterAgentEgressRule

`ClusterAgentEgressRule` defines allowed network egress (outbound traffic) for agent pods.

- **API Group**: `creuzz.io`
- **Version**: `v1alpha1`
- **Kind**: `ClusterAgentEgressRule`
- **Scope**: Cluster

## Description

Egress rules define which external domains and ports an agent can access. By default, all egress is denied except DNS resolution. Egress rules explicitly allow specific destinations.

> **Note**: For FQDN-based filtering (domain names), a CNI that supports it is required (e.g., Cilium with `CiliumNetworkPolicy`). Standard Kubernetes `NetworkPolicy` only supports IP-based rules.

## Specification

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `description` | string | No | Human-readable description |
| `domains` | []string | No | List of allowed FQDNs (requires Cilium) |
| `ports` | []EgressPort | No | List of allowed ports |

### `ports[]` (EgressPort)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `protocol` | string | No | `TCP`, `UDP`, or `SCTP` (default: `TCP`) |
| `port` | integer | Yes | Port number (1-65535) |

## Examples

### Allow Anthropic API

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentEgressRule
metadata:
  name: allow-anthropic
spec:
  description: "API Claude Anthropic"
  domains:
    - "api.anthropic.com"
  ports:
    - protocol: TCP
      port: 443
```

### Allow GitHub (API + Git)

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentEgressRule
metadata:
  name: allow-github
spec:
  description: "GitHub API and git operations"
  domains:
    - "github.com"
    - "api.github.com"
    - "*.githubusercontent.com"
  ports:
    - protocol: TCP
      port: 443
    - protocol: TCP
      port: 22
    - protocol: TCP
      port: 9418
```

### Allow Package Registries

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentEgressRule
metadata:
  name: allow-package-registries
spec:
  description: "npm, PyPI, Maven Central"
  domains:
    - "registry.npmjs.org"
    - "pypi.org"
    - "files.pythonhosted.org"
    - "repo1.maven.org"
    - "plugins.gradle.org"
  ports:
    - protocol: TCP
      port: 443
```

### Allow AWS Services

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentEgressRule
metadata:
  name: allow-aws
spec:
  description: "AWS S3 and DynamoDB"
  domains:
    - "*.amazonaws.com"
  ports:
    - protocol: TCP
      port: 443
```

### Allow Container Registries

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentEgressRule
metadata:
  name: allow-docker-hub
spec:
  description: "Docker Hub and registries"
  domains:
    - "registry-1.docker.io"
    - "auth.docker.io"
    - "production.cloudflare.docker.com"
    - "registry.internal"
  ports:
    - protocol: TCP
      port: 443
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
  clusterEgressRefs:
    - name: "allow-anthropic"
    - name: "allow-github"
    - name: "allow-package-registries"
```

## Generated NetworkPolicy

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: my-agent-netpol
  namespace: team-backend
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: opencode-agent
      app.kubernetes.io/instance: my-agent
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - ports:
    - port: 4096
      protocol: TCP
  egress:
  # DNS (always allowed)
  - ports:
    - port: 53
      protocol: UDP
    - port: 53
      protocol: TCP
    to:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: kube-system
      podSelector:
        matchLabels:
          k8s-app: kube-dns
  # From allow-anthropic rule
  - ports:
    - port: 443
      protocol: TCP
  # From allow-github rule
  - ports:
    - port: 443
      protocol: TCP
    - port: 22
      protocol: TCP
    - port: 9418
      protocol: TCP
```

## Default Egress (Always Allowed)

These are always allowed regardless of egress rules:

| Destination | Ports | Purpose |
|-------------|-------|---------|
| kube-dns (kube-system) | 53 UDP/TCP | DNS resolution |
| localhost | all | Intra-pod communication (MCP sidecars) |

## CiliumNetworkPolicy (FQDN Filtering)

For true FQDN-based filtering, use Cilium:

```yaml
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: my-agent-egress
  namespace: team-backend
spec:
  endpointSelector:
    matchLabels:
      app.kubernetes.io/name: opencode-agent
  egress:
  - toFQDNs:
    - matchName: "api.anthropic.com"
    - matchName: "github.com"
    toPorts:
    - ports:
      - port: "443"
        protocol: TCP
```

> **Note**: The Creuzz operator generates standard `NetworkPolicy`. For Cilium support, you would need to extend the operator or use a CNI that supports FQDN matching.

## Best Practices

1. **Principle of least privilege** — Only allow necessary destinations
2. **Use specific domains** — Avoid wildcards (`*.com`) when possible
3. **Document each rule** — Use the `description` field
4. **Group by purpose** — One rule per service/category
5. **Regular audits** — Review and remove unused rules

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Agent can't reach API | Add egress rule for the API domain |
| DNS resolution fails | Check kube-dns is running in kube-system |
| FQDN not working | Use Cilium CNI or resolve to IPs |
| Port blocked | Verify port is in the rule's `ports` list |

## See Also

- [OpenCodeAgent](opencodeagent.md) — Use egress rules in agents
- [Network Policies](../guides/network-policies.md) — Detailed guide
