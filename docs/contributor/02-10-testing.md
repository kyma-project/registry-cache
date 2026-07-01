# Testing Strategy

## Test Types

| Type | Framework | Location | Run with |
|---|---|---|---|
| Unit and controller tests | Ginkgo v2 + Gomega, `controller-runtime/envtest` | `internal/**/`, `api/` | `make test` |

## Unit and Controller Tests

Unit tests use `controller-runtime/envtest` to spin up a local Kubernetes API server using real kube-apiserver/etcd child processes. No external cluster is required, but the envtest binaries must be present locally. Before running tests for the first time, run:

```bash
make setup-envtest
```

Then run the tests:

```bash
make test
```

Test coverage is automatically uploaded to [Coveralls](https://coveralls.io/github/kyma-project/registry-cache) when the CI pipeline runs on the `main` branch.

## Linting

The project uses `golangci-lint` (v2.11.4).

```bash
make lint        # Check for linting issues
make lint-fix    # Apply auto-fixable suggestions
```

## CI Pipelines

Unit tests and linting run automatically on every pull request and on every push to `main`:

- Unit tests: `.github/workflows/unit-tests.yaml` — runs `make test`, uploads coverage to Coveralls
- Linting: `.github/workflows/lint.yaml` — runs `golangci-lint`

Container images are built on pushes to `main` and on version tags using `.github/workflows/build_registry_cache.yml`.
