// Copyright 2020-2021 Politecnico di Torino
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

package instance_controller_test

import (
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2/klogr"
	virtv1 "kubevirt.io/client-go/api/v1"
	cdiv1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	instance_controller "github.com/netgroup-polito/CrownLabs/operators/pkg/instance-controller"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var (
	instanceReconciler instance_controller.InstanceReconciler
	k8sClient          client.Client
	testEnv            = envtest.Environment{CRDDirectoryPaths: []string{
		filepath.Join("..", "..", "deploy", "crds"),
		filepath.Join("..", "..", "tests", "crds"),
	}}
	whiteListMap = map[string]string{"production": "true"}
)

const webdavSecretName = "webdav-secret"

var _ = BeforeSuite(func() {
	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	Expect(clv1alpha2.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(clv1alpha1.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())
	Expect(virtv1.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())
	Expect(cdiv1.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())

	ctrl.SetLogger(klogr.NewWithOptions())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())

	instanceReconciler = instance_controller.InstanceReconciler{
		Client:             k8sClient,
		Scheme:             scheme.Scheme,
		EventsRecorder:     record.NewFakeRecorder(1024),
		NamespaceWhitelist: metav1.LabelSelector{MatchLabels: whiteListMap},
		WebdavSecretName:   webdavSecretName,
		ReconcileDeferHook: GinkgoRecover,
		ServiceUrls: instance_controller.ServiceUrls{
			NextcloudBaseURL: "fake.com",
			WebsiteBaseURL:   "fakesite.com",
			InstancesAuthURL: "fake.com/auth",
		},
		ContainerEnvOpts: forge.ContainerEnvOpts{
			ImagesTag:            "v0.1.2",
			XVncImg:              "fake-xvnc",
			WebsockifyImg:        "fake-wskfy",
			MyDriveImgAndTag:     "fake-filebrowser",
			ContentDownloaderImg: "fake-archdl",
		},
	}
})

var _ = AfterSuite(func() {
	Expect(testEnv.Stop()).To(Succeed())
})
