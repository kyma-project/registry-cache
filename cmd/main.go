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

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"os"
	"path"

	rccontroller "github.com/kyma-project/registry-cache/internal/controller"
	"github.com/kyma-project/registry-cache/internal/webhook/certificate"
	"github.com/kyma-project/registry-cache/internal/webhook/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	webhook "github.com/kyma-project/registry-cache/internal/webhook/server"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	registrycachetypes "github.com/kyma-project/registry-cache/api/v1beta1"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const (
	certDir                  = "/tmp/"
	certificateAuthorityName = "ca.crt"
	flagWebhookName          = "webhook-name"
	webhookServerKeyName     = "tls.key"
	webhookServerCertName    = "tls.crt"
	patchFieldManagerName    = "registry-cache-webhook"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(registrycachetypes.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

// nolint:gocyclo
func main() {
	var metricsAddr string
	var probeAddr string
	var enableHTTP2 bool
	var tlsOpts []func(*tls.Config)
	var webhookCfgName string

	flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. "+
		"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	flag.StringVar(&webhookCfgName, flagWebhookName, "registry-cache-validating-webhook-configuration", "The name of the validating webhook configuration to be updated.")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		setupLog.Error(err, "unable to create rest configuration")
		os.Exit(1)
	}

	rtClient, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		setupLog.Error(err, "unable to create client")
		os.Exit(1)
	}

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts:  tlsOpts,
		CertDir:  certDir,
		KeyName:  webhookServerKeyName,
		CertName: webhookServerCertName,
		Callback: func(cert tls.Certificate) {
			certPath := path.Join(certDir, certificateAuthorityName)
			data, err := os.ReadFile(certPath)
			if err != nil {
				setupLog.Error(err, "unable to read certificate")
				os.Exit(1)
			}
			setupLog.Info("certificate loaded")

			updateCABundle := certificate.BuildUpdateCABundle(
				context.Background(),
				rtClient,
				certificate.BuildUpdateCABundleOpts{
					Name:         webhookCfgName,
					CABundle:     data,
					FieldManager: patchFieldManagerName,
				})

			if err := retry.RetryOnConflict(retry.DefaultBackoff, updateCABundle); err != nil {
				setupLog.Error(err, "unable to patch validating webhook configuration")
				os.Exit(1)
			}
		},
	})

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         false,
		LeaderElectionID:       "9a54de5d.kyma-project.io",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := v1beta1.SetupRegistryCacheConfigWebhookWithManager(mgr, rtClient); err != nil {
		setupLog.Error(err, "unable to setup registry cache config webhook")
		os.Exit(1)
	}

	regCacheReconciler := rccontroller.NewRegistryCacheReconciller(mgr, webhookServer.StartedChecker())

	if err = regCacheReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RegistryCache")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", webhookServer.StartedChecker()); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", webhookServer.StartedChecker()); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
