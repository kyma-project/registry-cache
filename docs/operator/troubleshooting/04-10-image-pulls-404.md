# Image Pulls Fail with "404 manifest unknown"

## Symptom

Image pulls from a cached upstream registry fail consistently with `404 manifest unknown` errors.

## Cause

The upstream registry returns `404` instead of `401` when credentials are incorrect. See also the [user troubleshooting guide](../../user/troubleshooting/01-10-incorrect-credentials.md) for end-user steps.

## Solution

1. Check the Registry Cache Pod logs for the affected upstream:

   ```bash
   kubectl logs -n kube-system -l app=registry-cache --tail=50
   ```

2. A pull failure due to incorrect credentials looks like:

   ```
   level=error msg="response completed with error" err.code="manifest unknown" err.detail="unknown tag=<tag>" err.message="manifest unknown" ... http.response.status=404
   ```

3. Verify the credentials in the Secret referenced by `spec.secretReferenceName` are correct. Credential Secrets are immutable — to update them, delete and recreate the Secret with the corrected credentials.
