# Registry Cache module user documentation

This document describes how to provide valid configuration of the Registry Cache for your Kyma Runtime cluster.

## Table of Contents
- [Introduction](#introduction)
- [Providing Basic Configuration](#basic-config.md)
- [Provide credentials for upstream repository](#upstream-credentials.md)
- [Advanced configuration](#advanced-config.md)
- [Validation of Registry Cache configuration](#validation.md)
- [Troubleshooting](#troubleshooting.md)

## Introduction
The Registry Cache Kyma module adds a possibility to enable and configure a caching layer for container image registries used in your BTP managed Kyma Runtimes.  
This feature reduces the amount of outbound traffic from your runtimes to public registries, improving performance and reliability of image pulls.
Additionally, it allows to configure access to private registries by providing credentials that will be used by the caching layer to authenticate against those registries.

## Prerequisites
- A managed Kyma Runtime instance running on BTP platform. 
- Administrative access to Kyma Runtime with kubeconfig and `kubectl` tool.
- Registry Cache module is installed on your Kyma Runtime cluster.

## Basic Configuration
To configure the Registry Cache for your Kyma Runtime cluster, create a custom resource of kind `RegistryCacheConfig`.

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

When this resource is applied to your Kyma Runtime cluster, the Kyma Control Plane service will process it and configure a caching layer for the specified upstream registry (in this case `docker.io`).  
The `volume.size` field specifies the size of the persistent volume that will be used to store cached images.

You can define multiple `RegistryCacheConfig` resources in your cluster to configure different caching providers for different upstream registries.    
Each create `RegistryCacheConfig` resource must have a unique name.
Every specified upstream registry must be unique across all `RegistryCacheConfig` resources in the cluster.

## Providing credentials for upstream repository

If the upstream registry requires authentication, you can provide credentials with a Kubernetes Secret in the same namespace as the `RegistryCacheConfig` resource and referencing it in the `spec.secretReferenceName` field.  
The referenced secret must be immutable and of type `generic`.

**Note:**
> The credential secret must exist on the cluster **before** applying the `RegistryCacheConfig` resource.

1.Create an immutable secret named `rc-secret` with username `admin` and password `admin` in the `test` namespace and export it to a YAML file:

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

For Artifact Registry, the username is _json_key and the password is the service account key in JSON format.   
To base64 encode the service account key, copy it and run:

```bash
echo -nE $SERVICE_ACCOUNT_KEY_JSON | base64 -w0
```
2.Apply Registry Cache configuration referencing the created secret:

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
### Advanced configuration

Following table describes all fields in the `RegistryCacheConfig` resource that can be used to customize the behavior of the Registry Cache:

| Field                          | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |   | Default value | Required |
|--------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---|---------------|----------|
| `spec.upstream`                | The URL of the upstream container image registry to cache images from.                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |   | N/A           | Yes      |
| `spec.RemoteURL`               | RemoteURL is the remote registry URL. If defined, the value is set as `proxy.remoteurl` in the registry [configuration](https://github.com/distribution/distribution/blob/main/docs/content/recipes/mirror.md#configure-the-cache); and in containerd configuration as `server` field in [hosts.toml](https://github.com/containerd/containerd/blob/main/docs/hosts.md#server-field) file. |   | N/A           | No       |
| `spec.secretReferenceName`     | The name of the Kubernetes Secret containing credentials for authenticating against the upstream registry.                                                                                                                                                                                                                                                                                                                                                                                                                                  |   | N/A           | No       |
| `spec.volume.size`             | The size of the persistent volume to use for storing cached images.                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |   | 10 Gi         | No       |
| `spec.volume.storageClassName` | The storage class name to use for the persistent volume. If not specified, the default storage class of the cluster will be used.                                                                                                                                                                                                                                                                                                                                                                                                           |   | N/A           | No       |
| `spec.garbageCollection.ttl`   | The time-to-live (TTL) duration for cached images. Images that have not been accessed within this duration will be eligible for garbage collection. Set to 0s to disable the garbage collection.                                                                                                                                                                                                                    |   | 168h (7 days) | No       |
| `spec.proxy.httpProxy`         | Proxy server for HTTP connections which is used by the registry cache                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |   | N/A           | No       |
| `spec.proxy.httpProxy`         | Proxy server for HTTPS connections which is used by the registry cache                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |   | N/A           | No       |
| `spec.http.tls`                | Indicates whether TLS is enabled for the HTTP server of the registry cache.                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |   | N/A           | No       |

## Validation of provided Registry Cache configuration

After applying the `RegistryCacheConfig` resource to your Kyma Runtime cluster, the registry cache webhook validates the configuration before it affects cluster configuration.  
If the configura the `RegistryCacheConfig` resource to your Kyma Runtime cluster, the registry cache webhook validates the configuration before it affects cluster configuration.  
If the configuration is valid, the resource status will be updated to `Ready` and the caching layer will be configured accordingly.
If there are any issues with the configuration, the status will be updated to `Error`, and an error message will be provided in the status conditions.

Example: invalid upstream URL

Example: 


You can validate if the configuration was applied successfully by checking the status of the resource.

| Field                          | Validation                                                                                                                                          |   | Example |
|--------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------|---|---------|
| `spec.upstream`                | URL must be be valid, resolvable with DNS and reachable via Internet                                                                                |   | N/A     |
| `spec.RemoteURL`               | String with the format of `<scheme><host>[:<port>]` where `<scheme>` is `https://` or `http://` and `<host>[:<port>]` corresponds to the Upstream.  |   | N/A     |
| `spec.secretReferenceName`     | Referenced secret name must exist in the same namespace as `RegistryCacheConfig`                                                                    |   | N/A     |
| `spec.volume.size`             | String in a format recognized by Go's resource.Quantity function (e.g. "10Gi")                                                                      |   | 10 Gi   |
| `spec.volume.storageClassName` | Referenced storage class must be available                                                                                                          |   | N/A     |
| `spec.garbageCollection.ttl`   | String in a format recognized by Go's `time.ParseDuration` function (e.g., "24h" for 24 hours). Set to 0s to disable the garbage collection.        |   | "168h"   |
| `spec.proxy.httpProxy`         | Proxy server for HTTP connections which is used by the registry cache                                                                               |   | N/A     |
| `spec.proxy.httpProxy`         | Proxy server for HTTPS connections which is used by the registry cache                                                                              |   | N/A     |
| `spec.http.tls`                | Indicates whether TLS is enabled for the HTTP server of the registry cache.                                                                         |   | N/A     |


## Troubleshooting

## Useful Links
- [Gardener Registry cache documentation](https://gardener.cloud/docs/extensions/others/gardener-extension-registry-cache/registry-cache/configuration/)
- [Gardener Registry Cache GitHub repository](https://github.com/gardener/gardener-extension-registry-cache/tree/main)


