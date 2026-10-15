# Network Policies

This guide covers network security and egress rules for Creuzz agents.

## Overview

Creuzz implements a **zero-trust network model** for agents:

- **Ingress**: Only port 4096 (OpenCode API) is allowed
- **Egress**: Denied by default, explicitly allowed via `ClusterAgentEgressRule`
- **DNS**: Always allowed (required for name resolution)
- **Localhost**: Always allowed (for MCP sidecar communication)

## Default Network Policy

Every agent gets a `NetworkPolicy` like this:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: my-agent-netpol
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: opencode-agent
      app.kubernetes.io/instance: my-agent
  policyTypes:
  - Ingress
  - Egress
  ingress:
  # Only OpenCode API port
  - ports:
    - port: 4096
      protocol: TCP
  egress:
  # DNS always allowed
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
  # Additional rules from ClusterAgentEgressRule references...
```

## Creating Egress Rules

### Allow HTTPS to Specific Domain

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentEgressRule
metadata:
  name: allow-anthropic
spec:
  description: "Anthropic Claude API"
  domains:
    - "api.anthropic.com"
  ports:
    - protocol: TCP
      port: 443
```

### Allow Multiple Domains

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
    - "raw.githubusercontent.com"
    - "codeload.github.com"
  ports:
    - protocol: TCP
      port: 443
    - protocol: TCP
      port: 22    # SSH
    - protocol: TCP
      port: 9418  # Git protocol
```

### Allow Package Registries

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentEgressRule
metadata:
  name: allow-registries
spec:
  description: "Package registries for builds"
  domains:
    - "registry.npmjs.org"
    - "pypi.org"
    - "files.pythonhosted.org"
    - "repo1.maven.org"
    - "plugins.gradle.org"
    - "registry-1.docker.io"
    - "auth.docker.io"
  ports:
    - protocol: TCP
      port: 443
```

### Allow Cloud Provider APIs

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentEgressRule
metadata:
  name: allow-aws
spec:
  description: "AWS services"
  domains:
    - "*.amazonaws.com"
  ports:
    - protocol: TCP
      port: 443
---
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentEgressRule
metadata:
  name: allow-gcp
spec:
  description: "Google Cloud services"
  domains:
    - "*.googleapis.com"
    - "*.gcr.io"
  ports:
    - protocol: TCP
      port: 443
---
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentEgressRule
metadata:
  name: allow-azure
spec:
  description: "Azure services"
  domains:
    - "*.azure.com"
    - "*.azurewebsites.net"
    - "*.blob.core.windows.net"
  ports:
    - protocol: TCP
      port: 443
```

## Using Egress Rules in Agents

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
    - name: "allow-registries"
    - name: "allow-aws"
```

## FQDN vs IP-Based Filtering

### Standard Kubernetes NetworkPolicy

Standard `NetworkPolicy` only supports **IP-based** rules. Domain names in `ClusterAgentEgressRule` are informational and require a CNI that supports FQDN filtering.

### Cilium (Recommended for FQDN)

Cilium supports true FQDN-based filtering via `CiliumNetworkPolicy`:

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
      app.kubernetes.io/instance: my-agent
  egress:
  - toFQDNs:
    - matchName: "api.anthropic.com"
    - matchName: "github.com"
    - matchPattern: "*.githubusercontent.com"
    toPorts:
    - ports:
      - port: "443"
        protocol: TCP
  # DNS
  - toEndpoints:
    - matchLabels:
        k8s-app: kube-dns
    toPorts:
    - ports:
      - port: "53"
        protocol: UDP
      - port: "53"
        protocol: TCP
```

> **Note**: Creuzz currently generates standard `NetworkPolicy`. Cilium support would require operator extension.

## Ingress Configuration

By default, only the OpenCode API port (4096) is exposed internally. For external access, configure an Ingress:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-agent-ingress
  namespace: team-backend
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt"
spec:
  tls:
  - hosts:
    - agent-my-agent.internal.company.com
    secretName: my-agent-tls
  rules:
  - host: agent-my-agent.internal.company.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: my-agent-svc
            port:
              number: 4096
```

### With Authentication

Add authentication via annotations:

```yaml
metadata:
  annotations:
    nginx.ingress.kubernetes.io/auth-type: basic
    nginx.ingress.kubernetes.io/auth-secret: basic-auth
    nginx.ingress.kubernetes.io/auth-realm: "Authentication Required"
```

Or use OAuth2 Proxy:

```yaml
metadata:
  annotations:
    nginx.ingress.kubernetes.io/auth-url: "https://oauth2-proxy.internal/oauth2/auth"
    nginx.ingress.kubernetes.io/auth-signin: "https://oauth2-proxy.internal/oauth2/start?rd=$escaped_request_uri"
```

## Testing Network Policies

### Test Egress (should succeed)

```bash
kubectl exec <agent-pod> -n <namespace> -- curl -v https://api.anthropic.com
```

### Test Egress (should fail)

```bash
kubectl exec <agent-pod> -n <namespace> -- curl -v https://example.com
# Should timeout or connection refused
```

### Test DNS

```bash
kubectl exec <agent-pod> -n <namespace> -- nslookup github.com
# Should resolve
```

### Test Internal Service

```bash
kubectl exec <agent-pod> -n <namespace> -- curl http://localhost:8001/health
# Should connect to MCP server
```

## Monitoring

### Network Policy Metrics (Cilium)

```bash
# View dropped connections
kubectl exec -n kube-system <cilium-pod> -- cilium monitor --type drop

# View policy verdicts
kubectl exec -n kube-system <cilium-pod> -- cilium policy trace \
  --src-k8s-pod team-backend:my-agent-agent \
  --dst-fqdn api.anthropic.com
```

### Audit Logging

Enable network policy audit logging in your CNI.

## Best Practices

1. **Principle of least privilege** — Only allow necessary destinations
2. **Use specific domains** — Avoid overly broad wildcards
3. **Document each rule** — Use the `description` field
4. **Regular audits** — Review and remove unused rules
5. **Test before deploying** — Verify connectivity in staging
6. **Monitor denials** — Alert on unexpected blocked traffic

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Can't reach external API | Add egress rule for the domain |
| DNS resolution fails | Check kube-dns is running |
| MCP server unreachable | Verify localhost is working (it's always allowed) |
| Ingress not working | Check NetworkPolicy allows ingress controller |
| FQDN not working | Use Cilium CNI or resolve to IPs |

## See Also

- [ClusterAgentEgressRule API Reference](../api-reference/clusteragentegressrule.md)
- [Architecture](../architecture.md) — Security model
