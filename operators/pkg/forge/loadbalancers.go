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
	"strings"

	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// PublicExposureOpts contains configuration for LoadBalancer services and public exposure.
type PublicExposureOpts struct {
	IPPool             []string          // List of available IPs for assignment
	CommonAnnotations  map[string]string // Common annotations to add to all LoadBalancer services
	LoadBalancerIPsKey string            // Annotation key for specifying the IP (e.g., "metallb.universe.tf/loadBalancerIPs")
}

// Metallb annotations, values, and labels for LoadBalancer services related to public exposure.
const (
	AllowSharedIPValue             = "true"
	BasePortForAutomaticAssignment = 49152
	LabelPublicExposureValue       = "pe"

	// Cilium SharedIpAnnotation | lbipam.cilium.io/sharing-key = "pe" .
	// Cilium SharedIpAcrossNamespace | lbipam.cilium.io/sharing-cross-namespace = "*" .
	// Cilium LoadBalancerIPsAnnotation | lbipam.cilium.io/ips .

	// MetalLB SharedIpAnnotation | metallb.universe.tf/allow-shared-ip = "pe" .
	// MetalLB LoadBalancerIPsAnnotation | metallb.universe.tf/loadBalancerIPs .
)

// ConfigureLoadBalancerAnnotationKeys parses a raw string containing two comma-separated annotation keys.
func ConfigureLoadBalancerAnnotationKeys(raw string) (s, ip string, err error) {
	parts := strings.Split(raw, ",")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid annotation format")
	}
	sharedKey := strings.TrimSpace(parts[0])
	ipKey := strings.TrimSpace(parts[1])
	// Based on MetalLB, in case of multiple annotation keys, like Cilium with the SharedIpAcrossNamespace, add LoC

	return sharedKey, ipKey, nil
}

// LoadBalancerServiceSpec forges the spec for a LoadBalancer service for public exposure.
func LoadBalancerServiceSpec(instance *clv1alpha2.Instance, ports []clv1alpha2.PublicServicePort) v1.ServiceSpec {
	svcPorts := make([]v1.ServicePort, len(ports))
	for i, p := range ports {
		protocol := v1.ProtocolTCP
		if p.Protocol == "UDP" {
			protocol = v1.ProtocolUDP
		}
		svcPorts[i] = v1.ServicePort{
			Name:       p.Name,
			Port:       p.Port,
			TargetPort: intstr.FromInt32(p.TargetPort),
			Protocol:   protocol,
		}
	}
	return v1.ServiceSpec{
		Type:     v1.ServiceTypeLoadBalancer,
		Selector: InstanceSelectorLabels(instance),
		Ports:    svcPorts,
	}
}

// LoadBalancerServiceAnnotations forges the annotations for a LoadBalancer service using the provided options.
func LoadBalancerServiceAnnotations(externalIP string, opts *PublicExposureOpts) map[string]string {
	annotations := make(map[string]string)

	// Add common annotations first
	for key, value := range opts.CommonAnnotations {
		annotations[key] = value
	}

	// Add the LoadBalancer IP annotation using the configured key
	annotations[opts.LoadBalancerIPsKey] = externalIP

	return annotations
}

// LoadBalancerServiceLabels forges the labels for a LoadBalancer service.
func LoadBalancerServiceLabels() map[string]string {
	labels := map[string]string{
		labelComponentKey: LabelPublicExposureValue,
	}

	return labels
}

// LoadBalancerServiceName forges the name for a LoadBalancer service based on the instance name.
func LoadBalancerServiceName(instance *clv1alpha2.Instance) string {
	return instance.Name + "-" + LabelPublicExposureValue
}

// PublicExposureNetworkPolicyName forges the name for a NetworkPolicy for public exposure.
func PublicExposureNetworkPolicyName(instance *clv1alpha2.Instance) string {
	return "crownlabs-allow-publicexposure-ingress-traffic-" + instance.Name
}

// ParseAnnotations parses a string like "key1=val1,key2=val2" into a map.
func ParseAnnotations(raw string) (map[string]string, error) {
	annotations := make(map[string]string)
	if raw == "" {
		return annotations, nil
	}

	pairs := strings.Split(raw, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid annotation format: %s", pair)
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		if key == "" {
			return nil, fmt.Errorf("empty annotation key in: %s", pair)
		}
		annotations[key] = value
	}
	return annotations, nil
}

// PublicExposureNetworkPolicy forges the NetworkPolicy object for a given instance.
func PublicExposureNetworkPolicy(instance *clv1alpha2.Instance, netpol *netv1.NetworkPolicy) {
	netpol.SetLabels(InstanceObjectLabels(netpol.GetLabels(), instance))

	ingressPorts := []netv1.NetworkPolicyPort{}
	// Use ports from Status, as they contain the actual assigned ports.
	if instance.Status.PublicExposure != nil {
		for _, p := range instance.Status.PublicExposure.Ports {
			port := intstr.FromInt32(p.TargetPort)
			protocol := v1.ProtocolTCP
			if p.Protocol == "UDP" {
				protocol = v1.ProtocolUDP
			}
			ingressPorts = append(ingressPorts, netv1.NetworkPolicyPort{
				Port:     &port,
				Protocol: &protocol,
			})
		}
	}

	netpol.Spec = netv1.NetworkPolicySpec{
		PodSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{
				"crownlabs.polito.it/instance": instance.Name,
			},
		},
		Ingress: []netv1.NetworkPolicyIngressRule{{
			Ports: ingressPorts,
		}},
		PolicyTypes: []netv1.PolicyType{netv1.PolicyTypeIngress},
	}
}
