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
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	cdiv1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	instance_creation "github.com/netgroup-polito/CrownLabs/operators/pkg/instance-creation"
)

// ContainerEnvOpts contains images name and tag for container environment.
type ContainerEnvOpts struct {
	ImagesTag     string
	VncImg        string
	WebsockifyImg string
	NovncImg      string
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
	Oauth2ProxyImage   string
	OidcClientSecret   string
	OidcProviderURL    string
	ContainerEnvOpts   ContainerEnvOpts

	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

// Reconcile reconciles the state of an Instance resource.
func (r *InstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if r.ReconcileDeferHook != nil {
		defer r.ReconcileDeferHook()
	}

	VMstart := time.Now()

	// get instance
	var instance crownlabsv1alpha2.Instance
	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		// reconcile was triggered by a delete request
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	ns := v1.Namespace{}
	namespaceName := types.NamespacedName{
		Name:      instance.Namespace,
		Namespace: "",
	}

	// It performs reconciliation only if the Instance belongs to whitelisted namespaces
	// by checking the existence of keys in the instance namespace
	if err := r.Get(ctx, namespaceName, &ns); err == nil {
		if !instance_creation.CheckLabels(&ns, r.NamespaceWhitelist.MatchLabels) {
			klog.Info("Namespace " + req.Namespace + " does not meet the selector labels")
			return ctrl.Result{}, nil
		}
	} else {
		klog.Error("Unable to get Instance namespace")
		klog.Error(err)
	}
	klog.Info("Namespace " + req.Namespace + " met the selector labels")

	// check if the Template exists
	templateName := types.NamespacedName{
		Namespace: instance.Spec.Template.Namespace,
		Name:      instance.Spec.Template.Name,
	}
	var template crownlabsv1alpha2.Template
	if err := r.Get(ctx, templateName, &template); err != nil {
		// no Template related exists
		klog.Info("Template " + templateName.Name + " doesn't exist.")
		r.EventsRecorder.Event(&instance, "Warning", "TemplateNotFound", "Template "+templateName.Name+" not found in namespace "+template.Namespace)
		return ctrl.Result{}, err
	}

	r.EventsRecorder.Event(&instance, "Normal", "TemplateFound", "Template "+templateName.Name+" found in namespace "+template.Namespace)

	labeledInstance := *instance.DeepCopy()
	labeledInstance.Labels = map[string]string{
		"crownlabs.polito.it/workspace":  strings.ReplaceAll(template.Spec.WorkspaceRef.Name, ".", "-"),
		"crownlabs.polito.it/template":   template.Name,
		"crownlabs.polito.it/managed-by": "instance",
	}
	if err := r.Patch(ctx, &labeledInstance, client.MergeFrom(&instance)); err != nil {
		klog.Error("Unable to update Instance labels")
		klog.Error(err)
	}

	if _, err := r.generateEnvironments(&template, &instance, VMstart); err != nil {
		klog.Error(err)
		return ctrl.Result{}, err
	}

	// create secret referenced by VirtualMachineInstance (Cloudinit)
	// To be extracted in a configuration flag
	VMElaborationTimestamp := time.Now()
	VMElaborationDuration := VMElaborationTimestamp.Sub(VMstart)
	elaborationTimes.Observe(VMElaborationDuration.Seconds())

	return ctrl.Result{}, nil
}

func (r *InstanceReconciler) generateEnvironments(template *crownlabsv1alpha2.Template, instance *crownlabsv1alpha2.Instance, vmstart time.Time) (ctrl.Result, error) {
	namespace := instance.Namespace
	name := strings.ReplaceAll(instance.Name, ".", "-")
	for i := range template.Spec.EnvironmentList {
		// prepare variables common to all resources
		switch template.Spec.EnvironmentList[i].EnvironmentType {
		case crownlabsv1alpha2.ClassVM:

			if err := r.CreateVMEnvironment(instance, &template.Spec.EnvironmentList[i], namespace, name, vmstart); err != nil {
				return ctrl.Result{}, err
			}
		case crownlabsv1alpha2.ClassContainer:
			if err := r.CreateContainerEnvironment(instance, &template.Spec.EnvironmentList[i], namespace, name, vmstart); err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager registers a new controller for Instance resources.
func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	klog.Info("setup manager")
	return ctrl.NewControllerManagedBy(mgr).
		For(&crownlabsv1alpha2.Instance{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		// Also Deployments are watched in order to better handle container environment.
		Owns(&appsv1.Deployment{}).
		Owns(&cdiv1.DataVolume{}, builder.WithPredicates(dataVolumePredicate())).
		Complete(r)
}

func dataVolumePredicate() predicate.Predicate {
	return predicate.NewPredicateFuncs(func(object client.Object) bool {
		dv, ok := object.(*cdiv1.DataVolume)
		return ok && dv.Status.Phase == cdiv1.DataVolumePhase("Succeeded")
	})
}

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
	statusInstance.Status.Phase = eventReason
	statusInstance.Status.URL = url

	if err := r.Status().Patch(ctx, &statusInstance, client.MergeFrom(instance)); err != nil {
		klog.Error("Unable to update Instance status")
		klog.Error(err)
	}
}
