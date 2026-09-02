# Registry Cache Does Not Cache Images from a Private Registry

## Symptom

You configured Registry Cache with credentials for a private upstream registry. Image pulls in your workloads still succeed, but you suspect or observe that images are not being served from the cache.

## Cause

When Registry Cache credentials are incorrect, the registry cache Pod in `kube-system` receives an error response from the upstream registry. Depending on the upstream registry, the error may be indistinguishable from a missing image in the logs — for example, JFrog Artifactory returns `404` instead of an authentication error. Meanwhile, image pulls in workloads continue to succeed using the `imagePullSecret` fallback, so the misconfiguration is not immediately visible.

> ### Note:
> Registry Cache is designed not to impair operations if its configuration is incorrect. If you have configured an `imagePullSecret` on your workloads (recommended), image pulls still succeed using direct fallback to the upstream registry even when the Registry Cache credentials are incorrect. This means misconfigured credentials are not immediately visible. To verify that the cache is working correctly, check the registry cache Pod logs as described in this document.

## Solution

The Gardener extension creates the registry cache Pods in `kube-system`. Each Pod is named after the upstream registry host whose images it caches.

1. List the Registry Cache Pods for the affected upstream.

   ```bash
   kubectl get pods -n kube-system | grep registry-<upstream-host>
   ```

2. Check the logs of the relevant Pod.

   ```bash
   kubectl logs -n kube-system <pod-name> --tail=50
   ```

   A pull failure due to incorrect credentials produces log output similar to the following:

   ```
   level=error msg="response completed with error" err.code="manifest unknown" err.detail="unknown tag=<tag>" err.message="manifest unknown" ... http.response.status=404
   ```

3. If you see this pattern repeating, verify that the credentials in the Secret referenced by **spec.secretReferenceName** are correct and up to date. To update the credentials, see [Rotating Credentials](../03-10-rotate-credentials.md).
