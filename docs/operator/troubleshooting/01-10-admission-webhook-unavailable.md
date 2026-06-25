# Admission Webhook Is Unavailable

## Symptom

New `RegistryCacheConfig` resources are rejected with an error such as:

```
Error from server (InternalError): error when creating "...": Internal error occurred:
failed calling webhook "registrycacheconfig-v1beta1.kb.io": ...
```

## Cause

The Registry Cache admission webhook server is unreachable or has not yet started.

## Solution

1. Check that the controller Pod is running:

   ```bash
   kubectl get pods -n kyma-system -l control-plane=controller-manager
   ```

2. Check the Pod logs for errors:

   ```bash
   kubectl logs -n kyma-system -l control-plane=controller-manager --tail=50
   ```

3. Verify the controller Deployment status:

   ```bash
   kubectl get deployment -n kyma-system -l control-plane=controller-manager
   ```
