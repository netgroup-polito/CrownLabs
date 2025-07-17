package instctrl

import (
	"context"
	"reflect"

	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// EnforcePublicExposure ensures the presence or absence of the LoadBalancer service for public exposure.
func (r *InstanceReconciler) EnforcePublicExposure(ctx context.Context) error {
	instance := clctx.InstanceFrom(ctx)
	template := clctx.TemplateFrom(ctx)

	if template.Spec.AllowPublicExposure &&
		instance.Spec.Running && instance.Spec.PublicExposure != nil && len(instance.Spec.PublicExposure.Ports) > 0 {
		return r.enforcePublicExposurePresence(ctx)
	}
	return r.enforcePublicExposureAbsence(ctx)
}

// enforcePublicExposurePresence ensures the presence and correctness of the LoadBalancer service.
func (r *InstanceReconciler) enforcePublicExposurePresence(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)

	svcName := forge.LoadBalancerServiceName(instance)
	service := &v1.Service{}
	service.SetName(svcName)
	service.SetNamespace(instance.Namespace)

	// Try to get the existing service
	err := r.Client.Get(ctx, client.ObjectKey{Name: svcName, Namespace: instance.Namespace}, service)
	serviceExists := err == nil

	// If the service exists, check if its current spec matches the desired spec
	if serviceExists {
		desiredPorts := instance.Spec.PublicExposure.Ports
		currentPorts := []clv1alpha2.PublicServicePort{}
		for _, p := range service.Spec.Ports {
			currentPorts = append(currentPorts, clv1alpha2.PublicServicePort{
				Name:       p.Name,
				Port:       p.Port,
				TargetPort: p.TargetPort.IntVal,
			})
		}
		currentIP := service.Annotations[forge.MetallbLoadBalancerIPsAnnotation]

		// If the current IP and ports match the desired, skip update
		if reflect.DeepEqual(desiredPorts, currentPorts) && currentIP != "" {
			// Also update status if needed
			newStatus := &clv1alpha2.InstancePublicExposureStatus{
				ExternalIP: currentIP,
				Ports:      currentPorts,
				Phase:      clv1alpha2.PublicExposurePhaseReady,
			}
			if !reflect.DeepEqual(instance.Status.PublicExposure, newStatus) {
				instance.Status.PublicExposure = newStatus
			}
			return nil
		}
	}

	// 1. Retrieve the map of used ports by other LoadBalancer services
	usedPortsByIP, err := UpdateUsedPortsByIP(ctx, r.Client, svcName, instance.Namespace)
	if err != nil {
		log.Error(err, "failed to get used ports by IP")
		return err
	}

	// 2. Find the best IP and ports to assign using the logic from ip_manager.go
	targetIP, assignedPorts, err := r.FindBestIPAndAssignPorts(ctx, r.Client, instance, usedPortsByIP)
	if err != nil {
		log.Error(err, "failed to assign IP and ports for public exposure")
		return err
	}

	// 3. Create or update the LoadBalancer Service
	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		// Set owner reference
		if err := controllerutil.SetControllerReference(instance, service, r.Scheme); err != nil {
			return err
		}

		// Set labels
		if service.Labels == nil {
			service.Labels = forge.LoadBalancerServiceLabels()
		}

		// Set annotations
		if service.Annotations == nil {
			service.Annotations = forge.LoadBalancerServiceAnnotations(targetIP)
		}

		// Set spec
		service.Spec = forge.LoadBalancerServiceSpec(instance, assignedPorts)

		return nil
	})

	if err != nil {
		log.Error(err, "failed to create or update LoadBalancer service", "service", service.GetName())
		return err
	}
	log.V(utils.FromResult(op)).Info("LoadBalancer service enforced", "service", service.GetName(), "result", op)

	// 4. Update the instance status
	newStatus := &clv1alpha2.InstancePublicExposureStatus{
		ExternalIP: targetIP,
		Ports:      assignedPorts,
		Phase:      clv1alpha2.PublicExposurePhaseReady,
	}
	if !reflect.DeepEqual(instance.Status.PublicExposure, newStatus) {
		instance.Status.PublicExposure = newStatus
	}

	return nil
}

// enforcePublicExposureAbsence ensures the absence of the LoadBalancer service.
func (r *InstanceReconciler) enforcePublicExposureAbsence(ctx context.Context) error {
	instance := clctx.InstanceFrom(ctx)
	svcName := forge.LoadBalancerServiceName(instance)
	service := &v1.Service{}
	service.SetName(svcName)
	service.SetNamespace(instance.Namespace)

	// Remove the service if it exists
	if err := utils.EnforceObjectAbsence(ctx, r.Client, service, "service"); err != nil {
		return err
	}

	// Clean up the status
	instance.Status.PublicExposure = nil
	return nil
}
