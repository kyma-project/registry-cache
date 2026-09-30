# Rotating Credentials

Update the credentials used by Registry Cache to authenticate against a private upstream registry.

## Context

Credential Secrets are immutable and you cannot update them in place.

## Procedure

1. Create a new Secret with the updated credentials. Use a different name (for example, `rc-secret-v2`).

   ```bash
   kubectl create -f - <<EOF
   apiVersion: v1
   kind: Secret
   metadata:
     name: rc-secret-v2
     namespace: <namespace>
   type: Opaque
   immutable: true
   data:
     username: $(echo -n $USERNAME | base64 | tr -d '\n')
     password: $(echo -n $PASSWORD | base64 | tr -d '\n')
   EOF
   ```

2. To reference the new Secret, update **spec.secretReferenceName** in the existing `RegistryCacheConfig` resource.

   ```bash
   kubectl patch registrycacheconfig <name> -n <namespace> \
     --type=merge -p '{"spec":{"secretReferenceName":"rc-secret-v2"}}'
   ```

3. When the `RegistryCacheConfig` is in `Ready` state, verify that image pulls succeed with the new Secret, and delete the old Secret.

   ```bash
   kubectl delete secret rc-secret -n <namespace>
   ```
