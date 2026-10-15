# Creating Your First Agent

This guide walks you through creating a complete OpenCode agent from scratch, including all required catalog resources.

## Scenario

You want to create an agent that:
- Works on the `billing-service` repository
- Has access to Terraform and Node.js tools
- Follows clean architecture principles
- Can query a PostgreSQL database via MCP
- Uses AWS credentials for deployment
- Can access the Anthropic API and GitHub

## Step 1: Prerequisites

Ensure Creuzz is installed:

```bash
kubectl get crds | grep creuzz.io
kubectl get deployment -n creuzz-system creuzz-controller-manager
```

## Step 2: Create Namespace and Secrets

```bash
# Create namespace
kubectl create namespace team-backend

# Create Git SSH secret
kubectl create secret generic github-ssh-creds \
  --namespace team-backend \
  --from-file=id_ed25519=~/.ssh/id_ed25519 \
  --from-file=known_hosts=~/.ssh/known_hosts

# Create AWS credentials
kubectl create secret generic aws-dev-creds \
  --namespace team-backend \
  --from-literal=access_key="$AWS_ACCESS_KEY_ID" \
  --from-literal=secret_key="$AWS_SECRET_ACCESS_KEY"

# Create database credentials for MCP
kubectl create secret generic staging-db-creds \
  --namespace team-backend \
  --from-literal=POSTGRES_HOST="postgres.staging.internal" \
  --from-literal=POSTGRES_PORT="5432" \
  --from-literal=POSTGRES_USER="agent" \
  --from-literal=POSTGRES_PASSWORD="$DB_PASSWORD" \
  --from-literal=POSTGRES_DATABASE="billing"
```

## Step 3: Create Catalog Resources

### Tool: Terraform

```yaml
# terraform-tool.yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentTool
metadata:
  name: terraform-1-5
spec:
  image: "hashicorp/terraform:1.5.7"
  mountPath: "/opt/tools/terraform"
  binPath: "/opt/tools/terraform/bin"
```

### Tool: Node.js

```yaml
# node-tool.yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentTool
metadata:
  name: node-20
spec:
  image: "node:20-alpine"
  mountPath: "/opt/tools/node"
  binPath: "/opt/tools/node/bin"
```

### Skill: Clean Architecture

```yaml
# clean-arch-skill.yaml
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
    - No framework dependencies in domain
    - Use interfaces for external dependencies
```

### MCP Server: PostgreSQL

```yaml
# postgres-mcp.yaml
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

### Environment: AWS Profile

```yaml
# aws-env.yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentEnv
metadata:
  name: aws-dev-profile
spec:
  variables:
    - name: AWS_DEFAULT_REGION
      value: "eu-west-3"
    - name: LOG_LEVEL
      value: "info"
  secrets:
    - name: AWS_ACCESS_KEY_ID
      secretRef: "aws-dev-creds"
      key: "access_key"
    - name: AWS_SECRET_ACCESS_KEY
      secretRef: "aws-dev-creds"
      key: "secret_key"
```

### Egress Rules

```yaml
# egress-rules.yaml
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
---
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentEgressRule
metadata:
  name: allow-github
spec:
  description: "GitHub API and git"
  domains:
    - "github.com"
    - "api.github.com"
  ports:
    - protocol: TCP
      port: 443
    - protocol: TCP
      port: 22
```

Apply all catalog resources:

```bash
kubectl apply -f terraform-tool.yaml
kubectl apply -f node-tool.yaml
kubectl apply -f clean-arch-skill.yaml
kubectl apply -f postgres-mcp.yaml
kubectl apply -f aws-env.yaml
kubectl apply -f egress-rules.yaml
```

## Step 4: Create the Agent

```yaml
# billing-agent.yaml
apiVersion: creuzz.io/v1alpha1
kind: OpenCodeAgent
metadata:
  name: billing-migration
  namespace: team-backend
spec:
  project:
    repoUrl: "git@github.com:orga/billing-service.git"
    branch: "refactor/v2"
    secretRef: "github-ssh-creds"
  runtime:
    image: "registry.internal/opencode-base:v2.0.0"
    cpu: "2"
    memory: "4Gi"
    storage: "10Gi"
  clusterToolRefs:
    - name: "terraform-1-5"
    - name: "node-20"
  clusterSkillRefs:
    - name: "clean-architecture"
  clusterMcpRefs:
    - name: "postgres-staging"
  clusterEnvRefs:
    - name: "aws-dev-profile"
  clusterEgressRefs:
    - name: "allow-anthropic"
    - name: "allow-github"
```

```bash
kubectl apply -f billing-agent.yaml
```

## Step 5: Verify

```bash
# Check agent status
kubectl get opencodeagent billing-migration -n team-backend

# Expected:
# NAME                PHASE   URL
# billing-migration   Ready   https://agent-team-backend-billing-migration.internal.local

# Check all resources
kubectl get all,configmap,networkpolicy,pvc -n team-backend

# View pod details
kubectl describe pod billing-migration-agent -n team-backend
```

## Step 6: Access the Agent

```bash
# Port-forward to access the API
kubectl port-forward -n team-backend service/billing-migration-svc 4096:4096

# In another terminal, test the API
curl http://localhost:4096/health
```

## What's Running?

The agent Pod contains:

```
┌─────────────────────────────────────────┐
│  Pod: billing-migration-agent           │
│                                         │
│  Init: git-clone                        │
│  └── Clones billing-service repo        │
│                                         │
│  Containers:                            │
│  ├── mcp-postgres-staging (sidecar)     │
│  │   └── PostgreSQL MCP on :8001         │
│  └── opencode (main)                    │
│      └── OpenCode API on :4096          │
│      ├── Tools: terraform, node         │
│      ├── Skills: clean-arch.md          │
│      └── Env: AWS_*, LOG_LEVEL          │
│                                         │
│  Volumes:                               │
│  ├── workspace (PVC 10Gi)               │
│  ├── opencode-config (ConfigMap)        │
│  ├── git-ssh (Secret)                   │
│  ├── tool-terraform-1-5 (image)         │
│  └── tool-node-20 (image)               │
└─────────────────────────────────────────┘
```

## Cleanup

```bash
# Delete the agent (cascades to all owned resources)
kubectl delete opencodeagent billing-migration -n team-backend

# Delete catalog resources (optional)
kubectl delete clusteragenttool terraform-1-5 node-20
kubectl delete clusteragentskill clean-architecture
kubectl delete clusteragentmcpserver postgres-staging
kubectl delete clusteragentenv aws-dev-profile
kubectl delete clusteragentegressrule allow-anthropic allow-github

# Delete namespace and secrets
kubectl delete namespace team-backend
```

## Next Steps

- [Managing Tools](managing-tools.md) — Add more tools
- [Configuring Skills](configuring-skills.md) — Create custom skills
- [Setting Up MCP Servers](mcp-servers.md) — Add more MCP servers
- [Environment Variables](environment-variables.md) — Advanced env config
- [Network Policies](network-policies.md) — Fine-tune egress rules
