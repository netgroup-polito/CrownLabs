package publicexposure

import (
	"context"
	"fmt"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	metallbPoolName = "my-ip-pool" // Unified with the controller
	sharedIPValue   = "true"       // Unified with the controller
	basePort        = 30000
)

// Manager manages the public exposure of instances
type Manager struct {
	client.Client
	Scheme *runtime.Scheme
}

// NewManager creates a new manager for public exposure
func NewManager(client client.Client, scheme *runtime.Scheme) *Manager {
	return &Manager{
		Client: client,
		Scheme: scheme,
	}
}

// ReconcileExposure reconciles the public exposure for an instance
func (m *Manager) ReconcileExposure(ctx context.Context, instance *clv1alpha2.Instance) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Reconciling instance exposure", "instance", instance.Name)

	svcName := fmt.Sprintf("instance-lb-%s", instance.Name)
	existingSvc := &v1.Service{}
	err := m.Get(ctx, types.NamespacedName{Name: svcName, Namespace: instance.Namespace}, existingSvc)
	svcExists := err == nil

	// Check if exposure is required
	if instance.Spec.PublicExposure == nil || len(instance.Spec.PublicExposure.ServicesPortMappings) == 0 {
		if svcExists {
			log.Info("Removing LoadBalancer service as exposure is no longer required", "service", svcName)
			return m.Delete(ctx, existingSvc)
		}
		return nil
	}

	// IF THE SERVICE ALREADY EXISTS, CHECK IF IT'S ALREADY CORRECT BEFORE DOING ANY OTHER OPERATION
	if svcExists {
		// Check if the current configuration is already the desired one
		if m.serviceMatchesDesiredConfig(existingSvc, instance) {
			log.Info("Service already has correct configuration, skipping update", "service", svcName)
			return nil
		}
	}

	// Get the used ports (excluding the current service)
	usedPortsByIP, err := m.updateUsedPortsByIP(ctx, instance.Namespace, svcName)
	if err != nil {
		return err
	}

	// Find the best IP and assign ports
	targetIP, assignedPorts, err := m.findBestIPAndAssignPorts(ctx, instance, usedPortsByIP)
	// fmt.Println("targetIP", targetIP, "assignedPorts", assignedPorts, "err", err)
	if err != nil {
		return err
	}

	// Create or update the service
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: instance.Namespace,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, m.Client, svc, func() error {
		// Set owner reference
		if err := controllerutil.SetControllerReference(instance, svc, m.Scheme); err != nil {
			return err
		}

		// Configure the service
		svc.Spec.Type = v1.ServiceTypeLoadBalancer
		svc.Spec.Selector = map[string]string{
			"crownlabs.polito.it/instance": instance.Name,
		}

		// Set annotations for MetalLB
		if svc.Annotations == nil {
			svc.Annotations = make(map[string]string)
		}
		svc.Annotations["metallb.universe.tf/address-pool"] = metallbPoolName
		svc.Annotations["metallb.universe.tf/allow-shared-ip"] = sharedIPValue
		svc.Annotations["metallb.universe.tf/loadBalancerIPs"] = targetIP

		// Configure service ports
		svc.Spec.Ports = []v1.ServicePort{}
		for _, p := range assignedPorts {
			svc.Spec.Ports = append(svc.Spec.Ports, v1.ServicePort{
				Name:       p.Name,
				Port:       p.AssignedPort,
				TargetPort: intstr.FromInt(int(p.TargetPort)),
				Protocol:   v1.ProtocolTCP,
			})
		}

		return nil
	})

	if err != nil {
		return err
	}

	log.Info("Service exposure reconciled", "service", svcName, "ip", targetIP)
	return nil
}

// serviceMatchesDesiredConfig checks if the existing service already matches the desired configuration
func (m *Manager) serviceMatchesDesiredConfig(svc *v1.Service, instance *clv1alpha2.Instance) bool {
	// Check that the service has the correct ports
	expectedPorts := make(map[string]clv1alpha2.ServicePortMapping)
	for _, p := range instance.Spec.PublicExposure.ServicesPortMappings {
		expectedPorts[p.Name] = p
	}

	// If the number of ports doesn't match, the service must be updated
	if len(svc.Spec.Ports) != len(expectedPorts) {
		return false
	}

	// Check each port
	for _, port := range svc.Spec.Ports {
		expectedPort, exists := expectedPorts[port.Name]
		if !exists {
			return false
		}

		// For specified ports (not 0), check that they match
		if expectedPort.Port != 0 && port.Port != expectedPort.Port {
			return false
		}

		// Check that the target port matches
		if port.TargetPort.IntVal != expectedPort.TargetPort {
			return false
		}
	}

	return true
}

// CleanupExposure removes the LoadBalancer service if no longer needed
func (m *Manager) CleanupExposure(ctx context.Context, instance *clv1alpha2.Instance) error {
	log := ctrl.LoggerFrom(ctx)
	svcName := fmt.Sprintf("instance-lb-%s", instance.Name)

	svc := &v1.Service{}
	err := m.Get(ctx, types.NamespacedName{Name: svcName, Namespace: instance.Namespace}, svc)
	if err == nil {
		log.Info("Removing LoadBalancer service", "service", svcName)
		return m.Delete(ctx, svc)
	}
	return client.IgnoreNotFound(err)
}
