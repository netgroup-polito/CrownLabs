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

// Package main contains the entrypoint for the bastion operator.
package main

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	bastion_controller "github.com/netgroup-polito/CrownLabs/operators/pkg/bastion-controller"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = crownlabsv1alpha1.AddToScheme(scheme)
	_ = crownlabsv1alpha2.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	klog.InitFlags(nil)
	flag.Parse()

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
		klog.Fatal("Unable to start manager", err)
	}

	authorizedKeysPath, isEnvSet := os.LookupEnv("AUTHORIZED_KEYS_PATH")
	if !isEnvSet {
		klog.Info("AUTHORIZED_KEYS_PATH env var is not set. Using default path \"/auth-keys-vol/authorized_keys\"")
		authorizedKeysPath = "/auth-keys-vol/authorized_keys"
	} else {
		klog.Infof("AUTHORIZED_KEYS_PATH env var found. Using path %v", authorizedKeysPath)
	}

	if err = (&bastion_controller.BastionReconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		AuthorizedKeysPath: authorizedKeysPath,
	}).SetupWithManager(mgr); err != nil {
		klog.Fatal("unable to create controller", "controller", "Bastion", err)
	}

	// +kubebuilder:scaffold:builder
	// Add readiness probe
	err = mgr.AddReadyzCheck("ready-ping", healthz.Ping)
	if err != nil {
		klog.Fatal("Unable to add a readiness check", err)
	}

	// Add liveness probe
	err = mgr.AddHealthzCheck("health-ping", healthz.Ping)
	if err != nil {
		klog.Fatal("Unable add an health check", err)
	}

	klog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		klog.Fatal("problem running manager", err)
	}
}
