# ClusterAgentSkill

`ClusterAgentSkill` injects context, system prompts, or architecture guidelines into agent environments.

- **API Group**: `creuzz.io`
- **Version**: `v1alpha1`
- **Kind**: `ClusterAgentSkill`
- **Scope**: Cluster

## Description

Skills are markdown files mounted into the agent's configuration directory at `~/.config/opencode/skills/<fileName>`. They provide context, coding standards, and architectural guidance to the OpenCode agent.

## Specification

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `fileName` | string | Yes | Name of the skill file (alphanumeric, `-`, `_`, `.` only) |
| `content` | string | Yes | Markdown content of the skill |

## Examples

### Clean Architecture Guidelines

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentSkill
metadata:
  name: clean-architecture
spec:
  fileName: "clean-arch.md"
  content: |
    # Clean Architecture Guidelines

    ## Layer Structure
    - **Domain**: Pure business logic, no dependencies
    - **Application**: Use cases and orchestration
    - **Infrastructure**: External concerns (DB, HTTP, etc.)
    - **Interface**: Controllers, presenters, gateways

    ## Rules
    - DTOs required at controller boundaries
    - No framework dependencies in domain layer
    - Use interfaces (ports) for external dependencies
    - Dependency rule: dependencies point inward only
```

### Security Best Practices

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentSkill
metadata:
  name: security-best-practices
spec:
  fileName: "security.md"
  content: |
    # Security Best Practices

    ## Secrets Management
    - Never commit secrets to version control
    - Use environment variables for sensitive data
    - Rotate credentials regularly

    ## Input Validation
    - Validate all user inputs
    - Use allowlists, not blocklists
    - Sanitize output to prevent injection

    ## Dependencies
    - Keep dependencies updated
    - Scan for vulnerabilities regularly
    - Use minimal base images
```

### API Design Standards

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentSkill
metadata:
  name: api-design
spec:
  fileName: "api-standards.md"
  content: |
    # API Design Standards

    ## REST Conventions
    - Use plural nouns for resources: `/users`, `/orders`
    - Use kebab-case for multi-word resources: `/order-items`
    - HTTP methods: GET (read), POST (create), PUT (replace), PATCH (update), DELETE
    - Return appropriate status codes: 200, 201, 400, 401, 403, 404, 500

    ## Versioning
    - URL versioning: `/v1/users`, `/v2/users`
    - Deprecate old versions with Sunset header

    ## Pagination
    - Use cursor-based pagination for large datasets
    - Include `next` and `previous` links
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
  clusterSkillRefs:
    - name: "clean-architecture"
    - name: "security-best-practices"
    - name: "api-design"
```

## Resulting ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-agent-config
  namespace: team-backend
data:
  opencode.json: |
    {"mcp": {"servers": {}}}
  skill-clean-arch.md: |
    # Clean Architecture Guidelines
    ...
  skill-security.md: |
    # Security Best Practices
    ...
  skill-api-standards.md: |
    # API Design Standards
    ...
```

## Resulting Volume Mount

The ConfigMap is mounted with `items` mapping to create the directory structure:

```yaml
volumes:
- name: opencode-config
  configMap:
    name: my-agent-config
    items:
    - key: opencode.json
      path: opencode.json
    - key: skill-clean-arch.md
      path: skills/clean-arch.md
    - key: skill-security.md
      path: skills/security.md
```

Files appear at:
- `~/.config/opencode/opencode.json`
- `~/.config/opencode/skills/clean-arch.md`
- `~/.config/opencode/skills/security.md`

## Best Practices

1. **Keep skills focused** — One skill per topic (architecture, security, testing, etc.)
2. **Use clear file names** — `clean-arch.md`, not `skill1.md`
3. **Version control skills** — Store skill definitions in Git
4. **Update safely** — Changes propagate to all referencing agents automatically

## File Name Restrictions

The `fileName` must match the pattern: `^[a-zA-Z0-9._-]+$`

Valid: `clean-arch.md`, `security_v2.md`, `api.standards.md`
Invalid: `my skill.md` (spaces), `skill/file.md` (slashes)

## See Also

- [OpenCodeAgent](opencodeagent.md) — Use skills in agents
- [Configuring Skills](../guides/configuring-skills.md) — Guide for skill management
