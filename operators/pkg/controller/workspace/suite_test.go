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

package workspace_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/mock"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/workspace"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomegaTypes "github.com/onsi/gomega/types"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	timeout      = time.Second * 10
	interval     = time.Millisecond * 250
	wsName       = "test-workspace"
	wsPrettyName = "Test Workspace"
)

var (
	ctx                 context.Context
	builder             fake.ClientBuilder
	cl                  client.Client
	mockCtrl            *gomock.Controller
	keycloakActor       *mock.MockKeycloakActorIface
	workspaceReconciler workspace.WorkspaceReconciler

	wsResource             *v1alpha1.Workspace
	wsReconcileErrExpected gomegaTypes.GomegaMatcher
)

func TestWorkspace(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workspace Controller Suite")
}

var _ = BeforeSuite(func() {
	Expect(v1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(v1alpha2.AddToScheme(scheme.Scheme)).To(Succeed())
})

var _ = BeforeEach(func() {
	ctx = ctrl.LoggerInto(context.Background(), logr.Discard())
	builder = *fake.NewClientBuilder().WithScheme(scheme.Scheme)

	mockCtrl = gomock.NewController(GinkgoT())
	keycloakActor = mock.NewMockKeycloakActorIface(mockCtrl)

	keycloakActor.EXPECT().IsInitialized().Return(false).AnyTimes()

	wsResource = &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: wsName,
			Labels: map[string]string{
				"crownlabs.polito.it/target": "test",
			},
		},
		Spec: v1alpha1.WorkspaceSpec{
			PrettyName: wsPrettyName,
		},
	}
	wsReconcileErrExpected = Not(HaveOccurred())
})

var _ = AfterEach(func() {
	mockCtrl.Finish()
})

var _ = JustBeforeEach(func() {
	cl = builder.WithObjects(wsResource).WithStatusSubresource(wsResource).Build()

	workspaceReconciler = workspace.WorkspaceReconciler{
		Client:        cl,
		Scheme:        scheme.Scheme,
		KeycloakActor: keycloakActor,
		TargetLabel:   common.NewLabel("crownlabs.polito.it/target", "test"),
	}

	_, err := workspaceReconciler.Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: wsName,
		},
	})
	Expect(err).To(wsReconcileErrExpected)
})
