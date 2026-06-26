# Registry Cache Module

## What Is Registry Cache?

The Registry Cache Kyma module adds a caching layer for container image registries in SAP BTP, Kyma runtime instances. It reduces outbound traffic to public registries, improving performance and reliability of image pulls. It also supports access to private registries by allowing you to provide credentials for the caching layer to use when authenticating against those registries.

The Registry Cache feature is built on top of [Gardener's Registry Cache extension](https://gardener.cloud/docs/extensions/others/gardener-extension-registry-cache/registry-cache/configuration/).

## Features

- Caches container images from upstream registries to reduce outbound network traffic.
- Supports private registries via credential Secrets referenced in `RegistryCacheConfig`.
- Configurable cache volume size and storage class per upstream registry.
- Configurable garbage collection TTL; garbage collection can be disabled.
- Proxy support for HTTP and HTTPS connections used by the cache.
- TLS-enabled HTTP server for the registry cache endpoint.

## Architecture

The Registry Cache module consists of two main runtime components: the **RegistryCache controller** and the **RegistryCacheConfig admission webhook**. Both run in the same Registry Cache Manager process.

```
┌─────────────────────────────────────────────────────-──-──┐
│                  Registry Cache Manager                   │
│                                                           │
│  ┌──────────────────────┐   ┌──────────────────────────┐  │
│  │ RegistryCache        │──►│  Webhook Server          │  │
│  │ Reconciler           │   │  (TLS :9443)             │  │
│  │                      │◄──│                          │  │
│  │  state machine:      │   │  RegistryCacheConfig     │  │
│  │  ─ → Processing      │   │  Webhook (validate)      │  │
│  │      → Ready         │   │                          │  │
│  │      → Error         │   │  cert renewal            │  │
│  │      → Deleting      │   └──────────┬───────────────┘  │
│  └──────────────────────┘              │                  │
│                                        ▼                  │
│                             ┌─────────────────────────-─┐ │
│                             │  Certificate Manager     -│ │
│                             │  patches ValidatingWebhook│ │
│                             │  Configuration CA bundle  │ │
│                             └──────────────────────────-┘ │
│                                                           │
│  /healthz  /readyz  ──► webhook.StartedChecker()          │
└───────────────────────────────────────────────────────--──┘
```

- **RegistryCache controller** — reconciles `RegistryCache` custom resources (CRs) and manages the installation state machine (see table below).
- **Webhook Server** — TLS server on port 9443 that validates `RegistryCacheConfig` resources on create and update.
- **Certificate Manager** — watches TLS certificate files and rotates the CA bundle in `ValidatingWebhookConfiguration` on renewal.

### RegistryCache State Machine

The controller drives the `RegistryCache` CR through the following states:

| Current state       | Condition              | Next state           |
|---------------------|------------------------|----------------------|
| _(empty)_           | Resource just created  | `Processing`         |
| `Processing`        | Webhook healthy        | `Ready`              |
| `Processing`        | Webhook not healthy    | `Processing` (retry) |
| `Ready`             | Webhook healthy        | `Ready` (no change)  |
| `Ready`             | Webhook not healthy    | `Error`              |
| `Error`             | Webhook healthy        | `Ready`              |
| `Error`             | Webhook not healthy    | `Error` (retry)      |
| Any                 | Deletion timestamp set | `Deleting`           |
| `Deleting`          | Finalizer removed      | _(resource gone)_    |

## API / Custom Resource Definitions

The Registry Cache module defines two custom resources:

| CRD                                                       | Scope      | Description                                                                                                     |
|-----------------------------------------------------------|------------|-----------------------------------------------------------------------------------------------------------------|
| [`RegistryCache`](resources/RegistryCache.md)             | Namespaced | Module CR managed by the lifecycle infrastructure. Tracks the installation health of the Registry Cache module. |
| [`RegistryCacheConfig`](resources/RegistryCacheConfig.md) | Namespaced | User-created CR that configures a caching layer for a specific upstream container image registry.               |

## Authorization

<!-- TODO: Update when https://github.com/kyma-project/registry-cache/issues/77 is implemented -->

## Resource Consumption

<!-- TODO: link to SAP Help Portal Sizing topic when available -->
