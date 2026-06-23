# RegistryCacheConfig

The `registrycacheconfigs.core.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to configure a caching layer for a specific upstream container image registry. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```bash
kubectl get crd registrycacheconfigs.core.kyma-project.io -o yaml
```

## Overview

A `RegistryCacheConfig` resource is created by the user to configure a caching layer for one upstream container image registry. Each resource targets a single upstream (for example, `docker.io` or `my-registry.example.com:5000`). Multiple `RegistryCacheConfig` resources can coexist in a cluster, but each upstream must be unique across all resources.

The resource is namespace-scoped and can be created in any namespace.

## Sample Custom Resource

This is a sample `RegistryCacheConfig` resource that configures a cache for `docker.io` with a 100Gi volume and custom garbage collection TTL:

```yaml
apiVersion: core.kyma-project.io/v1beta1
kind: RegistryCacheConfig
metadata:
  name: docker-cache
  namespace: my-namespace
spec:
  upstream: docker.io
  volume:
    size: 100Gi
  garbageCollection:
    ttl: 72h
```

## Custom Resource Parameters

This table lists all the possible parameters of a `RegistryCacheConfig` resource together with their descriptions:

| Parameter | Required | Default | Description |
|---|:---:|---|---|
| **metadata.name** | Yes | — | Specifies the name of the CR. |
| **metadata.namespace** | Yes | — | The namespace in which the CR is created. |
| **spec.upstream** | Yes | — | The host (and optional port) of the upstream registry to cache. No scheme — for example, `docker.io` or `my-registry.example.com:5000`. Must be DNS-resolvable and unique across all `RegistryCacheConfig` resources in the cluster. |
| **spec.remoteURL** | No | `https://<upstream>` | The remote registry URL in `<scheme><host>[:<port>]` format, where `<scheme>` is `https://` or `http://`. If set, used as `proxy.remoteurl` in the registry configuration and as the `server` field in the containerd hosts.toml file. |
| **spec.secretReferenceName** | No | — | The name of a Kubernetes Secret in the same namespace containing credentials for the upstream registry. The secret must be immutable and contain exactly the `username` and `password` data keys. |
| **spec.volume.size** | No | `10Gi` | The size of the persistent volume for storing cached images. Immutable after creation. |
| **spec.volume.storageClassName** | No | cluster default | The storage class for the persistent volume. Immutable after creation. |
| **spec.garbageCollection.ttl** | No | `168h` | The time-to-live for cached images. Images not accessed within this duration are eligible for garbage collection. Set to `0s` to disable. Cannot be re-enabled once disabled. |
| **spec.proxy.httpProxy** | No | — | Proxy server URL for HTTP connections used by the registry cache. Must start with `http://` or `https://`. |
| **spec.proxy.httpsProxy** | No | — | Proxy server URL for HTTPS connections used by the registry cache. Must start with `http://` or `https://`. |
| **spec.http.tls** | No | `true` | Whether TLS is enabled for the HTTP server of the registry cache. |

## Status Fields

| Field | Description |
|---|---|
| **status.state** | Current state of the resource. See [State Values](#state-values). |
| **status.conditions** | A list of Kubernetes standard conditions. Condition types: `RegistryCacheValidated`, `RegistryCacheConfigured`. |

## State Values

| State | Description |
|---|---|
| `Pending` | The resource has been accepted; the Kyma Control Plane is processing the configuration. |
| `Ready` | The caching layer has been successfully configured for the upstream registry. |
| `Failed` | The configuration failed. Check `status.conditions` for the error message. |

## Related Resources and Components

These are the resources related to this CR:

| Custom resource | Description |
|---|---|
| `RegistryCache` | Module CR that tracks the overall installation health of the Registry Cache module. See [RegistryCache](RegistryCache.md). |

These components use this CR:

| Component | Description |
|---|---|
| `RegistryCacheConfig` webhook | Validates the CR on create and update before it is persisted. |
| Kyma Control Plane | Processes the CR and configures the caching layer on the target cluster. |
