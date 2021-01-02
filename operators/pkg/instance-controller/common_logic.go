package instance_controller

import (
	"context"

	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	instance_creation "github.com/netgroup-polito/CrownLabs/operators/pkg/instance-creation"
)

// CreateInstanceExpositionEnvironment creates the components necessary to access the environment (service, ingress and oauth2-proxy related resources).
func (r *InstanceReconciler) CreateInstanceExpositionEnvironment(
	ctx context.Context,
	instance *crownlabsv1alpha2.Instance,
	name string,
) (v1.Service, networkingv1.Ingress, error) {
	// create Service to expose the pod
	service := instance_creation.ForgeService(name, instance.Namespace)

	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, &service, func() error {
		return ctrl.SetControllerReference(instance, &service, r.Scheme)
	}); err != nil {
		r.setInstanceStatus(ctx, "Could not create service "+service.Name+" in namespace "+service.Namespace+": "+err.Error(), "Error", "ServiceNotCreated", instance, "", "")
		return v1.Service{}, networkingv1.Ingress{}, err
	}

	r.setInstanceStatus(ctx, "Service "+service.Name+" correctly created in namespace "+service.Namespace, "Normal", "ServiceCreated", instance, "", "")

	urlUUID := uuid.New().String()

	// create Ingress to manage the service
	ingress := instance_creation.ForgeIngress(name, instance.Namespace, &service, urlUUID, r.WebsiteBaseURL)

	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, &ingress, func() error {
		return ctrl.SetControllerReference(instance, &ingress, r.Scheme)
	}); err != nil {
		r.setInstanceStatus(ctx, "Could not create ingress "+ingress.Name+" in namespace "+ingress.Namespace+": "+err.Error(), "Error", "IngressNotCreated", instance, "", "")
		return service, networkingv1.Ingress{}, err
	}

	r.setInstanceStatus(ctx, "Ingress "+ingress.Name+" correctly created in namespace "+ingress.Namespace, "Normal", "IngressCreated", instance, "", "")

	if err := r.createOAUTHlogic(name, instance, instance.Namespace, urlUUID); err != nil {
		return service, ingress, err
	}

	return service, ingress, nil
}
