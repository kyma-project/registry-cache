# Image Pulls Fail with "404 manifest unknown" Despite Correct Image Name

## Symptom

Image pulls from a cached upstream registry fail consistently with `404 manifest unknown` errors, even though the image exists in the upstream registry and the image name is correct.

## Cause

The upstream registry returns `404` instead of `401` when credentials are incorrect. This makes a credential failure indistinguishable from a missing image at the log level.

## Solution

1. Check the Registry Cache Pod logs for the affected upstream:

   ```bash
   kubectl logs -n kube-system -l app=registry-cache --tail=50
   ```

2. To filter logs for a specific upstream, use the Pod name pattern (Pods are named after the upstream host):

   ```bash
   kubectl logs -n kube-system $(kubectl get pods -n kube-system -o name | grep registry-<upstream-host>) --tail=50
   ```

3. A pull failure due to incorrect credentials looks similar to this one:

   ```
   level=error msg="response completed with error" err.code="manifest unknown" err.detail="unknown tag=<tag>" err.message="manifest unknown" ... http.response.status=404
   ```

4. If you see this pattern repeating, verify that the credentials in the referenced Secret are correct and that the Secret is up to date.
