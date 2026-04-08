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

// Package main contains the entrypoint for the instance automation operator.
package main

import (
	"flag"
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
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/mail"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/restcfg"
)

var (
	scheme     = runtime.NewScheme()
	mailClient *mail.Client
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

	prometheusURL := flag.String("monitoring-prometheus-url", "http://kube-prometheus-stack-prometheus.monitoring:9090", "The URL of the Prometheus instance to use for the Inactive Termination")
	prometheusNginxAvailability := flag.String("monitoring-nginx-availability", `count(up{service="ingress-nginx-external-controller-metrics"})`, "Prometheus Query to understand if Nginx Metrics are available in Prometheus.")
	prometheusBastionSSHAvailability := flag.String("monitoring-bastion-ssh-availability", `count(up{container="bastion-operator-tracker-sidecar"})`, "Prometheus Query to understand if SSH (custom metric) Metrics are available in Prometheus.")
	prometheusWebSSHAvailability := flag.String("monitoring-web-ssh-availability", `count(up{container="webssh"})`, "Prometheus Query to understand if WebSSH (custom metric) Metrics are available in Prometheus.")
	prometheusNginxData := flag.String("monitoring-nginx-data", `nginx_ingress_controller_requests{exported_namespace=%q, exported_service=%q}`, "Prometheus Query to retrieve metrics about the last (frontend) access to a specific instance.")
	prometheusBastionSSHData := flag.String("monitoring-bastion-ssh-data", `bastion_ssh_connections{destination_ip=%q}`, "Prometheus Query to retrieve metrics about the last (SSH) access to a specific instance.")
	prometheusWebSSHData := flag.String("monitoring-web-ssh-data", `bastion_web_ssh_connections{destination_ip=%q}`, "Prometheus Query to retrieve metrics about the last (WebSSH) access to a specific instance.")
	queryStep := flag.Duration("prometheus-query-step", 5*time.Minute, "The step to use when querying range data from Prometheus.")

	instanceTerminationStatusCheckTimeout := flag.Duration("instance-termination-status-check-timeout", 3*time.Second, "The maximum time to wait for the status check for Instances that require it")
	instanceTerminationStatusCheckInterval := flag.Duration("instance-termination-status-check-interval", 24*time.Hour, "The interval to check the status of Instances that require it")

	maxConcurrentTerminationReconciles := flag.Int("max-concurrent-reconciles-termination", 1, "The maximum number of concurrent Reconciles which can be run for the Instance Termination controller")
	maxConcurrentSubmissionReconciles := flag.Int("max-concurrent-reconciles-submission", 1, "The maximum number of concurrent Reconciles which can be run for the Instance Submission controller")
	maxConcurrentInactiveTerminationReconciles := flag.Int("max-concurrent-reconciles-inactive-termination", 1, "The maximum number of concurrent Reconciles which can be run for the Instance Inactive Termination controller")
	maxConcurrentExpirationReconciles := flag.Int("max-concurrent-reconciles-expiration", 1, "The maximum number of concurrent Reconciles which can be run for the Instance Expiration controller")

	instanceInactiveTerminationStatusCheckTimeout := flag.Duration("instance-inactive-termination-status-check-timeout", 5*time.Second, "The maximum time to wait for the status check for Instances that require it")
	instanceInactiveTerminationMaxNumberOfAlerts := flag.Int("instance-inactive-termination-max-number-of-alerts", 3, "the maximum number of notification that Crownlabs can send before stopping/deleting the Instance. It can be overrided by the AlertAnnotationNum annotation that can be in the Template resource.")
	instanceInactiveTerminationNotificationInterval := flag.Duration("instance-inactive-termination-notification-interval", 24*time.Hour, "It represent how long before the instance is deleted the notification email should be sent to the user.")
	expirationNotificationInterval := flag.Duration("expiration-notification-interval", 24*time.Hour, "It represent how long before the instance is deleted the notification email should be sent to the user.")

	marginTime := flag.Duration("margin-time", 1*time.Minute, "The margin time to add to operations involving time comparisons to avoid edge cases due to delays")

	flag.StringVar(&containerEnvOpts.ImagesTag, "container-env-sidecars-tag", "latest", "The tag for service containers (such as gui sidecar containers)")
	flag.StringVar(&containerEnvOpts.ContentUploaderImg, "container-env-content-uploader-img", "latest", "The image name for the job to compress and upload instance content from a persistent instance.")

	enableInactivityNotifications := flag.Bool("enable-inactivity-notifications", false, "Enable the sending of inactivity notifications to users on instance inactivity")
	enableExpirationNotifications := flag.Bool("enable-expiration-notifications", false, "Enable the sending of expiration notifications to users on instance expiration")

	mailTemplateDir := flag.String("mail-template-dir", "/etc/crownmail/templates", "The directory containing email templates and configuration (typically through a mounted ConfigMap)")
	mailConfigDir := flag.String("mail-config-dir", "/etc/crownmail/configs", "The directory containing email configuration (typically through a mounted Secret)")

	restcfg.InitFlags(nil)
	klog.InitFlags(nil)
	flag.Parse()

	ctrl.SetLogger(textlogger.NewLogger(textlogger.NewConfig()))

	log := ctrl.Log.WithName("setup")

	whiteListMap := parseMap(*namespaceWhiteList)
	log.Info("restricting reconciled namespaces", "labels", *namespaceWhiteList)

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

	mailClient, err = mail.NewMailClientFromFilesystem(*mailConfigDir, *mailTemplateDir)
	if err != nil {
		log.Error(err, "unable to create mail client from filesystem", "templateDir", *mailTemplateDir)
		os.Exit(1)
	}
	log.Info("CrownLabs Email client created", "templateDir", *mailTemplateDir)

	prometheus, err := instautoctrl.NewPrometheusObj(*prometheusURL, *prometheusNginxAvailability, *prometheusBastionSSHAvailability, *prometheusWebSSHAvailability,
		*prometheusNginxData, *prometheusBastionSSHData, *prometheusWebSSHData, *queryStep)

	if err != nil {
		log.Error(err, "unable to create Prometheus client")
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
			Client:                        mgr.GetClient(),
			Scheme:                        mgr.GetScheme(),
			EventsRecorder:                mgr.GetEventRecorderFor(instanceInactiveTermination),
			NamespaceWhitelist:            nsWhitelist,
			StatusCheckRequestTimeout:     *instanceInactiveTerminationStatusCheckTimeout,
			InstanceMaxNumberOfAlerts:     *instanceInactiveTerminationMaxNumberOfAlerts,
			EnableInactivityNotifications: *enableInactivityNotifications,
			MailClient:                    mailClient,
			Prometheus:                    prometheus,
			NotificationInterval:          *instanceInactiveTerminationNotificationInterval,
			MarginTime:                    *marginTime,
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
			Client:                        mgr.GetClient(),
			Scheme:                        mgr.GetScheme(),
			EventsRecorder:                mgr.GetEventRecorderFor(instanceExpiration),
			NamespaceWhitelist:            nsWhitelist,
			EnableExpirationNotifications: *enableExpirationNotifications,
			MailClient:                    mailClient,
			NotificationInterval:          *expirationNotificationInterval,
			MarginTime:                    *marginTime,
		}).SetupWithManager(mgr, *maxConcurrentExpirationReconciles); err != nil {
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
