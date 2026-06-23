# Troubleshooting

The Registry Cache configuration is validated before being applied to the cluster. Invalid configuration will be rejected by the webhook.
If the configuration is valid but the Registry Cache setup fails on the KCP side, the `RegistryCacheConfig` resource status transitions to `Error` with an error message in the status conditions. In this case, contact the Kyma support team for assistance.

For specific issues, see the guides below.
