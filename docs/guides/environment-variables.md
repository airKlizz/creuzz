# Environment Variables

This guide covers how to configure environment variables for agents using `ClusterAgentEnv`.

## Overview

`ClusterAgentEnv` allows you to define reusable environment profiles that can be shared across multiple agents. It supports:

- **Static variables**: Fixed values (region, log level, etc.)
- **Secret references**: Values from Kubernetes Secrets (API keys, passwords)

## Creating an Environment Profile

### Basic Profile

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentEnv
metadata:
  name: basic-config
spec:
  variables:
    - name: LOG_LEVEL
      value: "info"
    - name: ENVIRONMENT
      value: "development"
```

### AWS Profile

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentEnv
metadata:
  name: aws-dev-profile
spec:
  variables:
    - name: AWS_DEFAULT_REGION
      value: "eu-west-3"
    - name: AWS_PROFILE
      value: "dev"
  secrets:
    - name: AWS_ACCESS_KEY_ID
      secretRef: "aws-creds"
      key: "access_key"
    - name: AWS_SECRET_ACCESS_KEY
      secretRef: "aws-creds"
      key: "secret_key"
```

### LLM Configuration

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentEnv
metadata:
  name: llm-config
spec:
  variables:
    - name: OPENCODE_MODEL
      value: "claude-sonnet-4"
    - name: OPENCODE_MAX_TOKENS
      value: "8192"
  secrets:
    - name: ANTHROPIC_API_KEY
      secretRef: "llm-keys"
      key: "anthropic"
    - name: OPENAI_API_KEY
      secretRef: "llm-keys"
      key: "openai"
```

## Managing Secrets

### Creating Secrets

Secrets must exist in the **agent's namespace** (not cluster-scoped):

```bash
# Create AWS credentials
kubectl create secret generic aws-creds \
  --namespace team-backend \
  --from-literal=access_key="$AWS_ACCESS_KEY_ID" \
  --from-literal=secret_key="$AWS_SECRET_ACCESS_KEY"

# Create LLM API keys
kubectl create secret generic llm-keys \
  --namespace team-backend \
  --from-literal=anthropic="$ANTHROPIC_API_KEY" \
  --from-literal=openai="$OPENAI_API_KEY"
```

### Secret Best Practices

1. **Never commit secrets** to version control
2. **Use external secret managers** (Vault, AWS Secrets Manager, etc.) with CSI drivers
3. **Rotate credentials** regularly
4. **Use least privilege** — Only grant necessary permissions
5. **Audit access** — Enable audit logging for secret access

## Environment Composition

Combine multiple profiles in one agent:

```yaml
apiVersion: creuzz.io/v1alpha1
kind: OpenCodeAgent
metadata:
  name: my-agent
  namespace: team-backend
spec:
  # ...
  clusterEnvRefs:
    - name: "aws-dev-profile"
    - name: "llm-config"
    - name: "database-config"
    - name: "cicd-config"
```

Resulting environment:
```
AWS_DEFAULT_REGION=eu-west-3
AWS_PROFILE=dev
AWS_ACCESS_KEY_ID=<from secret>
AWS_SECRET_ACCESS_KEY=<from secret>
OPENCODE_MODEL=claude-sonnet-4
OPENCODE_MAX_TOKENS=8192
ANTHROPIC_API_KEY=<from secret>
DATABASE_URL=<from secret>
GITHUB_TOKEN=<from secret>
...
```

## Variable Precedence

If multiple profiles define the same variable, the **last one wins**:

```yaml
clusterEnvRefs:
  - name: "base-config"      # LOG_LEVEL=info
  - name: "debug-config"     # LOG_LEVEL=debug (overrides)
```

Result: `LOG_LEVEL=debug`

## Debugging

### View Environment Variables

```bash
# Get pod spec
kubectl get pod <agent-pod> -n <namespace> -o yaml | grep -A 50 "env:"

# Exec into container and check
kubectl exec <agent-pod> -n <namespace> -- env | sort
```

### Check Secret Resolution

```bash
# Verify secret exists
kubectl get secret <secret-name> -n <namespace>

# Check secret keys
kubectl get secret <secret-name> -n <namespace> -o jsonpath='{.data}' | jq 'keys'
```

### Common Issues

| Issue | Solution |
|-------|----------|
| Variable not set | Check profile is referenced in agent spec |
| Secret not found | Create secret in agent's namespace |
| Key not found | Verify key name in secret matches spec |
| Wrong value | Check for typos in variable name |

## Examples

### Database Configuration

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentEnv
metadata:
  name: database-config
spec:
  variables:
    - name: DB_POOL_SIZE
      value: "10"
    - name: DB_TIMEOUT
      value: "30s"
    - name: DB_SSLMODE
      value: "require"
  secrets:
    - name: DATABASE_URL
      secretRef: "db-creds"
      key: "url"
    - name: DB_PASSWORD
      secretRef: "db-creds"
      key: "password"
```

### Feature Flags

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentEnv
metadata:
  name: feature-flags
spec:
  variables:
    - name: FEATURE_NEW_UI
      value: "true"
    - name: FEATURE_EXPERIMENTAL_API
      value: "false"
    - name: FEATURE_MAX_UPLOAD_SIZE
      value: "10485760"  # 10MB
```

### CI/CD Configuration

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentEnv
metadata:
  name: cicd-config
spec:
  variables:
    - name: CI
      value: "true"
    - name: GIT_AUTHOR_NAME
      value: "OpenCode Agent"
    - name: GIT_AUTHOR_EMAIL
      value: "agent@company.com"
    - name: GIT_COMMITTER_NAME
      value: "OpenCode Agent"
    - name: GIT_COMMITTER_EMAIL
      value: "agent@company.com"
  secrets:
    - name: GITHUB_TOKEN
      secretRef: "github-creds"
      key: "token"
```

## Best Practices

1. **Group by purpose** — AWS, LLM, DB, CI/CD in separate profiles
2. **Use consistent naming** — Prefix related variables (`AWS_*`, `DB_*`)
3. **Document required secrets** — List in profile description
4. **Use defaults wisely** — Provide sensible defaults in variables
5. **Validate early** — Test profiles before deploying agents

## See Also

- [ClusterAgentEnv API Reference](../api-reference/clusteragentenv.md)
- [Setting Up MCP Servers](mcp-servers.md)
