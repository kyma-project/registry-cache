package rccontroller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	rcapi "github.com/kyma-project/registry-cache/api/v1beta1"
)

var _ = Describe("RegistryCache controller", func() {
	Context("When reconciling a resource", func() {
		const ResourceName = "test-resource"
		const NamespaceName = "default"
		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      ResourceName,
			Namespace: NamespaceName,
		}

		It("Should successfully process full lifecycle of Registry Cache resource - its status should be updated to Processing, Ready and Deleting", func() {
			By("By creating a new RegistryCache CR")
			registryCacheStub := newRegistryCacheStub(ResourceName)
			Expect(k8sClient.Create(ctx, registryCacheStub)).To(Succeed())

			By("By checking if RegistryCache CR has finalizer")
			Eventually(func() bool {
				registryCache := rcapi.RegistryCache{}
				if err := k8sClient.Get(ctx, typeNamespacedName, &registryCache); err != nil {
					return false
				}

				return controllerutil.ContainsFinalizer(&registryCache, "registry-cache.kyma-project.io/finalizer")
			}).Should(BeTrue())

			By("By waiting for RegistryCache to process creation and reach Processing state")
			Eventually(func() bool {
				registryCache := rcapi.RegistryCache{}
				if err := k8sClient.Get(ctx, typeNamespacedName, &registryCache); err != nil {
					return false
				}

				return registryCache.Status.State == rcapi.StateProcessing
			}, time.Second*60, time.Second*3).Should(BeTrue())

			By("By waiting for RegistryCache to reach Ready state")
			Eventually(func() bool {
				registryCache := rcapi.RegistryCache{}
				if err := k8sClient.Get(ctx, typeNamespacedName, &registryCache); err != nil {
					return false
				}

				return registryCache.Status.State == rcapi.StateReady
			}, time.Second*60, time.Second*3).Should(BeTrue())

			By("By deleting the RegistryCache CR")
			registryCache := rcapi.RegistryCache{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, &registryCache)).To(Succeed())
			Expect(k8sClient.Delete(ctx, &registryCache)).To(Succeed())

			By("By waiting for RegistryCache to reach Deleting state")
			Eventually(func() bool {
				registryCache := rcapi.RegistryCache{}
				if err := k8sClient.Get(ctx, typeNamespacedName, &registryCache); err != nil {
					return false
				}

				return registryCache.Status.State == rcapi.StateDeleting
			}, time.Second*60, time.Millisecond*500).Should(BeTrue())

			By("By checking if RegistryCache CR is deleted")
			Eventually(func() bool {
				registryCache := rcapi.RegistryCache{}
				err := k8sClient.Get(ctx, typeNamespacedName, &registryCache)
				return err != nil && apierrors.IsNotFound(err)
			}, time.Second*60, time.Second*3).Should(BeTrue())
		})
	})
})

func newRegistryCacheStub(name string) *rcapi.RegistryCache {
	return &rcapi.RegistryCache{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
	}
}
