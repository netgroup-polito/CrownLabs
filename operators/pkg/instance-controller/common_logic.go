package instance_controller

import (
	"context"

	"github.com/google/uuid"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	instance_creation "github.com/netgroup-polito/CrownLabs/operators/pkg/instance-creation"
)

// CreateInstanceExpositionEnvironment creates the components necessary to access the environment (service, ingress and oauth2-proxy related resources).
// Additionally, it makes the service expose another port and creates an ingress for FileBrowser sidecar container (only for container environments).
func (r *InstanceReconciler) CreateInstanceExpositionEnvironment(
	ctx context.Context,
	instance *crownlabsv1alpha2.Instance,
	hasFileBrowser bool,
) (v1.Service, networkingv1.Ingress, string, error) {
	namespacedName := forge.NamespaceName(instance)

	// create Service to expose the pod
	service := instance_creation.ForgeService(namespacedName.Name, namespacedName.Namespace)

	fileBrowserPortName := "filebrowser"
	if hasFileBrowser {
		service.Spec.Ports = append(service.Spec.Ports, v1.ServicePort{
			Name:     fileBrowserPortName,
			Protocol: v1.ProtocolTCP,
			Port:     8080,
		})
	}

	op, err := ctrl.CreateOrUpdate(ctx, r.Client, &service, func() error {
		return ctrl.SetControllerReference(instance, &service, r.Scheme)
	})

	if err != nil {
		r.setInstanceStatus(ctx, "Could not create service "+service.Name+" in namespace "+service.Namespace+": "+err.Error(), "Error", "ServiceNotCreated", instance, "", "")
		return v1.Service{}, networkingv1.Ingress{}, "", err
	}
	klog.Infof("Service for instance %s/%s %s", instance.GetNamespace(), instance.GetName(), op)

	urlUUID := uuid.New().String()

	// create Ingress to manage the service
	ingress := instance_creation.ForgeIngress(namespacedName.Name, namespacedName.Namespace, &service, r.WebsiteBaseURL, urlUUID, r.InstancesAuthURL)
	op, err = ctrl.CreateOrUpdate(ctx, r.Client, &ingress, func() error {
		return ctrl.SetControllerReference(instance, &ingress, r.Scheme)
	})

	if err != nil {
		r.setInstanceStatus(ctx, "Could not create ingress "+ingress.Name+" in namespace "+ingress.Namespace+": "+err.Error(), "Error", "IngressNotCreated", instance, "", "")
		return service, networkingv1.Ingress{}, "", err
	}
	klog.Infof("Ingress (gui) for instance %s/%s %s", instance.GetNamespace(), instance.GetName(), op)

	// Overwrite using the actual ingress' url UUID
	urlUUID = ingress.Annotations["crownlabs.polito.it/url-uuid"]

	if hasFileBrowser {
		// create separate Ingress for FileBrowser to manage the same service
		fileBrowserIngress := instance_creation.ForgeFileBrowserIngress(namespacedName.Name, namespacedName.Namespace, &service, urlUUID, r.WebsiteBaseURL, fileBrowserPortName, r.InstancesAuthURL)
		op, err := ctrl.CreateOrUpdate(ctx, r.Client, &fileBrowserIngress, func() error {
			return ctrl.SetControllerReference(instance, &fileBrowserIngress, r.Scheme)
		})

		if err != nil {
			r.setInstanceStatus(ctx, "Could not create ingress "+fileBrowserIngress.Name+" in namespace "+fileBrowserIngress.Namespace+": "+err.Error(), "Error", "IngressNotCreated", instance, "", "")
			return service, networkingv1.Ingress{}, urlUUID, err
		}
		klog.Infof("Ingress (filebrowser) for instance %s/%s %s", instance.GetNamespace(), instance.GetName(), op)
	}

	instance.Status.IP = service.Spec.ClusterIP
	instance.Status.URL = ingress.GetAnnotations()["crownlabs.polito.it/probe-url"]

	return service, ingress, urlUUID, nil
}
