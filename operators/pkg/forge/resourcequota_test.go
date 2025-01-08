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

package forge_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("Resource quota spec forging", func() {

	Describe("The forge.TenantResourceList function", func() {
		var (
			tenant clv1alpha2.Tenant
		)
		Describe("Forging the tenant resource quota with a defined spec value", func() {

			BeforeEach(func() {
				tenant = clv1alpha2.Tenant{
					Spec: clv1alpha2.TenantSpec{
						Quota: &clv1alpha2.TenantResourceQuota{
							CPU:       *resource.NewQuantity(25, resource.DecimalSI),
							Memory:    *resource.NewScaledQuantity(50, resource.Giga),
							Instances: 6,
						},
					},
				}
			})

			JustBeforeEach(func() {
				tenant.Status.Quota = forge.TenantResourceList(nil, tenant.Spec.Quota)
			})

			When("Forging resource quota in tenant status", func() {
				It("It should forge tenant resource quota by overriding from the tenant spec", func() {
					Expect(tenant.Status.Quota).To(Equal(*tenant.Spec.Quota))
				})
			})
		})
		Describe("Forging the tenant resource quota with a sample workspaces", func() {
			var (
				workspaces []clv1alpha1.Workspace
			)

			BeforeEach(func() {
				tenant = clv1alpha2.Tenant{}
				workspaces = make([]clv1alpha1.Workspace, 0)
				var sampleWorkspace1, sampleWorkspace2 clv1alpha1.Workspace
				// sample resource quota spec for each workspace.
				quota1 := clv1alpha1.WorkspaceResourceQuota{
					CPU:       *resource.NewQuantity(10, resource.DecimalSI),
					Memory:    *resource.NewScaledQuantity(15, resource.Giga),
					Instances: 2,
				}

				quota2 := clv1alpha1.WorkspaceResourceQuota{
					CPU:       *resource.NewQuantity(20, resource.DecimalSI),
					Memory:    *resource.NewScaledQuantity(25, resource.Giga),
					Instances: 3,
				}
				sampleWorkspace1.Spec.Quota = quota1
				sampleWorkspace2.Spec.Quota = quota2
				workspaces = append(workspaces, sampleWorkspace1, sampleWorkspace2)
				forge.CapCPU = 25
				forge.CapMemoryGiga = 40
				forge.CapInstance = 5
			})

			JustBeforeEach(func() {
				tenant.Status.Quota = forge.TenantResourceList(workspaces, tenant.Spec.Quota)
			})

			When("Forging resource quota in tenant status", func() {
				It("Should have total amount of CPU equal to the defined cap, because the sum for each workspace exceedes it", func() {
					Expect(tenant.Status.Quota.CPU).To(Equal(*resource.NewQuantity(25, resource.DecimalSI)))
				})

				It("Should have total amount of memory equal to the sum for each workspace", func() {
					Expect(tenant.Status.Quota.Memory).To(Equal(*resource.NewScaledQuantity(40, resource.Giga)))
				})

				It("Should have total number of instances equal to the sum for each workspace", func() {
					Expect(tenant.Status.Quota.Instances).To(Equal(uint32(5)))
				})
			})
		})
	})

	Describe("The forge.TenantResourceQuotaSpec function", func() {
		var (
			spec   corev1.ResourceList
			tenant clv1alpha2.Tenant
		)

		BeforeEach(func() {
			tenant = clv1alpha2.Tenant{
				Spec: clv1alpha2.TenantSpec{
					Quota: &clv1alpha2.TenantResourceQuota{
						CPU:       *resource.NewQuantity(15, resource.DecimalSI),
						Memory:    *resource.NewScaledQuantity(20, resource.Giga),
						Instances: 3,
					},
				},
			}
		})

		JustBeforeEach(func() {
			spec = forge.TenantResourceQuotaSpec(&tenant.Status.Quota)
		})

		When("Forging the resource quota specifications", func() {
			It("Should have total amount of CPU requests and limits equal to the ones associated with the Tenant", func() {
				Expect(spec[corev1.ResourceLimitsCPU]).To(Equal(tenant.Status.Quota.CPU))
				Expect(spec[corev1.ResourceRequestsCPU]).To(Equal(tenant.Status.Quota.CPU))
			})

			It("Should have total amount of memory requests and limits equal to the ones associated with the Tenant", func() {
				Expect(spec[corev1.ResourceLimitsMemory]).To(Equal(tenant.Status.Quota.Memory))
				Expect(spec[corev1.ResourceRequestsMemory]).To(Equal(tenant.Status.Quota.Memory))
			})

			It("Should have total number of instances equal to the one associated with the Tenant", func() {
				Expect(spec[forge.InstancesCountKey]).To(Equal(*resource.NewQuantity(int64(tenant.Status.Quota.Instances), resource.DecimalSI)))
			})
		})
	})

	Describe("The forge.SandboxResourceQuotaSpec function", func() {
		var (
			spec corev1.ResourceList
		)

		JustBeforeEach(func() {
			spec = forge.SandboxResourceQuotaSpec()
		})

		When("Forging the resource quota specifications", func() {
			It("Should configure the correct amount of CPU requests and limits", func() {
				Expect(spec[corev1.ResourceLimitsCPU]).To(Equal(*resource.NewQuantity(4, resource.DecimalSI)))
				Expect(spec[corev1.ResourceRequestsCPU]).To(Equal(*resource.NewQuantity(2, resource.DecimalSI)))
			})

			It("Should configure the correct amount of memory requests and limits", func() {
				Expect(spec[corev1.ResourceLimitsMemory]).To(Equal(*resource.NewScaledQuantity(8, resource.Giga)))
				Expect(spec[corev1.ResourceRequestsMemory]).To(Equal(*resource.NewScaledQuantity(8, resource.Giga)))
			})

		})
	})
})
