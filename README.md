[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/registry-cache)](https://api.reuse.software/info/github.com/kyma-project/registry-cache)
[![Go Report Card](https://goreportcard.com/badge/github.com/kyma-project/registry-cache)](https://goreportcard.com/report/github.com/kyma-project/registry-cache)
[![unit tests](https://badgen.net/github/checks/kyma-project/registry-cache/main/unit-tests)](https://github.com/kyma-project/registry-cache/actions/workflows/unit-tests.yaml)
[![Coverage Status](https://coveralls.io/repos/github/kyma-project/registry-cache/badge.svg?branch=main)](https://coveralls.io/github/kyma-project/registry-cache?branch=main)
[![golangci lint](https://badgen.net/github/checks/kyma-project/registry-cache/main/golangci-lint)](https://github.com/kyma-project/registry-cache/actions/workflows/lint.yaml)
[![latest release](https://badgen.net/github/release/kyma-project/registry-cache)](https://github.com/kyma-project/registry-cache/releases/latest)

# Registry Cache Kyma Module

This repository contains the source code for the Registry Cache Kyma Module.

## Overview

The Registry Cache Kyma module adds a possibility to enable and configure a caching layer for container image registries used in your BTP managed Kyma Runtimes.  
This feature reduces the amount of outbound traffic from your runtimes to public registries, improving performance and reliability of image pulls.  
Additionally, it allows to configure access to private registries by providing credentials that will be used by the caching layer to authenticate against those registries.  

For information how to use registry cache configuration, see the [user documentation](./docs/user/Readme.md).

**Note:** 
> As this feature is implemented as part of Kyma Control Plane it is available only for BTP managed Kyma Runtimes.  
> Installing his module in self-managed Kyma Runtime cluster and providing registry cache configuration will have no effect.

## Prerequisites

- A managed Kyma Runtime instance running on BTP platform.
- Access to Kyma console (Busola) or kubectl with kubeconfig for the Kyma Runtime cluster.

## Installation with kubectl

Enable the Registry Cache module in your Kyma cluster with kubectl by applying a custom resource.  
Apply the following script to install Registry Cache module operator:

```bash
kubectl apply -f https://github.com/kyma-project/registry-cache/releases/latest/download/registry-cache.yaml
```
To get Registry Cache configuration types installed, apply the sample Registry Cache CR:

```bash
kubectl apply -f https://github.com/kyma-project/registry-cache/releases/latest/download/default_registry_cache_cr.yaml
``` 

## Installation with Busola
To enable the Registry Cache module in your Kyma cluster with Busola find the list of "Modules" section in the main navigation panel.    
Then, click on "Modify Module" button and select "Registry Cache" from the list:

## Development

### Prerequisites

- Access to a Kubernetes cluster
- [Go](https://go.dev/)
- [k3d](https://k3d.io/)
- [Docker](https://www.docker.com/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Kubebuilder](https://book.kubebuilder.io/)
- [yq](https://mikefarah.gitbook.io/yq)

### Installation in the k3d Cluster Using Make Targets

1. Clone the project.

    ```bash
    git clone https://github.com/kyma-project/registry-cache.git && cd registry-cache/
    ```

2. Create a new k3d cluster and run registry-cache from the main branch:

    ```bash
    k3d cluster create test-cluster
    make deploy
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

For information on implementation details of registry cache module, see the [contributor documentation](./docs/contributor/Readme.md).
<!--- mandatory section - do not change this! --->

For information how to contribute see the [Contributing Rules](CONTRIBUTING.md).

## Code of Conduct
<!--- mandatory section - do not change this! --->

See the [Code of Conduct](CODE_OF_CONDUCT.md) document.

## Licensing
<!--- mandatory section - do not change this! --->

See the [license](./LICENSE) file.
