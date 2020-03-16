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
	virtv1 "github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/kubeVirt/api/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

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

	vm := labTemplate.Spec.Vm
	vm.Name = labTemplate.Name + "-" + labInstance.Spec.StudentID
	// this is added so that all resources created for this LabInstance are destroyed when the LabInstance is deleted
	ownerRef := []metav1.OwnerReference{
		{
			APIVersion: labInstance.APIVersion,
			Kind:       labInstance.Kind,
			Name:       labInstance.Name,
			UID:        labInstance.UID,
		},
	}
	vm.SetOwnerReferences(ownerRef)

	if err := CreateOrUpdate(r.Client, ctx, log, vm); err != nil {
		log.Info("Could not create vm " + vm.Name)
		return ctrl.Result{}, err
	}

	ingress := createIngress()
	ingress.Name = labTemplate.Name + "-" + labInstance.Spec.StudentID
	ingress.Namespace = "test-vm-ns"
	ingress.SetOwnerReferences(ownerRef)

	if err := CreateOrUpdate(r.Client, ctx, log, ingress); err != nil {
		log.Info("Could not create ingress " + ingress.Name)
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


func createIngress() v1beta1.Ingress {

	ingress := v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      nil,
			Annotations: map[string]string{"cert-manager.io/cluster-issuer": "letsencrypt-production"},
		},
		Spec: v1beta1.IngressSpec{
			Backend: nil,
			TLS: []v1beta1.IngressTLS{
				{
					Hosts:      []string{"lab2-test-vm.crown-labs.ipv6.polito.it"},
					SecretName: "lab2-test-vm-cert",
				},
			},
			Rules: []v1beta1.IngressRule{
				{
					Host: "lab2-test-vm.crown-labs.ipv6.polito.it",
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: v1beta1.IngressBackend{
										ServiceName: "vm-access-svc",
										ServicePort: intstr.IntOrString{IntVal: 6080},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return ingress
}

// create a resource or update it if already exists
func CreateOrUpdate(c client.Client, ctx context.Context, log logr.Logger, object interface{}) error {

	switch obj := object.(type) {
	case virtv1.VirtualMachineInstance:
		var vmi virtv1.VirtualMachineInstance
		err := c.Get(ctx, types.NamespacedName{
			Namespace: obj.Namespace,
			Name:      obj.Name,
		}, &vmi)
		if err != nil {
			err = c.Create(ctx, &obj, &client.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				log.Error(err, "unable to create virtual machine "+obj.Name)
				return err
			}
		}
		// CHECK: no need for update??
		//} else {
		//	obj.SetResourceVersion(vmi.ResourceVersion)
		//	err = c.Update(ctx, &obj, &client.UpdateOptions{})
		//	if err != nil {
		//		log.Error(err, "unable to update virtual machine "+obj.Name)
		//		return err
		//	}
		//}
	case v1beta1.Ingress:
		var ing v1beta1.Ingress
		err := c.Get(ctx, types.NamespacedName{
			Namespace: obj.Namespace,
			Name:      obj.Name,
		}, &ing)
		if err != nil {
			err = c.Create(ctx, &obj, &client.CreateOptions{})
			if err != nil && !errors.IsAlreadyExists(err) {
				log.Error(err, "unable to create ingress "+obj.Name)
				return err
			}
		} else {
			obj.SetResourceVersion(ing.ResourceVersion)
			err = c.Update(ctx, &obj, &client.UpdateOptions{})
			if err != nil {
				log.Error(err, "unable to update ingress "+obj.Name)
				return err
			}
		}
	}


	return nil
}


