/*

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
	"flag"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	virtv1 "kubevirt.io/client-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog/v2"
	cdiv1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	instance_controller "github.com/netgroup-polito/CrownLabs/operators/pkg/instance-controller"
	// +kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = crownlabsv1alpha1.AddToScheme(scheme)

	_ = crownlabsv1alpha2.AddToScheme(scheme)

	_ = virtv1.AddToScheme(scheme)

	_ = cdiv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var namespaceWhiteList string
	var webdavSecret string
	var websiteBaseURL string
	var nextcloudBaseURL string
	var oauth2ProxyImage string
	var oidcClientSecret string
	var oidcProviderURL string

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&namespaceWhiteList, "namespace-whitelist", "production=true", "The whitelist of the namespaces on "+
		"which the controller will work. Different labels (key=value) can be specified, by separating them with a &"+
		"( e.g. key1=value1&key2=value2")
	flag.StringVar(&websiteBaseURL, "website-base-url", "crownlabs.polito.it", "Base URL of crownlabs website instance")
	flag.StringVar(&nextcloudBaseURL, "nextcloud-base-url", "", "Base URL of NextCloud website to use")
	flag.StringVar(&webdavSecret, "webdav-secret-name", "webdav", "The name of the secret containing webdav credentials")
	flag.StringVar(&oauth2ProxyImage, "oauth2-proxy-image", "", "The docker image used for the oauth2-proxy deployment")
	flag.StringVar(&oidcClientSecret, "oidc-client-secret", "", "The oidc client secret used by oauth2-proxy")
	flag.StringVar(&oidcProviderURL, "oidc-provider-url", "", "The url of the oidc provider used by oauth2-proxy")
	klog.InitFlags(nil)
	flag.Parse()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		LeaderElection:         enableLeaderElection,
		Port:                   9443,
		HealthProbeBindAddress: ":8081",
		LivenessEndpointName:   "/healthz",
		ReadinessEndpointName:  "/ready",
	})
	if err != nil {
		klog.Fatal(err, "unable to start manager")
	}
	whiteListMap := parseMap(namespaceWhiteList)
	klog.Info("Reconciling only namespaces with the following labels: ")
	if err = (&instance_controller.InstanceReconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		EventsRecorder:     mgr.GetEventRecorderFor("InstanceOperator"),
		NamespaceWhitelist: metav1.LabelSelector{MatchLabels: whiteListMap, MatchExpressions: []metav1.LabelSelectorRequirement{}},
		NextcloudBaseURL:   nextcloudBaseURL,
		WebsiteBaseURL:     websiteBaseURL,
		WebdavSecretName:   webdavSecret,
		Oauth2ProxyImage:   oauth2ProxyImage,
		OidcClientSecret:   oidcClientSecret,
		OidcProviderURL:    oidcProviderURL,
	}).SetupWithManager(mgr); err != nil {
		klog.Fatal(err, "unable to create controller", "controller", "Instance")
	}

	// +kubebuilder:scaffold:builder
	// Add readiness probe
	err = mgr.AddReadyzCheck("ready-ping", healthz.Ping)
	if err != nil {
		klog.Error("Unable to add a readiness check")
		os.Exit(1)
	}

	// Add liveness probe
	err = mgr.AddHealthzCheck("health-ping", healthz.Ping)
	if err != nil {
		klog.Fatal("Unable add an health check")
	}
	klog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		klog.Fatal("Unable to start manager")
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
