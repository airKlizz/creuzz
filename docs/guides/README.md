# Guides

Task-oriented guides for common Creuzz operations.

## Available Guides

| Guide | Description | Level |
|-------|-------------|-------|
| [Creating Your First Agent](creating-your-first-agent.md) | Complete walkthrough from scratch | Beginner |
| [Managing Tools](managing-tools.md) | Create, update, and version CLI tools | Beginner |
| [Configuring Skills](configuring-skills.md) | Inject context and guidelines | Beginner |
| [Setting Up MCP Servers](mcp-servers.md) | Configure Model Context Protocol servers | Intermediate |
| [Environment Variables](environment-variables.md) | Manage env vars and secrets | Intermediate |
| [Network Policies](network-policies.md) | Configure egress and ingress rules | Advanced |

## Quick Reference

### Common Commands

```bash
# List all agents
kubectl get opencodeagents --all-namespaces

# List catalog resources
kubectl get clusteragenttools,clusteragentskills,clusteragentmcpservers,clusteragentenvs,clusteragentegressrules

# Describe an agent
kubectl describe opencodeagent <name> -n <namespace>

# View agent logs
kubectl logs <agent-pod> -n <namespace> -c opencode

# View MCP server logs
kubectl logs <agent-pod> -n <namespace> -c mcp-<server-name>

# Check network policy
kubectl get networkpolicy <agent-name>-netpol -n <namespace> -o yaml
```

### Common Workflows

1. **Add a new tool**: Create `ClusterAgentTool` → Reference in agent
2. **Update a skill**: Patch `ClusterAgentSkill` → Auto-propagates to agents
3. **Add MCP server**: Create `ClusterAgentMCPServer` + Secret → Reference in agent
4. **Allow new API**: Create `ClusterAgentEgressRule` → Reference in agent
5. **Update environment**: Patch `ClusterAgentEnv` → Auto-propagates to agents

## Need More Help?

- [API Reference](../api-reference/) — Complete CRD documentation
- [Troubleshooting](../troubleshooting.md) — Common issues
- [Examples](../examples/) — Ready-to-use YAML
