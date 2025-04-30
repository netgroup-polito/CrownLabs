// This file shows the modifications needed for operators/pkg/instctrl/exposition.go
// to support RDP via Guacamole instead of VNC

// enforceInstanceExpositionPresence ensures the presence of the objects required to expose an environment (i.e. service, ingress).
func (r *InstanceReconciler) enforceInstanceExpositionPresence(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)

	// Enforce the service presence
	service := v1.Service{ObjectMeta: forge.ObjectMeta(instance)}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, &service, func() error {
		// Service specifications are forged only at creation time, to prevent issues in case of updates.
		// Indeed, enforcing the specs may cause service disruption if they diverge from the backend
		// (i.e., VMI or Pod) configuration, which nonetheless cannot be changed without a restart.
		if service.CreationTimestamp.IsZero() {
			service.Spec = forge.ServiceSpec(instance, environment)
		}

		labels := forge.InstanceObjectLabels(service.GetLabels(), instance)
		if environment.EnvironmentType == clv1alpha2.ClassContainer {
			labels = forge.MonitorableServiceLabels(labels)
		}
		service.SetLabels(labels)

		return ctrl.SetControllerReference(instance, &service, r.Scheme)
	})

	if err != nil {
		log.Error(err, "failed to create object", "service", klog.KObj(&service))
		return err
	}
	log.V(utils.FromResult(res)).Info("object enforced", "service", klog.KObj(&service), "result", res)
	instance.Status.IP = service.Spec.ClusterIP

	// No need to create ingress resources in case of gui-less VMs.
	if (environment.EnvironmentType == clv1alpha2.ClassVM || environment.EnvironmentType == clv1alpha2.ClassCloudVM) && !environment.GuiEnabled {
		return nil
	}

	// Enforce the ingress to access the environment GUI
	host := forge.HostName(r.ServiceUrls.WebsiteBaseURL, environment.Mode)

	// MODIFICATION: For VMs with GUI, we now use Guacamole for RDP access instead of direct VNC
	if (environment.EnvironmentType == clv1alpha2.ClassVM || environment.EnvironmentType == clv1alpha2.ClassCloudVM) && environment.GuiEnabled {
		return r.enforceGuacamoleIngress(ctx, instance, environment, service, host)
	}

	// For other environment types (container, standalone), continue using the original method
	ingressGUI := netv1.Ingress{ObjectMeta: forge.ObjectMetaWithSuffix(instance, forge.IngressGUIName(environment))}

	res, err = ctrl.CreateOrUpdate(ctx, r.Client, &ingressGUI, func() error {
		// Ingress specifications are forged only at creation time, to prevent issues in case of updates.
		// Indeed, enforcing the specs may cause service disruption if they diverge from the service configuration.
		if ingressGUI.CreationTimestamp.IsZero() {
			ingressGUI.Spec = forge.IngressSpec(host, forge.IngressGUIPath(instance, environment),
				forge.IngressDefaultCertificateName, service.GetName(), forge.GUIPortName)
		}
		ingressGUI.SetLabels(forge.InstanceObjectLabels(ingressGUI.GetLabels(), instance))

		ingressGUI.SetAnnotations(forge.IngressGUIAnnotations(environment, ingressGUI.GetAnnotations()))

		if environment.Mode == clv1alpha2.ModeStandard {
			ingressGUI.SetAnnotations(forge.IngressAuthenticationAnnotations(ingressGUI.GetAnnotations(), r.ServiceUrls.InstancesAuthURL))
		}

		return ctrl.SetControllerReference(instance, &ingressGUI, r.Scheme)
	})

	if err != nil {
		log.Error(err, "failed to create object", "ingress", klog.KObj(&ingressGUI))
		return err
	}

	log.V(utils.FromResult(res)).Info("object enforced", "ingress", klog.KObj(&ingressGUI), "result", res)
	instance.Status.URL = forge.IngressGuiStatusURL(host, environment, instance)

	return nil
}

// NEW FUNCTION: enforceGuacamoleIngress creates an ingress that routes through Guacamole for RDP access
func (r *InstanceReconciler) enforceGuacamoleIngress(ctx context.Context, instance *clv1alpha2.Instance, 
	environment *clv1alpha2.Environment, service v1.Service, host string) error {
	
	log := ctrl.LoggerFrom(ctx)
	
	// Create ingress for Guacamole access
	ingressRDP := netv1.Ingress{ObjectMeta: forge.ObjectMetaWithSuffix(instance, forge.IngressGUIName(environment))}
	
	// Configure Guacamole service name and connection details
	guacamoleServiceName := "guacamole" // Name of the Guacamole service
	guacamoleNamespace := r.ServiceUrls.GuacamoleNamespace  // Namespace where Guacamole is deployed
	
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, &ingressRDP, func() error {
		// Set up path, annotations and labels for the ingress
		if ingressRDP.CreationTimestamp.IsZero() {
			// Create a path that includes the necessary parameters for Guacamole to connect to this VM via RDP
			rdpPath := forge.GuacamoleRDPPath(instance, service.Spec.ClusterIP)
			ingressRDP.Spec = forge.GuacamoleIngressSpec(host, rdpPath, forge.IngressDefaultCertificateName, 
				guacamoleServiceName, guacamoleNamespace, "http")
		}
		
		ingressRDP.SetLabels(forge.InstanceObjectLabels(ingressRDP.GetLabels(), instance))
		
		// Add annotations for Guacamole
		annotations := map[string]string{
			"nginx.ingress.kubernetes.io/proxy-read-timeout": "3600",
			"nginx.ingress.kubernetes.io/proxy-send-timeout": "3600",
			"nginx.ingress.kubernetes.io/proxy-connect-timeout": "60",
			"nginx.ingress.kubernetes.io/rewrite-target": forge.GuacamoleRewriteTarget(instance),
		}
		ingressRDP.SetAnnotations(annotations)
		
		// Add authentication annotations if in standard mode
		if environment.Mode == clv1alpha2.ModeStandard {
			ingressRDP.SetAnnotations(forge.IngressAuthenticationAnnotations(ingressRDP.GetAnnotations(), r.ServiceUrls.InstancesAuthURL))
		}
		
		return ctrl.SetControllerReference(instance, &ingressRDP, r.Scheme)
	})
	
	if err != nil {
		log.Error(err, "failed to create RDP ingress object", "ingress", klog.KObj(&ingressRDP))
		return err
	}
	
	log.V(utils.FromResult(res)).Info("RDP ingress object enforced", "ingress", klog.KObj(&ingressRDP), "result", res)
	
	// Set the URL that users will use to access the VM
	instance.Status.URL = forge.GuacamoleStatusURL(host, instance)
	
	return nil
}

// Additional functions that need to be added to the forge package:

// In forge/ingresses.go:

// GuacamoleRDPPath returns the path for accessing a VM via Guacamole RDP
func GuacamoleRDPPath(instance *clv1alpha2.Instance, vmIP string) string {
	return fmt.Sprintf("/guacamole/#/client/c/%s", instance.UID)
}

// GuacamoleIngressSpec creates an ingress spec that routes to the Guacamole service
func GuacamoleIngressSpec(host, path, certificateName, serviceName, serviceNamespace, servicePort string) netv1.IngressSpec {
	pathTypePrefix := netv1.PathTypePrefix
	return netv1.IngressSpec{
		TLS: []netv1.IngressTLS{{Hosts: []string{host}, SecretName: certificateName}},
		Rules: []netv1.IngressRule{{
			Host: host,
			IngressRuleValue: netv1.IngressRuleValue{
				HTTP: &netv1.HTTPIngressRuleValue{
					Paths: []netv1.HTTPIngressPath{{
						Path:     path,
						PathType: &pathTypePrefix,
						Backend: netv1.IngressBackend{
							Service: &netv1.IngressServiceBackend{
								Name: serviceName,
								Port: netv1.ServiceBackendPort{Name: servicePort},
							},
						},
					}},
				},
			},
		}},
	}
}

// GuacamoleRewriteTarget returns the rewrite target for Guacamole
func GuacamoleRewriteTarget(instance *clv1alpha2.Instance) string {
	return "/guacamole/"
}

// GuacamoleStatusURL returns the URL to access a VM via Guacamole
func GuacamoleStatusURL(host string, instance *clv1alpha2.Instance) string {
	return fmt.Sprintf("https://%s/guacamole/#/client/c/%s", host, instance.UID)
}

// Additional changes required in the ServiceUrls struct in operator/pkg/instctrl/controller.go:

/*
type ServiceUrls struct {
	WebsiteBaseURL   string
	InstancesAuthURL string
	GuacamoleNamespace string // Add this field
}
*/ 