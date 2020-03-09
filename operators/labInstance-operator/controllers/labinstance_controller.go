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
	"k8s.io/apimachinery/pkg/types"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	instancev1 "github.com/netgroup-polito/CrownLabs/operators/labInstance-operator/api/v1"
	templatev1 "github.com/netgroup-polito/CrownLabs/operators/labTemplate-operator/api/v1"
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

	// set labInstance status to DEPLOYED
	labInstance = setPhase(labInstance)
	if err := r.Status().Update(ctx, &labInstance); err != nil {
		log.Error(err, "unable to update Advertisement status")
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
