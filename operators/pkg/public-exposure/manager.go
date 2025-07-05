package publicexposure

import (
	"context"
	"fmt"
	"reflect"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// Manager manages the public exposure of instances.
type Manager struct {
	client.Client
	Scheme *runtime.Scheme
}

// NewManager creates a new manager for public exposure.
func NewManager(client client.Client, scheme *runtime.Scheme) *Manager {
	return &Manager{
		Client: client,
		Scheme: scheme,
	}
}

// serviceName generates the predictable name for the LoadBalancer service.
func (m *Manager) serviceName(instance *clv1alpha2.Instance) string {
	return fmt.Sprintf("inst-lb-%s", instance.Name)
}

// ReconcileExposure reconciles the public exposure for an instance.
func (m *Manager) ReconcileExposure(ctx context.Context, instance *clv1alpha2.Instance) error {
	log := ctrl.LoggerFrom(ctx).WithName("exposure-manager")

	// Use the helper function to get the service name.
	svcName := m.serviceName(instance)
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: instance.Namespace,
		},
	}

	// Case 1: No public exposure is required. Ensure the service is deleted.
	if instance.Spec.PublicExposure == nil || len(instance.Spec.PublicExposure.Ports) == 0 {
		log.Info("Public exposure not requested, ensuring service is absent", "service", svcName)
		if err := m.Delete(ctx, svc); err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete service %s: %w", svcName, err)
		}
		// Clear status if it was previously set
		if instance.Status.PublicExposure != nil {
			instance.Status.PublicExposure = nil
		}
		return nil
	}

	// Case 2: Public exposure is required.
	log.Info("Public exposure requested, reconciling service", "service", svcName)

	// Get all currently used ports in the cluster to avoid collisions.
	// We exclude the service we're about to create/update from this check.
	usedPortsByIP, err := m.updateUsedPortsByIP(ctx, svcName, instance.Namespace)
	if err != nil {
		return fmt.Errorf("failed to check for used ports: %w", err)
	}

	// Find the best IP and assign ports.
	targetIP, assignedPorts, err := m.findBestIPAndAssignPorts(ctx, instance, usedPortsByIP)
	if err != nil {
		return fmt.Errorf("failed to assign IP and ports: %w", err)
	}

	// Create or Update the service.
	op, err := controllerutil.CreateOrUpdate(ctx, m.Client, svc, func() error {
		// Set owner reference to tie the service's lifecycle to the instance.
		if err := controllerutil.SetControllerReference(instance, svc, m.Scheme); err != nil {
			return err
		}

		// Set the specific label for this component.
		if svc.Labels == nil {
			svc.Labels = make(map[string]string)
		}
		svc.Labels[PublicExposureComponentLabelKey] = PublicExposureComponentLabelValue

		// Configure the service spec.
		svc.Spec.Type = v1.ServiceTypeLoadBalancer
		svc.Spec.Selector = map[string]string{
			"crownlabs.polito.it/instance": instance.Name,
		}

		// Set annotations for MetalLB.
		if svc.Annotations == nil {
			svc.Annotations = make(map[string]string)
		}
		svc.Annotations[MetallbAddressPoolAnnotation] = DefaultAddressPool
		svc.Annotations[MetallbAllowSharedIPAnnotation] = AllowSharedIPValue
		svc.Annotations[MetallbLoadBalancerIPsAnnotation] = targetIP

		// Configure service ports.
		svcPorts := make([]v1.ServicePort, len(assignedPorts))
		for i, p := range assignedPorts {
			svcPorts[i] = v1.ServicePort{
				Name:       p.Name,
				Port:       p.Port,
				TargetPort: intstr.FromInt(int(p.TargetPort)),
				Protocol:   v1.ProtocolTCP,
			}
		}
		svc.Spec.Ports = svcPorts

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to reconcile service %s (operation: %s): %w", svcName, op, err)
	}

	log.Info("Service exposure reconciled successfully", "service", svcName, "ip", targetIP, "operation", op)

	// Update the instance status with the exposure details.
	newStatus := &clv1alpha2.InstancePublicExposureStatus{
		ExternalIP: targetIP,
		Ports:      assignedPorts,
	}

	// Only update status if it has changed to avoid unnecessary reconciles.
	if !reflect.DeepEqual(instance.Status.PublicExposure, newStatus) {
		instance.Status.PublicExposure = newStatus
		log.Info("Updating instance status with new exposure details")
	}

	return nil
}
