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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// Metallb annotations, values, and labels for LoadBalancer services related to public exposure.
const (
	MetallbAddressPoolAnnotation     = "metallb.universe.tf/address-pool"
	MetallbAllowSharedIPAnnotation   = "metallb.universe.tf/allow-shared-ip"
	MetallbLoadBalancerIPsAnnotation = "metallb.universe.tf/loadBalancerIPs"
	AllowSharedIPValue               = "true"
	BasePortForAutomaticAssignment   = 30000

	labelPublicExposureValue = "public-exposure"
)

// Default value, can be overwritten by the args passed to the operator.
var DefaultAddressPool = "public-exposure-ip-pool"

// SetDefaultAddressPool sets the default address pool.
func SetDefaultAddressPool(v string) {
	if v != "" {
		DefaultAddressPool = v
	}
}

// LoadBalancerServiceSpec forges the spec for a LoadBalancer service for public exposure.
func LoadBalancerServiceSpec(instance *clv1alpha2.Instance, ports []clv1alpha2.PublicServicePort) v1.ServiceSpec {
	svcPorts := make([]v1.ServicePort, len(ports))
	for i, p := range ports {
		svcPorts[i] = v1.ServicePort{
			Name:       p.Name,
			Port:       p.Port,
			TargetPort: intstr.FromInt32(p.TargetPort),
			Protocol:   v1.ProtocolTCP,
		}
	}
	return v1.ServiceSpec{
		Type:     v1.ServiceTypeLoadBalancer,
		Selector: InstanceSelectorLabels(instance),
		Ports:    svcPorts,
	}
}

// LoadBalancerServiceAnnotations forges the annotations for a LoadBalancer service.
func LoadBalancerServiceAnnotations(externalIP string) map[string]string {
	annotations := map[string]string{
		MetallbAddressPoolAnnotation:     DefaultAddressPool,
		MetallbAllowSharedIPAnnotation:   AllowSharedIPValue,
		MetallbLoadBalancerIPsAnnotation: externalIP,
	}

	return annotations
}

// LoadBalancerServiceLabels forges the labels for a LoadBalancer service.
func LoadBalancerServiceLabels() map[string]string {
	labels := map[string]string{
		labelComponentKey: labelPublicExposureValue,
	}

	return labels
}

// LoadBalancerServiceName forges the name for a LoadBalancer service based on the instance name.
func LoadBalancerServiceName(instance *clv1alpha2.Instance) string {
	return instance.Name + "-public-exposure"
}
