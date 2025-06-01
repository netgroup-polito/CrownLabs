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

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/textlogger"

	"k8s.io/apimachinery/pkg/runtime"
	// virtv1 "kubevirt.io/api/core/v1"
	// cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/utils"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(crownlabsv1alpha1.AddToScheme(scheme))
	utilruntime.Must(crownlabsv1alpha2.AddToScheme(scheme))

	// utilruntime.Must(virtv1.AddToScheme(scheme))
	// utilruntime.Must(cdiv1beta1.AddToScheme(scheme))
}

func main() {
	// General settings
	var metricsAddr string
	var enableLeaderElection bool
	var targetLabelStr string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&targetLabelStr, "target-label", "", "The key=value pair label that needs to be in the resource to be reconciled. A single pair in the format key=value")

	// Enabling modules
	var enableTenant bool
	flag.BoolVar(&enableTenant, "enable-tenant", true, "Enable the tenant controller.")

	klog.InitFlags(nil)
	flag.Parse()

	ctrl.SetLogger(textlogger.NewLogger(textlogger.NewConfig()))

	ctx := ctrl.SetupSignalHandler()
	log := ctrl.Log.WithName("setup")

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                server.Options{BindAddress: metricsAddr},
		WebhookServer:          webhook.NewServer(webhook.Options{Port: 9443}),
		LeaderElection:         enableLeaderElection,
		HealthProbeBindAddress: ":8081",
		LivenessEndpointName:   "/healthz",
		ReadinessEndpointName:  "/ready",
	})
	if err != nil {
		klog.Fatal(err, "Unable to create manager")
	}

	targetLabel, err := utils.ParseLabel(targetLabelStr)
	if err != nil {
		klog.Fatal(err, "Unable to parse target label")
	}
	log.Info("Selecting resources with label", "label", targetLabelStr)

	// enabling Keycloak if modules that needs it are enabled
	enableKeycloak := enableTenant // TODO || enableWorkspace
	if enableKeycloak {
		err := setup_keycloak(log)
		if err != nil {
			klog.Fatal(err, "Unable to setup Keycloak actor")
		}
	} else {
		log.Info("Keycloak actor will not be initialized (not needed)")
	}

	if enableTenant {
		log.Info("Starting the tenant controller")
		err := setup_tenant(mgr, targetLabel)
		if err != nil {
			klog.Fatal(err, "Unable to create tenant controller")
		}
	}

	// TODO setup workspace reconciler

	// Setup operator probes
	if err := addOperatorProbes(mgr); err != nil {
		klog.Fatal(err, "Unable to set up operator probes")
	}

	// Start the operator
	klog.Info("Starting manager")
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
