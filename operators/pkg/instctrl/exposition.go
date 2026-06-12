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
// an environment (i.e. service, ingress or gateway), depending on whether the instance is running or not.
func (r *InstanceReconciler) EnforceInstanceExposition(ctx context.Context) error {
	instance := clctx.InstanceFrom(ctx)

	if instance.Spec.Running && r.ExpositionConfig.GatewayAPIMode {
		return r.enforceInstanceExpositionHTTPRoutePresence(ctx)
	}
	if instance.Spec.Running && !r.ExpositionConfig.GatewayAPIMode {
		return r.enforceInstanceExpositionIngressPresence(ctx)
	}

	return r.enforceInstanceExpositionAbsence(ctx)
}

// enforceInstanceExpositionHTTPRoutePresence ensures the presence of the HTTPRoute required to expose an environment.
func (r *InstanceReconciler) enforceInstanceExpositionHTTPRoutePresence(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)

	svc, err := r.enforceInstanceExpositionServicePresence(ctx)
	if err != nil {
		return err
	}
	// If service presence couldn't be ensured due to out-of-range index, nothing to do.
	if svc == nil {
		return nil
	}

	// No need to create ingress resources in case of gui-less VMs.
	if (environment.EnvironmentType == clv1alpha2.ClassVM || environment.EnvironmentType == clv1alpha2.ClassCloudVM) && !environment.GuiEnabled {
		return nil
	}

	// Enforce the HTTPRoute presence
	httpRoute := gatewayv1.HTTPRoute{ObjectMeta: forge.ObjectMetaWithSuffix(instance, environment.Name)}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, &httpRoute, func() error {
		// HTTPRoute specifications are forged only at creation time, to prevent issues in case of updates.
		// Indeed, enforcing the specs may cause service disruption if they diverge from the service configuration.
		if httpRoute.CreationTimestamp.IsZero() {
			tpl := &forge.HTTPRouteTemplate{
				Path:        forge.ExpositionGUIPath(instance, environment),
				ServiceName: svc.GetName(),
			}
			httpRoute.Spec = forge.HTTPRouteSpec(tpl, &r.ExpositionConfig, environment, forge.GUIPortNumber)
		}
		httpRoute.SetLabels(forge.EnvironmentObjectLabels(httpRoute.GetLabels(), instance, environment))

		return ctrl.SetControllerReference(instance, &httpRoute, r.Scheme)
	})

	if err != nil {
		log.Error(err, "failed to create object", "httproute", klog.KObj(&httpRoute))
		return err
	}

	log.V(utils.FromResult(res)).Info("object enforced", "httproute", klog.KObj(&httpRoute), "result", res)

	// Destroy any Ingress
	if err := r.enforceInstanceExpositionIngressAbsence(ctx); err != nil {
		log.Error(err, "failed to remove conflicting ingress", "httproute", klog.KObj(&httpRoute))
		return err
	}

	return nil
}

// enforceInstanceExpositionIngressPresence ensures the presence of the Ingress required to expose an environment.
func (r *InstanceReconciler) enforceInstanceExpositionIngressPresence(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)

	svc, err := r.enforceInstanceExpositionServicePresence(ctx)
	if err != nil {
		return err
	}
	if svc == nil {
		return nil
	}

	// No need to create ingress resources in case of gui-less VMs.
	if (environment.EnvironmentType == clv1alpha2.ClassVM || environment.EnvironmentType == clv1alpha2.ClassCloudVM) && !environment.GuiEnabled {
		return nil
	}

	// Enforce the Ingress presence
	host := r.ExpositionConfig.WebsiteBaseURL
	ingressGUI := netv1.Ingress{ObjectMeta: forge.ObjectMetaWithSuffix(instance, environment.Name)}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, &ingressGUI, func() error {
		// Ingress specifications are forged only at creation time, to prevent issues in case of updates.
		// Indeed, enforcing the specs may cause service disruption if they diverge from the service configuration.
		if ingressGUI.CreationTimestamp.IsZero() {
			ingressGUI.Spec = forge.IngressSpec(host, forge.ExpositionGUIPath(instance, environment),
				forge.IngressDefaultCertificateName, svc.GetName(), forge.GUIPortName)
		}
		ingressGUI.SetLabels(forge.EnvironmentObjectLabels(ingressGUI.GetLabels(), instance, environment))

		ingressGUI.SetAnnotations(forge.IngressGUIAnnotations(environment, ingressGUI.GetAnnotations()))

		// Add authentication annotations only if enabled.
		if r.EnableAuthentication {
			ingressGUI.SetAnnotations(forge.IngressAuthenticationAnnotations(ingressGUI.GetAnnotations(), r.ExpositionConfig.InstancesAuthURL))
		}

		return ctrl.SetControllerReference(instance, &ingressGUI, r.Scheme)
	})

	if err != nil {
		log.Error(err, "failed to create object", "ingress", klog.KObj(&ingressGUI))
		return err
	}

	log.V(utils.FromResult(res)).Info("object enforced", "ingress", klog.KObj(&ingressGUI), "result", res)

	// Destroy any HTTPRoute
	if err := r.enforceInstanceExpositionHTTPRouteAbsence(ctx); err != nil {
		log.Error(err, "failed to remove conflicting httproute", "ingress", klog.KObj(&ingressGUI))
		return err
	}

	return nil
}

// enforceInstanceExpositionServicePresence ensures the presence of the Service required to expose an environment, and returns it.
func (r *InstanceReconciler) enforceInstanceExpositionServicePresence(ctx context.Context) (*corev1.Service, error) {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)
	envIndex := clctx.EnvironmentIndexFrom(ctx)

	// Check if index is out of range
	if envIndex >= len(instance.Status.Environments) {
		return nil, nil
	}

	service := corev1.Service{ObjectMeta: forge.ObjectMetaWithSuffix(instance, environment.Name)}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, &service, func() error {
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
		return nil, err
	}

	log.V(utils.FromResult(res)).Info("object enforced", "service", klog.KObj(&service), "result", res)

	instance.Status.Environments[envIndex].IP = service.Spec.ClusterIP
	return &service, nil
}

// enforceInstanceExpositionAbsence ensures the absence of the objects required to expose an environment (i.e. service, ingress).
func (r *InstanceReconciler) enforceInstanceExpositionAbsence(ctx context.Context) error {
	instance := clctx.InstanceFrom(ctx)
	envIndex := clctx.EnvironmentIndexFrom(ctx)

	// check if Status.Environments has enough capacity
	if len(instance.Status.Environments) <= envIndex {
		// extend the array to envIndex+1 elements
		newEnvs := make([]clv1alpha2.InstanceStatusEnv, envIndex+1)
		copy(newEnvs, instance.Status.Environments)
		instance.Status.Environments = newEnvs
	}

	instance.Status.Environments[envIndex].IP = ""

	// Enforce Service absence
	if err := r.enforceInstanceExpositionServiceAbsence(ctx); err != nil {
		return err
	}
	// Enforce HTTPRoute absence
	if err := r.enforceInstanceExpositionHTTPRouteAbsence(ctx); err != nil {
		return err
	}
	// Enforce Ingress absence
	if err := r.enforceInstanceExpositionIngressAbsence(ctx); err != nil {
		return err
	}

	return nil
}

// enforceInstanceExpositionHTTPRouteAbsence removes the HTTPRoute exposed resource for the environment.
func (r *InstanceReconciler) enforceInstanceExpositionHTTPRouteAbsence(ctx context.Context) error {
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)

	httpRoute := gatewayv1.HTTPRoute{ObjectMeta: forge.ObjectMetaWithSuffix(instance, environment.Name)}
	if err := utils.EnforceObjectAbsence(ctx, r.Client, &httpRoute, "httproute"); err != nil {
		return err
	}
	return nil
}

// enforceInstanceExpositionIngressAbsence removes the Ingress exposed resource for the environment.
func (r *InstanceReconciler) enforceInstanceExpositionIngressAbsence(ctx context.Context) error {
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)

	ingressGUI := netv1.Ingress{ObjectMeta: forge.ObjectMetaWithSuffix(instance, environment.Name)}
	if err := utils.EnforceObjectAbsence(ctx, r.Client, &ingressGUI, "ingress"); err != nil {
		return err
	}
	return nil
}

// enforceInstanceExpositionServiceAbsence removes the Service exposed resource for the environment.
func (r *InstanceReconciler) enforceInstanceExpositionServiceAbsence(ctx context.Context) error {
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)

	service := corev1.Service{ObjectMeta: forge.ObjectMetaWithSuffix(instance, environment.Name)}
	if err := utils.EnforceObjectAbsence(ctx, r.Client, &service, "service"); err != nil {
		return err
	}
	return nil
}
