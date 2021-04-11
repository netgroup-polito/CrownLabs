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
// Additionally, it makes the service expose another port and creates an ingress for FileBrowser sidecar container (only for container environments).
func (r *InstanceReconciler) CreateInstanceExpositionEnvironment(
	ctx context.Context,
	instance *crownlabsv1alpha2.Instance,
	name string, hasFileBrowser bool,
) (v1.Service, networkingv1.Ingress, string, error) {
	// create Service to expose the pod
	service := instance_creation.ForgeService(name, instance.Namespace)

	fileBrowserPortName := "filebrowser"
	if hasFileBrowser {
		service.Spec.Ports = append(service.Spec.Ports, v1.ServicePort{
			Name:     fileBrowserPortName,
			Protocol: v1.ProtocolTCP,
			Port:     8080,
		})
	}

	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, &service, func() error {
		return ctrl.SetControllerReference(instance, &service, r.Scheme)
	}); err != nil {
		r.setInstanceStatus(ctx, "Could not create service "+service.Name+" in namespace "+service.Namespace+": "+err.Error(), "Error", "ServiceNotCreated", instance, "", "")
		return v1.Service{}, networkingv1.Ingress{}, "", err
	}

	msg := "Service " + service.Name + " correctly created in namespace " + service.Namespace
	if hasFileBrowser {
		msg += " (includes FileBrowser exposition)"
	}

	r.setInstanceStatus(ctx, msg, "Normal", "ServiceCreated", instance, "", "")

	urlUUID := uuid.New().String()

	// create Ingress to manage the service
	ingress := instance_creation.ForgeIngress(name, instance.Namespace, &service, urlUUID, r.WebsiteBaseURL)

	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, &ingress, func() error {
		return ctrl.SetControllerReference(instance, &ingress, r.Scheme)
	}); err != nil {
		r.setInstanceStatus(ctx, "Could not create ingress "+ingress.Name+" in namespace "+ingress.Namespace+": "+err.Error(), "Error", "IngressNotCreated", instance, "", "")
		return service, networkingv1.Ingress{}, "", err
	}

	r.setInstanceStatus(ctx, "Ingress "+ingress.Name+" correctly created in namespace "+ingress.Namespace, "Normal", "IngressCreated", instance, "", "")

	// Overwrite using the actual ingress' url UUID
	urlUUID = ingress.Annotations["crownlabs.polito.it/url-uuid"]

	if hasFileBrowser {
		// create separate Ingress for FileBrowser to manage the same service
		fileBrowserIngress := instance_creation.ForgeFileBrowserIngress(name, instance.Namespace, &service, urlUUID, r.WebsiteBaseURL, fileBrowserPortName)

		if _, err := ctrl.CreateOrUpdate(ctx, r.Client, &fileBrowserIngress, func() error {
			return ctrl.SetControllerReference(instance, &fileBrowserIngress, r.Scheme)
		}); err != nil {
			r.setInstanceStatus(ctx, "Could not create ingress "+fileBrowserIngress.Name+" in namespace "+fileBrowserIngress.Namespace+": "+err.Error(), "Error", "IngressNotCreated", instance, "", "")
			return service, networkingv1.Ingress{}, urlUUID, err
		}

		r.setInstanceStatus(ctx, "Ingress "+fileBrowserIngress.Name+" correctly created in namespace "+fileBrowserIngress.Namespace, "Normal", "IngressCreated", instance, "", "")
	}

	if err := r.createOAUTHlogic(name, instance, instance.Namespace, urlUUID); err != nil {
		return service, ingress, urlUUID, err
	}

	return service, ingress, urlUUID, nil
}
