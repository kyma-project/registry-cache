package rccontroller

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	rcapi "github.com/kyma-project/registry-cache/api/v1beta1"
)

var (
	testEnv    *envtest.Environment
	suiteCtx   context.Context
	cancelFunc context.CancelFunc
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{"../../config/crd/bases"},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	// Initialize clients, schemes, and other resources here as needed for tests

	k8sClient, err := client.New(cfg, client.Options{Scheme: runtime.NewScheme()})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	err = rcapi.AddToScheme(k8sClient.Scheme())
	Expect(err).NotTo(HaveOccurred())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: k8sClient.Scheme(),
		Metrics: server.Options{
			BindAddress: ":8084",
		},
	})
	Expect(err).ToNot(HaveOccurred())

	reconciler := NewRegistryCacheReconciler(mgr, healthz.Ping)
	Expect(reconciler).NotTo(BeNil())
	err = reconciler.SetupWithManager(mgr)
	Expect(err).To(BeNil())

	go func() {
		defer GinkgoRecover()
		suiteCtx, cancelFunc = context.WithCancel(context.Background())
		err = mgr.Start(suiteCtx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()

})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancelFunc()
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
