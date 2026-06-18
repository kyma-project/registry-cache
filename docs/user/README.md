# Registry Cache Module User Documentation

This document describes how to configure the Registry Cache for your Kyma Runtime cluster.

## Table of Contents
- [Introduction](#introduction)
- [Providing Basic Configuration](#basic-config.md)
- [Providing Credentials for Upstream Repository](#upstream-credentials.md)
- [Advanced Configuration](#advanced-config.md)
- [Validation of Registry Cache Configuration](#validation.md)
- [Troubleshooting](#troubleshooting.md)

## Introduction
The Registry Cache Kyma module adds a caching layer for container image registries used in your BTP managed Kyma Runtimes.
This reduces outbound traffic to public registries, improving performance and reliability of image pulls.
It also supports access to private registries by allowing you to provide credentials for the caching layer to use when authenticating against those registries.

## Prerequisites
- A managed Kyma Runtime instance running on the BTP platform.
- Administrative access to the Kyma Runtime with kubeconfig and the `kubectl` tool.
- The Registry Cache module installed on your Kyma Runtime cluster.

## Basic Configuration
`RegistryCacheConfig` is a namespace-scoped resource and can be created in any namespace.

To configure the Registry Cache, create a `RegistryCacheConfig` custom resource. The following example uses the `test` namespace — create it first if it doesn't exist:

```bash
kubectl create namespace test
```

```bash
kubectl create -f - <<EOF 
apiVersion: core.kyma-project.io/v1beta1
kind: RegistryCacheConfig
metadata:
  name: config
  namespace: test
spec:
  upstream: docker.io
  volume:
    size: 100Gi
EOF
```

Once applied, the Kyma Control Plane processes the resource and configures a caching layer for the specified upstream registry (in this case `docker.io`).
The `volume.size` field specifies the size of the persistent volume used to store cached images.

You can create multiple `RegistryCacheConfig` resources to cache different upstream registries. Each resource must have a unique name, and each upstream registry must be unique across all resources in the cluster.

## Providing Credentials for Upstream Repository

If the upstream registry requires authentication, create a Kubernetes Secret in the same namespace as the `RegistryCacheConfig` resource and reference it in the `spec.secretReferenceName` field.
The secret must be immutable and of type `generic`.

**Note:**
> The credential secret must exist on the cluster **before** applying the `RegistryCacheConfig` resource.

1. Set environment variables with the upstream registry credentials:

```bash
export USERNAME=<your username>
export PASSWORD=<your password>
```

2. Create the namespace if it doesn't exist:

```bash
kubectl create namespace test
```

3. Create an immutable secret named `rc-secret` in the `test` namespace:

```bash
kubectl create -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: rc-secret
  namespace: test
type: Opaque
immutable: true
data:
  username: $(echo -n $USERNAME | base64 -w0)
  password: $(echo -n $PASSWORD | base64 -w0)
EOF
```

For Artifact Registry, use `_json_key` as the username and the service account key in JSON format as the password.
To base64-encode the service account key, run:

```bash
echo -nE $SERVICE_ACCOUNT_KEY_JSON | base64 -w0
```

4. Apply the Registry Cache configuration referencing the created secret:

```bash
kubectl create -f - <<EOF
apiVersion: core.kyma-project.io/v1beta1
kind: RegistryCacheConfig
metadata:
  name: config
  namespace: test
spec:
  upstream: docker.io
  secretReferenceName: rc-secret
  volume:
    size: 100Gi
EOF
```

## Advanced Configuration

The following table describes all fields in the `RegistryCacheConfig` resource:

| Field                          | Description                                                                                                                                                                                                                                                                                                                                                                                                            | Default value | Required |
|--------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------|----------|
| `spec.upstream`                | The host (and optional port) of the upstream container image registry to cache images from. No scheme — e.g. `docker.io` or `my-registry.example.com:5000`.                                                                                                                                                                                                                                                            | N/A           | Yes      |
| `spec.remoteURL`               | The remote registry URL. If defined, it is set as `proxy.remoteurl` in the registry [configuration](https://github.com/distribution/distribution/blob/main/docs/content/recipes/mirror.md#configure-the-cache) and as the `server` field in the containerd [hosts.toml](https://github.com/containerd/containerd/blob/main/docs/hosts.md#server-field) file. Defaults to `https://<upstream>`. | N/A           | No       |
| `spec.secretReferenceName`     | The name of the Kubernetes Secret containing credentials for the upstream registry.                                                                                                                                                                                                                                                                                                                                    | N/A           | No       |
| `spec.volume.size`             | The size of the persistent volume for storing cached images.                                                                                                                                                                                                                                                                                                                                                           | 10Gi          | No       |
| `spec.volume.storageClassName` | The storage class for the persistent volume. If not specified, the cluster's default storage class is used.                                                                                                                                                                                                                                                                                                            | N/A           | No       |
| `spec.garbageCollection.ttl`   | The time-to-live (TTL) for cached images. Images not accessed within this duration are eligible for garbage collection. Set to `0s` to disable garbage collection.                                                                                                                                                                                                                                                     | 168h (7 days) | No       |
| `spec.proxy.httpProxy`         | Proxy server for HTTP connections used by the registry cache.                                                                                                                                                                                                                                                                                                                                                          | N/A           | No       |
| `spec.proxy.httpsProxy`        | Proxy server for HTTPS connections used by the registry cache.                                                                                                                                                                                                                                                                                                                                                         | N/A           | No       |
| `spec.http.tls`                | Indicates whether TLS is enabled for the HTTP server of the registry cache.                                                                                                                                                                                                                                                                                                                                            | true          | No       |

## Validation of Registry Cache Configuration

After applying the `RegistryCacheConfig` resource, the registry cache webhook validates the configuration before it takes effect.
If the configuration is valid, the resource status is updated to `Ready` and the caching layer is configured.
If there are issues, the status is updated to `Error` and an error message is provided in the status conditions.

Example error message:
```
admission webhook "registrycacheconfig-v1beta1.kb.io" denied the request: spec.upstream: Invalid value: "dockerrrrr.io": upstream is not DNS resolvable
```

The following table describes the validation rules for each field:

| Field                          | Validation                                                                                                                                         | Example |
|--------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------|---------|
| `spec.upstream`                | Must be a valid DNS-resolvable host (no scheme). Must be unique across all `RegistryCacheConfig` resources in the cluster. Port, if specified, must be in the range 1–65535. | N/A     |
| `spec.remoteURL`               | Must have the format `<scheme><host>[:<port>]` where `<scheme>` is `https://` or `http://` and `<host>[:<port>]` corresponds to the upstream. Must be DNS resolvable.         | N/A     |
| `spec.secretReferenceName`     | The referenced secret must exist in the same namespace as the `RegistryCacheConfig` resource, be immutable, and contain exactly the `username` and `password` data keys.      | N/A     |
| `spec.volume.size`             | Must be a positive value in a format recognized by Go's `resource.Quantity` (e.g. `10Gi`). Immutable after creation.                                                          | 10Gi    |
| `spec.volume.storageClassName` | The referenced storage class must be available. Immutable after creation.                                                                                                     | N/A     |
| `spec.garbageCollection.ttl`   | Must be in a format recognized by Go's `time.ParseDuration` (e.g. `24h`). Set to `0s` to disable garbage collection. Cannot be re-enabled once disabled.                     | 168h    |
| `spec.proxy.httpProxy`         | Must be a valid URL starting with `http://` or `https://`.                                                                                                                    | N/A     |
| `spec.proxy.httpsProxy`        | Must be a valid URL starting with `http://` or `https://`.                                                                                                                    | N/A     |
| `spec.http.tls`                | Must be a valid boolean indicating whether TLS is enabled.                                                                                                                    | N/A     |

## Troubleshooting

The Registry Cache configuration is validated before being applied to the cluster. Invalid configuration will be rejected by the webhook.
If the configuration is valid but the Registry Cache setup fails on the KCP side, the `RegistryCacheConfig` resource status is updated to `Error` with an error message in the status conditions. In this case, contact the Kyma support team for assistance.

## Useful Links
- [Gardener Registry Cache documentation](https://gardener.cloud/docs/extensions/others/gardener-extension-registry-cache/registry-cache/configuration/)
- [Gardener Registry Cache GitHub repository](https://github.com/gardener/gardener-extension-registry-cache/tree/main)
