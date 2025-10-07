package validations

import (
	registrycacheext "github.com/gardener/gardener-extension-registry-cache/pkg/apis/registry/v1alpha3"
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
	return nil
}

func (v Validator) DoOnUpdate(newConfig, oldConfig *registrycache.RegistryCacheConfig) field.ErrorList {
	return nil
}

func toExtensionConfig(rc registrycache.RegistryCacheConfig) registrycacheext.RegistryConfig {
	return registrycacheext.RegistryConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "registry.extensions.gardener.cloud/v1alpha3",
			Kind:       "RegistryConfig",
		},
		Caches: []registrycacheext.RegistryCache{toExtensionCache(rc.Spec)},
	}
}

func toExtensionCache(registryCacheConfig registrycache.RegistryCacheConfigSpec) registrycacheext.RegistryCache {
	return registrycacheext.RegistryCache{}
}
