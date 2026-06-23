# Testing Strategy

## Test Types

| Type | Framework | Location | Run with |
|---|---|---|---|
| Unit and controller tests | Ginkgo v2 + Gomega, `controller-runtime/envtest` | `internal/**/`, `api/` | `make test` |
| End-to-end tests | Ginkgo v2 + Gomega | `test/e2e/` | `make test-e2e` |

## Unit and Controller Tests

Unit tests use `controller-runtime/envtest` to spin up a local Kubernetes API server in-process. No external cluster is required.

```bash
make test
```

Test coverage is automatically uploaded to [Coveralls](https://coveralls.io/github/kyma-project/registry-cache) when the CI pipeline runs on the `main` branch.

## End-to-End Tests

End-to-end tests require a running k3d cluster. The suite deploys the controller into the cluster and exercises real Kubernetes resources.

```bash
# Start a k3d cluster first, then:
make test-e2e
```

End-to-end tests are not run automatically in CI. They must be triggered manually or as part of a release pipeline.

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
