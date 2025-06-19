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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	tntctrl "github.com/netgroup-polito/CrownLabs/operators/pkg/tenant-controller"
	. "github.com/netgroup-polito/CrownLabs/operators/pkg/utils/tests"
)

var _ = Describe("Automatic namespace deletion", func() {
	var (
		ctx     context.Context
		builder fake.ClientBuilder
		cl      client.Client
		tenant  clv1alpha2.Tenant

		ns       corev1.Namespace
		instance clv1alpha2.Instance
		err      error
	)

	const (
		tenantName = "tester"
	)

	BeforeEach(func() {
		ctx = ctrl.LoggerInto(context.Background(), logr.Discard())
		builder = *fake.NewClientBuilder().WithScheme(scheme.Scheme)

		tenant = clv1alpha2.Tenant{ObjectMeta: metav1.ObjectMeta{Name: tenantName}}
		ns = corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "tenant-" + tenantName}}
		instance = clv1alpha2.Instance{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: ns.Name}}
	})

	JustBeforeEach(func() {
		cl = builder.WithObjects(&tenant).WithStatusSubresource(&tenant).Build()
		reconciler := tntctrl.TenantReconciler{
			Client: cl, Scheme: scheme.Scheme, TenantNSKeepAlive: 1 * time.Hour,
			RequeueTimeMinimum: 1 + time.Hour, RequeueTimeMaximum: 2 * time.Hour,
		}
		_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: tenant.Name}})
	})

	Context("the LastLogin flag has been refreshed recently", func() {
		BeforeEach(func() {
			tenant.Spec.LastLogin = metav1.Now()
		})

		When("the tenant namespace does not yet exist", func() {
			It("should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

			It("should enforce that the namespace is present", func() {
				Expect(cl.Get(ctx, types.NamespacedName{Name: ns.Name}, &ns)).ToNot(HaveOccurred())
			})
		})

		When("the tenant namespace does already exist", func() {
			BeforeEach(func() { builder.WithObjects(&ns) })

			It("should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

			It("should enforce that the namespace is present", func() {
				Expect(cl.Get(ctx, types.NamespacedName{Name: ns.Name}, &ns)).ToNot(HaveOccurred())
			})
		})
	})

	Context("the LastLogin flag has not been refreshed recently", func() {
		BeforeEach(func() {
			tenant.Spec.LastLogin = metav1.Time{Time: time.Now().Add(-2 * time.Hour)}
		})

		When("the tenant namespace does not yet exist", func() {
			It("should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

			It("should enforce that the namespace is not present", func() {
				Expect(cl.Get(ctx, types.NamespacedName{Name: ns.Name}, &ns)).To(FailBecauseNotFound())
			})
		})

		When("the tenant namespace does already exist", func() {
			BeforeEach(func() { builder.WithObjects(&ns) })

			When("it contains no instances", func() {
				It("should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

				It("should enforce that the namespace is not present", func() {
					Expect(cl.Get(ctx, types.NamespacedName{Name: ns.Name}, &ns)).To(FailBecauseNotFound())
				})
			})

			When("it contains an instance", func() {
				BeforeEach(func() { builder.WithObjects(&instance) })

				It("should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })

				It("should enforce that the namespace is still present", func() {
					Expect(cl.Get(ctx, types.NamespacedName{Name: ns.Name}, &ns)).ToNot(HaveOccurred())
				})
			})
		})
	})
})
