/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"context"
	"fmt"
	"github.com/kyma-project/registry-cache/internal/webhook/validations"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	corekymaprojectiov1beta1 "github.com/kyma-project/registry-cache/api/v1beta1"
)

// nolint:unused
// log is for logging in this package.
var registrycacheconfiglog = logf.Log.WithName("registrycacheconfig-resource")

// SetupRegistryCacheConfigWebhookWithManager registers the webhook for RegistryCacheConfig in the manager.
func SetupRegistryCacheConfigWebhookWithManager(mgr ctrl.Manager, client client.Client) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&corekymaprojectiov1beta1.RegistryCacheConfig{}).
		WithValidator(NewRegistryCacheConfigCustomValidator(client)).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validations.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-core-kyma-project-io-v1beta1-registrycacheconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=core.kyma-project.io,resources=registrycacheconfigs,verbs=create;update,versions=v1beta1,name=registrycacheconfig-v1beta1.kb.io,admissionReviewVersions=v1

// RegistryCacheConfigCustomValidator struct is responsible for validating the RegistryCacheConfig resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type RegistryCacheConfigCustomValidator struct {
	client client.Client
}

func NewRegistryCacheConfigCustomValidator(client client.Client) *RegistryCacheConfigCustomValidator {
	return &RegistryCacheConfigCustomValidator{
		client: client,
	}
}

var _ webhook.CustomValidator = &RegistryCacheConfigCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type RegistryCacheConfig.
func (v *RegistryCacheConfigCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {

	registrycacheconfig, ok := obj.(*corekymaprojectiov1beta1.RegistryCacheConfig)
	if !ok {
		return nil, fmt.Errorf("expected a RegistryCacheConfig object but got %T", obj)
	}
	registrycacheconfiglog.Info("Validation for RegistryCacheConfig upon creation", "name", registrycacheconfig.GetName())

	var registrycacheconfigs corekymaprojectiov1beta1.RegistryCacheConfigList
	err := v.client.List(context.Background(), &registrycacheconfigs, client.InNamespace(registrycacheconfig.Namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to list existing RegistryCacheConfig resources: %w", err)
	}

	var secretList v1.SecretList
	if err := v.client.List(context.Background(), &secretList, client.InNamespace(registrycacheconfig.Namespace)); err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	return nil, validations.NewValidator(validations.DefaultDNSValidator{}, v.client).Do(registrycacheconfig).ToAggregate()
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type RegistryCacheConfig.
func (v *RegistryCacheConfigCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {

	newRegistryCacheConfig, ok := newObj.(*corekymaprojectiov1beta1.RegistryCacheConfig)
	if !ok {
		return nil, fmt.Errorf("expected a RegistryCacheConfig object for the newObj but got %T", newObj)
	}

	oldRegistryCacheConfig, ok := oldObj.(*corekymaprojectiov1beta1.RegistryCacheConfig)
	if !ok {
		return nil, fmt.Errorf("expected a RegistryCacheConfig object for the newObj but got %T", newObj)
	}

	var registrycacheconfigs corekymaprojectiov1beta1.RegistryCacheConfigList
	err := v.client.List(context.Background(), &registrycacheconfigs, client.InNamespace(newRegistryCacheConfig.Namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to list existing RegistryCacheConfig resources: %w", err)
	}

	var secretList v1.SecretList
	if err := v.client.List(context.Background(), &secretList, client.InNamespace(newRegistryCacheConfig.Namespace)); err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	return nil, validations.NewValidator(validations.DefaultDNSValidator{}, v.client).DoOnUpdate(newRegistryCacheConfig, oldRegistryCacheConfig).ToAggregate()
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type RegistryCacheConfig.
func (v *RegistryCacheConfigCustomValidator) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}
