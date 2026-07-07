// Copyright 2020-2026 Politecnico di Torino
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

// Package instctrl groups the functionalities related to the Instance controller.
package instctrl

import (
	"context"
	"reflect"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/trace"
	virtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/clcontext"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// InstanceReconciler reconciles an Instance object.
type InstanceReconciler struct {
	client.Client
	Scheme                    *runtime.Scheme
	EventsRecorder            record.EventRecorder
	NamespaceWhitelist        metav1.LabelSelector
	ExpositionConfig          forge.ExpositionConfig
	ContainerEnvOpts          forge.ContainerEnvOpts
	WebSSHMasterPublicKey     []byte
	PublicExposureOpts        forge.PublicExposureOpts
	MirrorPVCStorageClassName string

	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

// calculateInstancePhase calculate the overall phase of the Instance based on the phases of its environments.
func (r *InstanceReconciler) calculateInstancePhase(environments []clv1alpha2.InstanceStatusEnv) clv1alpha2.EnvironmentPhase {
	total := len(environments)
	if total == 0 {
		return clv1alpha2.EnvironmentPhaseUnset
	}

	var (
		failed, creationLoopBackoff int
		resourceQuotaExceeded       int
		ready, running              int
		starting, importing         int
		stopping, off               int
	)

	for envIdx := range environments {
		switch environments[envIdx].Phase {
		case clv1alpha2.EnvironmentPhaseFailed:
			failed++

		case clv1alpha2.EnvironmentPhaseCreationLoopBackoff:
			creationLoopBackoff++

		case clv1alpha2.EnvironmentPhaseResourceQuotaExceeded:
			resourceQuotaExceeded++

		case clv1alpha2.EnvironmentPhaseReady:
			ready++

		case clv1alpha2.EnvironmentPhaseRunning:
			running++

		case clv1alpha2.EnvironmentPhaseStarting:
			starting++

		case clv1alpha2.EnvironmentPhaseImporting:
			importing++

		case clv1alpha2.EnvironmentPhaseStopping:
			stopping++

		case clv1alpha2.EnvironmentPhaseOff:
			off++

		default:
			continue
		}
	}

	if failed > 0 || creationLoopBackoff > 0 {
		return clv1alpha2.EnvironmentPhaseFailed
	}
	if resourceQuotaExceeded > 0 {
		return clv1alpha2.EnvironmentPhaseResourceQuotaExceeded
	}
	if ready == total {
		return clv1alpha2.EnvironmentPhaseReady
	}
	if ready > 0 || running > 0 {
		return clv1alpha2.EnvironmentPhaseRunning
	}
	if starting > 0 || importing > 0 {
		return clv1alpha2.EnvironmentPhaseStarting
	}
	if stopping > 0 {
		return clv1alpha2.EnvironmentPhaseStopping
	}
	if off == total {
		return clv1alpha2.EnvironmentPhaseOff
	}

	return clv1alpha2.EnvironmentPhaseUnset
}

// Reconcile reconciles the state of an Instance resource.
func (r *InstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	if r.ReconcileDeferHook != nil {
		defer r.ReconcileDeferHook()
	}

	log := ctrl.LoggerFrom(ctx, "instance", req.NamespacedName)

	tracer := trace.New("reconcile", trace.Field{Key: "instance", Value: req.NamespacedName})
	ctx = trace.ContextWithTrace(ctx, tracer)
	defer tracer.LogIfLong(utils.LongThreshold())

	// Get the instance object.
	var instance clv1alpha2.Instance
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

	// Add the retrieved instance as part of the context.
	ctx, _ = clctx.InstanceInto(ctx, &instance)
	tracer.Step("retrieved the instance")

	// Defer the function to update the instance status depending on the modifications
	// performed while enforcing the desired environments. This is deferred early to
	// allow setting the CreationLoopBackOff phase in case of errors.
	defer func(original, updated *clv1alpha2.Instance) {
		// If the reconciliation failed with an error, set the instance phase to CreationLoopBackOff.
		// Do not set the CreationLoopBackOff phase in case of conflicts, to prevent transients.
		if err != nil && !kerrors.IsConflict(err) {
			for i := range instance.Status.Environments {
				instance.Status.Environments[i].Phase = clv1alpha2.EnvironmentPhaseCreationLoopBackoff
			}
		}

		instance.Status.Phase = r.calculateInstancePhase(instance.Status.Environments)

		// Avoid triggering the status update if not necessary.
		if !reflect.DeepEqual(original.Status, updated.Status) {
			if err2 := r.Status().Patch(ctx, updated, client.MergeFrom(original)); err2 != nil {
				log.Error(err2, "failed to update the instance status")
				err = err2
			} else {
				tracer.Step("instance status updated")
				log.Info("instance status correctly updated")
			}
		}
	}(instance.DeepCopy(), &instance)

	// Retrieve the template associated with the current instance.
	templateName := forge.NamespacedNameFromGenericRef(instance.Spec.Template)
	var template clv1alpha2.Template
	if err := r.Get(ctx, templateName, &template); err != nil {
		log.Error(err, "failed retrieving the instance template", "template", templateName)
		r.EventsRecorder.Eventf(&instance, corev1.EventTypeWarning, EvTmplNotFound, EvTmplNotFoundMsg, templateName.Namespace, templateName.Name)
		return ctrl.Result{}, err
	}
	ctx, log = clctx.TemplateInto(ctx, &template)
	tracer.Step("retrieved the instance template")
	log.Info("successfully retrieved the instance template")

	// Retrieve the tenant associated with the current instance.
	tenantName := forge.NamespacedNameFromGenericRef(instance.Spec.Tenant)
	var tenant clv1alpha2.Tenant
	if err := r.Get(ctx, tenantName, &tenant); err != nil {
		log.Error(err, "failed retrieving the instance tenant", "tenant", tenantName)
		r.EventsRecorder.Eventf(&instance, corev1.EventTypeWarning, EvTntNotFound, EvTntNotFoundMsg, tenantName.Name)
		return ctrl.Result{}, err
	}
	ctx, log = clctx.TenantInto(ctx, &tenant)
	tracer.Step("retrieved the instance tenant")
	log.Info("successfully retrieved the instance tenant")

	// Check whether instance metadata (annotations, labels, pretty name) needs updating.
	hasTimestamp := instance.Annotations != nil && instance.Annotations[forge.LastPoweredOffTimestampAnnotation] != ""
	annotationsNeedUpdate := instance.Spec.Running == hasTimestamp

	labels, labelsNeedUpdate := forge.InstanceLabels(instance.GetLabels(), &template, &instance)

	prettyNameNeedUpdate := instance.Spec.PrettyName == ""

	if annotationsNeedUpdate || labelsNeedUpdate || prettyNameNeedUpdate {
		if err := utils.PatchObject(ctx, r.Client, &instance, enforceInstanceMetadata(labels, annotationsNeedUpdate, labelsNeedUpdate, prettyNameNeedUpdate)); err != nil {
			log.Error(err, "failed to update the instance metadata")
			return ctrl.Result{}, err
		}
		tracer.Step("instance metadata updated")
		log.Info("instance metadata correctly configured")
	}

	// Iterate over and enforce the instance environments.
	if err := r.enforceEnvironments(ctx); err != nil {
		log.Error(err, "failed to enforce instance environments")
		return ctrl.Result{}, err
	}

	// Check the public exposure configuration.
	if err := r.EnforcePublicExposure(ctx); err != nil {
		log.Error(err, "failed to enforce public exposure")
	}

	if err = r.podScheduleStatusIntoInstance(ctx, &instance); err != nil {
		log.Error(err, "unable to retrieve pod schedule status")
	}

	tracer.Step("instance environments enforced")
	log.Info("instance environments correctly enforced")

	return ctrl.Result{}, nil
}

func (r *InstanceReconciler) enforceEnvironments(ctx context.Context) error {
	instance := clctx.InstanceFrom(ctx)
	template := clctx.TemplateFrom(ctx)

	// Make sure the list of instance environments status is initialized
	tmplEnvCount := len(template.Spec.EnvironmentList)
	if len(instance.Status.Environments) != tmplEnvCount {
		instance.Status.Environments = make([]clv1alpha2.InstanceStatusEnv, tmplEnvCount)
	}
	// Set the name of the environment for the instance status
	// to the current template environment name.
	// This has to be done before it can crash for whatever reason,
	// since there is a validation on the name field.
	for i := range template.Spec.EnvironmentList {
		instance.Status.Environments[i].Name = template.Spec.EnvironmentList[i].Name
	}

	urlNeeded := false

	for i := range template.Spec.EnvironmentList {
		tmplEnv := &template.Spec.EnvironmentList[i]

		// Set an inner context for each environment
		innCtx, _ := clctx.EnvironmentInto(ctx, tmplEnv)
		innCtx = clctx.EnvironmentIndexInto(innCtx, i)

		if err := r.EnforceShVolMirrorPVCs(innCtx); err != nil {
			r.EventsRecorder.Eventf(instance, corev1.EventTypeWarning, EvEnvironmentErr, "Failed to enforce mirror SharedVolumes")
			return err
		}

		mountInfos, msg, err := forge.PVCMountInfosFromEnvironment(innCtx, r.Client)
		if err != nil {
			r.EventsRecorder.Eventf(instance, corev1.EventTypeWarning, EvEnvironmentErr, msg, tmplEnv.Name)
			return err
		}
		innCtx = clctx.VolumeMountInfosInto(innCtx, mountInfos)

		switch tmplEnv.EnvironmentType {
		case clv1alpha2.ClassVM, clv1alpha2.ClassCloudVM, clv1alpha2.ClassLocalVM:
			if err := r.EnforceVMEnvironment(innCtx); err != nil {
				r.EventsRecorder.Eventf(instance, corev1.EventTypeWarning, EvEnvironmentErr, EvEnvironmentErrMsg, tmplEnv.Name)
				return err
			}
			if tmplEnv.GuiEnabled {
				urlNeeded = true
			}

		case clv1alpha2.ClassStandalone, clv1alpha2.ClassContainer:
			if err := r.EnforceContainerEnvironment(innCtx); err != nil {
				r.EventsRecorder.Eventf(instance, corev1.EventTypeWarning, EvEnvironmentErr, EvEnvironmentErrMsg, tmplEnv.Name)
				return err
			}
			urlNeeded = true
		}

		r.setInitialReadyTimeIfNecessary(innCtx)
	}
	if urlNeeded {
		// Enforce the ingress to access the GUI
		// Use the configured website base URL
		host := r.ExpositionConfig.WebsiteBaseURL

		// Define url of the instance. This will be the root for the urls of the single environments
		instance.Status.URL = forge.ExpositionGuiStatusInstanceURL(host, instance)
	} else {
		instance.Status.URL = ""
	}

	return nil
}

// setInitialReadyTimeIfNecessary configures the instance InitialReadyTime status value and emits the corresponding
// prometheus metric, in case it was not already present and the instance is currently ready.
func (r *InstanceReconciler) setInitialReadyTimeIfNecessary(ctx context.Context) {
	instance := clctx.InstanceFrom(ctx)

	for i := range instance.Status.Environments {
		if instance.Status.Environments[i].Phase != clv1alpha2.EnvironmentPhaseReady || instance.Status.Environments[i].InitialReadyTime != "" {
			continue
		}
		duration := time.Since(instance.GetCreationTimestamp().Time).Truncate(time.Second)
		instance.Status.Environments[i].InitialReadyTime = duration.String()

		// Filter out possible outliers from the prometheus metrics.
		if duration > 30*time.Minute {
			continue
		}

		template := clctx.TemplateFrom(ctx)
		environment := clctx.EnvironmentFrom(ctx)

		metricInitialReadyTimes.With(prometheus.Labels{
			metricInitialReadyTimesLabelWorkspace:   template.Spec.WorkspaceRef.Name,
			metricInitialReadyTimesLabelTemplate:    template.GetName(),
			metricInitialReadyTimesLabelEnvironment: environment.Name,
			metricInitialReadyTimesLabelType:        string(environment.EnvironmentType),
			metricInitialReadyTimesLabelPersistent:  strconv.FormatBool(environment.Persistent),
		}).Observe(duration.Seconds())
	}
}

// SetupWithManager registers a new controller for Instance resources.
func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager, concurrency int) error {
	mgr.GetLogger().Info("setup manager")

	return ctrl.NewControllerManagedBy(mgr).
		For(&clv1alpha2.Instance{}).
		Owns(&appsv1.Deployment{}).
		Owns(&virtv1.VirtualMachine{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&gatewayv1.HTTPRoute{}).
		// Here, we use Watches instead of Owns since we need to react also in case a VMI generated from a VM is updated,
		// to correctly update the instance phase in case of persistent VMs with resource quota exceeded.
		Watches(&virtv1.VirtualMachineInstance{}, handler.EnqueueRequestsFromMapFunc(r.vmiToInstance)).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: concurrency,
		}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "Instance")).
		Complete(r)
}

// vmiToInstance returns a reconcile request for the instance associated with the given VMI object.
func (r *InstanceReconciler) vmiToInstance(_ context.Context, o client.Object) []reconcile.Request {
	if instance, found := forge.InstanceNameFromLabels(o.GetLabels()); found {
		return []reconcile.Request{{NamespacedName: types.NamespacedName{Namespace: o.GetNamespace(), Name: instance}}}
	}

	return nil
}

func enforceInstanceMetadata(labels map[string]string, annotationsNeedUpdate, labelsNeedUpdate, prettyNameNeedUpdate bool) func(*clv1alpha2.Instance) *clv1alpha2.Instance {
	return func(i *clv1alpha2.Instance) *clv1alpha2.Instance {
		if annotationsNeedUpdate {
			if i.Annotations == nil {
				i.Annotations = make(map[string]string)
			}
			if !i.Spec.Running {
				i.Annotations[forge.LastPoweredOffTimestampAnnotation] = time.Now().Format(time.RFC3339)
			} else {
				delete(i.Annotations, forge.LastPoweredOffTimestampAnnotation)
			}
		}

		if prettyNameNeedUpdate {
			i.Spec.PrettyName = forge.RandomInstancePrettyName()
		}

		if labelsNeedUpdate {
			i.SetLabels(labels)
		}

		return i
	}
}
