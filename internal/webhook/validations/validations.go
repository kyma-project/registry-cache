package validations

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"net"
	"net/url"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"

	registrycacheext "github.com/gardener/gardener-extension-registry-cache/pkg/apis/registry"
	registrycacheextvalidations "github.com/gardener/gardener-extension-registry-cache/pkg/apis/registry/validation"
	registrycache "github.com/kyma-project/registry-cache/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// Validator validates RegistryCacheConfig resources.
type Validator struct {
	dnsValidator  DNSValidator
	runtimeClient client.Client
}

// NewValidator constructs a Validator with provided secrets and existing configs.
func NewValidator(dnsValidator DNSValidator, runtimeClient client.Client) Validator {
	return Validator{
		dnsValidator:  dnsValidator,
		runtimeClient: runtimeClient,
	}
}

func (v Validator) Do(newConfig *registrycache.RegistryCacheConfig) field.ErrorList {
	if isEmptySpec(newConfig.Spec) {
		return field.ErrorList{field.Required(field.NewPath("spec"), "spec must not be empty")}
	}

	allErrs := v.validateCommon(newConfig)

	gardenerValidations := registrycacheextvalidations.ValidateRegistryConfig(toExtensionConfig(*newConfig), field.NewPath("spec"))

	return append(allErrs, transformFieldErrors(gardenerValidations)...)
}

func (v Validator) DoOnUpdate(newConfig, oldConfig *registrycache.RegistryCacheConfig) field.ErrorList {
	if isEmptySpec(newConfig.Spec) {
		return field.ErrorList{field.Required(field.NewPath("spec"), "spec must not be empty")}
	}

	allErrs := v.validateCommon(newConfig)

	gardenerValidations := registrycacheextvalidations.ValidateRegistryConfigUpdate(toExtensionConfig(*oldConfig), toExtensionConfig(*newConfig), field.NewPath("spec"))

	return append(allErrs, transformFieldErrors(gardenerValidations)...)
}

func (v Validator) validateCommon(newConfig *registrycache.RegistryCacheConfig) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, validateUpstreamUniqueness(newConfig, v.runtimeClient)...)
	allErrs = append(allErrs, validateUpstreamResolvability(newConfig, v.dnsValidator)...)
	allErrs = append(allErrs, validateRemoteURLResolvability(newConfig, v.dnsValidator)...)
	allErrs = append(allErrs, validateSecretReference(newConfig, v.runtimeClient)...)

	return allErrs
}

func toExtensionConfig(rc registrycache.RegistryCacheConfig) *registrycacheext.RegistryConfig {

	return &registrycacheext.RegistryConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "registry.extensions.gardener.cloud/v1alpha3",
			Kind:       "RegistryConfig",
		},
		Caches: []registrycacheext.RegistryCache{toExtensionCache(rc.Spec)},
	}
}

func toExtensionCache(c registrycache.RegistryCacheConfigSpec) registrycacheext.RegistryCache {
	ext := registrycacheext.RegistryCache{
		Upstream:            c.Upstream,
		RemoteURL:           c.RemoteURL,
		SecretReferenceName: c.SecretReferenceName,
	}
	if c.Volume != nil {
		ext.Volume = &registrycacheext.Volume{
			Size:             c.Volume.Size,
			StorageClassName: c.Volume.StorageClassName,
		}
	}
	if c.GarbageCollection != nil {
		ext.GarbageCollection = &registrycacheext.GarbageCollection{
			TTL: c.GarbageCollection.TTL,
		}
	}
	if c.Proxy != nil {
		ext.Proxy = &registrycacheext.Proxy{
			HTTPProxy:  c.Proxy.HTTPProxy,
			HTTPSProxy: c.Proxy.HTTPSProxy,
		}
	}
	if c.HTTP != nil {
		ext.HTTP = &registrycacheext.HTTP{
			TLS: c.HTTP.TLS,
		}
	}

	return ext
}

func validateUpstreamUniqueness(newConfig *registrycache.RegistryCacheConfig, runtimeClient client.Client) field.ErrorList {

	var existingConfigs registrycache.RegistryCacheConfigList
	err := runtimeClient.List(context.Background(), &existingConfigs)
	if err != nil {
		return field.ErrorList{field.InternalError(field.NewPath("spec").Child("upstream"), errors.Wrap(err, "failed to list existing registry cache configs"))}
	}

	for _, existingConfig := range existingConfigs.Items {
		if existingConfig.Name == newConfig.Name && existingConfig.Namespace == newConfig.Namespace {
			continue
		}

		if existingConfig.Spec.Upstream == newConfig.Spec.Upstream {
			return field.ErrorList{field.Duplicate(field.NewPath("spec").Child("upstream"), newConfig.Spec.Upstream)}
		}
	}

	return nil
}

func validateUpstreamResolvability(newConfig *registrycache.RegistryCacheConfig, dns DNSValidator) field.ErrorList {
	var allErrs field.ErrorList

	if newConfig.Spec.Upstream != "" {
		host := stripPort(newConfig.Spec.Upstream)
		if dns != nil && host != "" && !dns.IsResolvable(host) {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("upstream"), newConfig.Spec.Upstream, "upstream is not DNS resolvable"))
		}
	}

	return allErrs
}

func validateRemoteURLResolvability(newConfig *registrycache.RegistryCacheConfig, dns DNSValidator) field.ErrorList {
	var allErrs field.ErrorList

	if newConfig.Spec.RemoteURL != nil {
		parsed, err := url.Parse(*newConfig.Spec.RemoteURL)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("remoteURL"), newConfig.Spec.RemoteURL, "failed to parse remoteURL"))

			return allErrs
		}

		host := parsed.Hostname()
		if host != "" && dns != nil && !dns.IsResolvable(host) {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("remoteURL"), newConfig.Spec.RemoteURL, "remoteURL is not DNS resolvable"))
		}
	}

	return allErrs
}

func validateSecretReference(newConfig *registrycache.RegistryCacheConfig, runtimeClient client.Client) field.ErrorList {
	if newConfig.Spec.SecretReferenceName != nil {

		var registryCacheSecret v1.Secret
		err := runtimeClient.Get(context.Background(), types.NamespacedName{
			Name:      *newConfig.Spec.SecretReferenceName,
			Namespace: newConfig.Namespace,
		}, &registryCacheSecret)

		if err != nil {
			if k8serrors.IsNotFound(err) {
				return field.ErrorList{field.Invalid(field.NewPath("spec").Child("secretReferenceName"),
					*newConfig.Spec.SecretReferenceName, fmt.Sprintf("secret %s does not exist", *newConfig.Spec.SecretReferenceName))}
			}

			return field.ErrorList{field.InternalError(field.NewPath("spec").Child("secretReferenceName"), errors.Wrap(err, "failed to get secret"))}
		}

		return registrycacheextvalidations.ValidateUpstreamRegistrySecret(&registryCacheSecret, field.NewPath("spec").Child("secretReferenceName"), *newConfig.Spec.SecretReferenceName)
	}

	return nil
}

func isEmptySpec(s registrycache.RegistryCacheConfigSpec) bool {
	return s.Upstream == "" &&
		s.RemoteURL == nil &&
		s.Volume == nil &&
		s.GarbageCollection == nil &&
		s.Proxy == nil &&
		s.SecretReferenceName == nil &&
		s.HTTP == nil
}

func transformFieldErrors(errs field.ErrorList) field.ErrorList {
	var crdErrors field.ErrorList

	for _, extensionError := range errs {
		extensionError.Field = adjustPath(extensionError.Field)
		crdErrors = append(crdErrors, extensionError)
	}

	return crdErrors
}

func adjustPath(extensionPath string) string {
	parts := strings.Split(extensionPath, ".")

	if len(parts) < 2 {
		return extensionPath
	}

	// Get rid of the second part which is always "caches[0]". Please see this: https://github.com/gardener/gardener-extension-registry-cache/blob/75e59657a15811faafccfccfdc0e5930adfe622d/pkg/apis/registry/validation/validation.go#L40
	partsExtracted := append(parts[:1], parts[2:]...)

	return strings.Join(partsExtracted, ".")
}

func stripPort(s string) string {
	if h, _, err := net.SplitHostPort(s); err == nil && h != "" {
		return h
	}
	if strings.Count(s, ":") == 1 {
		if idx := strings.IndexByte(s, ':'); idx > 0 {
			return s[:idx]
		}
	}
	return s
}
