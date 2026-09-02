# Configuring Registry Cache

Create a RegistryCacheConfig custom resource (CR) to enable caching for an upstream container image registry in your SAP BTP, Kyma runtime instance.

## Prerequisites

- You have a Kyma instance created.
- You have administrative access to the Kyma runtime with kubeconfig and kubectl. See [Access a Kyma Instance Using kubectl](https://help.sap.com/docs/btp/sap-business-technology-platform/access-kyma-instance-using-kubectl?locale=en-US&version=Cloud&ai=true).
- You have the Registry Cache module added. See [Quick Install](https://kyma-project.io/02-get-started/01-quick-install.html).

## Context

`RegistryCacheConfig` is a namespace-scoped resource. You can create it in any namespace.

## Procedure

1. Create the `test` namespace if it doesn't exist.

    ```bash
    kubectl create namespace test
    ```

2. Create a `RegistryCacheConfig` CR.

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

    When applied, Kyma Control Plane (KCP) processes the resource and configures a caching layer for the specified upstream registry (in this case, `docker.io`).
    The **volume.size** field specifies the size of the persistent volume used to store cached images.

    You can create multiple `RegistryCacheConfig` resources to cache different upstream registries. Each resource must have a unique name, and each upstream registry must be unique across all resources in the cluster.

3. To verify that KCP processed it successfully, check the resource status.

    ```bash
    kubectl get registrycacheconfig <name> -n <namespace> -o jsonpath='{.status.state}'
    ```

    The expected output values are the following:
    - `Pending` — KCP is processing the configuration.
    - `Ready` — the caching layer has been configured successfully.
    - `Error` — KCP encountered an issue and is retrying. The state transitions to Ready automatically when processing succeeds.

## Related Information

[RegistryCacheConfig Custom Resource](resources/registry_cache_config_cr.md).
