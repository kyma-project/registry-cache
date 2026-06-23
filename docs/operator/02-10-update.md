# Updating Registry Cache

## Prerequisites

<!-- TODO: SME input required — confirm any prerequisites or approval steps required before updating -->

- Access to the Kyma runtime cluster with sufficient RBAC permissions.
- `kubectl` configured to point to the target cluster.
- The new release artifacts from the [releases page](https://github.com/kyma-project/registry-cache/releases).

## Procedure

<!-- TODO: SME input required — confirm the exact update procedure for the BTP operator environment, including release artifact paths -->

1. Apply the updated CRD manifests:

   ```bash
   kubectl apply -f <new-release-artifacts>/crds/
   ```

2. Apply the updated controller deployment manifests:

   ```bash
   kubectl apply -f <new-release-artifacts>/registry-cache.yaml
   ```

   The controller performs a rolling update automatically.

3. Monitor the rollout:

   ```bash
   kubectl rollout status deployment/registry-cache -n kyma-system
   ```

## Post-Update Steps

1. Confirm that the `RegistryCache` CR returns to `Ready` state:

   ```bash
   kubectl get registrycache -A
   ```

2. Check the status conditions for any warnings:

   ```bash
   kubectl get registrycache -A -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.status.state}{"\t"}{.status.conditions[*].message}{"\n"}{end}'
   ```

## What's Changed

<!-- Update this section for each release with notable behavioral or configuration changes. -->

See the [release notes](https://github.com/kyma-project/registry-cache/releases) for changes in each version.
