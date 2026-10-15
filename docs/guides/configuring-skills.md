# Configuring Skills

This guide covers how to create and manage `ClusterAgentSkill` resources to inject context and guidelines into agents.

## Overview

Skills are markdown files that provide context, coding standards, and architectural guidance to OpenCode agents. They are mounted at `~/.config/opencode/skills/<fileName>`.

## Creating a Skill

### Basic Skill

```yaml
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
```

### Skill with Code Examples

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentSkill
metadata:
  name: go-patterns
spec:
  fileName: "go-patterns.md"
  content: |
    # Go Design Patterns

    ## Repository Pattern

    ```go
    type UserRepository interface {
        FindByID(ctx context.Context, id string) (*User, error)
        Save(ctx context.Context, user *User) error
    }
    ```

    ## Error Handling

    Always wrap errors with context:

    ```go
    if err != nil {
        return fmt.Errorf("failed to fetch user: %w", err)
    }
    ```
```

## Skill Categories

### Architecture Skills

Define structural guidelines:

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentSkill
metadata:
  name: hexagonal-architecture
spec:
  fileName: "hexagonal.md"
  content: |
    # Hexagonal Architecture (Ports & Adapters)

    ## Core Principles
    - Domain logic is independent of external concerns
    - Communication through ports (interfaces)
    - Adapters implement ports for specific technologies

    ## Directory Structure
    ```
    /domain         # Pure business logic
    /application    # Use cases
    /infrastructure # Adapters (DB, HTTP, etc.)
    /interfaces     # Ports (interfaces)
    ```
```

### Security Skills

Define security requirements:

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentSkill
metadata:
  name: owasp-top-10
spec:
  fileName: "owasp.md"
  content: |
    # OWASP Top 10 Awareness

    ## A01: Broken Access Control
    - Deny by default
    - Check permissions on every request
    - Disable directory listing

    ## A02: Cryptographic Failures
    - Use TLS for all communications
    - Encrypt sensitive data at rest
    - Use strong algorithms (AES-256, RSA-2048+)

    ## A03: Injection
    - Use parameterized queries
    - Validate and sanitize all inputs
    - Use allowlists, not blocklists
```

### Testing Skills

Define testing standards:

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentSkill
metadata:
  name: testing-standards
spec:
  fileName: "testing.md"
  content: |
    # Testing Standards

    ## Unit Tests
    - Test one thing per test
    - Use table-driven tests in Go
    - Mock external dependencies
    - Aim for >80% coverage

    ## Integration Tests
    - Test with real dependencies (testcontainers)
    - Clean up resources after tests
    - Use unique names to avoid conflicts

    ## Test Structure (AAA)
    ```go
    func TestUserService_Create(t *testing.T) {
        // Arrange
        svc := NewUserService(mockRepo)

        // Act
        err := svc.Create(ctx, user)

        // Assert
        assert.NoError(t, err)
    }
    ```
```

### API Design Skills

```yaml
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentSkill
metadata:
  name: rest-api-design
spec:
  fileName: "rest-api.md"
  content: |
    # REST API Design

    ## Resource Naming
    - Use plural nouns: `/users`, `/orders`
    - Use kebab-case: `/order-items`
    - Nest resources: `/users/{id}/orders`

    ## HTTP Methods
    | Method | Purpose | Idempotent |
    |--------|---------|------------|
    | GET | Read | Yes |
    | POST | Create | No |
    | PUT | Replace | Yes |
    | PATCH | Update | No |
    | DELETE | Delete | Yes |

    ## Status Codes
    - 200 OK, 201 Created, 204 No Content
    - 400 Bad Request, 401 Unauthorized, 403 Forbidden
    - 404 Not Found, 409 Conflict
    - 500 Internal Server Error
```

## Updating Skills

Skills are **live-updated**: when a `ClusterAgentSkill` is modified, all referencing agents automatically receive the update in their ConfigMap.

```bash
# Update a skill
kubectl patch clusteragentskill clean-architecture --type=merge \
  -p '{"spec":{"content":"# Updated Guidelines\n- New rule added"}}'

# Verify propagation
kubectl get configmap <agent-name>-config -n <namespace> -o yaml | grep -A 5 "clean-arch"
```

## Multi-Language Skills

Create skills for different programming languages:

```yaml
# Python skill
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentSkill
metadata:
  name: python-standards
spec:
  fileName: "python.md"
  content: |
    # Python Standards
    - Follow PEP 8 style guide
    - Use type hints (PEP 484)
    - Use `black` for formatting
    - Use `mypy` for type checking
    - Prefer `pathlib` over `os.path`
---
# TypeScript skill
apiVersion: creuzz.io/v1alpha1
kind: ClusterAgentSkill
metadata:
  name: typescript-standards
spec:
  fileName: "typescript.md"
  content: |
    # TypeScript Standards
    - Use strict mode (`"strict": true`)
    - Avoid `any` type
    - Use interfaces over types for objects
    - Use `readonly` for immutable data
    - Prefer `const` over `let`
```

## Skill Composition

Agents can reference multiple skills:

```yaml
apiVersion: creuzz.io/v1alpha1
kind: OpenCodeAgent
metadata:
  name: full-stack-agent
  namespace: team-backend
spec:
  # ...
  clusterSkillRefs:
    - name: "clean-architecture"
    - name: "rest-api-design"
    - name: "typescript-standards"
    - name: "testing-standards"
    - name: "owasp-top-10"
```

## File Name Rules

The `fileName` field must match: `^[a-zA-Z0-9._-]+$`

| Valid | Invalid |
|-------|---------|
| `clean-arch.md` | `my skill.md` (space) |
| `api_v2.md` | `skill/file.md` (slash) |
| `standards.v2.md` | `skill@name.md` (special char) |

## Best Practices

1. **One topic per skill** — Don't mix architecture and testing
2. **Use clear names** — `clean-arch.md`, not `rules.md`
3. **Include examples** — Code snippets are more effective than prose
4. **Keep it concise** — Focus on actionable guidance
5. **Version in Git** — Store skill definitions in version control
6. **Test with agents** — Verify skills are understood by the LLM

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Skill not appearing | Check fileName pattern (no spaces/slashes) |
| Agent ignores skill | Verify skill is referenced in agent spec |
| Content not updating | Check ConfigMap: `kubectl get cm <agent>-config -o yaml` |
| Markdown not rendering | Ensure proper indentation in YAML |

## See Also

- [ClusterAgentSkill API Reference](../api-reference/clusteragentskill.md)
- [Creating Your First Agent](creating-your-first-agent.md)
