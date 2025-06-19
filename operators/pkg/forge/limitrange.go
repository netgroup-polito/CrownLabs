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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// SandboxLimitRangeSpec forges the Limit Range spec for sandbox namespaces.
func SandboxLimitRangeSpec() corev1.LimitRangeSpec {
	return corev1.LimitRangeSpec{
		Limits: []corev1.LimitRangeItem{{
			Type: corev1.LimitTypeContainer,
			DefaultRequest: corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewScaledQuantity(10, resource.Milli),
				corev1.ResourceMemory: *resource.NewScaledQuantity(50, resource.Mega),
			},
			Default: corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewScaledQuantity(100, resource.Milli),
				corev1.ResourceMemory: *resource.NewScaledQuantity(250, resource.Mega),
			},
		}},
	}
}
