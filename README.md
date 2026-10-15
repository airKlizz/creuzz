# Creuzz

> **Kubernetes Operator for Isolated OpenCode v2 Agent Environments**

Creuzz is a Kubernetes operator that provisions declarative, isolated, and ready-to-use [OpenCode v2](https://opencode.ai) agent environments. It provides a two-layer architecture: a **cluster-scoped catalog** managed by platform teams (CLI tools, network rules, skills, MCP servers, environment profiles) and a **namespaced instantiation resource** declared by developers.

## Features

- **Declarative Agent Provisioning** — Define agents as Kubernetes resources
- **OCI Image Volumes** — Mount CLI tools as read-only OCI image volumes (K8s 1.31+)
- **Skills Injection** — Inject context, system prompts, and architecture guidelines
- **MCP Sidecars** — Run Model Context Protocol servers alongside OpenCode
- **Strict Network Policies** — Zero-trust egress with fine-grained domain/port rules
- **Git Bootstrap** — Automatic repository cloning via init container
- **Dynamic PATH** — Automatic PATH construction from mounted tools
- **Reactive Watches** — Automatic rollout when cluster resources change

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Cluster-Scoped Catalog                     │
│  (Managed by Platform Team)                                   │
│                                                               │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐ │
│  │ClusterAgent │ │ClusterAgent │ │  ClusterAgentMCPServer  │ │
│  │    Tool     │ │    Skill    │ │                         │ │
│  └─────────────┘ └─────────────┘ └─────────────────────────┘ │
│  ┌─────────────┐ ┌─────────────────────────────────────────┐ │
│  │ClusterAgent │ │      ClusterAgentEgressRule             │ │
│  │    Env      │ │                                         │ │
│  └─────────────┘ └─────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼ references
┌─────────────────────────────────────────────────────────────┐
│                  Namespaced Instantiation                     │
│  (Declared by Developers)                                     │
│                                                               │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                    OpenCodeAgent                        │ │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────────┐ │ │
│  │  │   Pod   │  │ Service │  │ConfigMap│  │NetworkPolicy│ │ │
│  │  │         │  │         │  │         │  │             │ │ │
│  │  │ • Init  │  │ Port    │  │ • Skills│  │ • Deny All  │ │ │
│  │  │   Git   │  │ 4096    │  │ • MCP   │  │ • DNS       │ │ │
│  │  │ • MCP   │  │         │  │   Config│  │ • Egress    │ │ │
│  │  │   Side  │  │         │  │         │  │   Rules     │ │ │
│  │  │ • Open  │  │         │  │         │  │             │ │ │
│  │  │   Code  │  │         │  │         │  │             │ │ │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────────┘ │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Kubernetes 1.31+ (for OCI image volumes)
- `kubectl` configured with cluster access
- `docker` (for building the operator image)

### Install

```bash
# Install CRDs
make install

# Deploy the operator
make deploy IMG=creuzz-operator:latest
```

### Create Your First Agent

```yaml
apiVersion: creuzz.io/v1alpha1
kind: OpenCodeAgent
metadata:
  name: my-first-agent
  namespace: team-backend
spec:
  project:
    repoUrl: "https://github.com/myorg/myrepo.git"
    branch: "main"
    secretRef: "github-ssh-creds"
  runtime:
    image: "registry.internal/opencode-base:v2.0.0"
    cpu: "2"
    memory: "4Gi"
    storage: "10Gi"
  clusterSkillRefs:
    - name: "clean-architecture"
  clusterToolRefs:
    - name: "terraform-1-5"
    - name: "node-20"
```

```bash
kubectl apply -f my-agent.yaml
kubectl get opencodeagent my-first-agent -n team-backend
```

## Documentation

| Document | Description |
|----------|-------------|
| [Installation](docs/installation.md) | Detailed installation instructions |
| [Getting Started](docs/getting-started.md) | Create your first agent step-by-step |
| [Architecture](docs/architecture.md) | Deep dive into the operator design |
| [API Reference](docs/api-reference/) | Complete CRD specifications |
| [Guides](docs/guides/) | Task-oriented guides |
| [Troubleshooting](docs/troubleshooting.md) | Common issues and solutions |
| [Development](docs/development.md) | Contributing and building from source |

## API Overview

| Resource | Scope | Description |
|----------|-------|-------------|
| `ClusterAgentTool` | Cluster | OCI image tool mounted as volume |
| `ClusterAgentSkill` | Cluster | Markdown skill file injection |
| `ClusterAgentMCPServer` | Cluster | MCP server sidecar definition |
| `ClusterAgentEnv` | Cluster | Environment variables and secrets |
| `ClusterAgentEgressRule` | Cluster | Network egress rules |
| `OpenCodeAgent` | Namespaced | Agent instantiation resource |

All resources use API group `creuzz.io/v1alpha1`.

## Examples

See the [`config/samples/`](config/samples/) directory for complete examples, or the [`docs/examples/`](docs/examples/) directory for curated scenarios.

## Requirements

- **Kubernetes**: 1.31+ (for `image` volume support) or containerd 2.x
- **Container Runtime**: containerd with image volume support
- **Network Plugin**: Any CNI supporting NetworkPolicy (Calico, Cilium, etc.)

## Contributing

Contributions are welcome! Please see [Development Guide](docs/development.md) for instructions on building, testing, and submitting changes.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

---

**Creuzz** — *Declarative AI Agent Environments on Kubernetes*
