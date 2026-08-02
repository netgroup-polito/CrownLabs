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

// Package instsnapctrl contains the controllers for the instance snapshot feature.
package instsnapctrl

import (
	"context"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	virtv1 "kubevirt.io/api/core/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// InstanceSnapshotReconciler reconciles an InstanceSnapshot object.
type InstanceSnapshotReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	EventsRecorder record.EventRecorder
}

// Reconcile reconciles the state of an InstanceSnapshot resource.
func (r *InstanceSnapshotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "instancesnapshot", req.NamespacedName)

	var snapshot clv1alpha2.InstanceSnapshot
	if err := r.Get(ctx, req.NamespacedName, &snapshot); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Snapshot finalizer for DataVolume cleanup on deletion
	finalizerName := "instancesnapshot.crownlabs.polito.it/finalizer"
	if snapshot.DeletionTimestamp.IsZero() {
		if !ctrlutil.ContainsFinalizer(&snapshot, finalizerName) {
			ctrlutil.AddFinalizer(&snapshot, finalizerName)
			if err := r.Update(ctx, &snapshot); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// Deletion logic
		if ctrlutil.ContainsFinalizer(&snapshot, finalizerName) {
			if err := r.cleanupDataVolume(ctx, &snapshot); err != nil {
				return ctrl.Result{}, err
			}
			ctrlutil.RemoveFinalizer(&snapshot, finalizerName)
			if err := r.Update(ctx, &snapshot); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	defer func(original *clv1alpha2.InstanceSnapshot) {
		if !reflect.DeepEqual(original.Status, snapshot.Status) {
			if err := r.Status().Update(ctx, &snapshot); err != nil {
				log.Error(err, "failed to update snapshot status")
			}
		}
	}(snapshot.DeepCopy())

	if snapshot.Status.Phase == "" {
		snapshot.Status.Phase = clv1alpha2.SnapshotPhasePending
	}

	if snapshot.Status.Phase == clv1alpha2.SnapshotPhaseFailed || snapshot.Status.Phase == clv1alpha2.SnapshotPhaseCompleted {
		return ctrl.Result{}, nil // Already in a terminal state
	}

	// Verify source Instance and its VM state.
	var instance clv1alpha2.Instance
	instanceNN := types.NamespacedName{
		Namespace: snapshot.Spec.Instance.Namespace,
		Name:      snapshot.Spec.Instance.Name,
	}
	if err := r.Get(ctx, instanceNN, &instance); err != nil {
		log.Error(err, "failed to get source instance", "instance", instanceNN)
		if kerrors.IsNotFound(err) {
			snapshot.Status.Phase = clv1alpha2.SnapshotPhaseFailed
			r.EventsRecorder.Eventf(&snapshot, corev1.EventTypeWarning, "SourceInstanceDeleted", "Source Instance %s deleted", instanceNN.String())
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if instance.Spec.Running {
		err := fmt.Errorf("instance %s is running", instanceNN.String())
		log.Error(err, "cannot snapshot a running instance")
		return ctrl.Result{Requeue: true}, err
	}

	var vmi virtv1.VirtualMachineInstance
	vmiNN := types.NamespacedName{
		Namespace: instance.Namespace,
		Name:      fmt.Sprintf("%s-%s", instance.Name, snapshot.Spec.Environment),
	}
	err := r.Get(ctx, vmiNN, &vmi)
	if err == nil {
		err := fmt.Errorf("VMI %s is still running", vmiNN.String())
		log.Error(err, "cannot snapshot while VMI exists")
		return ctrl.Result{Requeue: true}, err
	} else if !kerrors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	// We are ready to create the DataVolume clone
	dvNN := types.NamespacedName{
		Namespace: snapshot.Namespace,
		Name:      fmt.Sprintf("%s-%s", snapshot.Name, string(snapshot.UID)[:5]),
	}

	var dv cdiv1beta1.DataVolume
	err = r.Get(ctx, dvNN, &dv)
	if err != nil && kerrors.IsNotFound(err) {
		// Fetch the source PVC to get its specifications
		pvcName := fmt.Sprintf("%s-%s", instance.Name, snapshot.Spec.Environment)
		var sourcePVC corev1.PersistentVolumeClaim
		pvcNN := types.NamespacedName{
			Namespace: instance.Namespace,
			Name:      pvcName,
		}
		if err := r.Get(ctx, pvcNN, &sourcePVC); err != nil {
			log.Error(err, "failed to get source PVC", "pvc", pvcNN)
			return ctrl.Result{}, err
		}

		storageQuantity := sourcePVC.Spec.Resources.Requests[corev1.ResourceStorage]

		// Create the DV
		dv = cdiv1beta1.DataVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name:      dvNN.Name,
				Namespace: dvNN.Namespace,
				Labels: map[string]string{
					"crownlabs.polito.it/snapshot-artifact": "true",
				},
			},
			Spec: cdiv1beta1.DataVolumeSpec{
				Source: &cdiv1beta1.DataVolumeSource{
					PVC: &cdiv1beta1.DataVolumeSourcePVC{
						Namespace: instance.Namespace,
						Name:      pvcName,
					},
				},
				PVC: &corev1.PersistentVolumeClaimSpec{
					StorageClassName: sourcePVC.Spec.StorageClassName,
					AccessModes:      sourcePVC.Spec.AccessModes,
					VolumeMode:       sourcePVC.Spec.VolumeMode,
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: storageQuantity,
						},
					},
				},
			},
		}

		if err := ctrlutil.SetControllerReference(&snapshot, &dv, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, &dv); err != nil {
			log.Error(err, "failed to create cloned DataVolume")
			return ctrl.Result{}, err
		}
		snapshot.Status.Artifact.DataVolumeRef = clv1alpha2.GenericRef{
			Name:      dv.Name,
			Namespace: dv.Namespace,
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// Check DataVolume status
	switch dv.Status.Phase {
	case cdiv1beta1.Succeeded:
		snapshot.Status.Phase = clv1alpha2.SnapshotPhaseCompleted
		snapshot.Status.Artifact.DataVolumeRef = clv1alpha2.GenericRef{
			Name:      dv.Name,
			Namespace: dv.Namespace,
		}
		if dv.Spec.PVC != nil {
			snapshot.Status.Artifact.VolumeSize = dv.Spec.PVC.Resources.Requests[corev1.ResourceStorage]
		}
		r.EventsRecorder.Eventf(&snapshot, corev1.EventTypeNormal, "SnapshotCompleted", "Snapshot clone succeeded")
		return ctrl.Result{}, nil
	case cdiv1beta1.Failed:
		snapshot.Status.Phase = clv1alpha2.SnapshotPhaseFailed
		r.EventsRecorder.Eventf(&snapshot, corev1.EventTypeWarning, "SnapshotFailed", "DataVolume clone failed")
		return ctrl.Result{}, nil
	default:
		// Not in a terminal state yet.
		snapshot.Status.Phase = clv1alpha2.SnapshotPhaseProcessing
	}

	return ctrl.Result{Requeue: true}, nil
}

func (r *InstanceSnapshotReconciler) cleanupDataVolume(ctx context.Context, snapshot *clv1alpha2.InstanceSnapshot) error {
	if snapshot.Status.Artifact.DataVolumeRef.Name != "" {
		dv := &cdiv1beta1.DataVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name:      snapshot.Status.Artifact.DataVolumeRef.Name,
				Namespace: snapshot.Status.Artifact.DataVolumeRef.Namespace,
			},
		}
		if err := r.Delete(ctx, dv); err != nil && !kerrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

// SetupWithManager registers the controller with the manager.
func (r *InstanceSnapshotReconciler) SetupWithManager(mgr ctrl.Manager, _ int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clv1alpha2.InstanceSnapshot{}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "InstanceSnapshot")).
		Complete(r)
}
