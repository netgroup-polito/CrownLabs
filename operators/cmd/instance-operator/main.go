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
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	virtv1 "kubevirt.io/client-go/api/v1"
	cdiv1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	instance_controller "github.com/netgroup-polito/CrownLabs/operators/pkg/instance-controller"
	instancesnapshot_controller "github.com/netgroup-polito/CrownLabs/operators/pkg/instancesnapshot-controller"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(crownlabsv1alpha1.AddToScheme(scheme))
	utilruntime.Must(crownlabsv1alpha2.AddToScheme(scheme))

	utilruntime.Must(virtv1.AddToScheme(scheme))
	utilruntime.Must(cdiv1.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var namespaceWhiteList string
	var webdavSecret string
	var websiteBaseURL string
	var nextcloudBaseURL string
	var instancesAuthURL string
	var containerEnvSidecarsTag string
	var containerEnvVncImg string
	var containerEnvWebsockifyImg string
	var containerEnvNovncImg string
	var vmRegistry string
	var vmRegistrySecret string
	var containerImgExport string
	var containerKaniko string
	var containerEnvFileBrowserImg string
	var containerEnvFileBrowserImgTag string
	var maxConcurrentReconciles int

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&namespaceWhiteList, "namespace-whitelist", "production=true", "The whitelist of the namespaces on "+
		"which the controller will work. Different labels (key=value) can be specified, by separating them with a &"+
		"( e.g. key1=value1&key2=value2")
	flag.StringVar(&websiteBaseURL, "website-base-url", "crownlabs.polito.it", "Base URL of crownlabs website instance")
	flag.StringVar(&nextcloudBaseURL, "nextcloud-base-url", "", "Base URL of NextCloud website to use")
	flag.StringVar(&instancesAuthURL, "instances-auth-url", "", "The base URL for user instances authentication (i.e., oauth2-proxy)")
	flag.StringVar(&webdavSecret, "webdav-secret-name", "webdav", "The name of the secret containing webdav credentials")

	flag.StringVar(&containerEnvSidecarsTag, "container-env-sidecars-tag", "latest", "The tag for service containers (such as gui sidecar containers)")
	flag.StringVar(&containerEnvVncImg, "container-env-vnc-img", "crownlabs/tigervnc", "The image name for the vnc image (sidecar for graphical container environment)")
	flag.StringVar(&containerEnvWebsockifyImg, "container-env-websockify-img", "crownlabs/websockify", "The image name for the websockify image (sidecar for graphical container environment)")
	flag.StringVar(&containerEnvNovncImg, "container-env-novnc-img", "crownlabs/novnc", "The image name for the novnc image (sidecar for graphical container environment)")

	flag.StringVar(&vmRegistry, "vm-registry", "", "The registry where VMs should be uploaded")
	flag.StringVar(&vmRegistrySecret, "vm-registry-secret", "", "The name of the secret for the VM registry")

	flag.StringVar(&containerImgExport, "container-export-img", "crownlabs/img-exporter", "The image for the img-exporter (container in charge of exporting the disk of a persistent vm)")
	flag.StringVar(&containerKaniko, "container-kaniko-img", "gcr.io/kaniko-project/executor", "The image for the Kaniko container to be deployed")
	flag.StringVar(&containerEnvFileBrowserImg, "container-env-filebrowser-img", "filebrowser/filebrowser", "The image name for the filebrowser image (sidecar for gui-based file manager)")
	flag.StringVar(&containerEnvFileBrowserImgTag, "container-env-filebrowser-img-tag", "latest", "The tag for the FileBrowser container (the gui-based file manager)")

	flag.IntVar(&maxConcurrentReconciles, "max-concurrent-reconciles", 1, "The maximum number of concurrent Reconciles which can be run")

	klog.InitFlags(nil)
	flag.Parse()

	if !klog.V(5).Enabled() {
		klog.SetLogFilter(utils.LogShortenerFilter{})
	}
	ctrl.SetLogger(klogr.NewWithOptions())

	log := ctrl.Log.WithName("setup")

	whiteListMap := parseMap(namespaceWhiteList)
	log.Info("restricting reconciled namespaces", "labels", namespaceWhiteList)

	// Configure the manager
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		LeaderElection:         enableLeaderElection,
		HealthProbeBindAddress: ":8081",
		LivenessEndpointName:   "/healthz",
		ReadinessEndpointName:  "/ready",
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Configure the Instance controller
	instanceCtrlName := "Instance"
	if err = (&instance_controller.InstanceReconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		EventsRecorder:     mgr.GetEventRecorderFor(instanceCtrlName),
		NamespaceWhitelist: metav1.LabelSelector{MatchLabels: whiteListMap, MatchExpressions: []metav1.LabelSelectorRequirement{}},
		NextcloudBaseURL:   nextcloudBaseURL,
		WebsiteBaseURL:     websiteBaseURL,
		WebdavSecretName:   webdavSecret,
		InstancesAuthURL:   instancesAuthURL,
		ContainerEnvOpts: instance_controller.ContainerEnvOpts{
			ImagesTag:         containerEnvSidecarsTag,
			VncImg:            containerEnvVncImg,
			WebsockifyImg:     containerEnvWebsockifyImg,
			NovncImg:          containerEnvNovncImg,
			FileBrowserImg:    containerEnvFileBrowserImg,
			FileBrowserImgTag: containerEnvFileBrowserImgTag,
		},
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", instanceCtrlName)
		os.Exit(1)
	}

	// Configure the InstanceSnapshot controller
	instanceSnapshotCtrl := "InstanceSnapshot"
	if err = (&instancesnapshot_controller.InstanceSnapshotReconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		EventsRecorder:     mgr.GetEventRecorderFor(instanceSnapshotCtrl),
		NamespaceWhitelist: metav1.LabelSelector{MatchLabels: whiteListMap, MatchExpressions: []metav1.LabelSelectorRequirement{}},
		VMRegistry:         vmRegistry,
		RegistrySecretName: vmRegistrySecret,
		ContainersSnapshot: instancesnapshot_controller.ContainersSnapshotOpts{
			ContainerKaniko:    containerKaniko,
			ContainerImgExport: containerImgExport,
		},
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", instanceSnapshotCtrl)
		os.Exit(1)
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
