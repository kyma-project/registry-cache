# Installing Registry Cache

## Prerequisites

<!-- TODO: SME input required — confirm required RBAC roles and access prerequisites -->

- Access to the Kyma runtime cluster with sufficient RBAC permissions to apply CustomResourceDefinitions (CRDs) and create cluster-scoped resources.
- `kubectl` configured to point to the target cluster.
- The Registry Cache module release artifacts from the [releases page](https://github.com/kyma-project/registry-cache/releases).

## Procedure

<!-- TODO: SME input required — confirm the exact installation steps for the BTP operator environment, including release artifact paths and any BTP-specific tooling or orchestration -->

1. Apply the CRD manifests:

   ```bash
   kubectl apply -f <release-artifacts>/crds/
   ```

2. Apply the controller deployment manifests:

   ```bash
   kubectl apply -f <release-artifacts>/registry-cache.yaml
   ```

3. Verify that the controller Pod is running:

   ```bash
   kubectl get pods -n kyma-system -l app=registry-cache
   ```

## Post-Installation Steps

1. Confirm that the `RegistryCache` custom resource (CR) reaches the `Ready` state:

   ```bash
   kubectl get registrycache -A
   ```

   The `STATE` column should show `Ready`.

2. Verify the health and readiness endpoints respond:

   ```bash
   kubectl get pod -n kyma-system -l app=registry-cache -o name \
     | xargs -I{} kubectl exec {} -n kyma-system -- wget -qO- http://localhost:8081/healthz
   kubectl get pod -n kyma-system -l app=registry-cache -o name \
     | xargs -I{} kubectl exec {} -n kyma-system -- wget -qO- http://localhost:8081/readyz
   ```

   Both endpoints should return `ok`.
