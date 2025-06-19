// Copyright 2020-2025 Politecnico di Torino
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
	"fmt"

	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
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
	cluster := environment.Cluster
	// Enforce the service presence
	service := v1.Service{ObjectMeta: forge.ObjectMeta(instance)}
	if environment.EnvironmentType == clv1alpha2.ClassCluster {
		if cluster.ControlPlane.Provider == clv1alpha2.ProviderKamaji {
			service = v1.Service{ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-control-plane", cluster.Name),
				Namespace: instance.Namespace,
			}}
		} else {
			service = v1.Service{ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-cluster-lb", cluster.Name),
				Namespace: instance.Namespace,
			}}
		}
		if err := r.Get(ctx, client.ObjectKeyFromObject(&service), &service); client.IgnoreNotFound(err) != nil {
			log.Error(err, "failed to retrieve clusterservice", "clusterservice", klog.KObj(&service))
			return err
		} else if err != nil {
			klog.Infof("clusterservice doesn't exist")
			return nil
		} else {
			labels := forge.InstanceObjectLabels(service.GetLabels(), instance)
			service.SetLabels(labels)
			if cluster.ControlPlane.Provider == clv1alpha2.ProviderKubeadm {
				service.Spec.Ports[0].Name = forge.ClusterPortName
			}
		}
	} else {
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
	}

	instance.Status.IP = service.Spec.ClusterIP

	// No need to create ingress resources in case of gui-less VMs.
	if (environment.EnvironmentType == clv1alpha2.ClassVM || environment.EnvironmentType == clv1alpha2.ClassCloudVM) && !environment.GuiEnabled {
		return nil
	}

	// Enforce the ingress to access the environment GUI

	host := forge.HostName(r.ServiceUrls.WebsiteBaseURL, environment.Mode)
	fmt.Println("host:", host)
	fmt.Println("service.GetName:", service.GetName())
	fmt.Println("forge.IngressGUIPath:", forge.IngressGUIPath(instance, environment))
	// cluster uses passthrough mode not ingress which will terminate in inress side.
	if environment.EnvironmentType == clv1alpha2.ClassCluster {
		configMap := v1.ConfigMap{ObjectMeta: forge.ObjectMetaWithSuffix(instance, forge.IngressGUIName(environment))}
		res, err := ctrl.CreateOrUpdate(ctx, r.Client, &configMap, func() error {
			if configMap.CreationTimestamp.IsZero() {
				var serviceName string
				if cluster.ControlPlane.Provider == clv1alpha2.ProviderKamaji {
					serviceName = fmt.Sprintf("%s-control-plane", cluster.Name)
				} else {
					serviceName = fmt.Sprintf("%s-cluster-lb", cluster.Name)
				}
				configMap.Data = forge.ConfigMapData(instance, serviceName, environment)
			}
			configMap.SetLabels(forge.InstanceObjectLabels(configMap.GetLabels(), instance))
			configMap.SetAnnotations(forge.IngressGUIAnnotations(environment, configMap.GetAnnotations()))
			if environment.Mode == clv1alpha2.ModeStandard {
				configMap.SetAnnotations(forge.IngressAuthenticationAnnotations(configMap.GetAnnotations(), r.ServiceUrls.InstancesAuthURL))
			}
			return ctrl.SetControllerReference(instance, &configMap, r.Scheme)
		})
		if err != nil {
			log.Error(err, "failed to create object", "configmap", klog.KObj(&configMap))
			return err
		}
		log.V(utils.FromResult(res)).Info("object enforced", "configmap", klog.KObj(&configMap), "result", res)
	} else {
		ingressGUI := netv1.Ingress{ObjectMeta: forge.ObjectMetaWithSuffix(instance, forge.IngressGUIName(environment))}
		res, err := ctrl.CreateOrUpdate(ctx, r.Client, &ingressGUI, func() error {
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
	}
	instance.Status.URL = forge.IngressGuiStatusURL(host, environment, instance)
	return nil
}

// enforceInstanceExpositionAbsence ensures the absence of the objects required to expose an environment (i.e. service, ingress).
func (r *InstanceReconciler) enforceInstanceExpositionAbsence(ctx context.Context) error {
	instance := clctx.InstanceFrom(ctx)
	instance.Status.IP = ""
	instance.Status.URL = ""

	// Enforce service absence
	service := v1.Service{ObjectMeta: forge.ObjectMeta(instance)}
	if err := utils.EnforceObjectAbsence(ctx, r.Client, &service, "service"); err != nil {
		return err
	}

	// Enforce gui ingress absence
	ingressGUI := netv1.Ingress{ObjectMeta: forge.ObjectMetaWithSuffix(instance, forge.IngressGUINameSuffix)}
	if err := utils.EnforceObjectAbsence(ctx, r.Client, &ingressGUI, "ingress"); err != nil {
		return err
	}

	// Enforce configmap absence
	configMap := v1.ConfigMap{ObjectMeta: forge.ObjectMetaWithSuffix(instance, forge.IngressdashboardPathSuffix)}
	if err := utils.EnforceObjectAbsence(ctx, r.Client, &configMap, "configmap"); err != nil {
		return err
	}
	return nil
}
