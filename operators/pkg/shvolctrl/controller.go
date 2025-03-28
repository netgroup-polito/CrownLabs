// Copyright 2020-2025 Politecnico di Torino
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

// Package shvolctrl groups the functionalities related to the SharedVolume controller.
package shvolctrl

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/trace"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	ctrlUtil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// SharedVolumeReconciler reconciles a SharedVolume object.
type SharedVolumeReconciler struct {
	client.Client
	EventsRecorder     record.EventRecorder
	NamespaceWhitelist metav1.LabelSelector
	PVCStorageClass    string

	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

// SetupWithManager registers a new controller for SharedVolume resources.
func (r *SharedVolumeReconciler) SetupWithManager(mgr ctrl.Manager, concurrency int) error {
	mgr.GetLogger().Info("setup manager")
	return ctrl.NewControllerManagedBy(mgr).
		For(&clv1alpha2.SharedVolume{}).
		Owns(&v1.PersistentVolumeClaim{}).
		Owns(&batchv1.Job{}).
		//FIXME: Should also Watch for Templates in case it's in Deleting phase
		WithOptions(controller.Options{
			MaxConcurrentReconciles: concurrency,
		}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "SharedVolume")).
		Complete(r)
}

// Reconcile reconciles the state of a SharedVolume resource.
func (r *SharedVolumeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	if r.ReconcileDeferHook != nil {
		defer r.ReconcileDeferHook()
	}

	log := ctrl.LoggerFrom(ctx, "sharedvolume", req.NamespacedName)

	tracer := trace.New("reconcile", trace.Field{Key: "sharedvolume", Value: req.NamespacedName})
	ctx = trace.ContextWithTrace(ctx, tracer)
	defer tracer.LogIfLong(utils.LongThreshold())

	// Get the shared volume object.
	var shvolume clv1alpha2.SharedVolume
	if err = r.Get(ctx, req.NamespacedName, &shvolume); err != nil {
		if !kerrors.IsNotFound(err) {
			log.Error(err, "Failed retrieving shared volume")
		}
		// Reconcile was triggered by a delete request and object is already deleted.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check the selector label, in order to know whether to perform or not reconciliation.
	if proceed, err := utils.CheckSelectorLabel(ctrl.LoggerInto(ctx, log), r.Client, shvolume.GetNamespace(), r.NamespaceWhitelist.MatchLabels); !proceed {
		// If there was an error while checking, show the error and try again.
		if err != nil {
			log.Error(err, "Failed checking selector labels")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Defer the function to update the SharedVolume status depending on the modifications
	// performed while enforcing the desired environments. This is deferred early to
	// allow setting the Error phase in case of errors.
	defer func(original, updated *clv1alpha2.SharedVolume) {
		// Avoid triggering the status update if not necessary.
		if !reflect.DeepEqual(original.Status, updated.Status) {
			if err2 := r.Status().Patch(ctx, updated, client.MergeFrom(original)); err2 != nil {
				log.Error(err2, "failed to update the sharedvolume status")
				err = err2
			} else {
				tracer.Step("sharedvolume status updated")
				log.Info("sharedvolume status correctly updated")
			}
		}
	}(shvolume.DeepCopy(), &shvolume)

	if !shvolume.GetDeletionTimestamp().IsZero() {
		log.Info("Processing delete request")
		shvolume.Status.Phase = clv1alpha2.SharedVolumePhaseDeleting

		if ctrlUtil.ContainsFinalizer(&shvolume, clv1alpha2.ShVolCtrlFinalizerName) {
			if err := r.handleDeletion(ctx, log, &shvolume); err != nil {
				log.Error(err, "failed handling deletion request")
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Patch the shared volume labels.
	labels, updated := forge.SharedVolumeLabels(shvolume.GetLabels())
	if updated {
		original := shvolume.DeepCopy()
		shvolume.SetLabels(labels)
		if err := r.Patch(ctx, &shvolume, client.MergeFrom(original)); err != nil {
			log.Error(err, "failed to update the sharedvolume labels")
			return ctrl.Result{}, err
		}
		tracer.Step("sharedvolume labels updated")
		log.Info("sharedvolume labels correctly configured")
	}

	// Change the Phase of the SharedVolume, so that if something happens it goes into Error.
	shvolume.Status.Phase = clv1alpha2.SharedVolumePhaseError

	// Create or Update the PVC, reconciling it with the SharedVolume spec.
	pvc := v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "shvol-" + shvolume.GetName(), Namespace: shvolume.GetNamespace()}}

	pvcOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &pvc, func() error {
		oldSize := *pvc.Spec.Resources.Requests.Storage()

		if pvc.CreationTimestamp.IsZero() {
			pvc.Spec = forge.SharedVolumePVCSpec(&r.PVCStorageClass)
		}

		pvc.SetLabels(forge.SharedVolumeObjectLabels(pvc.GetLabels()))

		// Set Error phase if ShVol size is forbidden (less than previous)
		if sizeDiff := shvolume.Spec.Size.Cmp(oldSize); sizeDiff > 0 || oldSize.IsZero() {
			pvc.Spec.Resources.Requests = v1.ResourceList{v1.ResourceStorage: shvolume.Spec.Size}

			log.V(utils.LogDebugLevel).Info("Size updated",
				"previous", oldSize, "current", shvolume.Spec.Size)
		} else if sizeDiff < 0 {
			shvolume.Status.Phase = clv1alpha2.SharedVolumePhaseError
			log.Error(fmt.Errorf("forbidden: size smaller than previous"), "Phase transitioned to Error")
			r.EventsRecorder.Eventf(&shvolume, v1.EventTypeWarning, EvPVCSmaller, EvPVCSmallerMsg)

			// Must return now (do not add code below or add return here).
		}

		return ctrl.SetControllerReference(&shvolume, &pvc, r.Scheme())
	})
	if err != nil {
		if isResourceQuotaExceeded(err) {
			shvolume.Status.Phase = clv1alpha2.SharedVolumePhaseResourceQuotaExceeded
			log.Error(fmt.Errorf("forbidden: resource quota exceeded"), "Phase transitioned to ResourceQuotaExceeded")
			r.EventsRecorder.Eventf(&shvolume, v1.EventTypeWarning, EvPVCResQuotaExceeded, EvPVCResQuotaExceededMsg)
			err = nil
		} else {
			log.Error(err, "Unable to create or update PVC")
		}

		return ctrl.Result{}, err
	}
	log.Info("PVC enforced", "result", pvcOpRes)

	// Check if PVC has been enforced correctly
	if shvolume.Spec.Size.Cmp(*pvc.Spec.Resources.Requests.Storage()) != 0 {
		return ctrl.Result{}, nil
	}

	// Update SharedVolume Status and start provisioning if PVC is Bound
	if pvc.Status.Phase == v1.ClaimBound {
		pv := v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: pvc.Spec.VolumeName}}
		if err := r.Get(ctx, types.NamespacedName{Name: pv.Name}, &pv); err != nil {
			log.Error(err, "Unable to get PV")
			return ctrl.Result{}, err
		}

		nfsServer, expPath := forge.NFSShVolSpec(&pv)
		if nfsServer == "" || expPath == "" {
			shvolume.Status.Phase = clv1alpha2.SharedVolumePhaseError
			log.Error(fmt.Errorf("pv does not have CSI params"), "Phase transitioned to Error")
			r.EventsRecorder.Eventf(&shvolume, v1.EventTypeWarning, EvPVNoCSI, EvPVNoCSIMsg)
			return ctrl.Result{}, nil
		}

		shvolume.Status.ServerAddress = nfsServer
		shvolume.Status.ExportPath = expPath

		shvolume.Status.Phase = clv1alpha2.SharedVolumePhaseProvisioning

		done, err := utils.NFSDriveProvisioning(ctx, log, r.Client, &pvc, &shvolume)
		if err != nil {
			return ctrl.Result{}, err
		} else if done {
			original := shvolume.DeepCopy()
			ctrlUtil.AddFinalizer(&shvolume, clv1alpha2.ShVolCtrlFinalizerName)
			if err := r.Patch(ctx, &shvolume, client.MergeFrom(original)); err != nil {
				log.Error(err, "failed adding finalizer")
				return ctrl.Result{}, err
			}

			shvolume.Status.Phase = clv1alpha2.SharedVolumePhaseReady
		}
	} else {
		shvolume.Status.Phase = clv1alpha2.SharedVolumePhasePending
	}

	return ctrl.Result{}, nil
}

func isResourceQuotaExceeded(err error) bool {
	return kerrors.IsForbidden(err) && strings.Contains(err.Error(), "exceeded quota")
}

func (r *SharedVolumeReconciler) handleDeletion(ctx context.Context, log logr.Logger, shvol *clv1alpha2.SharedVolume) error {
	var templates clv1alpha2.TemplateList
	if err := r.List(ctx, &templates, &client.ListOptions{}); err != nil {
		log.Error(err, "unable to retrieve template list")
		return err
	}

	found := false
	mountedList := []string{}

	for tmplKey := range templates.Items {
		tmpl := &templates.Items[tmplKey]

		for envKey := range tmpl.Spec.EnvironmentList {
			env := &tmpl.Spec.EnvironmentList[envKey]

			for _, shmount := range env.SharedVolumeMounts {
				if shmount.SharedVolumeRef.Name == shvol.GetName() && shmount.SharedVolumeRef.Namespace == shvol.GetNamespace() {
					mountedList = append(mountedList, tmpl.Name)
					found = true
				}
			}
		}
	}

	if found {
		r.EventsRecorder.Eventf(shvol, v1.EventTypeWarning, EvDeletionBlocked, EvDeletionBlockedMsg, mountedList)
		log.Info("blocked deletion, shvol is mounted on some templates")
	} else {
		ctrlUtil.RemoveFinalizer(shvol, clv1alpha2.ShVolCtrlFinalizerName)
		if err := r.Update(ctx, shvol); err != nil {
			log.Error(err, "failed removing finalizer")
			return err
		}
		log.Info("deletion ok, removed finalizer")
	}

	return nil
}
