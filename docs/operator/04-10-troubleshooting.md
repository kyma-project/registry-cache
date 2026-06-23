# Registry Cache - Troubleshooting

---

## Admission Webhook Is Unavailable

### Symptom

New `RegistryCacheConfig` resources are rejected with an error such as:

```
Error from server (InternalError): error when creating "...": Internal error occurred:
failed calling webhook "registrycacheconfig-v1beta1.kb.io": ...
```

### Cause

The Registry Cache admission webhook server is unreachable or has not yet started.

### Solution

1. Check that the controller pod is running:

   ```bash
   kubectl get pods -n kyma-system -l app=registry-cache
   ```

2. Check the pod logs for errors:

   ```bash
   kubectl logs -n kyma-system -l app=registry-cache --tail=50
   ```

3. Verify the health endpoint responds:

   ```bash
   kubectl get pod -n kyma-system -l app=registry-cache -o name \
     | xargs -I{} kubectl exec {} -n kyma-system -- wget -qO- http://localhost:8081/healthz
   ```

---

## Certificate Rotation Failure

### Symptom

Admission requests are rejected with a TLS or certificate error, or the `RegistryCache` CR transitions to `Error` state shortly after a certificate renewal.

### Cause

The `ValidatingWebhookConfiguration` CA bundle was not updated after the TLS certificate was rotated. The certificate manager retries with exponential backoff but may have exhausted retries due to persistent API server errors.

### Solution

1. Check the controller logs for certificate callback errors:

   ```bash
   kubectl logs -n kyma-system -l app=registry-cache --tail=100 | grep -i "cert\|certificate\|webhook"
   ```

2. If the CA bundle is stale, manually patch the `ValidatingWebhookConfiguration` with the current CA bundle:

   ```bash
   CA_BUNDLE=$(kubectl get secret <tls-secret-name> -n kyma-system -o jsonpath='{.data.ca\.crt}')
   kubectl patch validatingwebhookconfiguration registrycacheconfig-v1beta1 \
     --type='json' \
     -p="[{\"op\":\"replace\",\"path\":\"/webhooks/0/clientConfig/caBundle\",\"value\":\"${CA_BUNDLE}\"}]"
   ```

---

## RegistryCache CR Stuck in Error or Processing State

### Symptom

The `RegistryCache` CR remains in `Error` or `Processing` state for an extended period.

### Cause

The admission webhook server is not healthy, which prevents the controller from transitioning the CR to `Ready`.

### Solution

1. Inspect the status conditions of the `RegistryCache` CR for the error message:

   ```bash
   kubectl get registrycache -A -o yaml | grep -A 10 conditions
   ```

2. Check the controller pod health (see [Admission Webhook Is Unavailable](#admission-webhook-is-unavailable)).

3. If the pod is healthy but the state does not recover, restart the controller deployment:

   ```bash
   kubectl rollout restart deployment/registry-cache -n kyma-system
   ```

---

## Image Pulls Fail with "404 manifest unknown"

### Symptom

Image pulls from a cached upstream registry fail consistently with `404 manifest unknown` errors.

### Cause

The upstream registry returns `404` instead of `401` when credentials are incorrect. See also the [user troubleshooting guide](../user/troubleshooting/01-10-incorrect-credentials.md) for end-user steps.

### Solution

1. Check the registry cache pod logs for the affected upstream:

   ```bash
   kubectl logs -n kube-system -l app=registry-cache --tail=50
   ```

2. A pull failure due to incorrect credentials looks like:

   ```
   level=error msg="response completed with error" err.code="manifest unknown" err.detail="unknown tag=<tag>" err.message="manifest unknown" ... http.response.status=404
   ```

3. Verify the credentials in the secret referenced by `spec.secretReferenceName` are correct. Credential secrets are immutable — to update them, delete and recreate the secret with the corrected credentials.
