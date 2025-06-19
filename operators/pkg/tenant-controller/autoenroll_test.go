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

package tenant_controller_test

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gomegaTypes "github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	tntctrl "github.com/netgroup-polito/CrownLabs/operators/pkg/tenant-controller"
)

var _ = Describe("Workspace AutoEnroll Update", func() {
	var (
		ctx          context.Context
		builder      fake.ClientBuilder
		cl           client.Client
		tnReconciler tntctrl.TenantReconciler
		wsReconciler tntctrl.WorkspaceReconciler
		tenant       clv1alpha2.Tenant
		workspace    clv1alpha1.Workspace
	)

	const (
		tenantName    = "tester"
		workspaceName = "foo"

		labelkey   = "crownlabs.polito.it/operator-selector"
		labelvalue = "autoenroll-test"
	)

	BeforeEach(func() {
		ctx = ctrl.LoggerInto(context.Background(), logr.Discard())
		builder = *fake.NewClientBuilder().WithScheme(scheme.Scheme)

		workspace = clv1alpha1.Workspace{ObjectMeta: metav1.ObjectMeta{Name: workspaceName}}
		workspace.Spec.AutoEnroll = clv1alpha1.AutoenrollWithApproval
		tenant = clv1alpha2.Tenant{
			ObjectMeta: metav1.ObjectMeta{Name: tenantName},
			Spec: clv1alpha2.TenantSpec{
				FirstName: "mario",
				LastName:  "rossi",
				Email:     "mariorossi@email.com",
			},
		}
		tenant.Labels = map[string]string{labelkey: labelvalue}
		workspace.Labels = map[string]string{labelkey: labelvalue}
		tenant.Spec.Workspaces = []clv1alpha2.TenantWorkspaceEntry{{Name: workspaceName, Role: clv1alpha2.Candidate}}
	})

	JustBeforeEach(func() {
		cl = builder.WithObjects(&tenant, &workspace).WithStatusSubresource(&tenant, &workspace).Build()
		tnReconciler = tntctrl.TenantReconciler{
			Client: cl, Scheme: scheme.Scheme,
			TargetLabelKey: labelkey, TargetLabelValue: labelvalue,
		}
		wsReconciler = tntctrl.WorkspaceReconciler{
			Client: cl, Scheme: scheme.Scheme,
			TargetLabelKey: labelkey, TargetLabelValue: labelvalue,
		}
		_, err := tnReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: tenant.Name}})
		Expect(err).ToNot(HaveOccurred())

	})

	When("the AutoEnroll has been set to Immediate", func() {
		It("should update candidate tenants to user", func() {
			By("Updating the workspace")
			workspace.Spec.AutoEnroll = clv1alpha1.AutoenrollImmediate
			err := cl.Update(ctx, &workspace)
			Expect(err).ToNot(HaveOccurred())
			_, err = wsReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: workspace.Name}})
			Expect(err).ToNot(HaveOccurred())

			By("Checking that the tenant has been updated")
			tn := clv1alpha2.Tenant{}
			doesEventuallyExists(ctx, cl, client.ObjectKey{Name: tenantName}, &tn, BeTrue(), time.Second*10, time.Millisecond*250)

			Expect(tn.Spec.Workspaces[0].Role).To(Equal(clv1alpha2.User))
		})
	})

	When("the AutoEnroll has been set to None", func() {
		It("should update user and remove workspace", func() {
			By("Updating the workspace")
			workspace.Spec.AutoEnroll = clv1alpha1.AutoenrollNone
			err := cl.Update(ctx, &workspace)
			Expect(err).ToNot(HaveOccurred())
			_, err = wsReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: workspace.Name}})
			Expect(err).ToNot(HaveOccurred())

			By("Checking that the tenant has been updated")
			tn := clv1alpha2.Tenant{}
			doesEventuallyExists(ctx, cl, client.ObjectKey{Name: tenantName}, &tn, BeTrue(), time.Second*10, time.Millisecond*250)
			Expect(tn.Spec.Workspaces).To(BeEmpty())
		})
	})
})

func doesEventuallyExists(ctx context.Context, cl client.Client, objLookupKey client.ObjectKey, targetObj client.Object, expectedStatus gomegaTypes.GomegaMatcher, timeout, interval time.Duration) {
	Eventually(func() bool {
		err := cl.Get(ctx, objLookupKey, targetObj)
		return err == nil
	}, timeout, interval).Should(expectedStatus)
}
