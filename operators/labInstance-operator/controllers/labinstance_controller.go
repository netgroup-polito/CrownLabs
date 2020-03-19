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
	"github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/pkg"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	instancev1 "github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/api/v1"
	templatev1 "github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/labTemplate/api/v1"
)

// LabInstanceReconciler reconciles a LabInstance object
type LabInstanceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=instance.crown.team.com,resources=labinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=instance.crown.team.com,resources=labinstances/status,verbs=get;update;patch

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

	// The metadata.generation value is incremented for all changes, except for changes to .metadata or .status
	// if metadata.generation is not incremented there's no need to reconcile
	if labInstance.Status.ObservedGeneration == labInstance.ObjectMeta.Generation {
		return ctrl.Result{}, nil
	}

	// check if labTemplate exists
	templateName := types.NamespacedName{
		Namespace: labInstance.Namespace,
		Name:      labInstance.Spec.LabTemplateName,
	}
	var labTemplate templatev1.LabTemplate
	if err := r.Get(ctx, templateName, &labTemplate); err != nil {
		// no LabTemplate related exists
		log.Info("LabTemplate " + templateName.Name + " doesn't exist")
		return ctrl.Result{}, err
	}

	// prepare variables common to all resources
	name := labTemplate.Name + "-" + labInstance.Spec.StudentID
	namespace := labInstance.Namespace
	// this is added so that all resources created for this LabInstance are destroyed when the LabInstance is deleted
	labiOwnerRef := []metav1.OwnerReference{
		{
			APIVersion: labInstance.APIVersion,
			Kind:       labInstance.Kind,
			Name:       labInstance.Name,
			UID:        labInstance.UID,
		},
	}

	// create secret referenced by VirtualMachineInstance
	secret := pkg.CreateSecret(name, namespace)
	secret.SetOwnerReferences(labiOwnerRef)
	if err := pkg.CreateOrUpdate(r.Client, ctx, log, secret); err != nil {
		log.Info("Could not create secret " + secret.Name)
		return ctrl.Result{}, err
	} else {
		log.Info("Secret " + secret.Name + " correctly created")
	}
	// create VirtualMachineInstance
	vmi := pkg.CreateVirtualMachineInstance(name, namespace, labTemplate, secret.Name, name+"-pvc")
	vmi.SetOwnerReferences(labiOwnerRef)
	if err := pkg.CreateOrUpdate(r.Client, ctx, log, vmi); err != nil {
		log.Info("Could not create vm " + vmi.Name)
		return ctrl.Result{}, err
	} else {
		log.Info("VirtualMachineInstance " + vmi.Name + " correctly created")
	}
	// create Service to expose the vm
	service := pkg.CreateService(name, namespace)
	service.SetOwnerReferences(labiOwnerRef)
	if err := pkg.CreateOrUpdate(r.Client, ctx, log, service); err != nil {
		log.Info("Could not create service " + service.Name)
		return ctrl.Result{}, err
	} else {
		log.Info("Service " + service.Name + " correctly created")
	}
	// create Ingress to manage the service
	ingress := pkg.CreateIngress(name, namespace, secret.Name, service)
	ingress.SetOwnerReferences(labiOwnerRef)
	if err := pkg.CreateOrUpdate(r.Client, ctx, log, ingress); err != nil {
		log.Info("Could not create ingress " + ingress.Name)
		return ctrl.Result{}, err
	} else {
		log.Info("Ingress " + ingress.Name + " correctly created")
	}

	labInstance.Status.Url = ingress.Spec.Rules[0].Host + "/" + name
	if err := r.Status().Update(ctx, &labInstance); err != nil {
		log.Error(err, "unable to update LabInstance status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *LabInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&instancev1.LabInstance{}).
		Complete(r)
}

func setPhase(labInstance instancev1.LabInstance) instancev1.LabInstance {
	labInstance.Status.Phase = "DEPLOYED"
	labInstance.Status.ObservedGeneration = labInstance.ObjectMeta.Generation
	return labInstance
}
