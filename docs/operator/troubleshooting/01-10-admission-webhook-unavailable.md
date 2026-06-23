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
   kubectl get pods -n kyma-system -l app=registry-cache
   ```

2. Check the Pod logs for errors:

   ```bash
   kubectl logs -n kyma-system -l app=registry-cache --tail=50
   ```

3. Verify the health endpoint responds:

   ```bash
   kubectl get pod -n kyma-system -l app=registry-cache -o name \
     | xargs -I{} kubectl exec {} -n kyma-system -- wget -qO- http://localhost:8081/healthz
   ```
