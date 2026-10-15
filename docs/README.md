# Creuzz Documentation

Welcome to the Creuzz documentation. This guide will help you install, configure, and use the Creuzz operator to manage OpenCode v2 agent environments on Kubernetes.

## Table of Contents

### Getting Started
- [Installation](installation.md) — Install Creuzz on your cluster
- [Getting Started](getting-started.md) — Create your first agent in 5 minutes
- [Architecture](architecture.md) — Understand how Creuzz works

### Reference
- [API Reference](api-reference/) — Complete CRD documentation
  - [OpenCodeAgent](api-reference/opencodeagent.md) — Agent instantiation
  - [ClusterAgentTool](api-reference/clusteragenttool.md) — CLI tools
  - [ClusterAgentSkill](api-reference/clusteragentskill.md) — Skills and prompts
  - [ClusterAgentMCPServer](api-reference/clusteragentmcpserver.md) — MCP servers
  - [ClusterAgentEnv](api-reference/clusteragentenv.md) — Environment profiles
  - [ClusterAgentEgressRule](api-reference/clusteragentegressrule.md) — Network rules

### Guides
Task-oriented guides for common operations:
- [Creating Your First Agent](guides/creating-your-first-agent.md)
- [Managing Tools](guides/managing-tools.md)
- [Configuring Skills](guides/configuring-skills.md)
- [Setting Up MCP Servers](guides/mcp-servers.md)
- [Environment Variables](guides/environment-variables.md)
- [Network Policies](guides/network-policies.md)

### Operations
- [Troubleshooting](troubleshooting.md) — Common issues and solutions
- [Development](development.md) — Contributing and building from source
- [Examples](examples/) — Ready-to-use YAML examples

## Quick Links

- **GitHub Repository**: [github.com/airklizz/creuzz](https://github.com/airklizz/creuzz)
- **API Group**: `creuzz.io/v1alpha1`
- **Kubernetes Version**: 1.31+ required

## Concepts

### Two-Layer Architecture

Creuzz uses a **catalog + instantiation** pattern:

1. **Catalog (Cluster-scoped)**: Platform teams define reusable components
   - Tools (OCI images)
   - Skills (markdown files)
   - MCP servers
   - Environment profiles
   - Network egress rules

2. **Instantiation (Namespaced)**: Developers create agents by referencing catalog items

### Resource Hierarchy

```
OpenCodeAgent (namespaced)
├── ConfigMap (skills + opencode.json)
├── NetworkPolicy (zero-trust egress)
├── PersistentVolumeClaim (workspace)
├── Pod
│   ├── InitContainer (git clone)
│   ├── Sidecar: MCP servers
│   └── Main: opencode serve
└── Service (port 4096)
```

## Need Help?

- Check [Troubleshooting](troubleshooting.md) for common issues
- Review [Examples](examples/) for working configurations
- Open an issue on GitHub for bugs or feature requests
