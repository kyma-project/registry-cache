# Validation of Registry Cache Configuration

When you apply a `RegistryCacheConfig` resource, the Registry Cache webhook validates the configuration on the Kyma runtime side before the Kubernetes API accepts it. If the configuration is invalid, the API rejects the request and returns an error — no custom resource (CR) is created.

Example error message:
```
admission webhook "registrycacheconfig-v1beta1.kb.io" denied the request: spec.upstream: Invalid value: "dockerrrrr.io": upstream is not DNS resolvable
```

If the CR is accepted, KCP processes it. The status transitions from `Pending` to `Ready` on success, or to `Error` if KCP-side processing fails. For details, check `status.conditions`.

KCP periodically reconciles `RegistryCacheConfig` resources. During reconciliation, a CR in `Ready` state transitions back to `Pending` and then returns to `Ready` once reconciliation completes. This is expected behavior.

The following table describes the validation rules for each field:

| Field                            | Validation                                                                                                                                                                   |
|----------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **spec.upstream**                | Must be a valid DNS-resolvable host (no scheme). Must be unique across all `RegistryCacheConfig` resources in the cluster. Port, if specified, must be in the range 1–65535. |
| **spec.remoteURL**               | Must have the format `<scheme><host>[:<port>]` where `<scheme>` is `https://` or `http://` and `<host>[:<port>]` corresponds to the upstream. Must be DNS resolvable.        |
| **spec.secretReferenceName**     | The referenced Secret must exist in the same namespace as the `RegistryCacheConfig` resource, be immutable, and contain exactly the `username` and `password` data keys.     |
| **spec.volume.size**             | Must be a positive value in a format recognized by Go's `resource.Quantity` (for example, `10Gi`). Immutable after creation.                                                 |
| **spec.volume.storageClassName** | The referenced storage class must be available. Immutable after creation.                                                                                                    |
| **spec.garbageCollection.ttl**   | Must be in a format recognized by Go's `time.ParseDuration` (for example, `24h`). Set to `0s` to disable garbage collection. If disabled, you cannot be re-enable it.        |
| **spec.proxy.httpProxy**         | Must be a valid URL starting with `http://` or `https://`.                                                                                                                   |
| **spec.proxy.httpsProxy**        | Must be a valid URL starting with `http://` or `https://`.                                                                                                                   |
| **spec.http.tls**                | Must be a valid boolean indicating whether TLS is enabled.                                                                                                                   |
