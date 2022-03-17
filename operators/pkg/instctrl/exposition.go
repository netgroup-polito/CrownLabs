// Copyright 2020-2022 Politecnico di Torino
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

	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
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

	// No need to create the file-browser ingress resource in case of VM or restricted modes.
	if environment.EnvironmentType == clv1alpha2.ClassVM || environment.EnvironmentType == clv1alpha2.ClassCloudVM || environment.Mode != clv1alpha2.ModeStandard {
		return nil
	}

	// Enforce the ingress to access the environment "MyDrive"
	path := forge.IngressMyDrivePath(instance)
	ingressMyDrive := netv1.Ingress{ObjectMeta: forge.ObjectMetaWithSuffix(instance, forge.IngressMyDriveNameSuffix)}
	res, err = ctrl.CreateOrUpdate(ctx, r.Client, &ingressMyDrive, func() error {
		// Ingress specifications are forged only at creation time, to prevent issues in case of updates.
		// Indeed, enforcing the specs may cause service disruption if they diverge from the service configuration.
		if ingressMyDrive.CreationTimestamp.IsZero() {
			ingressMyDrive.Spec = forge.IngressSpec(host, path,
				forge.IngressDefaultCertificateName, service.GetName(), forge.MyDrivePortName)
		}

		ingressMyDrive.SetLabels(forge.InstanceObjectLabels(ingressMyDrive.GetLabels(), instance))
		ingressMyDrive.SetAnnotations(forge.IngressMyDriveAnnotations(ingressMyDrive.GetAnnotations()))
		ingressMyDrive.SetAnnotations(forge.IngressAuthenticationAnnotations(ingressMyDrive.GetAnnotations(), r.ServiceUrls.InstancesAuthURL))
		return ctrl.SetControllerReference(instance, &ingressMyDrive, r.Scheme)
	})

	if err != nil {
		log.Error(err, "failed to create object", "ingress", klog.KObj(&ingressMyDrive))
		return err
	}
	log.V(utils.FromResult(res)).Info("object enforced", "ingress", klog.KObj(&ingressMyDrive), "result", res)
	instance.Status.MyDriveURL = "https://" + host + path

	return nil
}

// enforceInstanceExpositionAbsence ensures the absence of the objects required to expose an environment (i.e. service, ingress).
func (r *InstanceReconciler) enforceInstanceExpositionAbsence(ctx context.Context) error {
	instance := clctx.InstanceFrom(ctx)
	instance.Status.IP = ""
	instance.Status.URL = ""
	instance.Status.MyDriveURL = ""

	// Enforce service absence
	service := v1.Service{ObjectMeta: forge.ObjectMeta(instance)}
	if err := r.enforceObjectAbsence(ctx, &service, "service"); err != nil {
		return err
	}

	// Enforce gui ingress absence
	ingressGUI := netv1.Ingress{ObjectMeta: forge.ObjectMetaWithSuffix(instance, forge.IngressGUINameSuffix)}
	if err := r.enforceObjectAbsence(ctx, &ingressGUI, "ingress"); err != nil {
		return err
	}

	// Enforce file-browser ingress absence
	ingressFB := netv1.Ingress{ObjectMeta: forge.ObjectMetaWithSuffix(instance, forge.IngressMyDriveNameSuffix)}
	return r.enforceObjectAbsence(ctx, &ingressFB, "ingress")
}

// enforceObjectAbsence deletes a Kubernetes object and prints the appropriate log messages, without failing if it does not exist.
func (r *InstanceReconciler) enforceObjectAbsence(ctx context.Context, obj client.Object, kind string) error {
	if err := r.Delete(ctx, obj); err != nil {
		if !kerrors.IsNotFound(err) {
			ctrl.LoggerFrom(ctx).Error(err, "failed to delete object", kind, klog.KObj(obj))
			return err
		}
		ctrl.LoggerFrom(ctx).V(utils.LogDebugLevel).Info("the object was already removed", kind, klog.KObj(obj))
	} else {
		ctrl.LoggerFrom(ctx).V(utils.LogInfoLevel).Info("object correctly removed", kind, klog.KObj(obj))
	}

	return nil
}
