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
func (r *InstanceReconciler) FindBestIPAndAssignPorts(ctx context.Context, instance *clv1alpha2.Instance, usedPortsByIP map[string]map[int32]bool, currentIP string) (string, []clv1alpha2.PublicServicePort, error) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Starting IP and port assignment for public exposure", "instance", instance.Name)
	prioritizedIPPool := r.BuildPrioritizedIPPool(r.PublicExposureOpts.IPPool, usedPortsByIP)
	log.Info("Prioritized IP pool for evaluation", "pool", prioritizedIPPool)

	// Move the preferred IP (if any) to the front of the pool.
	prioritizedIPPool = reorderIPPoolWithPreferredIP(prioritizedIPPool, currentIP)

	specifiedPorts, autoPorts := splitPorts(instance.Spec.PublicExposure.Ports)

	for _, ip := range prioritizedIPPool {
		log.Info("Checking IP for port availability", "ip", ip)
		portsInUse := usedPortsByIP[ip]
		if portsInUse == nil {
			portsInUse = make(map[int32]bool)
		}

		assignedPorts, ok := tryAssignPorts(specifiedPorts, autoPorts, portsInUse)
		if ok {
			log.Info("Found compatible IP and assigned all ports", "ip", ip, "ports", assignedPorts)
			return ip, assignedPorts, nil
		}
		log.Info("IP not compatible with requested ports", "ip", ip)
	}

	log.Error(fmt.Errorf("no available IP can support all requested ports"), "IP/port assignment failed for instance", "instance", instance.Name)
	return "", nil, fmt.Errorf("no available IP can support all requested ports")
}

// splitPorts separates specified and auto-assigned ports.
func splitPorts(ports []clv1alpha2.PublicServicePort) (specified, auto []clv1alpha2.PublicServicePort) {
	for _, p := range ports {
		if p.Port != 0 {
			specified = append(specified, p)
		} else {
			auto = append(auto, p)
		}
	}
	return
}

// tryAssignPorts tries to assign all specified and auto ports to the given IP.
// Returns the assigned ports and true if successful.
func tryAssignPorts(specified, auto []clv1alpha2.PublicServicePort, portsInUse map[int32]bool) ([]clv1alpha2.PublicServicePort, bool) {
	assigned := make([]clv1alpha2.PublicServicePort, 0, len(specified)+len(auto))

	// 1. Check specified ports
	for _, port := range specified {
		if portsInUse[port.Port] {
			return nil, false
		}
		portsInUse[port.Port] = true
		assigned = append(assigned, clv1alpha2.PublicServicePort{
			Name:       port.Name,
			Port:       port.Port,
			TargetPort: port.TargetPort,
			Protocol:   port.Protocol,
		})
	}

	// 2. Assign auto ports
	for _, port := range auto {
		assignedPort := int32(0)
		for p := int32(forge.BasePortForAutomaticAssignment); p <= 65535; p++ {
			if !portsInUse[p] {
				assignedPort = p
				portsInUse[p] = true
				break
			}
		}
		if assignedPort == 0 {
			return nil, false
		}
		assigned = append(assigned, clv1alpha2.PublicServicePort{
			Name:       port.Name,
			Port:       assignedPort,
			TargetPort: port.TargetPort,
			Protocol:   port.Protocol,
		})
	}
	return assigned, true
}

// reorderIPPoolWithPreferredIP moves the preferred IP (from existing service) to the front of the pool.
func reorderIPPoolWithPreferredIP(pool []string, preferredIP string) []string {
	if preferredIP == "" {
		return pool
	}
	var found bool
	var newPool []string
	for i, ip := range pool {
		if ip == preferredIP {
			found = true
			// Place preferredIP at the front, keeping the relative order of the others
			newPool = append([]string{preferredIP}, append(pool[:i], pool[i+1:]...)...)
			break
		}
	}
	if found {
		return newPool
	}
	return pool
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
		// Check for ServiceType is LoadBalancer, although the label should be enough for the purpose.
		if svc.Spec.Type != v1.ServiceTypeLoadBalancer {
			continue
		}

		// Exclude the service we are currently reconciling to avoid self-conflicts for IP/port assignment.
		if svc.Name == excludeSvcName && svc.Namespace == excludeSvcNs {
			continue
		}

		// Prefer the annotation as it's the desired state.
		externalIP, ok := svc.Annotations[opts.LoadBalancerIPsKey]
		if !ok || externalIP == "" {
			continue
		}

		if _, exists := usedPortsByIP[externalIP]; !exists {
			usedPortsByIP[externalIP] = make(map[int32]bool)
		}

		for _, port := range svc.Spec.Ports {
			usedPortsByIP[externalIP][port.Port] = true
			log.Info("Port marked as in use", "ip", externalIP, "port", port.Port, "service", fmt.Sprintf("%s/%s", svc.Namespace, svc.Name))
		}
	}

	return usedPortsByIP, nil
}
