# Certificate Rotation Failure

## Symptom

Admission requests are rejected with a TLS or certificate error, or the `RegistryCache` custom resource (CR) transitions to `Error` state shortly after a certificate renewal.

## Cause

The `ValidatingWebhookConfiguration` CA bundle was not updated after the TLS certificate was rotated. The certificate manager retries with exponential backoff but may have exhausted retries due to persistent API server errors.

## Solution

1. Check the controller logs for certificate callback errors:

   ```bash
   kubectl logs -n kyma-system -l control-plane=controller-manager --tail=100 | grep -i "cert\|certificate\|webhook"
   ```

2. If the CA bundle is stale, manually patch the `ValidatingWebhookConfiguration` with the current CA bundle:

   ```bash
   CA_BUNDLE=$(kubectl get secret <tls-secret-name> -n kyma-system -o jsonpath='{.data.ca\.crt}')
   kubectl patch validatingwebhookconfiguration registrycacheconfig-v1beta1 \
     --type='json' \
     -p="[{\"op\":\"replace\",\"path\":\"/webhooks/0/clientConfig/caBundle\",\"value\":\"${CA_BUNDLE}\"}]"
   ```
