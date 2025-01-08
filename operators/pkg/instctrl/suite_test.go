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

package instctrl_test

import (
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2/textlogger"
	virtv1 "kubevirt.io/api/core/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instctrl"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/tests"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Instance Controller Suite")
}

var (
	instanceReconciler instctrl.InstanceReconciler
	k8sClient          client.Client
	testEnv            = envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "deploy", "crds"),
			filepath.Join("..", "..", "tests", "crds"),
		},
		ErrorIfCRDPathMissing: true,
	}
	whiteListMap = map[string]string{"production": "true"}
)

var _ = BeforeSuite(func() {
	tests.LogsToGinkgoWriter()

	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	Expect(clv1alpha2.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())
	Expect(clv1alpha1.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())
	Expect(virtv1.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())
	Expect(cdiv1beta1.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())

	ctrl.SetLogger(textlogger.NewLogger(textlogger.NewConfig()))

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	instanceReconciler = instctrl.InstanceReconciler{
		Client:             k8sClient,
		Scheme:             scheme.Scheme,
		EventsRecorder:     record.NewFakeRecorder(1024),
		NamespaceWhitelist: metav1.LabelSelector{MatchLabels: whiteListMap},
		ReconcileDeferHook: GinkgoRecover,
		ServiceUrls: instctrl.ServiceUrls{
			WebsiteBaseURL:   "fakesite.com",
			InstancesAuthURL: "fake.com/auth",
		},
		ContainerEnvOpts: forge.ContainerEnvOpts{
			ImagesTag:            "v0.1.2",
			XVncImg:              "fake-xvnc",
			WebsockifyImg:        "fake-wskfy",
			ContentDownloaderImg: "fake-archdl",
		},
	}
})

var _ = AfterSuite(func() {
	Expect(testEnv.Stop()).To(Succeed())
})
