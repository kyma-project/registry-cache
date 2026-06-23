# Registry Cache Module

Learn how to configure the Registry Cache module for your Kyma runtime cluster.


## What Is Registry Cache?
The Registry Cache Kyma module adds a caching layer for container image registries used in your SAP BTP, Kyma runtime instances.
This reduces outbound traffic to public registries, improving performance and reliability of image pulls.
It also supports access to private registries by allowing you to provide credentials for the caching layer to use when authenticating against those registries.


> ### Note:
> The Registry Cache module is built on top of [Gardener's Registry Cache extension](https://gardener.cloud/docs/extensions/others/gardener-extension-registry-cache/registry-cache/configuration/).

## Prerequisites
- A SAP BTP, Kyma runtime instance running on the BTP platform.
- Administrative access to the Kyma runtime with kubeconfig and the `kubectl` tool.
- The Registry Cache module installed on your Kyma cluster.

## Basic Configuration
`RegistryCacheConfig` is a namespace-scoped resource and can be created in any namespace.

To configure Registry Cache, create a `RegistryCacheConfig` custom resource (CR). The following example uses the `test` namespace — create it first if it doesn't exist:

```bash
kubectl create namespace test
```

```bash
kubectl create -f - <<EOF 
apiVersion: core.kyma-project.io/v1beta1
kind: RegistryCacheConfig
metadata:
  name: config1
  namespace: test
spec:
  upstream: docker.io
  volume:
    size: 100Gi
EOF
```

Once applied, Kyma Control Plane (KCP) processes the resource and configures a caching layer for the specified upstream registry (in this case, `docker.io`).
The **volume.size** field specifies the size of the persistent volume used to store cached images.

You can create multiple `RegistryCacheConfig` resources to cache different upstream registries. Each resource must have a unique name, and each upstream registry must be unique across all resources in the cluster.

## Providing Credentials for Upstream Repository

If the upstream registry requires authentication, create a Kubernetes Secret in the same namespace as the `RegistryCacheConfig` resource and reference it in the **spec.secretReferenceName** field.
The Secret must be immutable and of type `generic`.

> ### Note:
> The credential Secret must exist on the cluster **before** applying the `RegistryCacheConfig` resource.

1. Set environment variables with the upstream registry credentials:

```bash
export USERNAME=<your username>
export PASSWORD=<your password>
```

2. Create the namespace if it doesn't exist:

```bash
kubectl create namespace test
```

3. Create an immutable Secret named `rc-secret` in the `test` namespace:

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
  username: $(echo -n $USERNAME | base64 | tr -d '\n')
  password: $(echo -n $PASSWORD | base64 | tr -d '\n')
EOF
```

For [Google Artifact Registry](https://cloud.google.com/artifact-registry/docs/docker/authentication), use `_json_key` as the username and the service account key in JSON format as the password.
To base64-encode the service account key, run:

```bash
echo -nE $SERVICE_ACCOUNT_KEY_JSON | base64 | tr -d '\n'
```

4. Apply the Registry Cache configuration referencing the created Secret:

```bash
kubectl create -f - <<EOF
apiVersion: core.kyma-project.io/v1beta1
kind: RegistryCacheConfig
metadata:
  name: config2
  namespace: test
spec:
  upstream: <protected registry URL>
  secretReferenceName: rc-secret
  volume:
    size: 100Gi
EOF
```

> ### Note:
> When using a private registry, the same credentials must be stored in **two** Kubernetes Secrets:
> - The Secret referenced in **spec.secretReferenceName** — used by Registry Cache to authenticate against the upstream registry when pulling images to cache.
> - An `imagePullSecret` on each workload — used by containerd to authenticate directly against the upstream registry as a fallback when Registry Cache is unavailable.
>
> Do not remove the `imagePullSecret` from your workloads when configuring credentials for Registry Cache. If the cache is unavailable, containerd falls back to the upstream registry and requires the credentials directly.

## Advanced Configuration

The following table describes all fields in the `RegistryCacheConfig` resource:

| Field                              | Required | Description                                                                                                                                                                                                                                                                                                                                                                                    | Default value |
|------------------------------------|----------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------|
| **spec.upstream**                  | Yes      | The host (and optional port) of the upstream container image registry to cache images from. No scheme — e.g. `docker.io` or `my-registry.example.com:5000`.                                                                                                                                                                                                                                    | None          |
| **spec.remoteURL**                 | No       | The remote registry URL. If defined, it is set as `proxy.remoteurl` in the registry [configuration](https://github.com/distribution/distribution/blob/main/docs/content/recipes/mirror.md#configure-the-cache) and as the `server` field in the containerd [hosts.toml](https://github.com/containerd/containerd/blob/main/docs/hosts.md#server-field) file. Defaults to `https://<upstream>`. | None          |
| **spec.secretReferenceName**       | No       | The name of the Kubernetes Secret containing credentials for the upstream registry.                                                                                                                                                                                                                                                                                                            | None          |
| **spec.volume.size**               | No       | The size of the persistent volume for storing cached images.                                                                                                                                                                                                                                                                                                                                   | 10Gi          |
| **spec.volume.storageClassName**   | No       | The storage class for the persistent volume. If not specified, the cluster's default storage class is used.                                                                                                                                                                                                                                                                                    | None          |
| **spec.garbageCollection.ttl**     | No       | The time-to-live (TTL) for cached images. Images not accessed within this duration are eligible for garbage collection. Set to `0s` to disable garbage collection.                                                                                                                                                                                                                             | 168h (7 days) |
| **spec.proxy.httpProxy**           | No       | Proxy server for HTTP connections used by the Registry Cache.                                                                                                                                                                                                                                                                                                                                  | None          |
| **spec.proxy.httpsProxy**          | No       | Proxy server for HTTPS connections used by the Registry Cache.                                                                                                                                                                                                                                                                                                                                 | None          |
| **spec.http.tls**                  | No       | Indicates whether TLS is enabled for the HTTP server of the Registry Cache.                                                                                                                                                                                                                                                                                                                    | true          |

## Validation of Registry Cache Configuration

After applying the `RegistryCacheConfig` resource, the Registry Cache webhook validates the configuration before it takes effect.
If the configuration is valid, the resource status transitions from `Pending` to `Ready` and the caching layer is configured.
If there are issues, the status transitions from `Pending` to `Error` and an error message is provided in the status conditions.

Example error message:
```
admission webhook "registrycacheconfig-v1beta1.kb.io" denied the request: spec.upstream: Invalid value: "dockerrrrr.io": upstream is not DNS resolvable
```

The following table describes the validation rules for each field:

| Field                          | Validation                                                                                                                                         | Example |
|--------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------|--|
| `spec.upstream`                | Must be a valid DNS-resolvable host (no scheme). Must be unique across all `RegistryCacheConfig` resources in the cluster. Port, if specified, must be in the range 1–65535. | None |
| `spec.remoteURL`               | Must have the format `<scheme><host>[:<port>]` where `<scheme>` is `https://` or `http://` and `<host>[:<port>]` corresponds to the upstream. Must be DNS resolvable.         | None |
| `spec.secretReferenceName`     | The referenced secret must exist in the same namespace as the `RegistryCacheConfig` resource, be immutable, and contain exactly the `username` and `password` data keys.      | None |
| `spec.volume.size`             | Must be a positive value in a format recognized by Go's `resource.Quantity` (e.g. `10Gi`). Immutable after creation.                                                          | 10Gi |
| `spec.volume.storageClassName` | The referenced storage class must be available. Immutable after creation.                                                                                                     | None |
| `spec.garbageCollection.ttl`   | Must be in a format recognized by Go's `time.ParseDuration` (e.g. `24h`). Set to `0s` to disable garbage collection. Cannot be re-enabled once disabled.                     | 168h |
| `spec.proxy.httpProxy`         | Must be a valid URL starting with `http://` or `https://`.                                                                                                                    | None |
| `spec.proxy.httpsProxy`        | Must be a valid URL starting with `http://` or `https://`.                                                                                                                    | None |
| `spec.http.tls`                | Must be a valid boolean indicating whether TLS is enabled.                                                                                                                    | None |

## Managing Registry Cache Configuration

### Listing Registry Cache Configurations

To list all `RegistryCacheConfig` resources across all namespaces, run:

```bash
kubectl get registrycacheconfig -A
```

To list resources in a specific namespace, run:

```bash
kubectl get registrycacheconfig -n <namespace>
```

### Deleting a Registry Cache Configuration

To delete a `RegistryCacheConfig` resource, run:

```bash
kubectl delete registrycacheconfig <name> -n <namespace>
```

For example:

```bash
kubectl delete registrycacheconfig config -n test
```

## Troubleshooting

The Registry Cache configuration is validated before being applied to the cluster. Invalid configuration will be rejected by the webhook.
If the configuration is valid but the Registry Cache setup fails on the KCP side, the `RegistryCacheConfig` resource status transitions to `Error` with an error message in the status conditions. In this case, contact the Kyma support team for assistance.

### Diagnosing Incorrect Credentials

The upstream registry does not return an explicit authentication error when credentials are wrong. Instead, image pulls fail with `404 manifest unknown`, which is indistinguishable from a missing image at the log level.

If image pulls fail consistently with `404` errors and you know the image exists in the upstream registry, check the Registry Cache Pod logs for the affected upstream:

```bash
kubectl logs -n kube-system -l app=registry-cache --tail=50
```

To filter logs for a specific upstream, use the Pod name pattern (Pods are named after the upstream host):

```bash
kubectl logs -n kube-system $(kubectl get pods -n kube-system -o name | grep registry-<upstream-host>) --tail=50
```

A pull failure due to incorrect credentials looks similar to this one:

```
level=error msg="response completed with error" err.code="manifest unknown" err.detail="unknown tag=<tag>" err.message="manifest unknown" ... http.response.status=404
```

If you see this pattern repeating, verify that the credentials in the referenced secret are correct and that the secret is up to date.

