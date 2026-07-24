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

package instsnapshotctrl

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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

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

	// Snapshot finalizer for DataVolume cleanup if cross-namespace or explicitly deleted
	finalizerName := "instancesnapshot.crownlabs.polito.it/finalizer"
	if snapshot.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&snapshot, finalizerName) {
			controllerutil.AddFinalizer(&snapshot, finalizerName)
			if err := r.Update(ctx, &snapshot); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// Deletion logic
		if controllerutil.ContainsFinalizer(&snapshot, finalizerName) {
			if err := r.cleanupDataVolume(ctx, &snapshot); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(&snapshot, finalizerName)
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

	if snapshot.Status.Phase == clv1alpha2.SnapshotPhaseFailed || snapshot.Status.Phase == clv1alpha2.SnapshotPhaseReady {
		return ctrl.Result{}, nil // Already in a terminal state
	}

	// Verify source Instance and its VM state.
	var instance clv1alpha2.Instance
	instanceNN := types.NamespacedName{
		Namespace: snapshot.Spec.Source.InstanceRef.Namespace,
		Name:      snapshot.Spec.Source.InstanceRef.Name,
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
		Name:      fmt.Sprintf("%s-%s", instance.Name, snapshot.Spec.Source.EnvironmentName),
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
		Namespace: snapshot.Spec.Destination.Namespace,
		Name:      fmt.Sprintf("%s-%s", snapshot.Name, string(snapshot.UID)[:5]),
	}

	var dv cdiv1beta1.DataVolume
	err = r.Get(ctx, dvNN, &dv)
	if err != nil && kerrors.IsNotFound(err) {
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
						Name:      snapshot.Spec.Source.PVCName,
					},
				},
				PVC: &corev1.PersistentVolumeClaimSpec{
					StorageClassName: snapshot.Spec.Source.Disk.StorageClassName,
					AccessModes:      snapshot.Spec.Source.Disk.AccessModes,
					VolumeMode:       snapshot.Spec.Source.Disk.VolumeMode,
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: snapshot.Spec.Source.Disk.Size,
						},
					},
				},
			},
		}

		if snapshot.Namespace == dvNN.Namespace {
			// Set owner ref if same namespace
			if err := controllerutil.SetControllerReference(&snapshot, &dv, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}
		}

		if err := r.Create(ctx, &dv); err != nil {
			log.Error(err, "failed to create cloned DataVolume")
			return ctrl.Result{}, err
		}
		snapshot.Status.Artifact.DataVolumeRef = corev1.ObjectReference{
			Name:      dv.Name,
			Namespace: dv.Namespace,
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// Check DataVolume status
	if dv.Status.Phase == cdiv1beta1.Succeeded {
		snapshot.Status.Phase = clv1alpha2.SnapshotPhaseReady
		snapshot.Status.Artifact.DataVolumeRef = corev1.ObjectReference{
			Name:      dv.Name,
			Namespace: dv.Namespace,
		}
		r.EventsRecorder.Eventf(&snapshot, corev1.EventTypeNormal, "SnapshotReady", "Snapshot clone succeeded")
		return ctrl.Result{}, nil
	} else if dv.Status.Phase == cdiv1beta1.Failed {
		snapshot.Status.Phase = clv1alpha2.SnapshotPhaseFailed
		r.EventsRecorder.Eventf(&snapshot, corev1.EventTypeWarning, "SnapshotFailed", "DataVolume clone failed")
		return ctrl.Result{}, nil
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
func (r *InstanceSnapshotReconciler) SetupWithManager(mgr ctrl.Manager, concurrency int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clv1alpha2.InstanceSnapshot{}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "InstanceSnapshot")).
		Complete(r)
}
