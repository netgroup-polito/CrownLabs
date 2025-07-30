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
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("Tenant forging", func() {

	Describe("The forge.GetWorkspaceTargetLabel function", func() {
		It("Should generate the correct workspace target label for a workspace name", func() {
			workspaceName := "test-workspace"
			expectedLabel := "crownlabs.polito.it/workspace-test-workspace"

			Expect(forge.GetWorkspaceTargetLabel(workspaceName)).To(Equal(expectedLabel))
		})

		It("Should handle workspace names with special characters", func() {
			workspaceName := "test.workspace-123"
			expectedLabel := "crownlabs.polito.it/workspace-test.workspace-123"

			Expect(forge.GetWorkspaceTargetLabel(workspaceName)).To(Equal(expectedLabel))
		})

		It("Should prepend the workspace label prefix correctly", func() {
			// Verify we're using the correct prefix from the v1alpha2 package
			Expect(v1alpha2.WorkspaceLabelPrefix).To(Equal("crownlabs.polito.it/workspace-"))

			workspaceName := "demo"
			expectedLabel := "crownlabs.polito.it/workspace-demo"

			Expect(forge.GetWorkspaceTargetLabel(workspaceName)).To(Equal(expectedLabel))
		})
	})

	var _ = Describe("GetWorkspaceTargetLabel", func() {
		It("Should correctly format workspace target label", func() {
			workspaceName := "netgroup"

			labelKey := forge.GetWorkspaceTargetLabel(workspaceName)

			Expect(labelKey).To(Equal("crownlabs.polito.it/workspace-netgroup"))
		})
	})

	var _ = Describe("CleanTenantName", func() {
		It("Should replace spaces with underscores", func() {
			name := "John Doe"

			cleanName := forge.CleanTenantName(name)

			Expect(cleanName).To(Equal("John_Doe"))
		})

		It("Should remove special characters", func() {
			name := "John!@#$%^&*()Doe"

			cleanName := forge.CleanTenantName(name)

			Expect(cleanName).To(Equal("JohnDoe"))
		})

		It("Should trim leading and trailing underscores", func() {
			name := "_JohnDoe_"

			cleanName := forge.CleanTenantName(name)

			Expect(cleanName).To(Equal("JohnDoe"))
		})

		It("Should handle combined cases", func() {
			name := "__John @#$ Doe__"

			cleanName := forge.CleanTenantName(name)

			Expect(cleanName).To(Equal("John__Doe"))
		})
	})

	var _ = Describe("ConfigureTenantResourceQuota", func() {
		It("Should initialize labels if nil and set quota spec", func() {
			rq := &corev1.ResourceQuota{}
			quota := &v1alpha2.TenantResourceQuota{}
			labels := map[string]string{
				"custom-label": "custom-value",
			}

			forge.ConfigureTenantResourceQuota(rq, quota, labels)

			Expect(rq.Labels).ToNot(BeNil())
			Expect(rq.Labels).To(HaveKeyWithValue("custom-label", "custom-value"))
			Expect(rq.Spec.Hard).ToNot(BeNil())
		})

		It("Should preserve existing labels", func() {
			rq := &corev1.ResourceQuota{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"existing-label": "existing-value",
					},
				},
			}

			quota := &v1alpha2.TenantResourceQuota{}
			labels := map[string]string{
				"custom-label": "custom-value",
			}

			forge.ConfigureTenantResourceQuota(rq, quota, labels)

			Expect(rq.Labels).To(HaveKeyWithValue("existing-label", "existing-value"))
			Expect(rq.Labels).To(HaveKeyWithValue("custom-label", "custom-value"))
		})
	})

	var _ = Describe("ConfigureTenantDenyNetworkPolicy", func() {
		It("Should initialize labels if nil and set policy spec", func() {
			policy := &netv1.NetworkPolicy{}
			labels := map[string]string{
				"custom-label": "custom-value",
			}

			forge.ConfigureTenantDenyNetworkPolicy(policy, labels)

			Expect(policy.Labels).ToNot(BeNil())
			Expect(policy.Labels).To(HaveKeyWithValue("custom-label", "custom-value"))
			Expect(policy.Spec.PodSelector.MatchLabels).To(BeEmpty())
			Expect(policy.Spec.Ingress).To(HaveLen(1))
			Expect(policy.Spec.Ingress[0].From).To(HaveLen(1))
			Expect(policy.Spec.Ingress[0].From[0].PodSelector).ToNot(BeNil())
		})

		It("Should preserve existing labels", func() {
			policy := &netv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"existing-label": "existing-value",
					},
				},
			}

			labels := map[string]string{
				"custom-label": "custom-value",
			}

			forge.ConfigureTenantDenyNetworkPolicy(policy, labels)

			Expect(policy.Labels).To(HaveKeyWithValue("existing-label", "existing-value"))
			Expect(policy.Labels).To(HaveKeyWithValue("custom-label", "custom-value"))
		})
	})

	var _ = Describe("ConfigureTenantAllowNetworkPolicy", func() {
		It("Should initialize labels if nil and set policy spec", func() {
			policy := &netv1.NetworkPolicy{}
			labels := map[string]string{
				"custom-label": "custom-value",
			}

			forge.ConfigureTenantAllowNetworkPolicy(policy, labels)

			Expect(policy.Labels).ToNot(BeNil())
			Expect(policy.Labels).To(HaveKeyWithValue("custom-label", "custom-value"))
			Expect(policy.Spec.PodSelector.MatchLabels).To(BeEmpty())
			Expect(policy.Spec.Ingress).To(HaveLen(1))
			Expect(policy.Spec.Ingress[0].From).To(HaveLen(1))
			Expect(policy.Spec.Ingress[0].From[0].NamespaceSelector).ToNot(BeNil())
			Expect(policy.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels).To(HaveKeyWithValue("crownlabs.polito.it/allow-instance-access", "true"))
		})

		It("Should preserve existing labels", func() {
			policy := &netv1.NetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"existing-label": "existing-value",
					},
				},
			}

			labels := map[string]string{
				"custom-label": "custom-value",
			}

			forge.ConfigureTenantAllowNetworkPolicy(policy, labels)

			Expect(policy.Labels).To(HaveKeyWithValue("existing-label", "existing-value"))
			Expect(policy.Labels).To(HaveKeyWithValue("custom-label", "custom-value"))
		})
	})

	Describe("The forge.UpdateTenantResourceCommonLabels function", func() {
		It("Should add target label and managed-by label to existing labels", func() {
			inputLabels := map[string]string{
				"existing-label": "existing-value",
			}

			targetLabel := common.NewLabel("test-key", "test-value")

			resultLabels := forge.UpdateTenantResourceCommonLabels(inputLabels, targetLabel)

			Expect(resultLabels).To(HaveLen(3))
			Expect(resultLabels).To(HaveKeyWithValue("existing-label", "existing-value"))
			Expect(resultLabels).To(HaveKeyWithValue("test-key", "test-value"))
			Expect(resultLabels).To(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))
		})

		It("Should initialize labels map when nil", func() {
			var inputLabels map[string]string

			targetLabel := common.NewLabel("test-key", "test-value")

			resultLabels := forge.UpdateTenantResourceCommonLabels(inputLabels, targetLabel)

			Expect(resultLabels).To(HaveLen(2))
			Expect(resultLabels).To(HaveKeyWithValue("test-key", "test-value"))
			Expect(resultLabels).To(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))
		})
	})
})
