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

package instancesnapshot_controller_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomegaTypes "github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	instancesnapshot_controller "github.com/netgroup-polito/CrownLabs/operators/pkg/instancesnapshot-controller"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/tests"
)

var ctx context.Context
var cancel context.CancelFunc
var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Instance Snapshot Controller Suite")
}

var _ = BeforeSuite(func() {
	ctx, cancel = context.WithCancel(context.Background())
	tests.LogsToGinkgoWriter()

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "deploy", "crds"),
			filepath.Join("..", "..", "tests", "crds")},
	}
	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = crownlabsv1alpha2.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:  scheme.Scheme,
		Metrics: server.Options{BindAddress: "0"},
	})
	Expect(err).ToNot(HaveOccurred())

	// Generate whitelist map for InstanceSnapshot controller reconciliation
	whiteListMap := map[string]string{
		"test-suite": "true",
	}

	err = (&instancesnapshot_controller.InstanceSnapshotReconciler{
		Client:             k8sManager.GetClient(),
		Scheme:             k8sManager.GetScheme(),
		EventsRecorder:     k8sManager.GetEventRecorderFor("instance-snapshot"),
		NamespaceWhitelist: metav1.LabelSelector{MatchLabels: whiteListMap, MatchExpressions: []metav1.LabelSelectorRequirement{}},
		ContainersSnapshot: instancesnapshot_controller.ContainersSnapshotOpts{
			VMRegistry:         "my-registry",
			RegistrySecretName: "kaniko-secret",
			ContainerKaniko:    "kaniko",
			ContainerImgExport: "crownlabs/img-export",
		},
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

func doesEventuallyExists(ctx context.Context, objLookupKey types.NamespacedName, targetObj client.Object, expectedStatus gomegaTypes.GomegaMatcher, timeout, interval time.Duration) {
	Eventually(func() bool {
		err := k8sClient.Get(ctx, objLookupKey, targetObj)
		return err == nil
	}, timeout, interval).Should(expectedStatus)
}
