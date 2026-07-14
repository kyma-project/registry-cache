package rccontroller

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-project/registry-cache/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	requeueInterval       = time.Second * 5
	requeueHealthInterval = time.Second * 30
	finalizer             = "registry-cache.kyma-project.io/finalizer"
	fieldOwner            = "registry-cache.kyma-project.io/owner"
)

type RegistryCacheReconciler struct {
	client.Client
	*runtime.Scheme
	record.EventRecorder
	healthz.Checker
}

func NewRegistryCacheReconciler(mgr ctrl.Manager, check healthz.Checker) *RegistryCacheReconciler {
	return &RegistryCacheReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		EventRecorder: mgr.GetEventRecorderFor("registry-cache-controller"),
		Checker:       check,
	}
}

func (r *RegistryCacheReconciler) SetupWithManager(mgr ctrl.Manager) error {
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

	if !instance.GetDeletionTimestamp().IsZero() && status.State != v1beta1.StateDeleting {
		return ctrl.Result{RequeueAfter: requeueInterval}, r.setStatusForObjectInstance(ctx, &instance, status.WithState(v1beta1.StateDeleting))
	}

	if instance.GetDeletionTimestamp().IsZero() {
		if controllerutil.AddFinalizer(&instance, finalizer) {
			logger.Info("Adding finalizer")
			return ctrl.Result{RequeueAfter: requeueInterval}, r.ssa(ctx, &instance)
		}
	}

	switch status.State {
	case "":
		return ctrl.Result{RequeueAfter: requeueInterval}, r.handleInitialState(ctx, &instance)
	case v1beta1.StateProcessing:
		return ctrl.Result{RequeueAfter: requeueInterval}, r.handleProcessingState(ctx, &instance)
	case v1beta1.StateDeleting:
		return ctrl.Result{RequeueAfter: requeueInterval}, r.handleDeletingState(ctx, &instance)
	case v1beta1.StateError:
		return ctrl.Result{RequeueAfter: requeueHealthInterval}, r.handleErrorState(ctx, &instance)
	case v1beta1.StateReady, v1beta1.StateWarning:
		return ctrl.Result{RequeueAfter: requeueHealthInterval}, r.handleReadyState(ctx, &instance)
	}
	return ctrl.Result{}, nil
}

func (r *RegistryCacheReconciler) handleInitialState(ctx context.Context, objectInstance *v1beta1.RegistryCache) error {
	return r.setInstanceStatus(ctx, objectInstance, v1beta1.StateProcessing, metav1.ConditionUnknown)
}

func (r *RegistryCacheReconciler) handleProcessingState(ctx context.Context, objectInstance *v1beta1.RegistryCache) error {
	return r.attemptToBecomeReady(ctx, objectInstance)
}

func (r *RegistryCacheReconciler) handleErrorState(ctx context.Context, objectInstance *v1beta1.RegistryCache) error {
	return r.attemptToBecomeReady(ctx, objectInstance)
}

func (r *RegistryCacheReconciler) handleReadyState(ctx context.Context, objectInstance *v1beta1.RegistryCache) error {
	// check if webhook is still ready
	if err := r.Checker(nil); err != nil {
		r.Event(objectInstance, "Error", "Webhook server not ready", err.Error())
		return r.setInstanceStatus(ctx, objectInstance, v1beta1.StateError, metav1.ConditionFalse)
	}
	return nil
}

func (r *RegistryCacheReconciler) handleDeletingState(ctx context.Context, objectInstance *v1beta1.RegistryCache) error {
	logger := log.FromContext(ctx)
	logger.Info("RegistryCache resource deleting state processing")

	r.Event(objectInstance, "Normal", "Deleting", "deleting webhook")

	if controllerutil.RemoveFinalizer(objectInstance, finalizer) {
		logger.Info("Removing finalizer")
		if err := r.Update(ctx, objectInstance); err != nil {
			return fmt.Errorf("error while removing finalizer: %w", err)
		}
		return nil
	}
	return nil
}

// stay in current state until we are ready
func (r *RegistryCacheReconciler) attemptToBecomeReady(ctx context.Context, objectInstance *v1beta1.RegistryCache) error {
	if err := r.Checker(nil); err != nil {
		logger := log.FromContext(ctx)
		logger.Info("Webhook server not ready!")
		return nil
	}
	return r.setInstanceStatus(ctx, objectInstance, v1beta1.StateReady, metav1.ConditionTrue)
}

func getInstanceStatus(objectInstance *v1beta1.RegistryCache) v1beta1.RegistryCacheStatus {
	return objectInstance.Status
}

func (r *RegistryCacheReconciler) setInstanceStatus(ctx context.Context, objectInstance *v1beta1.RegistryCache, state v1beta1.State, condStatus metav1.ConditionStatus) error {
	status := getInstanceStatus(objectInstance)
	return r.setStatusForObjectInstance(ctx, objectInstance, status.
		WithState(state).
		WithInstallConditionStatus(condStatus, objectInstance.GetGeneration()))
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
