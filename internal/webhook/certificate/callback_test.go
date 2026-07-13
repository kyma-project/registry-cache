package certificate_test

import (
	"context"
	"testing"

	"github.com/kyma-project/registry-cache/internal/webhook/certificate"
	"github.com/stretchr/testify/assert"
	admissionregistration "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func testScheme(t *testing.T) *runtime.Scheme {
	scheme := runtime.NewScheme()
	if err := admissionregistration.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	return scheme
}

func testMWhCfg(name string, caBundle []byte) admissionregistration.ValidatingWebhookConfiguration {
	return admissionregistration.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Webhooks: []admissionregistration.ValidatingWebhook{
			{
				ClientConfig: admissionregistration.WebhookClientConfig{
					CABundle: caBundle,
				},
			},
		},
	}
}

func Test_BuildUpdateCABundle_get_error(t *testing.T) {
	ctx := context.Background()
	scheme := testScheme(t)

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	err := certificate.BuildUpdateCABundle(ctx, fakeClient, certificate.BuildUpdateCABundleOpts{
		Name:     "test-me",
		CABundle: []byte("updated"),
	})()

	assert.ErrorContains(t, err, "unable to get validating webhook configuration")
}

func Test_BuildUpdateCABundle_patch_error(t *testing.T) {
	ctx := context.Background()
	scheme := testScheme(t)

	mWhCfg := testMWhCfg("test-me", []byte("test-me"))

	// use default fake client's patch error for server side apply to verify if
	// patch errors are propagated properly
	fakeClient := fake.NewClientBuilder().
		WithObjects(&mWhCfg).
		WithScheme(scheme).
		Build()

	err := certificate.BuildUpdateCABundle(ctx, fakeClient, certificate.BuildUpdateCABundleOpts{
		Name:     "test-me",
		CABundle: []byte("updated"),
	})()

	assert.Error(t, err)
}

func Test_BuildUpdateCABundle(t *testing.T) {
	ctx := context.Background()
	scheme := testScheme(t)

	mWhCfg := testMWhCfg("test-me", []byte("test-me"))

	fakeClient := fake.NewClientBuilder().
		WithObjects(&mWhCfg).
		WithScheme(scheme).
		WithInterceptorFuncs(interceptor.Funcs{
			Apply: buildApplyFake(&mWhCfg),
		}).Build()

	err := certificate.BuildUpdateCABundle(ctx, fakeClient, certificate.BuildUpdateCABundleOpts{
		Name:         "test-me",
		CABundle:     []byte("updated"),
		FieldManager: "test-manager",
	})()

	assert.NoError(t, err)
	assert.Equal(t, []byte("updated"), mWhCfg.Webhooks[0].ClientConfig.CABundle)
}

func buildApplyFake(c *admissionregistration.ValidatingWebhookConfiguration) func(context.Context,
	client.WithWatch,
	runtime.ApplyConfiguration,
	...client.ApplyOption) error {

	return func(ctx context.Context,
		clnt client.WithWatch,
		obj runtime.ApplyConfiguration,
		opts ...client.ApplyOption) error {

		type unstructuredContent interface {
			UnstructuredContent() map[string]interface{}
		}
		uc, ok := obj.(unstructuredContent)
		if !ok {
			return clnt.Apply(ctx, obj, opts...)
		}

		var updated admissionregistration.ValidatingWebhookConfiguration
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(uc.UnstructuredContent(), &updated); err != nil {
			return err
		}
		*c = updated
		return nil
	}
}
