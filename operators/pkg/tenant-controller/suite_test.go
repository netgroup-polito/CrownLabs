// Copyright 2020-2022 Politecnico di Torino
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

package tenant_controller

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	gocloak "github.com/Nerzal/gocloak/v7"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gomegaTypes "github.com/onsi/gomega/types"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/tenant-controller/mocks"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

const targetLabelKey = "reconcile"
const targetLabelValue = "true"

const kcAccessToken = "keycloak-token"
const kcTargetRealm = "targetRealm"
const kcTargetClientID = "targetClientId"

// keycloak variables.
var mKcClient *mocks.MockGoCloak
var mToken *gocloak.JWT = &gocloak.JWT{AccessToken: kcAccessToken}
var reqActions = []string{"UPDATE_PASSWORD", "VERIFY_EMAIL"}
var emailActionLifespan = 60 * 60 * 24 * 30
var kcA = KcActor{
	Client:                mKcClient,
	token:                 mToken,
	TargetRealm:           kcTargetRealm,
	TargetClientID:        kcTargetClientID,
	UserRequiredActions:   reqActions,
	EmailActionsLifeSpanS: emailActionLifespan,
}

// nextcloud variables.
var mNcA *mocks.NcHandlerMock

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "deploy", "crds")},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	Expect(crownlabsv1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(crownlabsv1alpha2.AddToScheme(scheme.Scheme)).To(Succeed())

	// +kubebuilder:scaffold:scheme

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		MetricsBindAddress: "0",
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&WorkspaceReconciler{
		Client:             k8sManager.GetClient(),
		Scheme:             k8sManager.GetScheme(),
		KcA:                &kcA,
		TargetLabelKey:     targetLabelKey,
		TargetLabelValue:   targetLabelValue,
		ReconcileDeferHook: GinkgoRecover,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	err = (&TenantReconciler{
		Client:             k8sManager.GetClient(),
		Scheme:             k8sManager.GetScheme(),
		KcA:                &kcA,
		NcA:                mNcA,
		TargetLabelKey:     targetLabelKey,
		TargetLabelValue:   targetLabelValue,
		ReconcileDeferHook: GinkgoRecover,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

func doesEventuallyExists(ctx context.Context, objLookupKey types.NamespacedName, targetObj client.Object, expectedStatus gomegaTypes.GomegaMatcher, timeout, interval time.Duration) {
	Eventually(func() bool {
		err := k8sClient.Get(ctx, objLookupKey, targetObj)
		return err == nil
	}, timeout, interval).Should(expectedStatus)
}
