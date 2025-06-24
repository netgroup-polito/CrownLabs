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

// Package main contains the entrypoint for the instance automation operator.
package main

import (
	"flag"
	"net/smtp"
	"os"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/textlogger"
	virtv1 "kubevirt.io/api/core/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instautoctrl"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/restcfg"
)

var (
	scheme     = runtime.NewScheme()
	mailClient *utils.MailClient
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(crownlabsv1alpha1.AddToScheme(scheme))
	utilruntime.Must(crownlabsv1alpha2.AddToScheme(scheme))

	utilruntime.Must(virtv1.AddToScheme(scheme))
	utilruntime.Must(cdiv1beta1.AddToScheme(scheme))
}

func main() {
	containerEnvOpts := forge.ContainerEnvOpts{}

	enableInstanceTermination := flag.Bool("enable-instance-termination", false, "Enable the Instance Termination controller")
	enableInstanceSubmission := flag.Bool("enable-instance-submission", false, "Enable the Instance Submission controller")
	enableInactiveTermination := flag.Bool("enable-instance-inactive-termination", false, "Enable the Instance Inactive Termination controller")
	enableInstanceExpiration := flag.Bool("enable-instance-expiration", false, "Enable the Instance Expiration controller")

	metricsAddr := flag.String("metrics-addr", ":8080", "The address the metric endpoint binds to.")
	enableLeaderElection := flag.Bool("enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")

	namespaceWhiteList := flag.String("namespace-whitelist", "production=true", "The whitelist of the namespaces on "+
		"which the controller will work. Different labels (key=value) can be specified, by separating them with a &"+
		"( e.g. key1=value1&key2=value2")

	prometheusURL := flag.String("monitoring-prometheus-url", "http://prometheus-kube-prometheus-prometheus.monitoring.svc.cluster.local", "The URL of the Prometheus instance to use for the Inactive Termination")

	instanceTerminationStatusCheckTimeout := flag.Duration("instance-termination-status-check-timeout", 3*time.Second, "The maximum time to wait for the status check for Instances that require it")
	instanceTerminationStatusCheckInterval := flag.Duration("instance-termination-status-check-interval", 2*time.Minute, "The interval to check the status of Instances that require it")

	maxConcurrentTerminationReconciles := flag.Int("max-concurrent-reconciles-termination", 1, "The maximum number of concurrent Reconciles which can be run for the Instance Termination controller")
	maxConcurrentSubmissionReconciles := flag.Int("max-concurrent-reconciles-submission", 1, "The maximum number of concurrent Reconciles which can be run for the Instance Submission controller")
	maxConcurrentInactiveTerminationReconciles := flag.Int("max-concurrent-reconciles-inactive-termination", 1, "The maximum number of concurrent Reconciles which can be run for the Instance Inactive Termination controller")

	instanceInactiveTerminationStatusCheckTimeout := flag.Duration("instance-inactive-termination-status-check-timeout", 5*time.Second, "The maximum time to wait for the status check for Instances that require it")
	instanceInactiveTerminationMaxNumberOfAlerts := flag.Int("instance-inactive-termination-max-number-of-alerts", 3, "The max number of alerts to send before terminating an inactive Instance")

	flag.StringVar(&containerEnvOpts.ImagesTag, "container-env-sidecars-tag", "latest", "The tag for service containers (such as gui sidecar containers)")
	flag.StringVar(&containerEnvOpts.ContentUploaderImg, "container-env-content-uploader-img", "latest", "The image name for the job to compress and upload instance content from a persistent instance.")

	smtpServer := flag.String("smtp-server", "smtp.polito.it", "SMTP server for sending emails")
	smtpPort := flag.Int("smtp-port", 587, "SMTP server port")
	smtpIdentity := flag.String("smtp-identity", "", "SMTP identity for authentication")
	smtpUsername := flag.String("smtp-username", "", "SMTP username for authentication")
	smtpPassword := flag.String("smtp-password", "", "SMTP password for authentication")
	smtpFrom := flag.String("smtp-from", "crownlabs@polito.it", "Email sender address")

	restcfg.InitFlags(nil)
	klog.InitFlags(nil)
	flag.Parse()

	mailClient = &utils.MailClient{
		SMTPServer: *smtpServer,
		SMTPPort:   *smtpPort,
		Auth:       smtp.PlainAuth(*smtpIdentity, *smtpUsername, *smtpPassword, *smtpServer),
		From:       *smtpFrom,
	}

	ctrl.SetLogger(textlogger.NewLogger(textlogger.NewConfig()))

	log := ctrl.Log.WithName("setup")

	whiteListMap := parseMap(*namespaceWhiteList)
	log.Info("restricting reconciled namespaces", "labels", *namespaceWhiteList)

	// Configure the manager
	mgr, err := ctrl.NewManager(restcfg.SetRateLimiter(ctrl.GetConfigOrDie()), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                server.Options{BindAddress: *metricsAddr},
		LeaderElection:         *enableLeaderElection,
		HealthProbeBindAddress: ":8081",
		LivenessEndpointName:   "/healthz",
		ReadinessEndpointName:  "/ready",
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	nsWhitelist := metav1.LabelSelector{MatchLabels: whiteListMap, MatchExpressions: []metav1.LabelSelectorRequirement{}}

	if *enableInstanceTermination {
		log.Info("Instance Termination controller enabled.")
		// Configure the Instance termination controller
		instanceTermination := "InstanceTermination"
		if err := (&instautoctrl.InstanceTerminationReconciler{
			Client:                      mgr.GetClient(),
			Scheme:                      mgr.GetScheme(),
			EventsRecorder:              mgr.GetEventRecorderFor(instanceTermination),
			NamespaceWhitelist:          nsWhitelist,
			StatusCheckRequestTimeout:   *instanceTerminationStatusCheckTimeout,
			InstanceStatusCheckInterval: *instanceTerminationStatusCheckInterval,
		}).SetupWithManager(mgr, *maxConcurrentTerminationReconciles); err != nil {
			log.Error(err, "unable to create controller", "controller", instanceTermination)
			os.Exit(1)
		}
	} else {
		log.Info("Instance Termination controller disabled.")
	}

	if *enableInstanceSubmission {
		log.Info("Instance Submission controller enabled.")

		// Configure the Instance submission controller
		instanceSubmission := "InstanceSubmission"
		if err := (&instautoctrl.InstanceSubmissionReconciler{
			Client:             mgr.GetClient(),
			Scheme:             mgr.GetScheme(),
			EventsRecorder:     mgr.GetEventRecorderFor(instanceSubmission),
			ContainerEnvOpts:   containerEnvOpts,
			NamespaceWhitelist: nsWhitelist,
		}).SetupWithManager(mgr, *maxConcurrentSubmissionReconciles); err != nil {
			log.Error(err, "unable to create controller", "controller", instanceSubmission)
			os.Exit(1)
		}
	} else {
		log.Info("Instance Submission controller disabled.")
	}

	if *enableInactiveTermination {
		log.Info("Instance Inactive Termination controller enabled.")

		// Configure the Instance Inactive termination controller
		instanceInactiveTermination := "InstanceInactiveTermination."
		if err := (&instautoctrl.InstanceInactiveTerminationReconciler{
			Client:                    mgr.GetClient(),
			Scheme:                    mgr.GetScheme(),
			EventsRecorder:            mgr.GetEventRecorderFor(instanceInactiveTermination),
			NamespaceWhitelist:        nsWhitelist,
			StatusCheckRequestTimeout: *instanceInactiveTerminationStatusCheckTimeout,
			InstanceMaxNumberOfAlerts: *instanceInactiveTerminationMaxNumberOfAlerts,
			MailClient:                mailClient,
			PrometheusURL:             *prometheusURL,
		}).SetupWithManager(mgr, *maxConcurrentInactiveTerminationReconciles); err != nil {
			log.Error(err, "unable to create controller", "controller", instanceInactiveTermination)
			os.Exit(1)
		}
	} else {
		log.Info("Instance Inactive Termination controller disabled.")
	}

	if *enableInstanceExpiration {
		log.Info("Instance Expiration controller enabled.")
		// Configure the Instance Expiration controller
		instanceExpiration := "InstanceExpiration"
		if err := (&instautoctrl.InstanceExpirationReconciler{
			Client:                    mgr.GetClient(),
			Scheme:                    mgr.GetScheme(),
			EventsRecorder:            mgr.GetEventRecorderFor(instanceExpiration),
			NamespaceWhitelist:        nsWhitelist,
			StatusCheckRequestTimeout: *instanceInactiveTerminationStatusCheckTimeout,
			MailClient:                mailClient,
		}).SetupWithManager(mgr, *maxConcurrentInactiveTerminationReconciles); err != nil {
			log.Error(err, "unable to create controller", "controller", instanceExpiration)
			os.Exit(1)
		}
	} else {
		log.Info("Instance Expiration controller disabled.")
	}

	// Add readiness probe
	err = mgr.AddReadyzCheck("ready-ping", healthz.Ping)
	if err != nil {
		log.Error(err, "unable to add a readiness check")
		os.Exit(1)
	}

	// Add liveness probe
	err = mgr.AddHealthzCheck("health-ping", healthz.Ping)
	if err != nil {
		log.Error(err, "unable to add an health check")
		os.Exit(1)
	}

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}
}

// This method parses a string to get a map. The different labels should divided by a &.
func parseMap(raw string) map[string]string {
	ss := strings.Split(raw, "&")
	m := make(map[string]string)
	for _, pair := range ss {
		z := strings.Split(pair, "=")
		m[z[0]] = z[1]
	}
	return m
}
