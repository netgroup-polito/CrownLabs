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

package instautoctrl_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomegaTypes "github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	virtv1 "kubevirt.io/api/core/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instautoctrl"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instautoctrl/mocks"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/tests"
)

var ctx context.Context
var cancel context.CancelFunc
var cfg *rest.Config
var k8sClient client.Client
var k8sClientExpiration client.Client
var testEnv *envtest.Environment
var mockCtrl *gomock.Controller
var mockProm *mocks.MockPrometheusClientInterface

// var log logr.Logger

func TestInstautoctrl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Instautoctrl Suite")
}

var _ = BeforeSuite(func() {
	ctx, cancel = context.WithCancel(context.Background())
	tests.LogsToGinkgoWriter()
	// opts := zap.Options{
	// 	Development: true,
	// }
	// opts.BindFlags(flag.CommandLine)
	// log = zap.New(zap.UseFlagOptions(&opts))
	// ctrl.SetLogger(log)

	By("bootstrapping test environment")

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "deploy", "crds"),
			filepath.Join("..", "..", "..", "tests", "crds")},
		ErrorIfCRDPathMissing: true,
	}
	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	Expect(crownlabsv1alpha2.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())
	Expect(virtv1.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())
	Expect(cdiv1beta1.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:  scheme.Scheme,
		Metrics: server.Options{BindAddress: "0"},
	})
	Expect(err).ToNot(HaveOccurred())

	// Generate whitelist map for InstanceSnapshot controller reconciliation
	whiteListMap := map[string]string{
		"crownlabs.polito.it/operator-selector": "test-suite",
	}

	mockCtrl = gomock.NewController(GinkgoT())
	mockProm = mocks.NewMockPrometheusClientInterface(mockCtrl)

	err = (&instautoctrl.InstanceInactiveTerminationReconciler{
		Client:                        k8sManager.GetClient(),
		Scheme:                        k8sManager.GetScheme(),
		EventsRecorder:                k8sManager.GetEventRecorderFor("instance-termination"),
		NamespaceWhitelist:            metav1.LabelSelector{MatchLabels: whiteListMap, MatchExpressions: []metav1.LabelSelectorRequirement{}},
		MailClient:                    nil,
		Prometheus:                    mockProm,
		InstanceMaxNumberOfAlerts:     3,
		NotificationInterval:          1 * time.Second,
		EnableInactivityNotifications: false,
		StatusCheckRequestTimeout:     10 * time.Second,
	}).SetupWithManager(k8sManager, 1)
	Expect(err).ToNot(HaveOccurred())

	err = (&instautoctrl.InstanceExpirationReconciler{
		Client:                        k8sManager.GetClient(),
		Scheme:                        k8sManager.GetScheme(),
		EventsRecorder:                k8sManager.GetEventRecorderFor("instance-expiration"),
		NamespaceWhitelist:            metav1.LabelSelector{MatchLabels: whiteListMap, MatchExpressions: []metav1.LabelSelectorRequirement{}},
		MailClient:                    nil,
		NotificationInterval:          1 * time.Second,
		EnableExpirationNotifications: false,
	}).SetupWithManager(k8sManager, 1)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	k8sClientExpiration = k8sManager.GetClient()
	Expect(k8sClientExpiration).ToNot(BeNil())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

func doesEventuallyExists(ctx context.Context, objLookupKey types.NamespacedName, targetObj client.Object, expectedStatus gomegaTypes.GomegaMatcher, timeout, interval time.Duration, k8sClient client.Client) {
	Eventually(func() bool {
		err := k8sClient.Get(ctx, objLookupKey, targetObj)
		return err == nil
	}, timeout, interval).Should(expectedStatus)
}
