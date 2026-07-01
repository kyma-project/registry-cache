# RegistryCache

The `registrycaches.core.kyma-project.io` CustomResourceDefinition (CRD) is a detailed description of the kind of data and the format used to track the installation state of the Registry Cache module on a Kyma runtime cluster. To get the up-to-date CRD and show the output in the `yaml` format, run this command:

```bash
kubectl get crd registrycaches.core.kyma-project.io -o yaml
```

## Overview

Kyma Lifecycle Manager (KLM) creates this resource automatically when you install the Registry Cache module on a Kyma runtime cluster. You do not create or delete this resource directly — the module lifecycle infrastructure manages it.

The `RegistryCache` custom resource (CR) tracks whether the Registry Cache admission webhook is healthy and the module is fully operational. The controller reconciles this resource and transitions it through a set of well-defined states.

## Sample Custom Resource

This is a sample `RegistryCache` resource in the `Ready` state:

```yaml
apiVersion: core.kyma-project.io/v1beta1
kind: RegistryCache
metadata:
  name: registry-cache
  namespace: kyma-system
  finalizers:
    - registry-cache.kyma-project.io/finalizer
status:
  state: Ready
  conditions:
    - type: Starting
      status: "True"
      reason: Ready
      message: Starting module
      observedGeneration: 1
```

## Custom Resource Parameters

This table lists all the parameters of a `RegistryCache` resource together with their descriptions:

| Parameter | Required | Description |
|---|:---:|---|
| **metadata.name** | Yes | Specifies the name of the CR. |
| **metadata.namespace** | Yes | The namespace in which the CR is created. |
| **spec** | No | Empty — the `RegistryCache` CR has no configurable spec fields. All configuration is done through `RegistryCacheConfig` resources. |

## Status Fields

| Field | Description |
|---|---|
| **status.state** | The current state of the Registry Cache module. See [State Lifecycle](#state-lifecycle). |
| **status.conditions** | A list of Kubernetes standard conditions. The condition type `Starting` reports the health of the admission webhook server. |

## State Lifecycle

The controller transitions the `RegistryCache` CR through the following states:

| State | Description |
|---|---|
| _(empty)_ | Initial state — the resource has just been created and has not yet been processed. |
| `Processing` | The controller is checking whether the admission webhook is ready. |
| `Ready` | The admission webhook is healthy and the module is fully operational. |
| `Error` | The admission webhook is unhealthy. The controller re-checks at the health interval (every 30 seconds). |
| `Deleting` | A deletion timestamp was set on the resource; the controller is removing the finalizer. |

The normal lifecycle is: _(empty)_ → `Processing` → `Ready`.

If the webhook becomes unhealthy while in `Ready`, the state transitions to `Error`. The controller retries every 30 seconds and returns to `Ready` once the webhook is healthy again.

## Related Resources and Components

These are the resources related to this CR:

| Custom resource | Description |
|---|---|
| `RegistryCacheConfig` | The user-facing resource for configuring a caching layer for a specific upstream registry. See [Registry Cache Module](../README.md). |

These components use this CR:

| Component | Description |
|---|---|
| Registry Cache controller | Reconciles `RegistryCache` CRs and drives status transitions. |
| Kyma Lifecycle Manager (KLM) | KCP component that creates and deletes the `RegistryCache` CR as part of module installation and removal. |
