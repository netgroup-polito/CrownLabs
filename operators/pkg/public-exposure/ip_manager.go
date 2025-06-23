package publicexposure

import (
	"context"
	"fmt"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// getMetalLBIPPool retrieves the IP pool configured in MetalLB
func (m *Manager) getMetalLBIPPool(ctx context.Context) ([]string, error) {
	// TODO: Implement logic to get the IP pool from MetalLB in the cluster
	// For now, returns a static pool
	return []string{
		"172.18.0.240", "172.18.0.241", "172.18.0.242", "172.18.0.243",
		"172.18.0.244", "172.18.0.245", "172.18.0.246", "172.18.0.247",
		"172.18.0.248", "172.18.0.249", "172.18.0.250",
	}, nil
}

// findBestIPAndAssignPorts finds the best IP for the requested ports and handles port assignment
func (m *Manager) findBestIPAndAssignPorts(ctx context.Context, instance *clv1alpha2.Instance, usedPortsByIP map[string]map[int]bool) (string, []clv1alpha2.ServicePortMapping, error) {
	log := ctrl.LoggerFrom(ctx)

	// Check if a service already exists for this instance and use the same IP if possible
	svcName := fmt.Sprintf("instance-lb-%s", instance.Name)
	existingSvc := &v1.Service{}
	err := m.Get(ctx, types.NamespacedName{Name: svcName, Namespace: instance.Namespace}, existingSvc)

	var preferredIP string
	if err == nil {
		// Try to use the already assigned IP
		if ip, ok := existingSvc.Annotations["metallb.universe.tf/loadBalancerIPs"]; ok {
			preferredIP = ip
		} else if len(existingSvc.Status.LoadBalancer.Ingress) > 0 {
			preferredIP = existingSvc.Status.LoadBalancer.Ingress[0].IP
		}
	}

	// 1. Get the available IP pool from MetalLB
	ipPool, err := m.getMetalLBIPPool(ctx)
	if err != nil {
		return "", nil, err
	}

	// If we have a preferred IP, put it at the beginning of the list
	if preferredIP != "" {
		// Remove the preferred IP from the list and add it at the beginning
		newPool := []string{preferredIP}
		for _, ip := range ipPool {
			if ip != preferredIP {
				newPool = append(newPool, ip)
			}
		}
		ipPool = newPool
		log.Info("Using preferred IP from existing service", "ip", preferredIP)
	}

	// Create a local copy of used ports for simulation
	simulatedUsedPorts := make(map[string]map[int]bool)
	for ip, ports := range usedPortsByIP {
		simulatedUsedPorts[ip] = make(map[int]bool)
		for port := range ports {
			simulatedUsedPorts[ip][port] = true
		}
	}

	// 2. Split service ports into specified and automatic ports
	var specifiedPorts, autoPorts []clv1alpha2.ServicePortMapping
	for _, svcPort := range instance.Spec.PublicExposure.ServicesPortMappings {
		if svcPort.Port != 0 {
			specifiedPorts = append(specifiedPorts, svcPort)
		} else {
			autoPorts = append(autoPorts, svcPort)
		}
	}

	// Choose the best IP considering specified ports first
	var bestIP string
	var allAssignedPorts []clv1alpha2.ServicePortMapping

	// 3. Examine each available IP
	for _, ip := range ipPool {
		// Initialize the port map if it doesn't exist
		if simulatedUsedPorts[ip] == nil {
			simulatedUsedPorts[ip] = make(map[int]bool)
		}

		// Flag to track if this IP is compatible with all specified ports
		isIPCompatible := true

		// 4. First check if all specified ports can be assigned
		var tempAssignedSpecific []clv1alpha2.ServicePortMapping
		for _, port := range specifiedPorts {
			// Check if the requested port is already in use
			if simulatedUsedPorts[ip][int(port.Port)] {
				isIPCompatible = false
				log.Info("Specified port already in use", "ip", ip, "port", port.Port)
				break
			}

			// Simulate port assignment
			simulatedUsedPorts[ip][int(port.Port)] = true
			assignedPort := clv1alpha2.ServicePortMapping{
				Name:         port.Name,
				TargetPort:   port.TargetPort,
				Port:         port.Port,
				AssignedPort: port.Port, // For specified ports, AssignedPort = Port
			}
			tempAssignedSpecific = append(tempAssignedSpecific, assignedPort)
		}

		// If not compatible with specified ports, try the next IP
		if !isIPCompatible {
			continue
		}

		// 5. Now assign automatic ports
		var tempAssignedAuto []clv1alpha2.ServicePortMapping
		allAutoPortsAssignable := true

		for _, port := range autoPorts {
			// Find a free port in the range 30000-32767
			var assignedPort int32
			for potentialPort := basePort; potentialPort <= 32767; potentialPort++ {
				// Verify that the port is not already used or requested by other ServiceRequests
				if !simulatedUsedPorts[ip][potentialPort] {
					assignedPort = int32(potentialPort)
					simulatedUsedPorts[ip][potentialPort] = true
					break
				}
			}

			if assignedPort == 0 {
				// Cannot find a free port
				allAutoPortsAssignable = false
				log.Info("Cannot find free port for automatic assignment", "ip", ip)
				break
			}

			tempAssignedAuto = append(tempAssignedAuto, clv1alpha2.ServicePortMapping{
				Name:         port.Name,
				TargetPort:   port.TargetPort,
				Port:         0, // The original port is 0 (automatic)
				AssignedPort: assignedPort,
			})
		}

		// 6. If all ports are assignable, this is the best IP
		if allAutoPortsAssignable {
			bestIP = ip
			allAssignedPorts = append(tempAssignedSpecific, tempAssignedAuto...)
			log.Info("Found compatible IP", "ip", bestIP)
			break
		}
	}

	if bestIP == "" {
		return "", nil, fmt.Errorf("no available IP can support all requested ports")
	}

	// Update the real map of used ports
	for _, port := range allAssignedPorts {
		if usedPortsByIP[bestIP] == nil {
			usedPortsByIP[bestIP] = make(map[int]bool)
		}
		usedPortsByIP[bestIP][int(port.AssignedPort)] = true
		log.Info("Port registered as in use", "ip", bestIP, "port", port.AssignedPort)
	}

	return bestIP, allAssignedPorts, nil
}

// updateUsedPortsByIP updates the map of ports in use for each IP
func (m *Manager) updateUsedPortsByIP(ctx context.Context, namespace string, excludeServiceName string) (map[string]map[int]bool, error) {
	usedPortsByIP := make(map[string]map[int]bool)
	logger := log.FromContext(ctx)

	// Get all LoadBalancer services
	svcList := &v1.ServiceList{}
	if err := m.List(ctx, svcList, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	for _, svc := range svcList.Items {
		// SKIP THE CURRENT SERVICE to avoid conflicts with itself
		if svc.Name == excludeServiceName {
			logger.Info("Skipping current service from used ports calculation", "service", svc.Name)
			continue
		}

		if svc.Spec.Type != v1.ServiceTypeLoadBalancer {
			continue
		}

		// Get the assigned external IP
		var externalIP string
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			externalIP = svc.Status.LoadBalancer.Ingress[0].IP
		} else {
			// If the service doesn't have an IP yet, try to find it in annotations
			if specIP, ok := svc.Annotations["metallb.universe.tf/loadBalancerIPs"]; ok {
				externalIP = specIP
			} else {
				continue // Cannot determine the IP
			}
		}

		// Initialize the map for this IP if it doesn't exist
		if _, exists := usedPortsByIP[externalIP]; !exists {
			usedPortsByIP[externalIP] = make(map[int]bool)
		}

		// Register the used ports
		for _, port := range svc.Spec.Ports {
			usedPortsByIP[externalIP][int(port.Port)] = true
			logger.Info("Port registered as in use", "ip", externalIP, "port", port.Port)
		}
	}

	return usedPortsByIP, nil
}
