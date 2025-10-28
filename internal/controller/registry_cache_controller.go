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
	"fmt"
	"time"

	"github.com/kyma-project/registry-cache/api/v1beta1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	requeueInterval = time.Second * 3
	finalizer       = "registry-cache.kyma-project.io/finalizer"
	debugLogLevel   = 2
	fieldOwner      = "registry-cache.kyma-project.io/owner"
)

type ManifestResources struct {
	Items []*unstructured.Unstructured
	Blobs [][]byte
}

type RegistryCacheReconciler struct {
	client.Client
	*runtime.Scheme
	record.EventRecorder
	*rest.Config
	resourceObjs       *ManifestResources
	FinalDeletionState v1beta1.State
}

func NewRegistryCacheReconciller(mgr ctrl.Manager) *RegistryCacheReconciler {
	return &RegistryCacheReconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		resourceObjs:       &ManifestResources{},
		FinalDeletionState: v1beta1.StateDeleting,
		//metrics: metrics,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *RegistryCacheReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Config = mgr.GetConfig()

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

func (r *RegistryCacheReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling RegistryCache resource", "namespace", req.Namespace, "name", req.Name)

	instance := v1beta1.RegistryCache{}
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		logger.Info(req.String() + " got deleted!")
		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, fmt.Errorf("error while getting object: %w", err)
		}
		return ctrl.Result{}, nil
	}

	status := getInstanceStatus(&instance)

	if !instance.GetDeletionTimestamp().IsZero() && status.State != r.FinalDeletionState {
		return ctrl.Result{}, r.setStatusForObjectInstance(ctx, &instance, status.WithState(r.FinalDeletionState))
	}

	if instance.GetDeletionTimestamp().IsZero() {
		if controllerutil.AddFinalizer(&instance, finalizer) {
			return ctrl.Result{}, r.ssa(ctx, &instance)
		}
	}

	switch status.State {
	case "":
		return ctrl.Result{}, r.HandleInitialState(ctx, &instance)
	case v1beta1.StateProcessing:
		return ctrl.Result{RequeueAfter: requeueInterval}, r.HandleProcessingState(ctx, &instance)
	case v1beta1.StateDeleting:
		return ctrl.Result{RequeueAfter: requeueInterval}, r.HandleDeletingState(ctx, &instance)
	case v1beta1.StateError:
		return ctrl.Result{RequeueAfter: requeueInterval}, r.HandleErrorState(ctx, &instance)
	case v1beta1.StateReady, v1beta1.StateWarning:
		return ctrl.Result{RequeueAfter: requeueInterval}, r.HandleReadyState(ctx, &instance)
	}
	return ctrl.Result{}, nil
}

func getInstanceStatus(objectInstance *v1beta1.RegistryCache) v1beta1.RegistryCacheStatus {
	return objectInstance.Status
}

func (r *RegistryCacheReconciler) processResources(ctx context.Context, objectInstance *v1beta1.RegistryCache) error {
	logger := log.FromContext(ctx)
	logger.Info("Installing Registry Cache resources")

	r.Event(objectInstance, "Normal", "ResourcesInstall", "installing resources")

	// the resources to be installed are unstructured,
	// so please make sure the types are available on the target cluster
	for _, obj := range r.resourceObjs.Items {
		if err := r.ssa(ctx, obj); err != nil && !errors2.IsAlreadyExists(err) {
			logger.Error(err, "error during installation of resources")
			return fmt.Errorf("error during installation of resources: %w", err)
		}
	}
	return nil
}

// HandleInitialState bootstraps state handling for the reconciled resource.
func (r *RegistryCacheReconciler) HandleInitialState(ctx context.Context, objectInstance *v1beta1.RegistryCache) error {
	status := getInstanceStatus(objectInstance)

	return r.setStatusForObjectInstance(ctx, objectInstance, status.
		WithState(v1beta1.StateProcessing).
		WithInstallConditionStatus(metav1.ConditionUnknown, objectInstance.GetGeneration()))
}

// HandleProcessingState processes the reconciled resource by processing the underlying resources.
// Based on the processing either a success or failure state is set on the reconciled resource.
func (r *RegistryCacheReconciler) HandleProcessingState(ctx context.Context, objectInstance *v1beta1.RegistryCache) error {
	status := getInstanceStatus(objectInstance)

	if err := r.processResources(ctx, objectInstance); err != nil {
		// stay in Processing state if FinalDeletionState is set to Processing
		if !objectInstance.GetDeletionTimestamp().IsZero() && r.FinalDeletionState == v1beta1.StateProcessing {
			return nil
		}

		r.Event(objectInstance, "Warning", "ResourcesInstall", err.Error())
		return r.setStatusForObjectInstance(ctx, objectInstance, status.
			WithState(v1beta1.StateError).
			WithInstallConditionStatus(metav1.ConditionFalse, objectInstance.GetGeneration()))
	}
	// set eventual state to Ready - if no errors were found
	return r.setStatusForObjectInstance(ctx, objectInstance, status.
		WithState(v1beta1.StateReady).
		WithInstallConditionStatus(metav1.ConditionTrue, objectInstance.GetGeneration()))
}

// HandleErrorState handles error recovery for the reconciled resource.
func (r *RegistryCacheReconciler) HandleErrorState(ctx context.Context, objectInstance *v1beta1.RegistryCache) error {
	status := getInstanceStatus(objectInstance)

	if err := r.processResources(ctx, objectInstance); err != nil {
		return err
	}

	// stay in Error state if FinalDeletionState is set to Error
	if !objectInstance.GetDeletionTimestamp().IsZero() && r.FinalDeletionState == v1beta1.StateError {
		return nil
	}
	// set eventual state to Ready - if no errors were found
	return r.setStatusForObjectInstance(ctx, objectInstance, status.
		WithState(v1beta1.StateReady).
		WithInstallConditionStatus(metav1.ConditionTrue, objectInstance.GetGeneration()))
}

// HandleDeletingState processed the deletion on the reconciled resource.
// Once the deletion if processed the relevant finalizers (if applied) are removed.
func (r *RegistryCacheReconciler) HandleDeletingState(ctx context.Context, objectInstance *v1beta1.RegistryCache) error {
	r.Event(objectInstance, "Normal", "Deleting", "resource deleting")
	logger := log.FromContext(ctx)
	status := getInstanceStatus(objectInstance)

	r.Event(objectInstance, "Normal", "ResourcesDelete", "deleting resources")

	// the resources to be installed are unstructured,
	// so please make sure the types are available on the target cluster
	for _, obj := range r.resourceObjs.Items {
		if err := r.Delete(ctx, obj); err != nil && !errors2.IsNotFound(err) {
			// stay in Deleting state if FinalDeletionState is set to Deleting
			if !objectInstance.GetDeletionTimestamp().IsZero() && r.FinalDeletionState == v1beta1.StateDeleting {
				return nil
			}

			logger.Error(err, "error during uninstallation of resources")
			r.Event(objectInstance, "Warning", "ResourcesDelete", "deleting resources error")
			return r.setStatusForObjectInstance(ctx, objectInstance, status.
				WithState(v1beta1.StateError).
				WithInstallConditionStatus(metav1.ConditionFalse, objectInstance.GetGeneration()))
		}
	}

	// if resources are ready to be deleted, remove finalizer
	if controllerutil.RemoveFinalizer(objectInstance, finalizer) {
		if err := r.Update(ctx, objectInstance); err != nil {
			return fmt.Errorf("error while removing finalizer: %w", err)
		}
		return nil
	}
	return nil
}

// HandleReadyState checks for the consistency of reconciled resource, by verifying the underlying resources.
func (r *RegistryCacheReconciler) HandleReadyState(ctx context.Context, objectInstance *v1beta1.RegistryCache) error {
	status := getInstanceStatus(objectInstance)
	if err := r.processResources(ctx, objectInstance); err != nil {
		// stay in Ready/Warning state if FinalDeletionState is set to Ready/Warning
		if !objectInstance.GetDeletionTimestamp().IsZero() &&
			(r.FinalDeletionState == v1beta1.StateReady || r.FinalDeletionState == v1beta1.StateWarning) {
			return nil
		}

		r.Event(objectInstance, "Warning", "ResourcesInstall", err.Error())
		return r.setStatusForObjectInstance(ctx, objectInstance, status.
			WithState(v1beta1.StateError).
			WithInstallConditionStatus(metav1.ConditionFalse, objectInstance.GetGeneration()))
	}
	return nil
}

func (r *RegistryCacheReconciler) setStatusForObjectInstance(ctx context.Context, objectInstance *v1beta1.RegistryCache, status *v1beta1.RegistryCacheStatus) error {
	objectInstance.Status = *status

	if err := r.ssaStatus(ctx, objectInstance); err != nil {
		r.Event(objectInstance, "Warning", "ErrorUpdatingStatus",
			fmt.Sprintf("updating state to %v", string(status.State)))
		return fmt.Errorf("error while updating status %s to: %w", status.State, err)
	}

	r.Event(objectInstance, "Normal", "StatusUpdated", fmt.Sprintf("updating state to %v", string(status.State)))
	return nil
}

// ssaStatus patches status using SSA on the passed object.
func (r *RegistryCacheReconciler) ssaStatus(ctx context.Context, obj client.Object) error {
	obj.SetManagedFields(nil)
	obj.SetResourceVersion("")
	if err := r.Status().Patch(ctx, obj, client.Apply,
		&client.SubResourcePatchOptions{PatchOptions: client.PatchOptions{FieldManager: fieldOwner}}); err != nil {
		return fmt.Errorf("error while patching status: %w", err)
	}

	return nil
}

// ssa patches the object using SSA.
func (r *RegistryCacheReconciler) ssa(ctx context.Context, obj client.Object) error {
	obj.SetManagedFields(nil)
	obj.SetResourceVersion("")
	if err := r.Patch(ctx, obj, client.Apply, client.ForceOwnership, client.FieldOwner(fieldOwner)); err != nil {
		return fmt.Errorf("error while patching object: %w", err)
	}
	return nil
}
