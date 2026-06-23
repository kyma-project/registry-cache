# Registry Cache - Operations

## Custom Resources

For the full parameter reference for both CustomResourceDefinitions (CRDs), see the [user documentation](../user/resources/).

### RegistryCache

The `RegistryCache` custom resource (CR) is created and managed by the module lifecycle infrastructure. It tracks the health of the Registry Cache installation.

| Field | Description |
|---|---|
| `status.state` | Current installation state: `Processing`, `Ready`, `Warning`, `Error`, or `Deleting`. |
| `status.conditions` | Standard Kubernetes conditions. Condition type `Starting` tracks admission webhook health. |

To check the current state:

```bash
kubectl get registrycache -A
```

### RegistryCacheConfig

The `RegistryCacheConfig` custom resource (CR) is created by end users to configure a caching layer for a specific upstream registry. It is namespace-scoped.

To list all `RegistryCacheConfig` resources across all namespaces:

```bash
kubectl get registrycacheconfig -A
```

## Configuration

<!-- TODO: SME input required — confirm the exact deployment configuration values used in the operator environment -->

| Parameter | Value | Description |
|---|---|---|
| Webhook server port | `9443` | TLS port for the admission webhook server. |
| HTTP/2 | Disabled | HTTP/2 is disabled on the webhook TLS listener. |
| Health probe bind address | `:8081` | Port for `/healthz` and `/readyz` endpoints. |

## Health Endpoints

The controller-runtime manager exposes health and readiness probes. Both delegate to `webhook.StartedChecker()`, which reports healthy only when the admission webhook TLS server is accepting connections.

| Endpoint | Path | Expected response |
|---|---|---|
| Liveness | `/healthz` | `ok` |
| Readiness | `/readyz` | `ok` |

If the webhook is unavailable, both endpoints return an error until the webhook recovers.

## Metrics

The controller exposes standard `controller-runtime` Prometheus metrics at the `/metrics` endpoint:

- `controller_runtime_reconcile_total` — total reconcile calls per controller
- `controller_runtime_reconcile_errors_total` — total reconcile errors per controller
- `workqueue_depth` — current depth of the work queue

No custom metrics are defined by the Registry Cache module.
