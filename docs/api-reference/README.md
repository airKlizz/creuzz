# API Reference

Complete reference documentation for all Creuzz Custom Resource Definitions (CRDs).

## Resources

### Namespaced

| Resource | Kind | Description |
|----------|------|-------------|
| [OpenCodeAgent](opencodeagent.md) | `OpenCodeAgent` | Agent instantiation resource |

### Cluster-Scoped (Catalog)

| Resource | Kind | Description |
|----------|------|-------------|
| [ClusterAgentTool](clusteragenttool.md) | `ClusterAgentTool` | OCI image CLI tool |
| [ClusterAgentSkill](clusteragentskill.md) | `ClusterAgentSkill` | Markdown skill injection |
| [ClusterAgentMCPServer](clusteragentmcpserver.md) | `ClusterAgentMCPServer` | MCP server sidecar |
| [ClusterAgentEnv](clusteragentenv.md) | `ClusterAgentEnv` | Environment variables and secrets |
| [ClusterAgentEgressRule](clusteragentegressrule.md) | `ClusterAgentEgressRule` | Network egress rules |

## API Group

All resources use API group `creuzz.io`:

```yaml
apiVersion: creuzz.io/v1alpha1
kind: <Kind>
```

## Common Patterns

### Referencing Cluster Resources

Namespaced resources reference cluster-scoped resources by name:

```yaml
spec:
  clusterToolRefs:
    - name: "terraform-1-5"
  clusterSkillRefs:
    - name: "clean-architecture"
```

### Status Conditions

All resources support standard Kubernetes conditions:

```yaml
status:
  conditions:
  - type: Ready
    status: "True"
    reason: Reconciliation
    message: "Resource is ready"
    lastTransitionTime: "2026-10-14T12:00:00Z"
```

## Validation

All CRDs include OpenAPI v3 schema validation. Invalid configurations are rejected by the API server:

```bash
# Example: missing required field
kubectl apply -f invalid-agent.yaml
# Error: spec.project.repoUrl: Required value
```

## Print Columns

Useful columns are configured for `kubectl get`:

```bash
kubectl get clusteragenttools
# NAME            IMAGE                          MOUNTPATH
# terraform-1-5   hashicorp/terraform:1.5.7      /opt/tools/terraform

kubectl get opencodeagents -n team-backend
# NAME               PHASE   URL                                          AGE
# billing-migration  Ready   https://agent-team-backend-billing...        5m
```
