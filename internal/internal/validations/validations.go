package validations

import (
	registrycache "github.com/kyma-project/registry-cache/api/v1beta1"
	v1 "k8s.io/api/core/v1"
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
