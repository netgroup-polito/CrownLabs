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

// Package instance_controller groups the functionalities related to the Instance controller.
package instance_controller

import (
	"context"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	virtv1 "kubevirt.io/client-go/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// ContainerEnvOpts contains images name and tag for container environment.
type ContainerEnvOpts struct {
	ImagesTag         string
	VncImg            string
	WebsockifyImg     string
	NovncImg          string
	FileBrowserImg    string
	FileBrowserImgTag string
}

// InstanceReconciler reconciles a Instance object.
type InstanceReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	EventsRecorder     record.EventRecorder
	NamespaceWhitelist metav1.LabelSelector
	WebsiteBaseURL     string
	NextcloudBaseURL   string
	WebdavSecretName   string
	InstancesAuthURL   string
	Concurrency        int
	ContainerEnvOpts   ContainerEnvOpts

	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

// Reconcile reconciles the state of an Instance resource.
func (r *InstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	if r.ReconcileDeferHook != nil {
		defer r.ReconcileDeferHook()
	}

	log := ctrl.LoggerFrom(ctx, "instance", req.NamespacedName)

	// Get the instance object.
	var instance crownlabsv1alpha2.Instance
	if err = r.Get(ctx, req.NamespacedName, &instance); err != nil {
		if !kerrors.IsNotFound(err) {
			log.Error(err, "failed retrieving instance")
		}
		// Reconcile was triggered by a delete request.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check the selector label, in order to know whether to perform or not reconciliation.
	if proceed, err := utils.CheckSelectorLabel(ctrl.LoggerInto(ctx, log), r.Client, instance.GetNamespace(), r.NamespaceWhitelist.MatchLabels); !proceed {
		// If there was an error while checking, show the error and try again.
		if err != nil {
			log.Error(err, "failed checking selector labels")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Retrieve the template associated with the current instance.
	templateName := types.NamespacedName{
		Namespace: instance.Spec.Template.Namespace,
		Name:      instance.Spec.Template.Name,
	}

	var template crownlabsv1alpha2.Template
	if err := r.Get(ctx, templateName, &template); err != nil {
		log.Error(err, "failed retrieving the instance template", "template", templateName)
		r.EventsRecorder.Eventf(&instance, v1.EventTypeWarning, EvTmplNotFound, EvTmplNotFoundMsg, template.Namespace, template.Name)
		return ctrl.Result{}, err
	}
	log = log.WithValues("template", templateName)
	log.Info("successfully retrieved the instance template")

	// Patch the instance labels to allow for easier categorization.
	labels, updated := forge.InstanceLabels(instance.GetLabels(), &template)
	if updated {
		original := instance.DeepCopy()
		instance.SetLabels(labels)
		if err := r.Patch(ctx, &instance, client.MergeFrom(original)); err != nil {
			log.Error(err, "failed to update the instance labels")
			return ctrl.Result{}, err
		}
		log.Info("instance labels correctly configured")
	}

	// Defer the function to patch the instance status depending on the modifications
	// performed while enforcing the desired environments.
	defer func(original, updated *crownlabsv1alpha2.Instance) {
		if !reflect.DeepEqual(original.Status, updated.Status) {
			if err2 := r.Status().Patch(ctx, updated, client.MergeFrom(original)); err2 != nil {
				log.Error(err2, "failed to update the instance status")
				err = err2
			} else {
				log.Info("instance status correctly updated")
			}
		}
	}(instance.DeepCopy(), &instance)

	// Iterate over and enforce the instance environments.
	if err := r.enforceEnvironments(ctrl.LoggerInto(ctx, log), &instance, &template); err != nil {
		log.Error(err, "failed to enforce instance environments")
		instance.Status.Phase = crownlabsv1alpha2.EnvironmentPhaseCreationLoopBackoff
		return ctrl.Result{}, err
	}
	log.Info("instance environments correctly enforced")

	return ctrl.Result{}, nil
}

func (r *InstanceReconciler) enforceEnvironments(ctx context.Context, instance *crownlabsv1alpha2.Instance, template *crownlabsv1alpha2.Template) error {
	for i := range template.Spec.EnvironmentList {
		// Currently, only instances composed of a single environment are supported.
		// Nonetheless, we return nil in the end, since it is useless to retry later.
		if i >= 1 {
			err := fmt.Errorf("instances composed of multiple environments are currently not supported")
			ctrl.LoggerFrom(ctx).Error(err, "failed to process environment")
			return nil
		}

		environment := &template.Spec.EnvironmentList[i]
		switch template.Spec.EnvironmentList[i].EnvironmentType {
		case crownlabsv1alpha2.ClassVM:
			if err := r.EnforceVMEnvironment(ctx, instance, environment); err != nil {
				r.EventsRecorder.Eventf(instance, v1.EventTypeWarning, EvEnvironmentErr, EvEnvironmentErrMsg, environment.Name)
				return err
			}
		case crownlabsv1alpha2.ClassContainer:
			if err := r.EnforceContainerEnvironment(ctx, instance, environment); err != nil {
				r.EventsRecorder.Eventf(instance, v1.EventTypeWarning, EvEnvironmentErr, EvEnvironmentErrMsg, environment.Name)
				return err
			}
		}
	}
	return nil
}

// SetupWithManager registers a new controller for Instance resources.
func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	mgr.GetLogger().Info("setup manager")
	return ctrl.NewControllerManagedBy(mgr).
		For(&crownlabsv1alpha2.Instance{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&appsv1.Deployment{}).
		Owns(&virtv1.VirtualMachine{}).
		Owns(&virtv1.VirtualMachineInstance{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.Concurrency,
		}).
		Complete(r)
}

// The setInstanceStatus function is deprecated and should no longer be used.
func (r *InstanceReconciler) setInstanceStatus(
	ctx context.Context,
	msg string, eventType string, eventReason string,
	instance *crownlabsv1alpha2.Instance, ip, url string) {
	klog.Info(msg)

	// Ignore other phases if currently ready and no error occurred
	// Do not return if the event is VmiReady, to avoid problems if the other parameters changed
	if instance.Status.Phase == "VmiReady" && eventReason != "VmiReady" && eventType == "Normal" && eventReason != "VmiOff" {
		return
	}
	r.EventsRecorder.Event(instance, eventType, eventReason, msg)

	statusInstance := *instance.DeepCopy()
	statusInstance.Status.IP = ip
	statusInstance.Status.Phase = crownlabsv1alpha2.EnvironmentPhase(eventReason)
	statusInstance.Status.URL = url

	if err := r.Status().Patch(ctx, &statusInstance, client.MergeFrom(instance)); err != nil {
		klog.Error("Unable to update Instance status")
		klog.Error(err)
	}
}
