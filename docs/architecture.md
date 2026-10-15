# Architecture

This document describes the architecture of the Creuzz operator, its components, and how they interact.

## Overview

Creuzz follows the **Kubernetes Operator pattern** with a **two-layer architecture**:

1. **Catalog Layer (Cluster-scoped)** — Platform team manages reusable building blocks
2. **Instantiation Layer (Namespaced)** — Developers assemble agents from catalog items

## Design Principles

| Principle | Implementation |
|-----------|---------------|
| **Declarative** | Agents defined as Kubernetes CRDs |
| **Isolated** | Each agent runs in its own Pod with strict NetworkPolicies |
| **Composable** | Agents assembled from reusable cluster-scoped components |
| **Reactive** | Watches on cluster resources trigger automatic updates |
| **Secure by Default** | Zero-trust networking, read-only tool mounts, secret injection |

## Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        Kubernetes Cluster                        │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                   Creuzz Operator (Deployment)              │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │ │
│  │  │  OpenCode   │  │   Cluster   │  │     Cluster         │  │ │
│  │  │   Agent     │  │   Agent     │  │     Agent           │  │ │
│  │  │ Controller  │  │   Tool      │  │     Skill           │  │ │
│  │  │             │  │ Controller  │  │     Controller      │  │ │
│  │  │ • Reconcile │  │             │  │                     │  │ │
│  │  │ • Watches   │  │ • No-op     │  │ • No-op             │  │ │
│  │  │ • Owns Pods │  │             │  │                     │  │ │
│  │  └─────────────┘  └─────────────┘  └─────────────────────┘  │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │ │
│  │  │   Cluster   │  │   Cluster   │  │     Cluster         │  │ │
│  │  │   Agent     │  │   Agent     │  │     Agent           │  │ │
│  │  │  MCPServer  │  │    Env      │  │     EgressRule      │  │ │
│  │  │ Controller  │  │ Controller  │  │     Controller      │  │ │
│  │  └─────────────┘  └─────────────┘  └─────────────────────┘  │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                              │                                   │
│                              ▼                                   │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                    Cluster-Scoped Catalog                    │ │
│  │  ClusterAgentTool  │  ClusterAgentSkill  │  ClusterAgent... │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                              │                                   │
│                              ▼ references                        │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │              Namespace: team-backend                        │ │
│  │  ┌─────────────────────────────────────────────────────────┐│ │
│  │  │  OpenCodeAgent: billing-migration                       ││ │
│  │  │  spec:                                                  ││ │
│  │  │    project: {repoUrl, branch, secretRef}                ││ │
│  │  │    runtime: {image, cpu, memory, storage}               ││ │
│  │  │    clusterToolRefs: [terraform-1-5, node-20]            ││ │
│  │  │    clusterSkillRefs: [clean-architecture]               ││ │
│  │  │    clusterMcpRefs: [postgres-staging]                   ││ │
│  │  │    clusterEnvRefs: [aws-dev-profile]                    ││ │
│  │  │    clusterEgressRefs: [allow-anthropic, allow-github]   ││ │
│  │  └─────────────────────────────────────────────────────────┘│ │
│  │                            │                                │ │
│  │                            ▼                                │ │
│  │  ┌─────────────────────────────────────────────────────────┐│ │
│  │  │  Generated Resources (owned by OpenCodeAgent)           ││ │
│  │  │                                                         ││ │
│  │  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐  ││ │
│  │  │  │  ConfigMap  │  │ NetworkPol  │  │      PVC        │  ││ │
│  │  │  │             │  │             │  │                 │  ││ │
│  │  │  │ • skills/   │  │ • Ingress   │  │ • workspace     │  ││ │
│  │  │  │ • opencode  │  │   :4096     │  │   10Gi          │  ││ │
│  │  │  │   .json     │  │ • Egress    │  │                 │  ││ │
│  │  │  │             │  │   DNS, rules│  │                 │  ││ │
│  │  │  └─────────────┘  └─────────────┘  └─────────────────┘  ││ │
│  │  │                                                         ││ │
│  │  │  ┌─────────────────────────────────────────────────────┐│ │
│  │  │  │  Pod: billing-migration-agent                       ││ │
│  │  │  │                                                     ││ │
│  │  │  │  Volumes:                                           ││ │
│  │  │  │  ├── workspace (PVC)                                ││ │
│  │  │  │  ├── opencode-config (ConfigMap)                    ││ │
│  │  │  │  ├── git-ssh (Secret)                               ││ │
│  │  │  │  ├── tool-terraform-1-5 (image volume)              ││ │
│  │  │  │  └── tool-node-20 (image volume)                   ││ │
│  │  │  │                                                     ││ │
│  │  │  │  InitContainers:                                    ││ │
│  │  │  │  └── git-clone (alpine/git)                         ││ │
│  │  │  │      Clones repo → /workspace                       ││ │
│  │  │  │                                                     ││ │
│  │  │  │  Containers:                                        ││ │
│  │  │  │  ├── mcp-postgres-staging (sidecar)                 ││ │
│  │  │  │  │   Port: 8001, envFrom: staging-db-creds         ││ │
│  │  │  │  └── opencode (main)                                ││ │
│  │  │  │      Command: opencode serve --port 4096           ││ │
│  │  │  │      PATH: /opt/tools/terraform/bin:...            ││ │
│  │  │  │      Env: AWS_*, LOG_LEVEL, etc.                  ││ │
│  │  │  │      Resources: 2 CPU, 4Gi memory                  ││ │
│  │  │  │                                                     ││ │
│  │  │  └─────────────────────────────────────────────────────┘│ │
│  │  │                                                         ││ │
│  │  │  ┌─────────────────────────────────────────────────────┐│ │
│  │  │  │  Service: billing-migration-svc                     ││ │
│  │  │  │  Type: ClusterIP, Port: 4096                        ││ │
│  │  │  └─────────────────────────────────────────────────────┘│ │
│  │  └─────────────────────────────────────────────────────────┘ │
│  └─────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Reconciliation Flow

The OpenCodeAgent controller follows this reconciliation loop:

```
┌─────────────────┐
│  OpenCodeAgent  │◄────────────────────────────────────────┐
│    Created      │                                         │
└────────┬────────┘                                         │
         │                                                  │
         ▼                                                  │
┌─────────────────┐     ┌─────────────────────────────────┐  │
│ 1. Fetch Agent  │────►│ Resolve Cluster References      │  │
│    from API     │     │ • ClusterAgentTool              │  │
└─────────────────┘     │ • ClusterAgentSkill             │  │
                        │ • ClusterAgentMCPServer         │  │
                        │ • ClusterAgentEnv               │  │
                        │ • ClusterAgentEgressRule        │  │
                        └─────────────┬───────────────────┘  │
                                      │ Missing?             │
                                      ▼                      │
                              ┌───────────────┐              │
                              │ Status: Failed│──────────────┘
                              │ (with reason) │
                              └───────────────┘
                                      │ All found
                                      ▼
┌─────────────────┐     ┌─────────────────────────────────┐
│ 5. Create       │◄────│ 2. Generate ConfigMap           │
│    Service      │     │ • skills/<fileName>             │
│    (port 4096)  │     │ • opencode.json (MCP config)    │
└─────────────────┘     └─────────────────────────────────┘
         │                        │
         │                        ▼
         │               ┌─────────────────────────────────┐
         │               │ 3. Generate NetworkPolicy       │
         │               │ • Ingress: port 4096 only       │
         │               │ • Egress: DNS + allowed rules   │
         │               │ • Default: deny all             │
         │               └─────────────────────────────────┘
         │                        │
         ▼                        ▼
┌─────────────────┐     ┌─────────────────────────────────┐
│ 6. Update       │◄────│ 4. Create Pod                   │
│    Status       │     │ • PVC (workspace)               │
│ • Phase: Ready  │     │ • InitContainer (git clone)     │
│ • URL           │     │ • Sidecars (MCP servers)        │
│ • PodName       │     │ • Main container (opencode)     │
│ • ServiceName   │     │ • Image volumes (tools)         │
└─────────────────┘     │ • Dynamic PATH                  │
                        │ • Env vars + secrets            │
                        └─────────────────────────────────┘
```

## Watches and Reactivity

The controller implements **inverse watches** using `handler.EnqueueRequestsFromMapFunc`:

| Trigger Event | Action |
|--------------|--------|
| `ClusterAgentSkill` modified | Update ConfigMap for all referencing agents |
| `ClusterAgentEgressRule` modified | Update NetworkPolicy for all referencing agents |
| `ClusterAgentTool` image changed | Roll out Pod (delete + recreate) |
| `ClusterAgentMCPServer` modified | Roll out Pod |
| `ClusterAgentEnv` modified | Roll out Pod |

## Security Model

### Network Isolation

```
┌─────────────────────────────────────────┐
│           Agent Pod                     │
│  ┌─────────────────────────────────┐    │
│  │  Ingress:                       │    │
│  │  • Port 4096 (OpenCode API)     │    │
│  │    from Ingress/ServiceMesh     │    │
│  │                                 │    │
│  │  Egress:                        │    │
│  │  • DNS (53 UDP/TCP) → kube-dns  │    │
│  │  • Allowed domains/ports        │    │
│  │    (from ClusterAgentEgressRule)│    │
│  │  • localhost (intra-pod)        │    │
│  │                                 │    │
│  │  ❌ Everything else: DENIED     │    │
│  └─────────────────────────────────┘    │
└─────────────────────────────────────────┘
```

### Secret Handling

- Git SSH keys: Mounted as read-only secret volume in init container
- Environment secrets: Injected via `secretKeyRef` (never logged)
- MCP secrets: Injected via `envFrom` in sidecar containers

### Volume Security

| Volume Type | Mount | Access |
|-------------|-------|--------|
| Tools (OCI image) | `/opt/tools/<name>` | Read-only |
| Config (ConfigMap) | `~/.config/opencode` | Read-only |
| Git SSH (Secret) | `/git-ssh` | Read-only |
| Workspace (PVC) | `/workspace` | Read-write |

## Resource Ownership

All generated resources have **owner references** to the OpenCodeAgent:

```yaml
ownerReferences:
- apiVersion: creuzz.io/v1alpha1
  kind: OpenCodeAgent
  name: billing-migration
  controller: true
  blockOwnerDeletion: true
```

This enables:
- **Garbage Collection**: Deleting the OpenCodeAgent deletes all owned resources
- **Cascading Updates**: Changes propagate through the ownership chain

## Scaling Considerations

| Aspect | Behavior |
|--------|----------|
| Horizontal | Multiple agents can run in parallel (one Pod per agent) |
| Vertical | Configure via `spec.runtime.{cpu,memory}` |
| Storage | PVC per agent, configurable size |
| TTL | `spec.ttlAfterIdle` for scale-to-zero (future) |

## API Group

All Creuzz resources use the API group `creuzz.io`:

| Resource | Full Name | Scope |
|----------|-----------|-------|
| OpenCodeAgent | `opencodeagents.creuzz.io` | Namespaced |
| ClusterAgentTool | `clusteragenttools.creuzz.io` | Cluster |
| ClusterAgentSkill | `clusteragentskills.creuzz.io` | Cluster |
| ClusterAgentMCPServer | `clusteragentmcpservers.creuzz.io` | Cluster |
| ClusterAgentEnv | `clusteragentenvs.creuzz.io` | Cluster |
| ClusterAgentEgressRule | `clusteragentegressrules.creuzz.io` | Cluster |
