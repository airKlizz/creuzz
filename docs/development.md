# Development Guide

This guide covers how to contribute to Creuzz, build from source, and run tests.

## Prerequisites

- Go 1.26+
- Docker
- kubectl
- kind (for local testing)
- make
- kubebuilder v4 (installed via `go install sigs.k8s.io/kubebuilder/v4@latest`)

## Project Structure

```
creuzz/
├── api/v1alpha1/           # CRD type definitions
│   ├── groupversion_info.go
│   ├── opencodeagent_types.go
│   ├── clusteragenttool_types.go
│   ├── clusteragentskill_types.go
│   ├── clusteragentmcpserver_types.go
│   ├── clusteragentenv_types.go
│   ├── clusteragentegressrule_types.go
│   └── zz_generated.deepcopy.go  # Auto-generated
├── internal/controller/    # Reconciliation logic
│   ├── opencodeagent_controller.go
│   ├── clusteragenttool_controller.go
│   └── ...
├── cmd/main.go             # Operator entrypoint
├── config/                 # Kustomize manifests
│   ├── crd/               # CRD definitions
│   ├── rbac/              # RBAC rules
│   ├── manager/           # Deployment
│   ├── default/           # Default kustomization
│   └── samples/           # Example resources
├── docs/                   # Documentation
├── test/                   # E2E tests
├── Makefile               # Build targets
├── Dockerfile             # Container image
├── go.mod
└── PROJECT                # Kubebuilder project config
```

## Building

### Local Build

```bash
# Build the binary
make build

# Run locally (requires kubeconfig)
make run
```

### Docker Build

```bash
# Build image
make docker-build IMG=creuzz-operator:dev

# Build and push
make docker-build docker-push IMG=registry.example.com/creuzz:dev
```

## Development Workflow

### 1. Make Changes

Edit Go files in `api/v1alpha1/` or `internal/controller/`.

### 2. Regenerate Code

After modifying types or RBAC markers:

```bash
# Generate deepcopy methods
make generate

# Generate CRDs and RBAC
make manifests

# Format code
make fmt

# Run vet
make vet
```

### 3. Run Tests

```bash
# Run unit tests (uses envtest)
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test ./internal/controller/... -run TestOpenCodeAgent -v
```

### 4. Test Locally with Kind

```bash
# Create Kind cluster
kind create cluster --name creuzz-dev

# Build and load image
make docker-build IMG=creuzz-operator:dev
kind load docker-image creuzz-operator:dev --name creuzz-dev

# Install CRDs
make install

# Deploy operator
make deploy IMG=creuzz-operator:dev

# Create test resources
kubectl apply -f config/samples/

# Watch logs
kubectl logs -n creuzz-system deployment/creuzz-controller-manager -f
```

### 5. Clean Up

```bash
# Delete Kind cluster
kind delete cluster --name creuzz-dev
```

## Adding a New CRD

Use kubebuilder to scaffold a new API:

```bash
# Create new API
kubebuilder create api \
  --group creuzz \
  --version v1alpha1 \
  --kind MyNewResource \
  --namespaced=false \
  --resource=true \
  --controller=true

# Update PROJECT file domain
# Edit api/v1alpha1/groupversion_info.go if needed

# Regenerate
make manifests generate
```

## Code Style

### Go Guidelines

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` and `goimports`
- Run `make fmt` and `make vet` before committing
- Add comments for exported types and functions

### Kubernetes Conventions

- Use `metav1.Condition` for status (not custom strings)
- Use `metav1.Time` for timestamps
- Follow [API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- Use structured logging: `log.Info("message", "key", value)`

### Logging

Follow [Kubernetes logging guidelines](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/logging.md):

```go
// Good
log.Info("Created Deployment", "name", deploy.Name)
log.Error(err, "Failed to create Pod", "name", name)

// Bad
log.Info("created deployment " + deploy.Name)
log.Error(err, "pod creation failed")
```

## Testing

### Unit Tests

Tests use Ginkgo + Gomega and envtest:

```go
var _ = Describe("OpenCodeAgent Controller", func() {
    Context("When reconciling a resource", func() {
        It("should successfully reconcile the resource", func() {
            // Test logic
        })
    })
})
```

### E2E Tests

E2E tests are in `test/e2e/` and require a Kind cluster:

```bash
# Run E2E tests
make test-e2e
```

## Pull Request Process

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make changes and commit: `git commit -m "Add my feature"`
4. Run tests: `make test`
5. Push: `git push origin feature/my-feature`
6. Open a Pull Request

### Commit Message Format

```
<type>: <description>

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Examples:
```
feat: add support for custom ingress annotations
fix: resolve race condition in pod creation
docs: update getting started guide
```

## Release Process

1. Update version in `Makefile` (IMG tag)
2. Update `CHANGELOG.md`
3. Create git tag: `git tag v0.1.0`
4. Push tag: `git push origin v0.1.0`
5. Build and push release image
6. Create GitHub release

## Resources

- [Kubebuilder Book](https://book.kubebuilder.io)
- [Controller Runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- [Kubernetes Logging Guidelines](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/logging.md)

## Getting Help

- Open an issue on GitHub
- Check existing issues and PRs
- Review the [Architecture](architecture.md) documentation
