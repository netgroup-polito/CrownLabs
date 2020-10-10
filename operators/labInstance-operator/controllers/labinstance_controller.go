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
	"net"
	"strings"
	"time"

	"github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/pkg/instanceCreation"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	instancev1alpha1 "github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/api/v1alpha1"
	virtv1 "github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/kubeVirt/api/v1"
	templatev1alpha1 "github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/labTemplate/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LabInstanceReconciler reconciles a LabInstance object
type LabInstanceReconciler struct {
	client.Client
	Log                logr.Logger
	Scheme             *runtime.Scheme
	EventsRecorder     record.EventRecorder
	NamespaceWhitelist metav1.LabelSelector
	WebsiteBaseUrl     string
	NextcloudBaseUrl   string
	WebdavSecretName   string
	Oauth2ProxyImage   string
	OidcClientSecret   string
	OidcProviderUrl    string
}

// +kubebuilder:rbac:groups=crownlabs.polito.it,resources=labinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=crownlabs.polito.it,resources=labinstances/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events/status,verbs=get

func (r *LabInstanceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	VMstart := time.Now()
	ctx := context.Background()
	log := r.Log.WithValues("labinstance", req.NamespacedName)

	// get labInstance
	var labInstance instancev1alpha1.LabInstance
	if err := r.Get(ctx, req.NamespacedName, &labInstance); err != nil {
		// reconcile was triggered by a delete request
		log.Info("LabInstance " + req.Name + " deleted")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	ns := v1.Namespace{}
	namespaceName := types.NamespacedName{
		Name:      labInstance.Namespace,
		Namespace: "",
	}

	// It performs reconciliation only if the LabInstance belongs to whitelisted namespaces
	// by checking the existence of keys in labInstance namespace
	if err := r.Get(ctx, namespaceName, &ns); err == nil {
		if !instanceCreation.CheckLabels(ns, r.NamespaceWhitelist.MatchLabels) {
			log.Info("Namespace " + req.Namespace + " does not meet " +
				"the selector labels")
			return ctrl.Result{}, nil
		}
	} else {
		log.Error(err, "unable to get LabInstance namespace")
	}
	log.Info("Namespace" + req.Namespace + " met the selector labels")
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
	var labTemplate templatev1alpha1.LabTemplate
	if err := r.Get(ctx, templateName, &labTemplate); err != nil {
		// no LabTemplate related exists
		log.Info("LabTemplate " + templateName.Name + " doesn't exist. Deleting LabInstance " + labInstance.Name)
		r.EventsRecorder.Event(&labInstance, "Warning", "LabTemplateNotFound", "LabTemplate "+templateName.Name+" not found in namespace "+labTemplate.Namespace)
		_ = r.Delete(ctx, &labInstance, &client.DeleteOptions{})
		return ctrl.Result{}, err
	}

	vmType := labTemplate.Spec.VmType
	if vmType != templatev1alpha1.TypeCLI {
		vmType = templatev1alpha1.TypeGUI
	}

	r.EventsRecorder.Event(&labInstance, "Normal", "LabTemplateFound", "LabTemplate "+templateName.Name+" found in namespace "+labTemplate.Namespace)
	labInstance.Labels = map[string]string{
		"course-name":        strings.ReplaceAll(strings.ToLower(labTemplate.Spec.CourseName), " ", "-"),
		"template-name":      labTemplate.Name,
		"template-namespace": labTemplate.Namespace,
	}
	labInstance.Status.ObservedGeneration = labInstance.ObjectMeta.Generation
	if err := r.Update(ctx, &labInstance); err != nil {
		log.Error(err, "unable to update LabInstance labels")
	}

	// prepare variables common to all resources
	name := fmt.Sprintf("%v-%.4s", strings.ReplaceAll(labInstance.Name, ".", "-"), uuid.New().String())
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

	// create secret referenced by VirtualMachineInstance (Cloudinit)
	// To be extracted in a configuration flag

	var user, password string
	err := instanceCreation.GetWebdavCredentials(r.Client, ctx, log, r.WebdavSecretName, labInstance.Namespace, &user, &password)
	if err != nil {
		log.Error(err, "unable to get Webdav Credentials")
	} else {
		log.Info("Webdav secrets obtained. Building cloud-init script." + labInstance.Name)
	}
	secret := instanceCreation.CreateSecret(name, namespace, user, password, r.NextcloudBaseUrl)
	secret.SetOwnerReferences(labiOwnerRef)
	if err := instanceCreation.CreateOrUpdate(r.Client, ctx, log, secret); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create secret "+secret.Name+" in namespace "+secret.Namespace, "Warning", "SecretNotCreated", &labInstance, "", "")
	} else {
		setLabInstanceStatus(r, ctx, log, "Secret "+secret.Name+" correctly created in namespace "+secret.Namespace, "Normal", "SecretCreated", &labInstance, "", "")
	}

	// create Service to expose the vm
	service := instanceCreation.CreateService(name, namespace)
	service.SetOwnerReferences(labiOwnerRef)
	if err := instanceCreation.CreateOrUpdate(r.Client, ctx, log, service); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create service "+service.Name+" in namespace "+service.Namespace, "Warning", "ServiceNotCreated", &labInstance, "", "")
		return ctrl.Result{}, err
	} else {
		setLabInstanceStatus(r, ctx, log, "Service "+service.Name+" correctly created in namespace "+service.Namespace, "Normal", "ServiceCreated", &labInstance, "", "")
	}

	urlUUID := uuid.New().String()
	// create Ingress to manage the service
	ingress := instanceCreation.CreateIngress(name, namespace, service, urlUUID, r.WebsiteBaseUrl)
	ingress.SetOwnerReferences(labiOwnerRef)
	if err := instanceCreation.CreateOrUpdate(r.Client, ctx, log, ingress); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create ingress "+ingress.Name+" in namespace "+ingress.Namespace, "Warning", "IngressNotCreated", &labInstance, "", "")
		return ctrl.Result{}, err
	} else {
		setLabInstanceStatus(r, ctx, log, "Ingress "+ingress.Name+" correctly created in namespace "+ingress.Namespace, "Normal", "IngressCreated", &labInstance, "", "")
	}

	// create Service for oauth2
	oauthService := instanceCreation.CreateOauth2Service(name, namespace)
	oauthService.SetOwnerReferences(labiOwnerRef)
	if err := instanceCreation.CreateOrUpdate(r.Client, ctx, log, oauthService); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create service "+oauthService.Name+" in namespace "+oauthService.Namespace, "Warning", "Oauth2ServiceNotCreated", &labInstance, "", "")
		return ctrl.Result{}, err
	} else {
		setLabInstanceStatus(r, ctx, log, "Service "+oauthService.Name+" correctly created in namespace "+oauthService.Namespace, "Normal", "Oauth2ServiceCreated", &labInstance, "", "")
	}

	// create Ingress to manage the oauth2 service
	oauthIngress := instanceCreation.CreateOauth2Ingress(name, namespace, oauthService, urlUUID, r.WebsiteBaseUrl)
	oauthIngress.SetOwnerReferences(labiOwnerRef)
	if err := instanceCreation.CreateOrUpdate(r.Client, ctx, log, oauthIngress); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create ingress "+oauthIngress.Name+" in namespace "+oauthIngress.Namespace, "Warning", "Oauth2IngressNotCreated", &labInstance, "", "")
		return ctrl.Result{}, err
	} else {
		setLabInstanceStatus(r, ctx, log, "Ingress "+oauthIngress.Name+" correctly created in namespace "+oauthIngress.Namespace, "Normal", "Oauth2IngressCreated", &labInstance, "", "")
	}

	// create Deployment for oauth2
	oauthDeploy := instanceCreation.CreateOauth2Deployment(name, namespace, urlUUID, r.Oauth2ProxyImage, r.OidcClientSecret, r.OidcProviderUrl)
	oauthDeploy.SetOwnerReferences(labiOwnerRef)
	if err := instanceCreation.CreateOrUpdate(r.Client, ctx, log, oauthDeploy); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create deployment "+oauthDeploy.Name+" in namespace "+oauthDeploy.Namespace, "Warning", "Oauth2DeployNotCreated", &labInstance, "", "")
		return ctrl.Result{}, err
	} else {
		setLabInstanceStatus(r, ctx, log, "Deployment "+oauthDeploy.Name+" correctly created in namespace "+oauthDeploy.Namespace, "Normal", "Oauth2DeployCreated", &labInstance, "", "")
	}

	// create VirtualMachineInstance
	vmi := instanceCreation.CreateVirtualMachineInstance(name, namespace, labTemplate, labInstance.Name, secret.Name)
	vmi.SetOwnerReferences(labiOwnerRef)
	if err := instanceCreation.CreateOrUpdate(r.Client, ctx, log, vmi); err != nil {
		setLabInstanceStatus(r, ctx, log, "Could not create vmi "+vmi.Name+" in namespace "+vmi.Namespace, "Warning", "VmiNotCreated", &labInstance, "", "")
		return ctrl.Result{}, err
	} else {
		setLabInstanceStatus(r, ctx, log, "VirtualMachineInstance "+vmi.Name+" correctly created in namespace "+vmi.Namespace, "Normal", "VmiCreated", &labInstance, "", "")
	}
	VmElaborationTimestamp := time.Now()
	VMElaborationDuration := VmElaborationTimestamp.Sub(VMstart)
	elaborationTimes.Observe(VMElaborationDuration.Seconds())
	go getVmiStatus(r, ctx, log, name, vmType, service, ingress, &labInstance, vmi, VMstart)

	return ctrl.Result{}, nil
}

func (r *LabInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&instancev1alpha1.LabInstance{}).
		Complete(r)
}

func setLabInstanceStatus(r *LabInstanceReconciler, ctx context.Context, log logr.Logger,
	msg string, eventType string, eventReason string,
	labInstance *instancev1alpha1.LabInstance, ip, url string) {

	log.Info(msg)
	r.EventsRecorder.Event(labInstance, eventType, eventReason, msg)

	labInstance.Status.Phase = eventReason
	labInstance.Status.IP = ip
	labInstance.Status.Url = url
	labInstance.Status.ObservedGeneration = labInstance.ObjectMeta.Generation
	if err := r.Status().Update(ctx, labInstance); err != nil {
		log.Error(err, "unable to update LabInstance status")
	}
}

func getVmiStatus(r *LabInstanceReconciler, ctx context.Context, log logr.Logger,
	name string, vmType templatev1alpha1.VmType, service v1.Service, ingress v1beta1.Ingress,
	labInstance *instancev1alpha1.LabInstance, vmi virtv1.VirtualMachineInstance, startTimeVM time.Time) {

	var vmStatus virtv1.VirtualMachineInstancePhase

	var ip string
	url := ingress.GetAnnotations()["crownlabs.polito.it/probe-url"]

	// iterate until the vm is running
	for {
		err := r.Client.Get(ctx, types.NamespacedName{
			Namespace: vmi.Namespace,
			Name:      vmi.Name,
		}, &vmi)
		if err == nil {
			if vmStatus != vmi.Status.Phase {
				vmStatus = vmi.Status.Phase
				if len(vmi.Status.Interfaces) > 0 {
					ip = vmi.Status.Interfaces[0].IP
				}

				msg := "VirtualMachineInstance " + vmi.Name + " in namespace " + vmi.Namespace + " status update to " + string(vmStatus)
				if vmStatus == virtv1.Failed {
					setLabInstanceStatus(r, ctx, log, msg, "Warning", "Vmi"+string(vmStatus), labInstance, "", "")
					return
				}

				setLabInstanceStatus(r, ctx, log, msg, "Normal", "Vmi"+string(vmStatus), labInstance, ip, url)
				if vmStatus == virtv1.Running {
					break
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	// when the vm status is Running, it is still not available for some seconds
	// hence, wait until it starts responding
	host := service.Name + "." + service.Namespace
	port := "6080" // VNC
	if vmType == templatev1alpha1.TypeCLI {
		port = "22" // SSH
	}

	err := waitForConnection(log, host, port)
	if err != nil {
		log.Error(err, fmt.Sprintf("Unable to check whether %v:%v is reachable", host, port))
	} else {
		msg := "VirtualMachineInstance " + vmi.Name + " in namespace " + vmi.Namespace + " status update to VmiReady."
		setLabInstanceStatus(r, ctx, log, msg, "Normal", "VmiReady", labInstance, ip, url)
		readyTime := time.Now()
		bootTime := readyTime.Sub(startTimeVM)
		bootTimes.Observe(bootTime.Seconds())
	}
}

func waitForConnection(log logr.Logger, host, port string) error {
	for retries := 0; retries < 120; retries++ {
		timeout := time.Second
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
		if err != nil {
			log.Info(fmt.Sprintf("Unable to check whether %v:%v is reachable: %v", host, port, err))
			time.Sleep(time.Second)
		} else {
			// The connection succeeded, hence the VM is ready
			defer conn.Close()
			return nil
		}
	}

	return fmt.Errorf("Timeout while checking whether %v:%v is reachable", host, port)
}
