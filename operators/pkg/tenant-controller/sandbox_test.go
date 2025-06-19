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
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	tntctrl "github.com/netgroup-polito/CrownLabs/operators/pkg/tenant-controller"
	. "github.com/netgroup-polito/CrownLabs/operators/pkg/utils/tests"
)

var _ = Describe("Sandbox", func() {
	var (
		ctx           context.Context
		clientBuilder fake.ClientBuilder
		reconciler    tntctrl.TenantReconciler
		tenant        clv1alpha2.Tenant

		sandboxNSname types.NamespacedName
		sbNamespace   corev1.Namespace

		roleBindingNSname types.NamespacedName
		sbRoleBinding     rbacv1.RoleBinding

		resQuotaNSname types.NamespacedName
		sbResQuota     corev1.ResourceQuota

		limitRangeNSname types.NamespacedName
		sbLimitRange     corev1.LimitRange

		ownerRef metav1.OwnerReference
		err      error
	)

	const (
		tenantName = "tester"
	)

	CreateSandboxObjects := func() {
		sbNamespace = corev1.Namespace{
			ObjectMeta: forge.NamespacedNameToObjectMeta(sandboxNSname),
		}
		sbRoleBinding = rbacv1.RoleBinding{
			ObjectMeta: forge.NamespacedNameToObjectMeta(roleBindingNSname),
		}
		sbResQuota = corev1.ResourceQuota{
			ObjectMeta: forge.NamespacedNameToObjectMeta(resQuotaNSname),
		}
		sbLimitRange = corev1.LimitRange{
			ObjectMeta: forge.NamespacedNameToObjectMeta(limitRangeNSname),
		}

		sbNamespace.SetCreationTimestamp(metav1.NewTime(time.Now()))
		sbRoleBinding.SetCreationTimestamp(metav1.NewTime(time.Now()))
		sbResQuota.SetCreationTimestamp(metav1.NewTime(time.Now()))
		sbLimitRange.SetCreationTimestamp(metav1.NewTime(time.Now()))
		clientBuilder.WithObjects(&sbNamespace, &sbRoleBinding, &sbResQuota, &sbLimitRange)
	}

	DescribeBodySandboxNamespacePresence := func() {
		It("Namespace should be present and have the expected labels", func() {
			Expect(reconciler.Get(ctx, sandboxNSname, &sbNamespace)).To(Succeed())
			for k, v := range forge.SandboxObjectLabels(nil, tenantName) {
				Expect(sbNamespace.GetLabels()).To(HaveKeyWithValue(k, v))
			}
			Expect(sbNamespace.GetOwnerReferences()).To(ContainElement(ownerRef))
		})

		It("Should fill the correct sandbox status value", func() {
			Expect(tenant.Status.SandboxNamespace).To(BeIdenticalTo(clv1alpha2.NameCreated{Name: sandboxNSname.Name, Created: true}))
		})
	}
	DescribeBodySandboxResourceQuotaPresence := func() {
		It("Resource quota should be present and have the expected labels", func() {
			Expect(reconciler.Get(ctx, resQuotaNSname, &sbResQuota)).To(Succeed())
			for k, v := range forge.SandboxObjectLabels(nil, tenantName) {
				Expect(sbResQuota.GetLabels()).To(HaveKeyWithValue(k, v))
			}
			Expect(sbResQuota.GetOwnerReferences()).To(ContainElement(ownerRef))
		})

		It("Resource quota should be present and have the expected spec", func() {
			Expect(reconciler.Get(ctx, resQuotaNSname, &sbResQuota)).To(Succeed())
			for k, v := range forge.SandboxResourceQuotaSpec() {
				Expect(sbResQuota.Spec.Hard).To(HaveKeyWithValue(k, WithTransform(
					func(q resource.Quantity) string { return q.String() }, Equal(v.String()))))
			}
		})
	}
	DescribeBodySandboxRoleBindingPresence := func() {
		It("Role binding should be present and have the expected labels", func() {
			Expect(reconciler.Get(ctx, roleBindingNSname, &sbRoleBinding)).To(Succeed())
			for k, v := range forge.SandboxObjectLabels(nil, tenantName) {
				Expect(sbRoleBinding.GetLabels()).To(HaveKeyWithValue(k, v))
			}
			Expect(sbRoleBinding.GetOwnerReferences()).To(ContainElement(ownerRef))
		})

		It("Role binding should be present and have the expected roleref", func() {
			Expect(reconciler.Get(ctx, roleBindingNSname, &sbRoleBinding)).To(Succeed())
			Expect(sbRoleBinding.RoleRef).To(BeIdenticalTo(rbacv1.RoleRef{Kind: "ClusterRole", Name: "crownlabs-sandbox", APIGroup: rbacv1.GroupName}))
		})

		It("Role binding should be present and have the expected subjects", func() {
			Expect(reconciler.Get(ctx, roleBindingNSname, &sbRoleBinding)).To(Succeed())
			for i, subject := range []rbacv1.Subject{{Kind: rbacv1.UserKind, Name: tenant.Name, APIGroup: rbacv1.GroupName}} {
				Expect(sbRoleBinding.Subjects[i]).To(BeIdenticalTo(subject))
			}
		})
	}
	DescribeBodySandboxLimitRangePresence := func() {
		It("Limit range should be present and have the expected labels", func() {
			Expect(reconciler.Get(ctx, limitRangeNSname, &sbLimitRange)).To(Succeed())
			for k, v := range forge.SandboxObjectLabels(nil, tenantName) {
				Expect(sbLimitRange.GetLabels()).To(HaveKeyWithValue(k, v))
			}
			Expect(sbLimitRange.GetOwnerReferences()).To(ContainElement(ownerRef))
		})

		It("Role binding should be present and have the expected spec", func() {
			Expect(reconciler.Get(ctx, limitRangeNSname, &sbLimitRange)).To(Succeed())
			// The following asserts the correctness of a single field, leaving more thorough checks to the appropriate unit tests.
			Expect(sbLimitRange.Spec.Limits).To(HaveLen(1))
			Expect(sbLimitRange.Spec.Limits[0].Type).To(BeIdenticalTo(corev1.LimitTypeContainer))
		})
	}
	DescribeBodySandboxNamespaceAbsence := func() {
		It("Should set the sandbox status empty", func() {
			Expect(tenant.Status.SandboxNamespace).To(BeIdenticalTo(clv1alpha2.NameCreated{Name: "", Created: false}))
		})

		It("Namespace should not be present", func() {
			Expect(reconciler.Get(ctx, sandboxNSname, &sbNamespace)).To(FailBecauseNotFound())
		})
	}

	BeforeEach(func() {
		ctx = ctrl.LoggerInto(context.Background(), logr.Discard())
		clientBuilder = *fake.NewClientBuilder().WithScheme(scheme.Scheme)
		tenant = clv1alpha2.Tenant{
			ObjectMeta: metav1.ObjectMeta{Name: tenantName},
		}
		ownerRef = metav1.OwnerReference{
			APIVersion:         clv1alpha2.GroupVersion.String(),
			Kind:               "Tenant",
			Name:               tenant.Name,
			UID:                tenant.GetUID(),
			BlockOwnerDeletion: ptr.To[bool](true),
			Controller:         ptr.To[bool](true),
		}
		sandboxNSname = types.NamespacedName{
			Name: forge.CanonicalSandboxName(tenantName),
		}
		roleBindingNSname = types.NamespacedName{
			Name:      "sandbox-editor",
			Namespace: sandboxNSname.Name,
		}
		resQuotaNSname = types.NamespacedName{
			Name:      "sandbox-resource-quota",
			Namespace: sandboxNSname.Name,
		}
		limitRangeNSname = types.NamespacedName{
			Name:      "sandbox-limit-range",
			Namespace: sandboxNSname.Name,
		}
		sbNamespace = corev1.Namespace{}
		sbRoleBinding = rbacv1.RoleBinding{}
		sbResQuota = corev1.ResourceQuota{}
		sbLimitRange = corev1.LimitRange{}
	})

	JustBeforeEach(func() {
		client := clientBuilder.Build()
		reconciler = tntctrl.TenantReconciler{Client: client, Scheme: scheme.Scheme, SandboxClusterRole: "crownlabs-sandbox"}
		err = reconciler.EnforceSandboxResources(ctx, &tenant)
	})

	Context("CreateSandbox flag is true", func() {
		BeforeEach(func() {
			tenant.Spec.CreateSandbox = true
		})

		When("Sandbox resources are not yet present", func() {
			It("Should not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
			Describe("Assessing the namespace presence", func() { DescribeBodySandboxNamespacePresence() })
			Describe("Assessing the resource quota presence", func() { DescribeBodySandboxResourceQuotaPresence() })
			Describe("Assessing the role binding presence", func() { DescribeBodySandboxRoleBindingPresence() })
			Describe("Assessing the limit range presence", func() { DescribeBodySandboxLimitRangePresence() })
		})

		When("Sandbox resources are already present", func() {
			BeforeEach(func() { CreateSandboxObjects() })

			It("Should not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
			Describe("Assessing the namespace presence", func() { DescribeBodySandboxNamespacePresence() })
			Describe("Assessing the resource quota presence", func() { DescribeBodySandboxResourceQuotaPresence() })
			Describe("Assessing the role binding presence", func() { DescribeBodySandboxRoleBindingPresence() })
			Describe("Assessing the limit range presence", func() { DescribeBodySandboxLimitRangePresence() })
		})
	})

	Context("CreateSandbox flag is false", func() {
		BeforeEach(func() {
			tenant.Spec.CreateSandbox = false
		})
		When("Sandbox Resources have not yet been deleted", func() {
			BeforeEach(func() {
				CreateSandboxObjects()
			})
			It("Should not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
			Describe("Assessing the namespace absence", func() { DescribeBodySandboxNamespaceAbsence() })
		})
		When("Sandbox Resources have already been deleted", func() {
			It("Should not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
			Describe("Assessing the namespace absence", func() { DescribeBodySandboxNamespaceAbsence() })
		})
	})
})
