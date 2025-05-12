// Copyright 2020-2025 Politecnico di Torino
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package main contains the entrypoint for the Crownlabs unified operator.
package main

import (
	"flag"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
)

var (
	scheme         = runtime.NewScheme()
	enableWebhooks bool
	reschedule     = common.Rescheduler{
		RequeueAfterMin: 1 * 24 * time.Hour,
		RequeueAfterMax: 7 * 24 * time.Hour,
	}
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(v1alpha2.AddToScheme(scheme))
}

func main() {
	// General settings
	var metricsAddr string
	var healthProbeAddr string
	var enableLeaderElection bool
	var targetLabelStr string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&healthProbeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&targetLabelStr, "target-label", "", "The key=value pair label that needs to be in the resource to be reconciled. A single pair in the format key=value")
	flag.DurationVar(&reschedule.RequeueAfterMin, "reschedule-min", reschedule.RequeueAfterMin,
		"Minimum duration to wait before requeuing the reconciliation. "+
			"Set to 0 to disable requeuing. "+
			"Default is 1 day.")
	flag.DurationVar(&reschedule.RequeueAfterMax, "reschedule-max", reschedule.RequeueAfterMax,
		"Maximum duration to wait before requeuing the reconciliation. "+
			"Set to 0 to disable requeuing. "+
			"Default is 7 days.")

	// Enabling modules
	var enableTenant bool
	var enableWorkspace bool
	var enableKeycloak bool
	flag.BoolVar(&enableTenant, "enable-tenant", true, "Enable the tenant controller.")
	flag.BoolVar(&enableWorkspace, "enable-workspace", true, "Enable the workspace controller.")
	flag.BoolVar(&enableKeycloak, "enable-keycloak", true, "Enable the Keycloak integration.")

	flag.BoolVar(&enableWebhooks, "enable-webhooks", true, "Enable the webhooks server.")

	klog.InitFlags(nil)
	flag.Parse()

	// Set the ctrl.Log to klogr, which is compatible with klog
	ctrl.SetLogger(klog.NewKlogr())

	ctx := ctrl.SetupSignalHandler()
	log := ctrl.Log.WithName("setup")

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                server.Options{BindAddress: metricsAddr},
		WebhookServer:          webhook.NewServer(webhook.Options{Port: 9443}),
		LeaderElection:         enableLeaderElection,
		HealthProbeBindAddress: healthProbeAddr,
		LivenessEndpointName:   "/healthz",
		ReadinessEndpointName:  "/ready",
	})
	if err != nil {
		klog.Fatal(err, "Unable to create manager")
	}

	targetLabel, err := common.ParseLabel(targetLabelStr)
	if err != nil {
		klog.Fatal(err, "Unable to parse target label")
	}
	log.Info("Selecting resources with label", "label", targetLabelStr)

	// enabling Keycloak if modules that needs it are enabled
	enableKeycloak = enableKeycloak && (enableTenant || enableWorkspace)
	if enableKeycloak {
		err := setupKeycloak(ctx, log)
		if err != nil {
			klog.Fatal(err, "Unable to setup Keycloak actor")
		}
	} else {
		log.Info("Keycloak actor will not be initialized (not needed)")
	}

	if enableTenant {
		log.Info("Starting the tenant controller")
		err := setupTenant(mgr, log, targetLabel)
		if err != nil {
			klog.Fatal(err, "Unable to create tenant controller")
		}
	}

	if enableWorkspace {
		log.Info("Starting the workspace controller")
		err := setupWorkspace(mgr, log, targetLabel)
		if err != nil {
			klog.Fatal(err, "Unable to create workspace controller")
		}
	}

	// Setup operator probes
	if err := addOperatorProbes(mgr); err != nil {
		klog.Fatal(err, "Unable to set up operator probes")
	}

	// Start the operator
	log.Info("Starting manager")
	if err := mgr.Start(ctx); err != nil {
		klog.Fatal(err, "Failed starting manager")
	}
}

func addOperatorProbes(mgr manager.Manager) error {
	// Add readiness probe
	err := mgr.AddReadyzCheck("ready-ping", healthz.Ping)
	if err != nil {
		return err
	}

	// Add liveness probe
	err = mgr.AddHealthzCheck("health-ping", healthz.Ping)
	if err != nil {
		return err
	}

	return nil
}
