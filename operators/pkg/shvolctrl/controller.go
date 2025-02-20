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

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	"k8s.io/utils/trace"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	tnctrl "github.com/netgroup-polito/CrownLabs/operators/pkg/tenant-controller"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// SharedVolumeReconciler reconciles a SharedVolume object.
type SharedVolumeReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	EventsRecorder     record.EventRecorder
	NamespaceWhitelist metav1.LabelSelector
	PVCStorageClass    string

	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

//TODO: Ma perché non ci sei nel team di CL sul sito? https://preprod.crownlabs.polito.it/about/

// SetupWithManager registers a new controller for SharedVolume resources.
func (r *SharedVolumeReconciler) SetupWithManager(mgr ctrl.Manager, concurrency int) error {
	mgr.GetLogger().Info("setup manager")
	return ctrl.NewControllerManagedBy(mgr).
		For(&clv1alpha2.SharedVolume{}). //TODO: Corretto? È ancora valida la wiki https://preprod.crownlabs.polito.it/learning/operator/ ?
		Owns(&v1.PersistentVolumeClaim{}).
		Owns(&batchv1.Job{}).
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
		// Reconcile was triggered by a delete request.
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

	// Create or Update the PVC, reconciling it with the SharedVolume spec.
	phaseToSet := clv1alpha2.SharedVolumePhaseUnset
	pvc := v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: shvolume.GetName(), Namespace: shvolume.GetNamespace()}}

	pvcOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &pvc, func() error {
		oldSize := *pvc.Spec.Resources.Requests.Storage()

		if pvc.CreationTimestamp.IsZero() {
			pvc.Spec = forge.SharedVolumePVCSpec(&r.PVCStorageClass)
			//TODO: Non servono label sul PVC, vero? (Tranne quelle generate dal Job, o quelle vanno sullo shvol?)
		}

		// Set Error phase if ShVol size is forbidden (less than previous)
		if shvolume.Spec.Size.Cmp(oldSize) >= 0 {
			pvc.Spec.Resources.Requests = v1.ResourceList{v1.ResourceStorage: shvolume.Spec.Size}
			log.Info("Size updated",
				"previous", oldSize, "current", shvolume.Spec.Size)
		} else {
			pvc.Spec.Resources.Requests = v1.ResourceList{v1.ResourceStorage: oldSize}
			phaseToSet = clv1alpha2.SharedVolumePhaseError
			shvolume.Status.ErrorReason = clv1alpha2.SharedVolumeErrorReasonSmaller
			log.Error(fmt.Errorf("forbidden: size smaller than previous"), "Phase transitioned to Error")
			return ctrl.SetControllerReference(&shvolume, &pvc, r.Scheme) //TODO: Così va bene per uscire? (Serve solo se la roba di sotto va spostata dentro qui)
		}

		return ctrl.SetControllerReference(&shvolume, &pvc, r.Scheme)
	})
	if err != nil {
		log.Error(err, "Unable to create or update PVC")
		return ctrl.Result{}, err
	}
	log.Info("PVC enforced", "result", pvcOpRes)

	// Update SharedVolume Status and start provisioning if PVC is Bound
	//TODO: Switch-Case?
	//TODO: Controlliamo qui o dentro alla MutatingFn?
	if pvc.Status.Phase == v1.ClaimBound {
		pv := v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: pvc.Spec.VolumeName}}
		if err := r.Get(ctx, types.NamespacedName{Name: pv.Name}, &pv); err != nil {
			log.Error(err, "Unable to get PV")
			return ctrl.Result{}, err
		}

		shvolume.Status.ServerAddress = fmt.Sprintf("%s.%s", pv.Spec.CSI.VolumeAttributes["server"], pv.Spec.CSI.VolumeAttributes["clusterID"])
		shvolume.Status.ExportPath = pv.Spec.CSI.VolumeAttributes["share"]

		phaseToSet = clv1alpha2.SharedVolumePhaseProvisioning

		done, err := r.doDriveProvisioning(ctx, log.WithName("provisioning-job"), &pvc, &shvolume)
		if err != nil {
			return ctrl.Result{}, err
		}
		if done {
			phaseToSet = clv1alpha2.SharedVolumePhaseReady
		}
	} else if pvc.Status.Phase == v1.ClaimPending {
		phaseToSet = clv1alpha2.SharedVolumePhasePending
	} else {
		phaseToSet = clv1alpha2.SharedVolumePhaseCreating //TODO: A questo punto serve davvero questa differenza (pending/creating)?
	}
	//TODO: Chi ce lo dovrebbe mettere in ResourceQuotaExceeded?

	/*
		switch pvc.Status.Phase {
		case v1.ClaimBound:
			pv := v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: pvc.Spec.VolumeName}}
			if err := r.Get(ctx, types.NamespacedName{Name: pv.Name}, &pv); err != nil {
				log.Error(err, "Unable to get PV")
				return ctrl.Result{}, err
			}

			shvolume.Status.ServerAddress = fmt.Sprintf("%s.%s", pv.Spec.CSI.VolumeAttributes["server"], pv.Spec.CSI.VolumeAttributes["clusterID"])
			shvolume.Status.ExportPath = pv.Spec.CSI.VolumeAttributes["share"]

			phaseToSet = clv1alpha2.SharedVolumePhaseProvisioning

			done, err := r.doDriveProvisioning(ctx, log.WithName("provisioning-job"), &pvc, &shvolume)
			if err != nil {
				return ctrl.Result{}, err
			}
			if done {
				phaseToSet = clv1alpha2.SharedVolumePhaseReady
			}
		case v1.ClaimPending:
			phaseToSet = clv1alpha2.SharedVolumePhasePending
		default:
			phaseToSet = clv1alpha2.SharedVolumePhaseCreating
		}
	*/

	// Update Phase according to what happened
	if phaseToSet != shvolume.Status.Phase {
		log.Info("Phase changed",
			"previous", shvolume.Status.Phase, "current", phaseToSet)

		shvolume.Status.Phase = phaseToSet
		if err := r.Update(ctx, &shvolume); err != nil {
			log.Error(err, "Failed to update phase")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *SharedVolumeReconciler) doDriveProvisioning(ctx context.Context, log logr.Logger, pvc *v1.PersistentVolumeClaim, shvol *clv1alpha2.SharedVolume) (bool, error) {
	val, found := pvc.Labels[forge.ProvisionJobLabel]
	if !found || val != forge.ProvisionJobValueOk {
		chownJob := batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: pvc.Name + "-provision", Namespace: pvc.Namespace}}
		labelToSet := forge.ProvisionJobValuePending

		chownJobOpRes, err := ctrl.CreateOrUpdate(ctx, r.Client, &chownJob, func() error {
			if chownJob.CreationTimestamp.IsZero() {
				log.Info("Created")
				r.updateProvisioningJob(&chownJob, pvc)
			} else if found && val == forge.ProvisionJobValuePending {
				if chownJob.Status.Succeeded == 1 {
					labelToSet = forge.ProvisionJobValueOk
					log.Info("Completed")
				} else if chownJob.Status.Failed == 1 {
					log.Info("Failed")
				}
			}

			return ctrl.SetControllerReference(shvol, &chownJob, r.Scheme)
		})
		if err != nil {
			log.Error(err, "Unable to create or update Job")
			return false, err
		}
		log.Info("Job enforced", "result", chownJobOpRes)

		if labelToSet != pvc.Labels[forge.ProvisionJobLabel] {
			log.Info("PVC labels changed",
				"previous", pvc.Labels[forge.ProvisionJobLabel], "current", labelToSet)

			pvc.Labels[forge.ProvisionJobLabel] = labelToSet
			if err := r.Update(ctx, pvc); err != nil {
				log.Error(err, "Failed to update PVC labels")
				return false, err
			}
		}
	}

	if pvc.Labels[forge.ProvisionJobLabel] == forge.ProvisionJobValueOk { //TODO: Va bene come escamotage?
		return true, nil
	}

	return false, nil
}

// TODO: Non sarebbe il caso di portarla fuori da qualche parte invece che ripeterla?
func (r *SharedVolumeReconciler) updateProvisioningJob(job *batchv1.Job, pvc *v1.PersistentVolumeClaim) {
	job.Spec.BackoffLimit = ptr.To[int32](tnctrl.ProvisionJobMaxRetries) //TODO: Brutto vero?
	job.Spec.TTLSecondsAfterFinished = ptr.To[int32](tnctrl.ProvisionJobTTLSeconds)
	job.Spec.Template.Spec.RestartPolicy = v1.RestartPolicyOnFailure
	job.Spec.Template.Spec.Containers = []v1.Container{{
		Name:    "chown-container",
		Image:   tnctrl.ProvisionJobBaseImage,
		Command: []string{"chown", "-R", fmt.Sprintf("%d:%d", forge.CrownLabsUserID, forge.CrownLabsUserID), forge.MyDriveVolumeMountPath},
		VolumeMounts: []v1.VolumeMount{{
			Name:      "drive",
			MountPath: forge.MyDriveVolumeMountPath,
		},
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				"cpu":    resource.MustParse("100m"),
				"memory": resource.MustParse("128Mi"),
			},
			Limits: v1.ResourceList{
				"cpu":    resource.MustParse("100m"),
				"memory": resource.MustParse("128Mi"),
			},
		},
	},
	}
	job.Spec.Template.Spec.Volumes = []v1.Volume{{
		Name: "drive",
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvc.Name,
			},
		},
	},
	}
}
