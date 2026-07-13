# Set Up the Development Environment

## Prerequisites

- Go 1.26.4 or later (the required version is defined in `go.mod`)
- Docker or a compatible container tool
- `kubectl` configured to point to a Kubernetes cluster

## Clone and Build

```bash
git clone https://github.com/kyma-project/registry-cache.git
cd registry-cache
make build
```

The compiled manager binary is placed in `bin/manager`.

## Available Make Targets

| Target | Description |
|---|---|
| `make build` | Compile the manager binary |
| `make run` | Run the controller locally against the cluster configured in `~/.kube/config` (requires a valid TLS certificate at `/tmp/tls.crt` — see [Installation in the k3d Cluster Using Make Targets](../../README.md#installation-in-the-k3d-cluster-using-make-targets) for the recommended local dev workflow) |
| `make manifests` | Regenerate CRD manifests and `WebhookConfiguration` from kubebuilder markers |
| `make generate` | Regenerate `DeepCopy` methods from Go type definitions |
| `make fmt` | Format Go source files with `gofmt` |
| `make vet` | Run `go vet` static analysis |
| `make lint` | Run `golangci-lint` |
| `make lint-fix` | Run `golangci-lint` and apply auto-fixable suggestions |
| `make test` | Run unit and controller tests (no cluster required) |

## Code Generation

After modifying Go type definitions in `api/v1beta1/`, regenerate the derived files:

```bash
make generate   # DeepCopy methods
make manifests  # CRD YAML and WebhookConfiguration
```

Commit the generated files together with the type changes.
