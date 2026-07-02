# Troubleshooting

The webhook validates the Registry Cache configuration before applying it to the cluster, and rejects an invalid configuration.
If the configuration is valid but the Registry Cache setup fails on the Kyma Control Plane (KCP) side, the `RegistryCacheConfig` resource status transitions to `Error` with an error message in the status conditions. In this case, contact the Kyma support team for assistance.

For specific issues, see the relevant guides.
