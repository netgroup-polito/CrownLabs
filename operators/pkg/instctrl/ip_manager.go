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

package instctrl

import (
	"context"
	"fmt"
	"sort"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

// BuildPrioritizedIPPool creates a sorted IP pool, prioritizing IPs that are already in use.
// This encourages IP reuse and reduces IP fragmentation.
func (r *InstanceReconciler) BuildPrioritizedIPPool(fullPool []string, usedPortsByIP map[string]map[int32]bool) []string {
	usedIPs := make([]string, 0, len(usedPortsByIP))
	for ip := range usedPortsByIP {
		usedIPs = append(usedIPs, ip)
	}
	sort.Strings(usedIPs) // Sort for consistent ordering.

	unusedIPs := []string{}
	isUsed := make(map[string]bool, len(usedIPs))
	for _, ip := range usedIPs {
		isUsed[ip] = true
	}

	for _, ip := range fullPool {
		if !isUsed[ip] {
			unusedIPs = append(unusedIPs, ip)
		}
	}
	sort.Strings(unusedIPs) // Sort for consistent ordering.

	// The final pool will have used IPs first, followed by unused IPs.
	return append(usedIPs, unusedIPs...)
}

// FindBestIPAndAssignPorts finds the best IP for the requested ports and handles port assignment.
func (r *InstanceReconciler) FindBestIPAndAssignPorts(ctx context.Context, c client.Client, instance *clv1alpha2.Instance, usedPortsByIP map[string]map[int32]bool) (string, []clv1alpha2.PublicServicePort, error) {
	log := ctrl.LoggerFrom(ctx)

	log.Info("Starting IP and port assignment for public exposure", "instance", instance.Name)
	prioritizedIPPool := r.BuildPrioritizedIPPool(r.PublicExposureOpts.IPPool, usedPortsByIP)
	log.V(1).Info("Prioritized IP pool for evaluation", "pool", prioritizedIPPool)

	// Check if a service already exists for this instance and try to reuse its IP.
	existingSvc := &v1.Service{
		ObjectMeta: forge.ObjectMetaWithSuffix(instance, forge.LabelPublicExposureValue),
	}

	err := c.Get(ctx, types.NamespacedName{Name: existingSvc.Name, Namespace: instance.Namespace}, existingSvc)
	if err == nil {
		var preferredIP string
		if ip, ok := existingSvc.Annotations[r.PublicExposureOpts.LoadBalancerIPsKey]; ok && ip != "" {
			preferredIP = ip
		} else if len(existingSvc.Status.LoadBalancer.Ingress) > 0 {
			preferredIP = existingSvc.Status.LoadBalancer.Ingress[0].IP
		}

		if preferredIP != "" {
			log.Info("Found preferred IP from existing service", "ip", preferredIP, "service", existingSvc.Name)
			// Move the preferred IP to the front of the pool to try it first.
			found := false
			for i, ip := range prioritizedIPPool {
				if ip == preferredIP {
					prioritizedIPPool[0], prioritizedIPPool[i] = prioritizedIPPool[i], prioritizedIPPool[0]
					found = true
					break
				}
			}
			if !found {
				// if the preferred IP is not in the pool, add it at the beginning
				prioritizedIPPool = append([]string{preferredIP}, prioritizedIPPool...)
			}
		}
	}

	var specifiedPorts, autoPorts []clv1alpha2.PublicServicePort
	for _, svcPort := range instance.Spec.PublicExposure.Ports {
		if svcPort.Port != 0 {
			specifiedPorts = append(specifiedPorts, svcPort)
		} else {
			autoPorts = append(autoPorts, svcPort)
		}
	}

	// Iterate over each available IP in the prioritized pool.
	for _, ip := range prioritizedIPPool {
		log.V(1).Info("Checking IP for port availability", "ip", ip)

		// Create a simulated copy of the ports in use for this IP to avoid modifying the original map.
		simulatedPortsInUse := make(map[int32]bool)
		if used, ok := usedPortsByIP[ip]; ok {
			for p := range used {
				simulatedPortsInUse[p] = true
			}
		}

		var tempAssignedPorts []clv1alpha2.PublicServicePort
		isIPCompatible := true

		// 1. Check if all specified ports can be assigned on this IP.
		for _, port := range specifiedPorts {
			if simulatedPortsInUse[port.Port] {
				isIPCompatible = false
				break
			}
			simulatedPortsInUse[port.Port] = true // Simulate port assignment.
			tempAssignedPorts = append(tempAssignedPorts, clv1alpha2.PublicServicePort{
				Name:       port.Name,
				Port:       port.Port,
				TargetPort: port.TargetPort,
				Protocol:   port.Protocol,
			})
		}

		if !isIPCompatible {
			log.V(1).Info("IP not compatible with specified ports", "ip", ip)
			continue // Try the next IP in the pool.
		}

		// 2. Assign ports for automatic requests.
		allAutoPortsAssignable := true
		for _, port := range autoPorts {
			assignedPort := int32(0)
			for p := int32(forge.BasePortForAutomaticAssignment); p <= 65535; p++ {
				if !simulatedPortsInUse[p] {
					assignedPort = p
					simulatedPortsInUse[p] = true // Simulate port assignment.
					break
				}
			}

			if assignedPort == 0 {
				allAutoPortsAssignable = false
				break
			}
			tempAssignedPorts = append(tempAssignedPorts, clv1alpha2.PublicServicePort{
				Name:       port.Name,
				Port:       assignedPort,
				TargetPort: port.TargetPort,
				Protocol:   port.Protocol,
			})
		}

		// 3. If all ports are assignable, this is the best IP.
		if allAutoPortsAssignable {
			log.Info("Found compatible IP and assigned all ports", "ip", ip, "ports", tempAssignedPorts)
			return ip, tempAssignedPorts, nil
		}
	}

	log.Error(fmt.Errorf("no available IP can support all requested ports"), "IP/port assignment failed for instance", "instance", instance.Name)
	return "", nil, fmt.Errorf("no available IP can support all requested ports")
}

// UpdateUsedPortsByIP scans LoadBalancer services with the specific public-exposure label
// to build a map of used ports per IP.
func UpdateUsedPortsByIP(ctx context.Context, c client.Client, excludeSvcName, excludeSvcNs string, opts *forge.PublicExposureOpts) (map[string]map[int32]bool, error) {
	usedPortsByIP := make(map[string]map[int32]bool)
	log := ctrl.LoggerFrom(ctx)

	svcList := &v1.ServiceList{}

	// Use a label selector to list only the relevant services.
	labels := forge.LoadBalancerServiceLabels()
	listOptions := []client.ListOption{
		client.MatchingLabels(labels),
	}

	if err := c.List(ctx, svcList, listOptions...); err != nil {
		return nil, fmt.Errorf("failed to list public exposure services: %w", err)
	}

	for i := range svcList.Items {
		svc := &svcList.Items[i]
		// The check for ServiceType is still a good practice, although the label should be enough.
		if svc.Spec.Type != v1.ServiceTypeLoadBalancer {
			continue
		}

		// Exclude the service we are currently reconciling to avoid self-conflicts.
		if svc.Name == excludeSvcName && svc.Namespace == excludeSvcNs {
			continue
		}

		var externalIP string
		// Prefer the annotation as it's the desired state.
		if ip, ok := svc.Annotations[opts.LoadBalancerIPsKey]; ok && ip != "" {
			externalIP = ip
		} else if len(svc.Status.LoadBalancer.Ingress) > 0 {
			externalIP = svc.Status.LoadBalancer.Ingress[0].IP
		}

		if externalIP == "" {
			continue
		}

		if _, exists := usedPortsByIP[externalIP]; !exists {
			usedPortsByIP[externalIP] = make(map[int32]bool)
		}

		for _, port := range svc.Spec.Ports {
			usedPortsByIP[externalIP][port.Port] = true
			log.V(1).Info("Port marked as in use", "ip", externalIP, "port", port.Port, "service", fmt.Sprintf("%s/%s", svc.Namespace, svc.Name))
		}
	}

	return usedPortsByIP, nil
}
