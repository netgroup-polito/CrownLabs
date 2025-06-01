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
	"os"
	"time"

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

	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/args"
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

	// Auth settings
	var keycloakURL string
	var keycloakRealm string
	var keycloakClientID string
	var keycloakClientSecret string

	flag.StringVar(&keycloakURL, "keycloak-url", "", "Keycloak URL.")
	flag.StringVar(&keycloakRealm, "keycloak-realm", "", "Keycloak realm.")
	flag.StringVar(&keycloakClientID, "keycloak-client-id", "", "Keycloak client ID.")
	flag.StringVar(&keycloakClientSecret, "keycloak-client-secret", "", "Keycloak client secret.")

	var tenantNSKeepAlive time.Duration
	flag.DurationVar(&tenantNSKeepAlive, "tenant-ns-keep-alive", 24*time.Hour,
		"Time elapsed after last login of tenant during which the tenant namespace should be kept alive")

	mydrivePVCsSize := args.NewQuantity("1Gi")
	var mydrivePVCsStorageClassName string
	var myDrivePVCsNamespace string
	flag.Var(&mydrivePVCsSize, "mydrive-pvcs-size", "The dimension of the user's personal space")
	flag.StringVar(&mydrivePVCsStorageClassName, "mydrive-pvcs-storage-class-name", "rook-nfs", "The name for the user's storage class")
	flag.StringVar(&myDrivePVCsNamespace, "mydrive-pvcs-namespace", "mydrive-pvcs", "The namespace where the PVCs are created")

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
		log.Error(err, "Unable to create manager")
		os.Exit(1)
	}

	targetLabel, err := utils.ParseLabel(targetLabelStr)
	if err != nil {
		log.Error(err, "Unable to parse target label")
		os.Exit(1)
	}
	log.Info("Selecting resources with label", "label", targetLabelStr)

	// enabling Keycloak if modules that needs it are enabled
	enableKeycloak := enableTenant // TODO || enableWorkspace
	// enabling Keycloak if all the settings are provided
	enableKeycloak = enableKeycloak && keycloakURL != ""
	enableKeycloak = enableKeycloak && keycloakRealm != ""
	enableKeycloak = enableKeycloak && keycloakClientID != ""
	enableKeycloak = enableKeycloak && keycloakClientSecret != ""
	if enableKeycloak {
		log.Info("Keycloak settings provided, initializing Keycloak actor")
		err = utils.SetupKeycloakActor(
			keycloakURL,
			keycloakClientID,
			keycloakClientSecret,
			keycloakRealm,
		)
		if err != nil {
			log.Error(err, "Unable to initialize Keycloak actor")
			os.Exit(1)
		}
	} else {
		log.Info("Keycloak actor will not be initialized (not needed or settings not provided)")
	}

	if enableTenant {
		log.Info("Starting the tenant controller")
		err := setup_tenant(mgr, targetLabel, tenantNSKeepAlive)
		if err != nil {
			log.Error(err, "Unable to create tenant controller")
			os.Exit(1)
		}
	}

	// TODO setup workspace reconciler

	// Setup operator probes
	if err := addOperatorProbes(mgr); err != nil {
		log.Error(err, "Unable to set up operator probes")
		os.Exit(1)
	}

	// Start the operator
	klog.Info("Starting manager")
	if err := mgr.Start(ctx); err != nil {
		log.Error(err, "Failed starting manager")
		os.Exit(1)
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
