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

package forge

import (
	"fmt"
	"maps"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
)

// GetWorkspaceTargetLabel returns the label key for identifying tenants with a specific workspace.
func GetWorkspaceTargetLabel(workspaceName string) string {
	return fmt.Sprintf("%s%s", v1alpha2.WorkspaceLabelPrefix, workspaceName)
}

// UpdateTenantResourceCommonLabels updates the common labels for resources managed by the tenant controller.
func UpdateTenantResourceCommonLabels(labels map[string]string, targetLabel common.KVLabel) map[string]string {
	if labels == nil {
		labels = make(map[string]string, 1)
	}
	labels[targetLabel.GetKey()] = targetLabel.GetValue()
	labels[labelManagedByKey] = labelManagedByTenantValue

	return labels
}

// CleanTenantName sanitizes a tenant name by replacing spaces with underscores and removing
// any characters that are not alphanumeric or underscores. It also trims leading
// and trailing underscores.
func CleanTenantName(name string) string {
	re := regexp.MustCompile("^[a-zA-Z0-9_]+$")
	name = strings.ReplaceAll(name, " ", "_")

	if !re.MatchString(name) {
		problemChars := make([]string, 0)
		for _, c := range name {
			if !re.MatchString(string(c)) {
				problemChars = append(problemChars, string(c))
			}
		}
		for _, v := range problemChars {
			name = strings.Replace(name, v, "", 1)
		}
	}

	return strings.Trim(name, "_")
}

// ConfigureTenantResourceQuota configures a ResourceQuota for a tenant namespace.
func ConfigureTenantResourceQuota(rq *corev1.ResourceQuota, quota *v1alpha2.TenantResourceQuota, labels map[string]string) {
	// Set the labels
	if rq.Labels == nil {
		rq.Labels = make(map[string]string)
	}

	// Copy the provided labels
	maps.Copy(rq.Labels, labels)

	// Configure the resource quota spec
	rq.Spec.Hard = TenantResourceQuotaSpec(quota)
}

// ConfigureTenantDenyNetworkPolicy configures a NetworkPolicy that denies all ingress traffic.
func ConfigureTenantDenyNetworkPolicy(policy *netv1.NetworkPolicy, labels map[string]string) {
	// Set the labels
	if policy.Labels == nil {
		policy.Labels = make(map[string]string)
	}

	// Copy the provided labels
	maps.Copy(policy.Labels, labels)

	// Configure the network policy spec
	policy.Spec.PodSelector.MatchLabels = make(map[string]string)
	policy.Spec.Ingress = []netv1.NetworkPolicyIngressRule{{
		From: []netv1.NetworkPolicyPeer{{PodSelector: &metav1.LabelSelector{}}},
	}}
}

// ConfigureTenantAllowNetworkPolicy configures a NetworkPolicy that allows trusted ingress traffic.
func ConfigureTenantAllowNetworkPolicy(policy *netv1.NetworkPolicy, labels map[string]string) {
	// Set the labels
	if policy.Labels == nil {
		policy.Labels = make(map[string]string)
	}

	// Copy the provided labels
	maps.Copy(policy.Labels, labels)

	// Configure the network policy spec
	policy.Spec.PodSelector.MatchLabels = make(map[string]string)
	policy.Spec.Ingress = []netv1.NetworkPolicyIngressRule{{
		From: []netv1.NetworkPolicyPeer{{
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					labelAllowInstanceAccessKey: labelAllowInstanceAccessValue,
				},
			},
		}},
	}}
}
