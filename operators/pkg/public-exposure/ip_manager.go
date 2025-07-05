package publicexposure

import (
	"context"
	"fmt"
	"sort"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// getMetalLBIPPool retrieves the IP pool configured in MetalLB.
// TODO: Implement logic to get the IP pool dynamically from MetalLB ConfigMap or CRDs if they exist.
func (m *Manager) getMetalLBIPPool(_ context.Context) ([]string, error) {
	// For now, this returns a static pool. This should be made configurable.
	return []string{
		"172.18.0.240", "172.18.0.241", "172.18.0.242", "172.18.0.243",
		"172.18.0.244", "172.18.0.245", "172.18.0.246", "172.18.0.247",
		"172.18.0.248", "172.18.0.249", "172.18.0.250",
	}, nil
}

// buildPrioritizedIPPool creates a sorted IP pool, prioritizing IPs that are already in use.
// This encourages IP reuse and reduces IP fragmentation.
func (m *Manager) buildPrioritizedIPPool(fullPool []string, usedPortsByIP map[string]map[int32]bool) []string {
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

// findBestIPAndAssignPorts finds the best IP for the requested ports and handles port assignment.
func (m *Manager) findBestIPAndAssignPorts(ctx context.Context, instance *clv1alpha2.Instance, usedPortsByIP map[string]map[int32]bool) (string, []clv1alpha2.PublicServicePort, error) {
	log := ctrl.LoggerFrom(ctx)

	fullIPPool, err := m.getMetalLBIPPool(ctx)
	if err != nil {
		return "", nil, err
	}

	// NEW LOGIC: Build a prioritized IP pool to check used IPs first.
	prioritizedIPPool := m.buildPrioritizedIPPool(fullIPPool, usedPortsByIP)
	log.V(1).Info("Prioritized IP pool for evaluation", "pool", prioritizedIPPool)

	// Check if a service already exists for this instance and try to reuse its IP.
	svcName := m.serviceName(instance)
	existingSvc := &v1.Service{}
	err = m.Get(ctx, types.NamespacedName{Name: svcName, Namespace: instance.Namespace}, existingSvc)
	if err == nil {
		var preferredIP string
		if ip, ok := existingSvc.Annotations[MetallbLoadBalancerIPsAnnotation]; ok && ip != "" {
			preferredIP = ip
		} else if len(existingSvc.Status.LoadBalancer.Ingress) > 0 {
			preferredIP = existingSvc.Status.LoadBalancer.Ingress[0].IP
		}

		if preferredIP != "" {
			// Move the preferred IP to the front of the pool to try it first.
			found := false
			for i, ip := range prioritizedIPPool {
				if ip == preferredIP {
					// Swap with the first element
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
			for p := int32(BasePortForAutomaticAssignment); p <= 32767; p++ {
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
			})
		}

		// 3. If all ports are assignable, this is the best IP.
		if allAutoPortsAssignable {
			log.Info("Found compatible IP and assigned all ports", "ip", ip, "ports", tempAssignedPorts)
			return ip, tempAssignedPorts, nil
		}
	}

	return "", nil, fmt.Errorf("no available IP can support all requested ports")
}

// updateUsedPortsByIP scans LoadBalancer services with the specific public-exposure label
// to build a map of used ports per IP.
func (m *Manager) updateUsedPortsByIP(ctx context.Context, excludeSvcName, excludeSvcNs string) (map[string]map[int32]bool, error) {
	usedPortsByIP := make(map[string]map[int32]bool)
	logger := log.FromContext(ctx)

	svcList := &v1.ServiceList{}

	// Use a label selector to list only the relevant services.
	listOptions := []client.ListOption{
		client.MatchingLabels{
			PublicExposureComponentLabelKey: PublicExposureComponentLabelValue,
		},
	}

	if err := m.List(ctx, svcList, listOptions...); err != nil {
		return nil, fmt.Errorf("failed to list public exposure services: %w", err)
	}

	for _, svc := range svcList.Items {
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
		if ip, ok := svc.Annotations[MetallbLoadBalancerIPsAnnotation]; ok && ip != "" {
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
			logger.V(1).Info("Port marked as in use", "ip", externalIP, "port", port.Port, "service", fmt.Sprintf("%s/%s", svc.Namespace, svc.Name))
		}
	}

	return usedPortsByIP, nil
}
