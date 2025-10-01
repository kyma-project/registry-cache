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
	"errors"
	"fmt"

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
func SetupRegistryCacheConfigWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&corekymaprojectiov1beta1.RegistryCacheConfig{}).
		WithValidator(&RegistryCacheConfigCustomValidator{}).
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
	// TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &RegistryCacheConfigCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type RegistryCacheConfig.
func (v *RegistryCacheConfigCustomValidator) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {

	registrycacheconfig, ok := obj.(*corekymaprojectiov1beta1.RegistryCacheConfig)
	if !ok {
		return nil, fmt.Errorf("expected a RegistryCacheConfig object but got %T", obj)
	}
	registrycacheconfiglog.Info("Validation for RegistryCacheConfig upon creation", "name", registrycacheconfig.GetName())

	// TODO(user): fill in your validation logic upon object creation.

	return nil, errors.New("not implemented temporarily")
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type RegistryCacheConfig.
func (v *RegistryCacheConfigCustomValidator) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {

	registrycacheconfig, ok := newObj.(*corekymaprojectiov1beta1.RegistryCacheConfig)
	if !ok {
		return nil, fmt.Errorf("expected a RegistryCacheConfig object for the newObj but got %T", newObj)
	}
	registrycacheconfiglog.Info("Validation for RegistryCacheConfig upon update", "name", registrycacheconfig.GetName())

	// TODO(user): fill in your validation logic upon object update.

	return nil, errors.New("not implemented temporarily")
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type RegistryCacheConfig.
func (v *RegistryCacheConfigCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	registrycacheconfig, ok := obj.(*corekymaprojectiov1beta1.RegistryCacheConfig)
	if !ok {
		return nil, fmt.Errorf("expected a RegistryCacheConfig object but got %T", obj)
	}
	registrycacheconfiglog.Info("Validation for RegistryCacheConfig upon deletion", "name", registrycacheconfig.GetName())

	// TODO(user): fill in your validation logic upon object deletion.

	return nil, nil
}
