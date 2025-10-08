package certificate

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"time"

	admissionregistration "k8s.io/api/admissionregistration/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BuildUpdateCABundleOpts struct {
	// Name of the validating webhook configuration to be updated
	Name string
	// CABundle the validating webhook configuration webhooks will be updated with
	CABundle []byte
	// FieldManager the name of the filed manager for patch operation
	FieldManager string
}

// buildUpdateCABundle - builds a function that will update certificate authority
func BuildUpdateCABundle(
	ctx context.Context,
	rtClient client.Client,
	opts BuildUpdateCABundleOpts) func() error {

	logger := slog.Default()
	return func() error {
		getCtx, cancelGet := context.WithTimeout(ctx, 5*time.Second)
		defer cancelGet()

		var validatingWebhook admissionregistration.ValidatingWebhookConfiguration
		if err := rtClient.Get(
			getCtx,
			client.ObjectKey{Name: opts.Name},
			&validatingWebhook); err != nil {
			return fmt.Errorf("unable to get validating webhook configuration: %w", err)
		}

		var updated bool
		for i := 0; i < len(validatingWebhook.Webhooks); i++ {
			if bytes.Equal(opts.CABundle, validatingWebhook.Webhooks[i].ClientConfig.CABundle) {
				continue
			}
			validatingWebhook.Webhooks[i].ClientConfig.CABundle = opts.CABundle
			updated = true
		}

		if !updated {
			logger.Info("validating webhook configuration up to date")
			return nil
		}

		validatingWebhook.Kind = "ValidatingWebhookConfiguration"
		validatingWebhook.APIVersion = "admissionregistration.k8s.io/v1"
		validatingWebhook.ManagedFields = nil

		patchCtx, cancelPatch := context.WithTimeout(ctx, 5*time.Second)
		defer cancelPatch()

		logger.Info("attempting to patch validating webhook configuration", "name", validatingWebhook.Name)

		return rtClient.Patch(patchCtx, &validatingWebhook, client.Apply, &client.PatchOptions{
			FieldManager: opts.FieldManager,
			Force:        ptr.To(true),
		})
	}
}
