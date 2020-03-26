/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	virtv1 "github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/kubeVirt/api/v1"
	"github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/pkg"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"

	instancev1 "github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/api/v1"
	templatev1 "github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/labTemplate/api/v1"
)

// LabInstanceReconciler reconciles a LabInstance object
type LabInstanceReconciler struct {
	client.Client
	Log            logr.Logger
	Scheme         *runtime.Scheme
	EventsRecorder record.EventRecorder
	NamespacePrefix string
}

// +kubebuilder:rbac:groups=instance.crown.team.com,resources=labinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=instance.crown.team.com,resources=labinstances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events/status,verbs=get

func (r *LabInstanceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("labinstance", req.NamespacedName)

	// get labInstance
	var labInstance instancev1.LabInstance
	if err := r.Get(ctx, req.NamespacedName, &labInstance); err != nil {
		// reconcile was triggered by a delete request
		log.Info("LabInstance " + req.Name + " deleted")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// perform reconcile only if the LabInstance belongs to the watched namespaces
	if !strings.HasPrefix(labInstance.Namespace, r.NamespacePrefix){
		return ctrl.Result{}, nil
	}

	// The metadata.generation value is incremented for all changes, except for changes to .metadata or .status
	// if metadata.generation is not incremented there's no need to reconcile
	if labInstance.Status.ObservedGeneration == labInstance.ObjectMeta.Generation {
		return ctrl.Result{}, nil
	}

	// check if labTemplate exists
	templateName := types.NamespacedName{
		Namespace: labInstance.Spec.LabTemplateNamespace,
		Name:      labInstance.Spec.LabTemplateName,
	}
	var labTemplate templatev1.LabTemplate
	if err := r.Get(ctx, templateName, &labTemplate); err != nil {
		// no LabTemplate related exists
		log.Info("LabTemplate " + templateName.Name + " doesn't exist. Deleting LabInstance " + labInstance.Name)
		r.EventsRecorder.Event(&labInstance, "Warning", "LabTemplateNotFound", "LabTemplate "+templateName.Name+" not found in namespace "+labTemplate.Namespace)
		_ = r.Delete(ctx, &labInstance, &client.DeleteOptions{})
		return ctrl.Result{}, err
	}
	r.EventsRecorder.Event(&labInstance, "Normal", "LabTemplateFound", "LabTemplate "+templateName.Name+" found in namespace "+labTemplate.Namespace)

	// prepare variables common to all resources
	name := labTemplate.Name + "-" + labInstance.Spec.StudentID + "-" + fmt.Sprintf("%.8s", uuid.New().String())
	namespace := labInstance.Namespace
	// this is added so that all resources created for this LabInstance are destroyed when the LabInstance is deleted
	b := true
	labiOwnerRef := []metav1.OwnerReference{
		{
			APIVersion:         labInstance.APIVersion,
			Kind:               labInstance.Kind,
			Name:               labInstance.Name,
			UID:                labInstance.UID,
			BlockOwnerDeletion: &b,
		},
	}

	// 1: create secret referenced by VirtualMachineInstance (Cloudinit)
	secret := pkg.CreateSecret(name, namespace)
	secret.SetOwnerReferences(labiOwnerRef)
	if err := pkg.CreateOrUpdate(r.Client, ctx, log, secret); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create secret "+secret.Name+"in namespace "+secret.Namespace, "Warning", "SecretNotCreated", &labInstance, "")
	} else {
		setLabInstanceStatus(r, ctx, log, "Secret "+secret.Name+" correctly created in namespace "+secret.Namespace, "Normal", "SecretCreated", &labInstance, "")
	}
	// 2: create pvc referenced by VirtualMachineInstance (Persistent Data)
	// Check if exists
	// If exists, can we attach?
	// If yes, attach
	// If not, update the status with error
	pvc := pkg.CreatePersistentVolumeClaim(labInstance.Namespace, namespace, "csi-ceph-fs")
	if err := pkg.CreateOrUpdate(r.Client, ctx, log, pvc); err != nil && err.Error() != "ALREADY EXISTS" {
		setLabInstanceStatus(r, ctx, log, "Could not create pvc "+pvc.Name+"in namespace "+pvc.Namespace, "Warning", "PvcNotCreated", &labInstance, "")
		return ctrl.Result{}, err
	} else if err != nil && err.Error() == "ALREADY EXISTS" {
		setLabInstanceStatus(r, ctx, log, "PersistentVolumeClaim "+pvc.Name+" already exists in namespace "+pvc.Namespace, "Warning", "PvcAlreadyExists", &labInstance, "")
	} else {
		setLabInstanceStatus(r, ctx, log, "PersistentVolumeClaim "+pvc.Name+" correctly created in namespace "+pvc.Namespace, "Normal", "PvcCreated", &labInstance, "")
	}

	// 3: create Service to expose the vm
	service := pkg.CreateService(name, namespace)
	service.SetOwnerReferences(labiOwnerRef)
	if err := pkg.CreateOrUpdate(r.Client, ctx, log, service); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create service "+service.Name+"in namespace "+service.Namespace, "Warning", "ServiceNotCreated", &labInstance, "")
		return ctrl.Result{}, err
	} else {
		setLabInstanceStatus(r, ctx, log, "Service "+service.Name+" correctly created in namespace "+service.Namespace, "Normal", "ServiceCreated", &labInstance, "")
	}

	// 4: create Ingress to manage the service
	ingress := pkg.CreateIngress(name, namespace, service)
	ingress.SetOwnerReferences(labiOwnerRef)
	if err := pkg.CreateOrUpdate(r.Client, ctx, log, ingress); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create ingress "+ingress.Name+"in namespace "+ingress.Namespace, "Warning", "IngressNotCreated", &labInstance, "")
		return ctrl.Result{}, err
	} else {
		setLabInstanceStatus(r, ctx, log, "Ingress "+ingress.Name+" correctly created in namespace "+ingress.Namespace, "Normal", "IngressCreated", &labInstance, "")
	}

	// 5: create Service for oauth2
	oauthService := pkg.CreateOauth2Service(name, namespace)
	oauthService.SetOwnerReferences(labiOwnerRef)
	if err := pkg.CreateOrUpdate(r.Client, ctx, log, oauthService); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create service "+oauthService.Name+"in namespace "+oauthService.Namespace, "Warning", "Oauth2ServiceNotCreated", &labInstance, "")
		return ctrl.Result{}, err
	} else {
		setLabInstanceStatus(r, ctx, log, "Service "+oauthService.Name+" correctly created in namespace "+oauthService.Namespace, "Normal", "Oauth2ServiceCreated", &labInstance, "")
	}

	// 6: create Ingress to manage the oauth2 service
	oauthIngress := pkg.CreateOauth2Ingress(name, namespace, oauthService)
	oauthIngress.SetOwnerReferences(labiOwnerRef)
	if err := pkg.CreateOrUpdate(r.Client, ctx, log, oauthIngress); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create ingress "+oauthIngress.Name+"in namespace "+oauthIngress.Namespace, "Warning", "Oauth2IngressNotCreated", &labInstance, "")
		return ctrl.Result{}, err
	} else {
		setLabInstanceStatus(r, ctx, log, "Ingress "+oauthIngress.Name+" correctly created in namespace "+oauthIngress.Namespace, "Normal", "Oauth2IngressCreated", &labInstance, "")
	}

	// 6: create Deployment for oauth2
	oauthDeploy := pkg.CreateOauth2Deployment(name, namespace)
	oauthDeploy.SetOwnerReferences(labiOwnerRef)
	if err := pkg.CreateOrUpdate(r.Client, ctx, log, oauthDeploy); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create deployment "+oauthDeploy.Name+"in namespace "+oauthDeploy.Namespace, "Warning", "Oauth2DeployNotCreated", &labInstance, "")
		return ctrl.Result{}, err
	} else {
		setLabInstanceStatus(r, ctx, log, "Deployment "+oauthDeploy.Name+" correctly created in namespace "+oauthDeploy.Namespace, "Normal", "Oauth2DeployCreated", &labInstance, "")
	}

	// 7: create VirtualMachineInstance
	vmi := pkg.CreateVirtualMachineInstance(name, namespace, labTemplate, secret.Name, pvc.Name)
	vmi.SetOwnerReferences(labiOwnerRef)
	if err := pkg.CreateOrUpdate(r.Client, ctx, log, vmi); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create vmi "+vmi.Name+" in namespace "+vmi.Namespace, "Warning", "VmiNotCreated", &labInstance, "")
		return ctrl.Result{}, err
	} else {
		setLabInstanceStatus(r, ctx, log, "VirtualMachineInstance "+vmi.Name+" correctly created in namespace "+vmi.Namespace, "Normal", "VmiCreated", &labInstance, "")
	}

	go getVmiStatus(r, ctx, log, name, ingress, &labInstance, vmi)

	return ctrl.Result{}, nil
}

func (r *LabInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&instancev1.LabInstance{}).
		Complete(r)
}

func setLabInstanceStatus(r *LabInstanceReconciler, ctx context.Context, log logr.Logger,
	msg string, eventType string, eventReason string,
	labInstance *instancev1.LabInstance, url string) {

	log.Info(msg)
	r.EventsRecorder.Event(labInstance, eventType, eventReason, msg)

	labInstance.Status.Phase = eventReason
	labInstance.Status.Url = url
	labInstance.Status.ObservedGeneration = labInstance.ObjectMeta.Generation
	if err := r.Status().Update(ctx, labInstance); err != nil {
		log.Error(err, "unable to update LabInstance status")
	}
	return
}

func getVmiStatus(r *LabInstanceReconciler, ctx context.Context, log logr.Logger,
	name string, ingress v1beta1.Ingress,
	labInstance *instancev1.LabInstance, vmi virtv1.VirtualMachineInstance) {

	var vmStatus virtv1.VirtualMachineInstancePhase
	// iterate until the vm is running
	for {
		err := r.Client.Get(ctx, types.NamespacedName{
			Namespace: vmi.Namespace,
			Name:      vmi.Name,
		}, &vmi)
		if err == nil {
			if vmStatus != vmi.Status.Phase {
				vmStatus = vmi.Status.Phase
				if vmStatus != virtv1.Running {
					setLabInstanceStatus(r, ctx, log, "VirtualMachineInstance "+vmi.Name+" in namespace "+vmi.Namespace+" status update to "+string(vmStatus), "Normal", "Vmi"+string(vmStatus), labInstance, "")
				} else {
					setLabInstanceStatus(r, ctx, log, "VirtualMachineInstance "+vmi.Name+" in namespace "+vmi.Namespace+" status update to "+string(vmStatus), "Normal", "Vmi"+string(vmStatus), labInstance, "")
					break
				}
			}
		}
		time.Sleep(10 * time.Second)
	}

	// when the vm status is Running, it is still not available for some seconds
	// curl the url until the vm is ready
	url := ingress.GetAnnotations()["crownlabs.polito.it/probe-url"]
	for {
		resp, err := http.Get(url)
		if err != nil || resp == nil {
			log.Error(err, "unable to perform get on "+url)
		} else {
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				setLabInstanceStatus(r, ctx, log, "VirtualMachineInstance "+vmi.Name+" in namespace "+vmi.Namespace+" status update to VmiReady", "Normal", "VmiReady", labInstance, url)
				resp.Body.Close()
				break
			}
		}
		time.Sleep(5 * time.Second)
	}

	return
}

