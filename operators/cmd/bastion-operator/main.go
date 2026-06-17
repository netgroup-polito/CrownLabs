// Copyright 2020-2026 Politecnico di Torino
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
	"fmt"
	"net/http"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/textlogger"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/sshctrl"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/restcfg"
)

var (
	rscheme = runtime.NewScheme()
)

const (
	sshPort uint16 = 22
)

func init() {
	_ = scheme.AddToScheme(rscheme)

	_ = clv1alpha1.AddToScheme(rscheme)
	_ = clv1alpha2.AddToScheme(rscheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr, sshTrackerMetricName string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&sshTrackerMetricName, "ssh-tracker-metric-name", "bastion_ssh_connections", "The name of the metric for SSH connections.")

	restcfg.InitFlags(nil)
	klog.InitFlags(nil)
	flag.Parse()

	ctrl.SetLogger(textlogger.NewLogger(textlogger.NewConfig()))

	log := ctrl.Log.WithName("setup")

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 rscheme,
		Metrics:                server.Options{BindAddress: metricsAddr},
		WebhookServer:          webhook.NewServer(webhook.Options{Port: 9443}),
		LeaderElection:         enableLeaderElection,
		HealthProbeBindAddress: ":8081",
		LivenessEndpointName:   "/healthz",
		ReadinessEndpointName:  "/ready",
	})

	if err != nil {
		klog.Fatal(err, "Unable to start manager")
	}

	authorizedKeysPath, isEnvSet := os.LookupEnv("AUTHORIZED_KEYS_PATH")
	if !isEnvSet {
		log.Info("AUTHORIZED_KEYS_PATH env var is not set. Using default path \"/auth-keys-vol/authorized_keys\"")
		authorizedKeysPath = "/auth-keys-vol/authorized_keys"
	} else {
		log.Info("AUTHORIZED_KEYS_PATH env var found", "path", authorizedKeysPath)
	}

	if err = (&sshctrl.BastionReconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		AuthorizedKeysPath: authorizedKeysPath,
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "Bastion")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder
	// Add readiness probe
	err = mgr.AddReadyzCheck("ready-ping", healthz.Ping)
	if err != nil {
		log.Error(err, "Unable to add a readiness check")
		os.Exit(1)
	}

	// Add liveness probe
	err = mgr.AddHealthzCheck("health-ping", healthz.Ping)
	if err != nil {
		log.Error(err, "Unable to add an health check")
		os.Exit(1)
	}

	sshTracker := sshctrl.NewSSHTracker(log.WithName("ssh-tracker"), sshPort, sshTrackerMetricName)

	err = mgr.AddHealthzCheck("ssh-tracker", func(_ *http.Request) error {
		if !sshTracker.IsRunning() {
			return fmt.Errorf("tracker is not running")
		}
		return nil
	})
	if err != nil {
		log.Error(err, "Unable to add ssh-tracker health check")
		os.Exit(1)
	}

	err = mgr.Add(sshTracker)
	if err != nil {
		log.Error(err, "Unable to add SSH tracker to manager")
		os.Exit(1)
	}

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")
		os.Exit(1)
	}
}
