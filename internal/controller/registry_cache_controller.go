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

package rccontroller

import (
	"context"
	"time"

	"github.com/kyma-project/registry-cache/api/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type RegistryCacheReconciller struct {
	client.Client
	*runtime.Scheme
}

func NewRegistryCacheReconciller(mgr ctrl.Manager, objects []unstructured.Unstructured) *RegistryCacheReconciller {
	return &RegistryCacheReconciller{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		//log:     logger,
		//metrics: metrics,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *RegistryCacheReconciller) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.RegistryCache{}).
		WithEventFilter(predicate.Or(
			predicate.GenerationChangedPredicate{},
			predicate.LabelChangedPredicate{},
			predicate.AnnotationChangedPredicate{},
		)).
		Named("registry-cache-controller").
		Complete(r)
}

func (r *RegistryCacheReconciller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var instance v1beta1.RegistryCache
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		return ctrl.Result{
			RequeueAfter: time.Second * 30,
		}, client.IgnoreNotFound(err)
	}

	return ctrl.Result{}, nil
}
