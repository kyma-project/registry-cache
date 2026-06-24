# Architecture

For the architecture diagram and component overview, see [Architecture](../user/README.md#architecture) in the user documentation.

## Components and Code Paths

| Component | Package | Responsibility |
|---|---|---|
| `RegistryCacheReconciler` | `internal/controller` | Reconciles `RegistryCache` CRs; drives status transitions (Processing → Ready / Error / Deleting) with 5s requeue on transitions and 30s on health checks |
| Webhook Server | `internal/webhook/server` | TLS server (port 9443) for admission webhooks; exposes `StartedChecker` for health probing |
| `RegistryCacheConfig` Webhook | `internal/webhook/v1beta1` | Validates `RegistryCacheConfig` resources on create and update |
| Validation Framework | `internal/webhook/validations` | Pluggable validation chain: DNS resolution, upstream uniqueness, secret existence and format |
| Certificate Manager | `internal/webhook/certificate` | Watches TLS cert files; rotates the CA bundle in `ValidatingWebhookConfiguration` on cert renewal |
| HTTP Server | `internal/httpserver` | Underlying HTTP server used by the webhook multiplexer |

## Health Probes

Both `/healthz` and `/readyz` endpoints delegate to `webhook.StartedChecker()`. The manager reports healthy only when the admission webhook TLS server is accepting connections.

If the webhook becomes unavailable, the controller detects this on the next health-check reconcile (every 30s) and transitions the `RegistryCache` custom resource (CR) to `Error`.

## Certificate Rotation

`internal/webhook/certificate/callback.go` watches the webhook's TLS certificate files on disk. When a renewal is detected, it patches the `ValidatingWebhookConfiguration` with the updated CA bundle using a retry loop with exponential backoff.

## Key Implementation Patterns

- Server-Side Apply (SSA): Status updates use `client.Apply` with field owner `registry-cache.kyma-project.io/owner` to avoid conflicts.
- Finalizer: `registry-cache.kyma-project.io/finalizer` ensures cleanup logic runs before the `RegistryCache` CR is removed from the API server.
- Graceful shutdown: Both the webhook server and the certificate watcher respect context cancellation on SIGTERM.
