# Configure Registry Cache

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

After creating a `RegistryCacheConfig` resource, verify that the configuration was processed successfully by checking the resource status:

```bash
kubectl get registrycacheconfig <name> -n <namespace>
```

The `STATUS` column shows the current state:
- `Pending` — KCP is processing the configuration.
- `Ready` — the caching layer has been configured successfully.
- `Error` — the configuration failed. Check `status.conditions` for details, or see [RegistryCacheConfig](resources/RegistryCacheConfig.md#state-values).

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

   For [Google Artifact Registry](https://cloud.google.com/artifact-registry/docs/docker/authentication), the username is `_json_key` and the password is the service account key in JSON format. Follow steps 3a–3b instead of step 3 above.

   3a. Base64-encode the service account key:

      ```bash
      export PASSWORD=$(echo -nE $SERVICE_ACCOUNT_KEY_JSON | base64 | tr -d '\n')
      ```

   3b. Create an immutable Secret with the encoded key as the password:

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
     username: $(echo -n "_json_key" | base64 | tr -d '\n')
     password: $PASSWORD
   EOF
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

## Rotating Credentials

Credential Secrets are immutable and cannot be updated in place. To rotate credentials:

1. Create a new Secret with the updated credentials. Use a different name (for example, `rc-secret-v2`):

   ```bash
   kubectl create -f - <<EOF
   apiVersion: v1
   kind: Secret
   metadata:
     name: rc-secret-v2
     namespace: <namespace>
   type: Opaque
   immutable: true
   data:
     username: $(echo -n $USERNAME | base64 | tr -d '\n')
     password: $(echo -n $PASSWORD | base64 | tr -d '\n')
   EOF
   ```

2. Delete the existing `RegistryCacheConfig` resource:

   ```bash
   kubectl delete registrycacheconfig <name> -n <namespace>
   ```

3. Recreate the `RegistryCacheConfig` resource referencing the new Secret:

   ```bash
   kubectl create -f - <<EOF
   apiVersion: core.kyma-project.io/v1beta1
   kind: RegistryCacheConfig
   metadata:
     name: <name>
     namespace: <namespace>
   spec:
     upstream: <upstream>
     secretReferenceName: rc-secret-v2
     volume:
       size: <size>
   EOF
   ```

4. Once the new `RegistryCacheConfig` is in `Ready` state, delete the old Secret:

   ```bash
   kubectl delete secret rc-secret -n <namespace>
   ```

## Advanced Configuration

For all available configuration fields and their defaults, see [RegistryCacheConfig](resources/RegistryCacheConfig.md).

## Validation of Registry Cache Configuration

After applying the `RegistryCacheConfig` resource, the Registry Cache webhook validates the configuration before it takes effect.
If the configuration is valid, the resource status transitions from `Pending` to `Ready` and the caching layer is configured.
If there are issues, the status transitions from `Pending` to `Error` and an error message is provided in the status conditions.

Example error message:
```
admission webhook "registrycacheconfig-v1beta1.kb.io" denied the request: spec.upstream: Invalid value: "dockerrrrr.io": upstream is not DNS resolvable
```

The following table describes the validation rules for each field:

| Field | Validation | Example |
|---|---|---|
| **spec.upstream** | Must be a valid DNS-resolvable host (no scheme). Must be unique across all `RegistryCacheConfig` resources in the cluster. Port, if specified, must be in the range 1–65535. | None |
| **spec.remoteURL** | Must have the format `<scheme><host>[:<port>]` where `<scheme>` is `https://` or `http://` and `<host>[:<port>]` corresponds to the upstream. Must be DNS resolvable. | None |
| **spec.secretReferenceName** | The referenced Secret must exist in the same namespace as the `RegistryCacheConfig` resource, be immutable, and contain exactly the `username` and `password` data keys. | None |
| **spec.volume.size** | Must be a positive value in a format recognized by Go's `resource.Quantity` (for example, `10Gi`). Immutable after creation. | 10Gi |
| **spec.volume.storageClassName** | The referenced storage class must be available. Immutable after creation. | None |
| **spec.garbageCollection.ttl** | Must be in a format recognized by Go's `time.ParseDuration` (for example, `24h`). Set to `0s` to disable garbage collection. Cannot be re-enabled once disabled. | 168h |
| **spec.proxy.httpProxy** | Must be a valid URL starting with `http://` or `https://`. | None |
| **spec.proxy.httpsProxy** | Must be a valid URL starting with `http://` or `https://`. | None |
| **spec.http.tls** | Must be a valid boolean indicating whether TLS is enabled. | None |

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
