# RegistryCache CR Stuck in Error or Processing State

## Symptom

The `RegistryCache` CR remains in `Error` or `Processing` state for an extended period.

## Cause

The admission webhook server is not healthy, which prevents the controller from transitioning the CR to `Ready`.

## Solution

1. Inspect the status conditions of the `RegistryCache` CR for the error message:

   ```bash
   kubectl get registrycache -A -o yaml | grep -A 10 conditions
   ```

2. Check the controller Pod health (see [Admission Webhook Is Unavailable](01-10-admission-webhook-unavailable.md)).

3. If the Pod is healthy but the state does not recover, restart the controller Deployment:

   ```bash
   kubectl rollout restart deployment/registry-cache -n kyma-system
   ```
