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

package instance_controller

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/api/errors"
	"strings"
	"time"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/instance-creation"

	"github.com/go-logr/logr"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LabInstanceReconciler reconciles a Instance object
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

func (r *LabInstanceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	VMstart := time.Now()
	ctx := context.Background()
	log := r.Log.WithValues("labinstance", req.NamespacedName)

	// get labInstance
	var labInstance crownlabsv1alpha2.Instance
	if err := r.Get(ctx, req.NamespacedName, &labInstance); err != nil {
		// reconcile was triggered by a delete request
		log.Info("Instance " + req.Name + " deleted")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	ns := v1.Namespace{}
	namespaceName := types.NamespacedName{
		Name:      labInstance.Namespace,
		Namespace: "",
	}

	// It performs reconciliation only if the Instance belongs to whitelisted namespaces
	// by checking the existence of keys in labInstance namespace
	if err := r.Get(ctx, namespaceName, &ns); err == nil {
		if !instance_creation.CheckLabels(ns, r.NamespaceWhitelist.MatchLabels) {
			log.Info("Namespace " + req.Namespace + " does not meet the selector labels")
			return ctrl.Result{}, nil
		}
	} else {
		log.Error(err, "unable to get Instance namespace")
	}
	log.Info("Namespace" + req.Namespace + " met the selector labels")
	// The metadata.generation value is incremented for all changes, except for changes to .metadata or .status
	// if metadata.generation is not incremented there's no need to reconcile
	if labInstance.Status.ObservedGeneration == labInstance.ObjectMeta.Generation {
		return ctrl.Result{}, nil
	}

	// check if labTemplate exists
	templateName := types.NamespacedName{
		Namespace: labInstance.Spec.Template.Namespace,
		Name:      labInstance.Spec.Template.Name,
	}
	var labTemplate crownlabsv1alpha2.Template
	if err := r.Get(ctx, templateName, &labTemplate); err != nil {
		// no Template related exists
		log.Info("Template " + templateName.Name + " doesn't exist. Deleting Instance " + labInstance.Name)
		r.EventsRecorder.Event(&labInstance, "Warning", "LabTemplateNotFound", "Template "+templateName.Name+" not found in namespace "+labTemplate.Namespace)
		return ctrl.Result{}, err
	}

	r.EventsRecorder.Event(&labInstance, "Normal", "LabTemplateFound", "Template "+templateName.Name+" found in namespace "+labTemplate.Namespace)
	labInstance.Labels = map[string]string{
		"course-name":        strings.ReplaceAll(strings.ToLower(labTemplate.Spec.WorkspaceRef.Name), " ", "-"),
		"template-name":      labTemplate.Name,
		"template-namespace": labTemplate.Namespace,
	}
	labInstance.Status.ObservedGeneration = labInstance.ObjectMeta.Generation
	if err := r.Update(ctx, &labInstance); err != nil {
		log.Error(err, "unable to update Instance labels")
	}

	if _, err := r.generateEnvironments(labTemplate, labInstance, VMstart); err != nil {
		return ctrl.Result{}, err
	}

	// create secret referenced by VirtualMachineInstance (Cloudinit)
	// To be extracted in a configuration flag
	VmElaborationTimestamp := time.Now()
	VMElaborationDuration := VmElaborationTimestamp.Sub(VMstart)
	elaborationTimes.Observe(VMElaborationDuration.Seconds())

	return ctrl.Result{}, nil
}

func (r *LabInstanceReconciler) generateEnvironments(template crownlabsv1alpha2.Template, instance crownlabsv1alpha2.Instance, vmstart time.Time) (ctrl.Result, error) {
	name := fmt.Sprintf("%v-%.4s", strings.ReplaceAll(instance.Name, ".", "-"), uuid.New().String())
	namespace := instance.Namespace
	for i := range template.Spec.EnvironmentList {
		// prepare variables common to all resources
		switch template.Spec.EnvironmentList[i].EnvironmentType {
		case crownlabsv1alpha2.ClassVM:
			if err := r.CreateVMEnvironment(&instance, &template.Spec.EnvironmentList[i], namespace, name, vmstart); err != nil {
				return ctrl.Result{}, err
			}
		case crownlabsv1alpha2.ClassContainer:
			return ctrl.Result{}, errors.NewBadRequest("Container Environments are not implemented")
		}
	}
	return ctrl.Result{}, nil
}

func (r *LabInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crownlabsv1alpha2.Instance{}).
		Complete(r)
}

func setLabInstanceStatus(r *LabInstanceReconciler, ctx context.Context, log logr.Logger,
	msg string, eventType string, eventReason string,
	labInstance *crownlabsv1alpha2.Instance, ip, url string) {

	log.Info(msg)
	r.EventsRecorder.Event(labInstance, eventType, eventReason, msg)

	labInstance.Status.Phase = eventReason
	labInstance.Status.IP = ip
	labInstance.Status.Url = url
	labInstance.Status.ObservedGeneration = labInstance.ObjectMeta.Generation
	if err := r.Status().Update(ctx, labInstance); err != nil {
		log.Error(err, "unable to update Instance status")
	}
}
