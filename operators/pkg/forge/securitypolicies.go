// Copyright 2020-2026 Politecnico di Torino
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
	egv1alpha1 "github.com/envoyproxy/gateway/api/v1alpha1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// SecurityPolicySpec forges the specification of a Kubernetes SecurityPolicy resource.
func SecurityPolicySpec(targetRouteName string, templateSpec *egv1alpha1.SecurityPolicySpec) egv1alpha1.SecurityPolicySpec {
	if templateSpec == nil {
		panic("Unexpected empty templatePolSpec")
	}
	spec := *templateSpec.DeepCopy()
	spec.PolicyTargetReferences = SecurityPolicyTargetRefs(targetRouteName)
	return spec
}

// SecurityPolicyTargetRefs creates PolicyTargetReferences pointing to the target route.
func SecurityPolicyTargetRefs(routeName string) egv1alpha1.PolicyTargetReferences {
	return egv1alpha1.PolicyTargetReferences{
		TargetRefs: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
			{
				LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
					Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
					Kind:  gwapiv1a2.Kind("HTTPRoute"),
					Name:  gwapiv1a2.ObjectName(routeName),
				},
			},
		},
	}
}
