[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/registry-cache)](https://api.reuse.software/info/github.com/kyma-project/registry-cache)
[![Go Report Card](https://goreportcard.com/badge/github.com/kyma-project/registry-cache)](https://goreportcard.com/report/github.com/kyma-project/registry-cache)
[![unit tests](https://badgers.space/github/checks/kyma-project/registry-cache/main/unit-tests)](https://github.com/kyma-project/registry-cache/actions/workflows/unit-tests.yaml)
[![Coverage Status](https://coveralls.io/repos/github/kyma-project/registry-cache/badge.svg?branch=main)](https://coveralls.io/github/kyma-project/registry-cache?branch=main)
[![golangci lint](https://badgers.space/github/checks/kyma-project/registry-cache/main/golangci-lint)](https://github.com/kyma-project/registry-cache/actions/workflows/lint.yaml)
[![latest release](https://badgers.space/github/release/kyma-project/registry-cache)](https://github.com/kyma-project/registry-cache/releases/latest)

# Registry Cache

## Overview

The Registry Cache Kyma module adds a caching layer for container image registries in BTP-managed Kyma Runtimes. It reduces outbound traffic to public registries, improving performance and reliability of image pulls. It also supports access to private registries by allowing you to provide credentials for the caching layer to use when authenticating against those registries.

The Registry Cache feature is built on top of [Gardener's Registry Cache extension](https://gardener.cloud/docs/extensions/others/gardener-extension-registry-cache/registry-cache/configuration/).

## Prerequisites

- A managed Kyma Runtime instance running on the BTP platform.
- Administrative access to the Kyma Runtime with kubeconfig and the `kubectl` tool.
- The Registry Cache module installed on your Kyma Runtime cluster.

## Installation

For installation instructions, see [docs/user/README.md](docs/user/README.md).

## Usage

The Registry Cache module is configured using `RegistryCacheConfig` custom resources. Each resource defines a caching layer for a specific upstream container image registry.

For a full configuration guide including advanced options, credentials, and validation, see [docs/user/README.md](docs/user/README.md).

## Development

For developer setup, architecture overview, and testing strategy, see [docs/contributor/](docs/contributor/).

## Contributing
<!--- mandatory section - do not change this! --->

See the [Contributing Rules](CONTRIBUTING.md).

## Code of Conduct
<!--- mandatory section - do not change this! --->

See the [Code of Conduct](CODE_OF_CONDUCT.md) document.

## Licensing
<!--- mandatory section - do not change this! --->

See the [license](./LICENSE) file.
