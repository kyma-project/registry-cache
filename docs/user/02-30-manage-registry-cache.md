# Managing Registry Cache Configuration

List and delete `RegistryCacheConfig` resources in your cluster.

## Procedure

- To list all `RegistryCacheConfig` resources across all namespaces, run:

    ```bash
    kubectl get registrycacheconfig -A
    ```

- To list resources in a specific namespace, run:

    ```bash
    kubectl get registrycacheconfig -n <namespace>
    ```

- To delete a `RegistryCacheConfig` resource, run:

    ```bash
    kubectl delete registrycacheconfig <name> -n <namespace>
    ```

    For example:

    ```bash
    kubectl delete registrycacheconfig config -n test
    ```
