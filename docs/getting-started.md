# Getting Started

This guide walks you through creating your first OpenCode agent with Creuzz in under 5 minutes.

## Overview

You will:
1. Create a namespace
2. Set up a Git SSH secret
3. Create a ClusterAgentSkill (cluster-scoped)
4. Create a ClusterAgentTool (cluster-scoped)
5. Create an OpenCodeAgent
6. Verify the agent is running

## Step 1: Create a Namespace

```bash
kubectl create namespace team-backend
```

## Step 2: Create Git SSH Secret

Create a secret containing your Git SSH credentials:

```bash
kubectl create secret generic github-ssh-creds \
  --namespace team-backend \
  --from-file=id_ed25519=~/.ssh/id_ed25519 \
  --from-file=known_hosts=~/.ssh/known_hosts
```

> **Note**: The secret must contain `id_ed25519` (private key) and `known_hosts`.

## Step 3: Create a Skill (Cluster-Scoped)

Skills inject context and guidelines into the agent:

```yaml
# skill.yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentSkill
metadata:
  name: clean-architecture
spec:
  fileName: "clean-arch.md"
  content: |
    # Clean Architecture Guidelines
    - Isolate Domain and Infrastructure layers
    - DTOs required at controller boundaries
    - No dependencies from domain to external
    - Use interfaces for ports (hexagonal architecture)
```

```bash
kubectl apply -f skill.yaml
```

## Step 4: Create a Tool (Cluster-Scoped)

Tools are OCI images mounted as read-only volumes:

```yaml
# tool.yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentTool
metadata:
  name: terraform-1-5
spec:
  image: "hashicorp/terraform:1.5.7"
  mountPath: "/opt/tools/terraform"
  binPath: "/opt/tools/terraform/bin"
```

```bash
kubectl apply -f tool.yaml
```

## Step 5: Create an Agent

Now create your first OpenCode agent:

```yaml
# agent.yaml
apiVersion: creuzz.io/v1alpha1
kind: OpenCodeAgent
metadata:
  name: my-first-agent
  namespace: team-backend
spec:
  project:
    repoUrl: "git@github.com:myorg/myrepo.git"
    branch: "main"
    secretRef: "github-ssh-creds"
  runtime:
    image: "registry.internal/opencode-base:v2.0.0"
    cpu: "2"
    memory: "4Gi"
    storage: "10Gi"
  clusterToolRefs:
    - name: "terraform-1-5"
  clusterSkillRefs:
    - name: "clean-architecture"
```

```bash
kubectl apply -f agent.yaml
```

## Step 6: Verify

Check the agent status:

```bash
kubectl get opencodeagent my-first-agent -n team-backend
```

Expected output:
```
NAME             PHASE   URL
my-first-agent   Ready   https://agent-team-backend-my-first-agent.internal.local
```

View detailed status:

```bash
kubectl describe opencodeagent my-first-agent -n team-backend
```

Check created resources:

```bash
# All resources in the namespace
kubectl get all -n team-backend

# ConfigMap with skills and MCP config
kubectl get configmap my-first-agent-config -n team-backend -o yaml

# Pod with init container, sidecars, and main container
kubectl get pod my-first-agent-agent -n team-backend -o yaml

# NetworkPolicy
kubectl get networkpolicy my-first-agent-netpol -n team-backend -o yaml
```

## What Was Created?

The operator created these resources (all with owner references to the OpenCodeAgent):

| Resource | Name | Purpose |
|----------|------|---------|
| ConfigMap | `my-first-agent-config` | Skills + `opencode.json` MCP config |
| NetworkPolicy | `my-first-agent-netpol` | Zero-trust network rules |
| PVC | `my-first-agent-workspace` | 10Gi workspace volume |
| Pod | `my-first-agent-agent` | Agent runtime (init + main) |
| Service | `my-first-agent-svc` | ClusterIP on port 4096 |

## Access the Agent

The agent API is available on port 4096. To access it:

```bash
# Port-forward to the service
kubectl port-forward -n team-backend service/my-first-agent-svc 4096:4096

# Test the API
curl http://localhost:4096/health
```

## Next Steps

- [Add MCP Servers](guides/mcp-servers.md) — Connect databases and APIs
- [Configure Environment](guides/environment-variables.md) — Set env vars and secrets
- [Network Policies](guides/network-policies.md) — Allow external API access
- [Managing Tools](guides/managing-tools.md) — Add more CLI tools

## Cleanup

```bash
# Delete the agent (cascades to all owned resources)
kubectl delete opencodeagent my-first-agent -n team-backend

# Delete cluster-scoped resources
kubectl delete clusteragentskill clean-architecture
kubectl delete clusteragenttool terraform-1-5

# Delete namespace
kubectl delete namespace team-backend
```
