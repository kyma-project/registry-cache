package validations

import (
	registrycacheext "github.com/gardener/gardener-extension-registry-cache/pkg/apis/registry"
	registrycacheextvalidations "github.com/gardener/gardener-extension-registry-cache/pkg/apis/registry/validation"
	registrycache "github.com/kyma-project/registry-cache/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type Validator struct {
	secrets         []v1.Secret
	existingConfigs []registrycache.RegistryCacheConfig
}

func NewValidator(secrets []v1.Secret, existingConfigs []registrycache.RegistryCacheConfig) Validator {
	return Validator{
		secrets:         secrets,
		existingConfigs: existingConfigs,
	}
}

func (v Validator) Do(newConfig *registrycache.RegistryCacheConfig) field.ErrorList {
	objectToValidate := toExtensionConfig(*newConfig)

	return registrycacheextvalidations.ValidateRegistryConfig(objectToValidate, field.NewPath("spec"))
}

func (v Validator) DoOnUpdate(newConfig, oldConfig *registrycache.RegistryCacheConfig) field.ErrorList {
	return nil
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
