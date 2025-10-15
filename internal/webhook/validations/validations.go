package validations

import (
	registrycacheext "github.com/gardener/gardener-extension-registry-cache/pkg/apis/registry"
	registrycacheextvalidations "github.com/gardener/gardener-extension-registry-cache/pkg/apis/registry/validation"
	registrycache "github.com/kyma-project/registry-cache/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"net/url"
	"slices"
	"strings"
)

// Validator validates RegistryCacheConfig resources.
// Add a DNS validator to make DNS checks testable.
type Validator struct {
	secrets         []v1.Secret
	existingConfigs []registrycache.RegistryCacheConfig
	dns             DNSValidator
}

// NewValidator constructs a Validator with provided secrets and existing configs.
// Initialize the DNS validator with a real implementation by default.
func NewValidator(secrets []v1.Secret, existing []registrycache.RegistryCacheConfig, dnsValidator DNSValidator) Validator {
	return Validator{
		secrets:         secrets,
		existingConfigs: existing,
		dns:             dnsValidator,
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

	allErrs = append(allErrs, validateUpstreamUniqueness(newConfig, v.existingConfigs)...)
	allErrs = append(allErrs, validateUpstreamResolvability(newConfig, v.dns)...)
	allErrs = append(allErrs, validateRemoteURLResolvability(newConfig, v.dns)...)
	allErrs = append(allErrs, validateSecretReferenceName(newConfig, v.secrets)...)

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

func validateUpstreamUniqueness(newConfig *registrycache.RegistryCacheConfig, existingConfigs []registrycache.RegistryCacheConfig) field.ErrorList {
	var allErrs field.ErrorList

	for _, existingConfig := range existingConfigs {
		if existingConfig.Name == newConfig.Name && existingConfig.Namespace == newConfig.Namespace {
			continue
		}

		if existingConfig.Spec.Upstream == newConfig.Spec.Upstream {
			appendedErr := field.Duplicate(field.NewPath("spec").Child("upstream"), newConfig.Spec.Upstream)
			allErrs = append(allErrs, appendedErr)
			break
		}
	}

	return allErrs
}

func validateUpstreamResolvability(newConfig *registrycache.RegistryCacheConfig, dns DNSValidator) field.ErrorList {
	var allErrs field.ErrorList

	if u := newConfig.Spec.Upstream; u != "" {
		if dns != nil && !dns.IsResolvable(u) {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("upstream"), u, "upstream is not DNS resolvable"))
		}
	}

	return allErrs
}

func validateRemoteURLResolvability(newConfig *registrycache.RegistryCacheConfig, dns DNSValidator) field.ErrorList {
	var allErrs field.ErrorList

	if newConfig.Spec.RemoteURL != nil {
		if parsed, err := url.Parse(*newConfig.Spec.RemoteURL); err == nil {
			host := parsed.Hostname()
			if host != "" && dns != nil && !dns.IsResolvable(host) {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("remoteURL"), newConfig.Spec.RemoteURL, "remoteURL is not DNS resolvable"))
			}
		}
	}

	return allErrs
}

func validateSecretReferenceName(newConfig *registrycache.RegistryCacheConfig, secrets []v1.Secret) field.ErrorList {
	var allErrs field.ErrorList

	if newConfig.Spec.SecretReferenceName != nil {
		secretIndex := slices.IndexFunc(secrets, func(secret v1.Secret) bool {
			return secret.Name == *newConfig.Spec.SecretReferenceName && secret.Namespace == newConfig.Namespace
		})
		if secretIndex == -1 {
			allErrs = append(allErrs,
				field.Invalid(field.NewPath("spec").Child("secretReferenceName"),
					*newConfig.Spec.SecretReferenceName, "secret does not exist"),
			)
		} else {
			secret := secrets[secretIndex]
			secretErrors := registrycacheextvalidations.ValidateUpstreamRegistrySecret(&secret, field.NewPath("spec").Child("secretReferenceName"), *newConfig.Spec.SecretReferenceName)
			if len(secretErrors) > 0 {
				allErrs = append(allErrs, secretErrors...)
			}
		}
	}

	return allErrs
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
