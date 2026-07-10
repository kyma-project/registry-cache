[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/registry-cache)](https://api.reuse.software/info/github.com/kyma-project/registry-cache)
[![Go Report Card](https://goreportcard.com/badge/github.com/kyma-project/registry-cache)](https://goreportcard.com/report/github.com/kyma-project/registry-cache)
[![unit tests](https://badgers.space/github/checks/kyma-project/registry-cache/main/unit-tests)](https://github.com/kyma-project/registry-cache/actions/workflows/unit-tests.yaml)
[![Coverage Status](https://coveralls.io/repos/github/kyma-project/registry-cache/badge.svg?branch=main)](https://coveralls.io/github/kyma-project/registry-cache?branch=main)
[![golangci lint](https://badgers.space/github/checks/kyma-project/registry-cache/main/golangci-lint)](https://github.com/kyma-project/registry-cache/actions/workflows/lint.yaml)
[![latest release](https://badgers.space/github/release/kyma-project/registry-cache)](https://github.com/kyma-project/registry-cache/releases/latest)

# Registry Cache

This repository contains the source code for the Registry Cache module.

## Overview

With the Registry Cache module, you can enable and configure a caching layer for container image registries used in your SAP BTP, Kyma runtime instances.  
This feature reduces the amount of outbound traffic from your runtimes to public registries, improving performance and reliability of image pulls.  
Additionally, it allows configuring access to private registries by providing credentials that the caching layer uses to authenticate against them.

For information on using the Registry Cache configuration, see the [user documentation](./docs/user/README.md).

> ### Note:
> Since this feature is implemented as part of Kyma Control Plane, it is available only for SAP BTP, Kyma runtime.  
> Installing this module in a self-managed Kyma cluster and providing Registry Cache configuration will have no effect.

## Prerequisites

- A managed Kyma runtime instance running on the SAP BTP platform.
- Access to Kyma dashboard (Busola) or kubectl with kubeconfig for the Kyma runtime cluster.

## Installation

For information on how to add a module to your Kyma cluster, see [Quick Install](https://kyma-project.io/02-get-started/01-quick-install.html).

## Development

For developer setup, architecture overview, and testing strategy, see the [contributor documentation](./docs/contributor/).

### Prerequisites

- Access to a Kubernetes cluster
- [Go](https://go.dev/) 1.26.4 or later
- [k3d](https://k3d.io/)
- [Docker](https://www.docker.com/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Kubebuilder](https://book.kubebuilder.io/)
- [yq](https://mikefarah.gitbook.io/yq)

### Installation in the k3d Cluster Using Make Targets

1. Clone the project:

    ```bash
    git clone https://github.com/kyma-project/registry-cache.git && cd registry-cache/
    ```

2. Create a new k3d cluster:

    ```bash
    k3d cluster create test-cluster
    ```

3. Build the controller image and load it into the k3d cluster:

    ```bash
    make docker-build
    k3d image import registry-cache-test:latest -c test-cluster
    ```

4. Apply the k3d installation manifest:

    ```bash
    make build-k3d-installer
    kubectl create ns kyma-system
    kubectl apply -f dist/k3d-install.yaml
    ```

5. Patch deployment:
    ```
    kubectl patch deployment registry-cache-controller-manager -n kyma-system \
        --type='json' \
        -p='[{"op":"replace","path":"/spec/template/spec/containers/0/imagePullPolicy","value":"Never"}]'
    ```

### Using Registry Cache Operator

- Create a Registry Cache instance.

    ```bash
    kubectl apply -f config/samples/default_registry_cache_cr.yaml
    ```

- Delete a Registry Cache instance.

    ```bash
    kubectl delete -f config/samples/default_registry_cache_cr.yaml
    ```

## Contributing

<!--- mandatory section - do not change this! --->

See the [Contributing Rules](CONTRIBUTING.md).

## Code of Conduct
<!--- mandatory section - do not change this! --->

See the [Code of Conduct](CODE_OF_CONDUCT.md) document.

## Licensing
<!--- mandatory section - do not change this! --->

See the [license](./LICENSE) file.
