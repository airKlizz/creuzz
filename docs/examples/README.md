# Examples

Ready-to-use YAML examples for common Creuzz scenarios.

## Directory Structure

```
examples/
├── README.md                          # This file
├── basic-agent.yaml                   # Minimal agent
├── full-stack-agent.yaml              # Complete agent with all features
├── catalog/                           # Catalog resources
│   ├── tools.yaml
│   ├── skills.yaml
│   ├── mcp-servers.yaml
│   ├── environments.yaml
│   └── egress-rules.yaml
└── namespaces/
    └── team-backend.yaml              # Namespace with secrets
```

## Quick Start

Apply all examples in order:

```bash
# 1. Create namespace
kubectl apply -f namespaces/team-backend.yaml

# 2. Create catalog resources
kubectl apply -f catalog/

# 3. Create secrets (edit first!)
# kubectl create secret generic github-ssh-creds -n team-backend ...
# kubectl create secret generic aws-dev-creds -n team-backend ...

# 4. Create an agent
kubectl apply -f basic-agent.yaml
# or
kubectl apply -f full-stack-agent.yaml
```

## Examples

| File | Description | Use Case |
|------|-------------|----------|
| [basic-agent.yaml](basic-agent.yaml) | Minimal agent with one skill | Getting started |
| [full-stack-agent.yaml](full-stack-agent.yaml) | Complete agent with tools, skills, MCP, env, egress | Production-like setup |
| [catalog/tools.yaml](catalog/tools.yaml) | Common CLI tools | Platform setup |
| [catalog/skills.yaml](catalog/skills.yaml) | Architecture and security skills | Platform setup |
| [catalog/mcp-servers.yaml](catalog/mcp-servers.yaml) | Database and API MCP servers | Platform setup |
| [catalog/environments.yaml](catalog/environments.yaml) | AWS, LLM, database profiles | Platform setup |
| [catalog/egress-rules.yaml](catalog/egress-rules.yaml) | Common egress rules | Platform setup |
| [namespaces/team-backend.yaml](namespaces/team-backend.yaml) | Namespace template | Team setup |

## Customizing Examples

1. **Copy** the example files to your own directory
2. **Edit** image references to match your registry
3. **Update** repository URLs and branches
4. **Create** required secrets in your namespace
5. **Apply** with `kubectl apply -f`

## Prerequisites

Before applying examples:

1. **Creuzz installed**: See [Installation](../installation.md)
2. **Secrets created**: Git SSH, AWS, database credentials
3. **Images available**: Tools and MCP server images accessible

## See Also

- [Getting Started](../getting-started.md)
- [API Reference](../api-reference/)
- [Guides](../guides/)
