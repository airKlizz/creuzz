# ClusterAgentEnv

`ClusterAgentEnv` groups environment variables and secret references for injection into agent containers.

- **API Group**: `creuzz.io`
- **Version**: `v1alpha1`
- **Kind**: `ClusterAgentEnv`
- **Scope**: Cluster

## Description

Environment profiles define reusable sets of environment variables (static values and secret references) that can be injected into multiple agents. This separates configuration from agent definitions.

## Specification

### `spec.variables`

List of static environment variables.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Environment variable name |
| `value` | string | Yes | Environment variable value |

### `spec.secrets`

List of environment variables sourced from Kubernetes Secrets.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Environment variable name |
| `secretRef` | string | Yes | Name of the Secret in the target namespace |
| `key` | string | Yes | Key within the Secret |

## Examples

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
    - name: LOG_LEVEL
      value: "debug"
  secrets:
    - name: AWS_ACCESS_KEY_ID
      secretRef: "aws-dev-creds"
      key: "access_key"
    - name: AWS_SECRET_ACCESS_KEY
      secretRef: "aws-dev-creds"
      key: "secret_key"
    - name: AWS_SESSION_TOKEN
      secretRef: "aws-dev-creds"
      key: "session_token"
```

Required secret:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-dev-creds
  namespace: team-backend
type: Opaque
stringData:
  access_key: "AKIAIOSFODNN7EXAMPLE"
  secret_key: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  session_token: ""
```

### LLM API Configuration

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
    - name: OPENCODE_TEMPERATURE
      value: "0.7"
  secrets:
    - name: ANTHROPIC_API_KEY
      secretRef: "llm-api-keys"
      key: "anthropic"
    - name: OPENAI_API_KEY
      secretRef: "llm-api-keys"
      key: "openai"
```

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
  secrets:
    - name: DATABASE_URL
      secretRef: "database-creds"
      key: "url"
    - name: DB_PASSWORD
      secretRef: "database-creds"
      key: "password"
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
      value: "agent@opencode.local"
  secrets:
    - name: GITHUB_TOKEN
      secretRef: "github-creds"
      key: "token"
    - name: GITLAB_TOKEN
      secretRef: "gitlab-creds"
      key: "token"
```

## Usage in OpenCodeAgent

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
```

## Generated Environment Variables

```yaml
containers:
- name: opencode
  env:
  # Static variables
  - name: AWS_DEFAULT_REGION
    value: "eu-west-3"
  - name: AWS_PROFILE
    value: "dev"
  - name: LOG_LEVEL
    value: "debug"
  # From secrets
  - name: AWS_ACCESS_KEY_ID
    valueFrom:
      secretKeyRef:
        name: aws-dev-creds
        key: access_key
  - name: AWS_SECRET_ACCESS_KEY
    valueFrom:
      secretKeyRef:
        name: aws-dev-creds
        key: secret_key
  # ...
```

## Secret Namespace

**Important**: Secrets referenced in `ClusterAgentEnv` must exist in the **same namespace** as the `OpenCodeAgent`, not in the cluster scope.

```
ClusterAgentEnv (cluster-scoped)     OpenCodeAgent (namespaced)
┌─────────────────┐                  ┌─────────────────┐
│  secrets:       │                  │  namespace:     │
│  - secretRef:   │─────────────────►│    team-backend │
│      aws-creds  │   looks up in    │                 │
│      key: id    │   agent's ns     │  Secret must    │
└─────────────────┘                  │  exist here!    │
                                     └─────────────────┘
```

## Best Practices

1. **Group by purpose** — One env profile per concern (AWS, LLM, DB, etc.)
2. **Use descriptive names** — `aws-dev-profile`, not `env1`
3. **Never commit secrets** — Use Kubernetes Secrets, not values
4. **Document required secrets** — List needed secrets in the profile description
5. **Use consistent naming** — Prefix related variables (e.g., `AWS_*`, `DB_*`)

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Secret not found | Create the secret in the agent's namespace |
| Key not found in secret | Verify the key exists: `kubectl get secret <name> -o yaml` |
| Env var not set | Check the agent Pod spec: `kubectl get pod <name> -o yaml` |

## See Also

- [OpenCodeAgent](opencodeagent.md) — Use env profiles in agents
- [Environment Variables](../guides/environment-variables.md) — Detailed guide
