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

// Package tenant_test implements tenant controller tests.
package tenant_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomegaTypes "github.com/onsi/gomega/types"
	"go.uber.org/mock/gomock"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/mock"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/tenant"
)

const (
	timeout  = time.Second * 10
	interval = time.Millisecond * 250
	tnName   = "testuser"
)

var (
	ctx              context.Context
	log              logr.Logger
	builder          fake.ClientBuilder
	cl               client.Client
	mockCtrl         *gomock.Controller
	keycloakActor    *mock.MockKeycloakActorIface
	tenantReconciler tenant.Reconciler

	tnResource             *v1alpha2.Tenant
	tnReconcileErrExpected gomegaTypes.GomegaMatcher

	objects []client.Object

	runReconcile = true
)

func TestTenant(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tenant Controller Suite")
}

var _ = BeforeSuite(func() {
	Expect(v1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
	Expect(v1alpha2.AddToScheme(scheme.Scheme)).To(Succeed())
})

var _ = BeforeEach(func() {
	log = logr.New(GinkgoLogWriter{})
	ctx = ctrl.LoggerInto(context.Background(), log)
	builder = *fake.NewClientBuilder().WithScheme(scheme.Scheme)

	mockCtrl = gomock.NewController(GinkgoT())
	keycloakActor = mock.NewMockKeycloakActorIface(mockCtrl)

	keycloakActor.EXPECT().IsInitialized().Return(false).AnyTimes()

	tnResource = &v1alpha2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: tnName,
			Labels: map[string]string{
				"crownlabs.polito.it/operator-selector": "test",
			},
		},
		Spec: v1alpha2.TenantSpec{
			FirstName: "Test",
			LastName:  "Tenant",
			Email:     "test@tenant.example",
			LastLogin: metav1.Now(),
		},
	}
	tnReconcileErrExpected = Not(HaveOccurred())
})

var _ = AfterEach(func() {
	mockCtrl.Finish()

	removeObjFromObjectsList(tnResource)
})

var _ = JustBeforeEach(func() {
	addObjToObjectsList(tnResource)
	cl = builder.WithObjects(objects...).WithStatusSubresource(objects...).Build()

	tenantReconciler = tenant.Reconciler{
		Client:                      cl,
		Scheme:                      scheme.Scheme,
		KeycloakActor:               keycloakActor,
		TargetLabel:                 common.NewLabel("crownlabs.polito.it/operator-selector", "test"),
		TenantNSKeepAlive:           24 * time.Hour,
		WaitUserVerification:        true,
		SandboxClusterRole:          "test-sandbox-editor",
		BaseWorkspaces:              []string{"base-ws1"},
		MyDrivePVCsNamespace:        "mydrive-pvcs",
		MyDrivePVCsSize:             resource.MustParse("5Gi"),
		MyDrivePVCsStorageClassName: "nfs",
	}

	if !runReconcile {
		return
	}

	_, err := tenantReconciler.Reconcile(ctx, reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: tnResource.Name,
		},
	})
	Expect(err).To(tnReconcileErrExpected)
})

func DoesEventuallyExists(ctx context.Context, cl client.Client, objLookupKey client.ObjectKey, targetObj client.Object, expectedStatus gomegaTypes.GomegaMatcher, timeout, interval time.Duration) {
	Eventually(func() bool {
		err := cl.Get(ctx, objLookupKey, targetObj)
		return err == nil
	}, timeout, interval).Should(expectedStatus)
}

func addObjToObjectsList(obj client.Object) {
	for _, o := range objects {
		if o.GetName() == obj.GetName() && o.GetNamespace() == obj.GetNamespace() {
			return // Object already exists in the list
		}
	}
	objects = append(objects, obj)
}

func removeObjFromObjectsList(obj client.Object) {
	for i, o := range objects {
		if o.GetName() == obj.GetName() && o.GetNamespace() == obj.GetNamespace() {
			objects = append(objects[:i], objects[i+1:]...)
			return
		}
	}
}

func tenantBeingDeleted() {
	tnResource.Finalizers = append(tnResource.Finalizers, v1alpha2.TnOperatorFinalizerName)
	tnResource.DeletionTimestamp = &metav1.Time{Time: time.Now().Add(10 * time.Second)}
}

// GinkgoLogWriter implements logr.LogSink.
type GinkgoLogWriter struct {
	kandv string
}

// Init initializes the logger with runtime information.
func (w GinkgoLogWriter) Init(_ logr.RuntimeInfo) {}

// Enabled returns whether logging is enabled at the specified level.
func (w GinkgoLogWriter) Enabled(_ int) bool { return true }

// Info logs an info message with the given key-value pairs.
func (w GinkgoLogWriter) Info(_ int, msg string, keysAndValues ...interface{}) {
	GinkgoWriter.Printf("%s -- %v %v\n", msg, w.kandv, keysAndValues)
}

// Error logs an error message with the given key-value pairs.
func (w GinkgoLogWriter) Error(err error, msg string, keysAndValues ...interface{}) {
	GinkgoWriter.Printf("ERROR: %s -- %v -- %v\n", msg, err, keysAndValues)
}

// WithValues returns a new LogSink with the specified key-value pairs.
func (w GinkgoLogWriter) WithValues(kv ...interface{}) logr.LogSink {
	w.kandv += fmt.Sprintf("%v", kv)
	return w
}

// WithName returns a new LogSink with the specified name.
func (w GinkgoLogWriter) WithName(_ string) logr.LogSink { return w }
