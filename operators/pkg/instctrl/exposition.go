// Copyright 2020-2026 Politecnico di Torino
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

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/clcontext"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// EnforceInstanceExposition ensures the presence/absence of the objects required to expose
// an environment (i.e. service, ingress), depending on whether the instance is running or not.
func (r *InstanceReconciler) EnforceInstanceExposition(ctx context.Context) error {
	instance := clctx.InstanceFrom(ctx)

	if instance.Spec.Running {
		return r.enforceInstanceExpositionPresence(ctx)
	}

	return r.enforceInstanceExpositionAbsence(ctx)
}

// enforceInstanceExpositionPresence ensures the presence of the objects required to expose an environment (i.e. service, ingress).
func (r *InstanceReconciler) enforceInstanceExpositionPresence(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)
	envIndex := clctx.EnvironmentIndexFrom(ctx)

	// Check if index is out of range
	if envIndex >= len(instance.Status.Environments) {
		return nil
	}

	// Enforce the service presence
	service := corev1.Service{ObjectMeta: forge.ObjectMetaWithSuffix(instance, environment.Name)}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, &service, func() error {
		// Service specifications are forged only at creation time, to prevent issues in case of updates.
		// Indeed, enforcing the specs may cause service disruption if they diverge from the backend
		// (i.e., VMI or Pod) configuration, which nonetheless cannot be changed without a restart.
		if service.CreationTimestamp.IsZero() {
			service.Spec = forge.ServiceSpec(instance, environment)
		}

		labels := forge.EnvironmentObjectLabels(service.GetLabels(), instance, environment)
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

	instance.Status.Environments[envIndex].IP = service.Spec.ClusterIP

	// No need to create ingress resources in case of gui-less VMs.
	if (environment.EnvironmentType == clv1alpha2.ClassVM || environment.EnvironmentType == clv1alpha2.ClassCloudVM) && !environment.GuiEnabled {
		return nil
	}

	// Use the configured website base URL
	host := r.ServiceUrls.WebsiteBaseURL

	// Enforce the external exposure presence (HTTProute if Gateway API is enabled, Ingress otherwise)
	if r.GatewayAPIMode {
		httpRoute := gatewayv1.HTTPRoute{ObjectMeta: forge.ObjectMetaWithSuffix(instance, environment.Name)}
		res, err = ctrl.CreateOrUpdate(ctx, r.Client, &httpRoute, func() error {
			// HTTPRoute specifications are forged only at creation time, to prevent issues in case of updates.
			// Indeed, enforcing the specs may cause service disruption if they diverge from the service configuration.
			if httpRoute.CreationTimestamp.IsZero() {
				gwNs, gwName, gwSection := utils.ParseGatewayParent(r.GatewayAPIRefsValues)
				params := &forge.HTTPRouteSpecParams{
					Host:               host,
					Path:               forge.ExposeGUIPath(instance, environment),
					ServiceName:        service.GetName(),
					GatewayName:        gwName,
					GatewayNamespace:   gwNs,
					GatewaySectionName: gwSection,
				}
				httpRoute.Spec = forge.HTTPRouteSpec(params, environment, forge.GUIPortNumber)
			}
			httpRoute.SetLabels(forge.EnvironmentObjectLabels(httpRoute.GetLabels(), instance, environment))

			return ctrl.SetControllerReference(instance, &httpRoute, r.Scheme)
		})

		if err != nil {
			log.Error(err, "failed to create object", "httproute", klog.KObj(&httpRoute))
			return err
		}

		log.V(utils.FromResult(res)).Info("object enforced", "httproute", klog.KObj(&httpRoute), "result", res)

		// TODO HTTPROUTE: it makes sense to enforce the absence of the Ingress resource even in Gateway API mode, to prevent conflicts in case of a live flag change?
		// // If an Ingress with the same ObjectMeta exists (e.g. flag changed live), remove it.
		// ingressGUI := netv1.Ingress{ObjectMeta: forge.ObjectMetaWithSuffix(instance, environment.Name)}
		// if err := utils.EnforceObjectAbsence(ctx, r.Client, &ingressGUI, "ingress"); err != nil {
		// 	log.Error(err, "failed to delete conflicting ingress", "ingress", klog.KObj(&ingressGUI))
		// 	return err
		// }
	} else {
		ingressGUI := netv1.Ingress{ObjectMeta: forge.ObjectMetaWithSuffix(instance, environment.Name)}
		res, err = ctrl.CreateOrUpdate(ctx, r.Client, &ingressGUI, func() error {
			// Ingress specifications are forged only at creation time, to prevent issues in case of updates.
			// Indeed, enforcing the specs may cause service disruption if they diverge from the service configuration.
			if ingressGUI.CreationTimestamp.IsZero() {
				ingressGUI.Spec = forge.IngressSpec(host, forge.ExposeGUIPath(instance, environment),
					forge.IngressDefaultCertificateName, service.GetName(), forge.GUIPortName)
			}
			ingressGUI.SetLabels(forge.EnvironmentObjectLabels(ingressGUI.GetLabels(), instance, environment))

			ingressGUI.SetAnnotations(forge.IngressGUIAnnotations(environment, ingressGUI.GetAnnotations()))

			// Add authentication annotations only if enabled.
			if r.EnableAuthentication {
				ingressGUI.SetAnnotations(forge.IngressAuthenticationAnnotations(ingressGUI.GetAnnotations(), r.ServiceUrls.InstancesAuthURL))
			}

			return ctrl.SetControllerReference(instance, &ingressGUI, r.Scheme)
		})

		if err != nil {
			log.Error(err, "failed to create object", "ingress", klog.KObj(&ingressGUI))
			return err
		}

		log.V(utils.FromResult(res)).Info("object enforced", "ingress", klog.KObj(&ingressGUI), "result", res)

		// TODO HTTPROUTE: it makes sense to enforce the absence of the HTTPRoute resource even in non-Gateway API mode, to prevent conflicts in case of a live flag change?
		// // If an HTTPRoute with the same ObjectMeta exists (e.g. flag changed live), remove it.
		// httpRoute := gatewayv1.HTTPRoute{ObjectMeta: forge.ObjectMetaWithSuffix(instance, environment.Name)}
		// if err := utils.EnforceObjectAbsence(ctx, r.Client, &httpRoute, "httproute"); err != nil {
		// 	log.Error(err, "failed to delete conflicting httproute", "httproute", klog.KObj(&httpRoute))
		// 	return err
		// }
	}

	return nil
}

// enforceInstanceExpositionAbsence ensures the absence of the objects required to expose an environment (i.e. service, ingress).
func (r *InstanceReconciler) enforceInstanceExpositionAbsence(ctx context.Context) error {
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)
	envIndex := clctx.EnvironmentIndexFrom(ctx)

	// check if Status.Environments has enough capacity
	if len(instance.Status.Environments) <= envIndex {
		// extend the array to envIndex+1 elements
		newEnvs := make([]clv1alpha2.InstanceStatusEnv, envIndex+1)
		copy(newEnvs, instance.Status.Environments)
		instance.Status.Environments = newEnvs
	}

	instance.Status.Environments[envIndex].IP = ""

	// Enforce service absence
	service := corev1.Service{ObjectMeta: forge.ObjectMetaWithSuffix(instance, environment.Name)}
	if err := utils.EnforceObjectAbsence(ctx, r.Client, &service, "service"); err != nil {
		return err
	}

	// Enforce the external exposure absence (HTTPRoute if Gateway API is enabled, Ingress otherwise)
	if r.GatewayAPIMode {
		httpRoute := gatewayv1.HTTPRoute{ObjectMeta: forge.ObjectMetaWithSuffix(instance, environment.Name)}
		if err := utils.EnforceObjectAbsence(ctx, r.Client, &httpRoute, "httproute"); err != nil {
			return err
		}
	} else {
		ingressGUI := netv1.Ingress{ObjectMeta: forge.ObjectMetaWithSuffix(instance, environment.Name)}
		if err := utils.EnforceObjectAbsence(ctx, r.Client, &ingressGUI, "ingress"); err != nil {
			return err
		}
	}

	return nil
}
