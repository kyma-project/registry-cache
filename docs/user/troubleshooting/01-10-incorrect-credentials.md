# Image Pulls Fail with "404 manifest unknown" Despite Correct Image Name

## Symptom

Image pulls from a cached upstream registry fail consistently with `404 manifest unknown` errors, even though the image exists in the upstream registry and the image name is correct.

> ### Note:
> Registry Cache is designed to not impair operations if its configuration is incorrect. If you have configured an `imagePullSecret` on your workloads (recommended), image pulls will still succeed via direct fallback to the upstream registry even when Registry Cache credentials are wrong. This means image pull failures may not be visible even with misconfigured credentials — the only way to verify the cache is working correctly is to check the registry cache Pod logs as described below.

## Cause

The upstream registry returns `404` instead of `401` when credentials are incorrect. This makes a credential failure indistinguishable from a missing image at the log level.

## Solution

The registry cache Pods are created in `kube-system` by the Gardener extension. They are named after the upstream registry host they cache.

1. List the registry cache Pods for the affected upstream:

   ```bash
   kubectl get pods -n kube-system | grep registry-<upstream-host>
   ```

2. Check the logs of the relevant Pod:

   ```bash
   kubectl logs -n kube-system <pod-name> --tail=50
   ```

3. A pull failure due to incorrect credentials looks similar to this one:

   ```
   level=error msg="response completed with error" err.code="manifest unknown" err.detail="unknown tag=<tag>" err.message="manifest unknown" ... http.response.status=404
   ```

4. If you see this pattern repeating, verify that the credentials in the referenced Secret are correct and that the Secret is up to date.
