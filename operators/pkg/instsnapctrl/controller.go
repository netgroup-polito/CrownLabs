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
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// InstanceSnapshotReconciler reconciles an InstanceSnapshot object.
type InstanceSnapshotReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	EventsRecorder record.EventRecorder
}

// Reconcile reconciles the state of an InstanceSnapshot resource.
//
//nolint:gocyclo // This method coordinates the snapshot state machine and its dependent resources.
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
			// Patch metadata only: a full Update can change omitted or empty spec fields
			// during JSON serialization and violate the spec's immutability validation.
			original := snapshot.DeepCopy()
			ctrlutil.AddFinalizer(&snapshot, finalizerName)
			if err := r.Patch(ctx, &snapshot, client.MergeFromWithOptions(original, client.MergeFromWithOptimisticLock{})); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// Deletion logic
		if ctrlutil.ContainsFinalizer(&snapshot, finalizerName) {
			if err := r.cleanupDataVolume(ctx, &snapshot); err != nil {
				return ctrl.Result{}, err
			}
			original := snapshot.DeepCopy()
			ctrlutil.RemoveFinalizer(&snapshot, finalizerName)
			if err := r.Patch(ctx, &snapshot, client.MergeFromWithOptions(original, client.MergeFromWithOptimisticLock{})); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	defer func(original *clv1alpha2.InstanceSnapshot) {
		if !reflect.DeepEqual(original.Status, snapshot.Status) {
			if err := r.Status().Patch(ctx, &snapshot, client.MergeFrom(original)); err != nil {
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

	// Fetch the Template associated with the source Instance to verify it is single-env.
	var template clv1alpha2.Template
	templateNN := types.NamespacedName{
		Namespace: instance.Spec.Template.Namespace,
		Name:      instance.Spec.Template.Name,
	}
	if err := r.Get(ctx, templateNN, &template); err != nil {
		log.Error(err, "failed to get source instance template", "template", templateNN)
		if kerrors.IsNotFound(err) {
			snapshot.Status.Phase = clv1alpha2.SnapshotPhaseFailed
			r.EventsRecorder.Eventf(&snapshot, corev1.EventTypeWarning, "TemplateNotFound",
				"Template %s for source Instance not found", templateNN.String())
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Deny snapshot on templates that do not have exactly one environment.
	if len(template.Spec.EnvironmentList) != 1 {
		snapshot.Status.Phase = clv1alpha2.SnapshotPhaseFailed
		r.EventsRecorder.Eventf(&snapshot, corev1.EventTypeWarning, "MultiEnvTemplateNotAllowed",
			"Snapshots are only supported on single-environment templates (template %s has %d environments)",
			templateNN.String(), len(template.Spec.EnvironmentList))
		log.Info("snapshot denied: template does not have exactly one environment",
			"template", templateNN, "envCount", len(template.Spec.EnvironmentList))
		return ctrl.Result{}, nil
	}

	if instance.Spec.Running {
		err := fmt.Errorf("instance %s is running", instanceNN.String())
		log.Error(err, "cannot snapshot a running instance")
		return ctrl.Result{Requeue: true}, err
	}

	// Applying the same naming convention used by the instance controller to find the PVC of the instance.
	environmentNN := forge.NamespacedNameWithSuffix(&instance, snapshot.Spec.Environment)

	var vmi virtv1.VirtualMachineInstance
	err := r.Get(ctx, environmentNN, &vmi)
	if err == nil {
		err := fmt.Errorf("VMI %s is still running", environmentNN.String())
		log.Error(err, "cannot snapshot while VMI exists")
		return ctrl.Result{Requeue: true}, err
	} else if !kerrors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	// The DataVolume and its PVC share the snapshot's name and namespace.
	dvNN := types.NamespacedName{
		Namespace: snapshot.Namespace,
		Name:      snapshot.Name,
	}

	var dv cdiv1beta1.DataVolume
	err = r.Get(ctx, dvNN, &dv)
	if err != nil && kerrors.IsNotFound(err) {
		// Fetch the source PVC to get its specifications
		var sourcePVC corev1.PersistentVolumeClaim
		if err := r.Get(ctx, environmentNN, &sourcePVC); err != nil {
			log.Error(err, "failed to get source PVC", "pvc", environmentNN)
			return ctrl.Result{}, err
		}

		storageQuantity := sourcePVC.Spec.Resources.Requests[corev1.ResourceStorage]

		// Create the DV
		dv = cdiv1beta1.DataVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name:      dvNN.Name,
				Namespace: dvNN.Namespace,
				Labels: map[string]string{
					forge.LabelSnapshotArtifactKey: forge.LabelSnapshotArtifactValue,
				},
				Annotations: map[string]string{},
			},
			Spec: cdiv1beta1.DataVolumeSpec{
				Source: &cdiv1beta1.DataVolumeSource{
					PVC: &cdiv1beta1.DataVolumeSourcePVC{
						Namespace: environmentNN.Namespace,
						Name:      environmentNN.Name,
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

		r.populateMetadata(&snapshot, &instance, &dv)

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
		snapshot.Status.Artifact.VolumeSize = storageQuantity
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// The DataVolume already exists, but this snapshot may not be the one that created it: adopting
	// it would publish, and on deletion destroy, a volume belonging to somebody else.
	if !metav1.IsControlledBy(&dv, &snapshot) {
		snapshot.Status.Phase = clv1alpha2.SnapshotPhaseFailed
		r.EventsRecorder.Eventf(&snapshot, corev1.EventTypeWarning, "ArtifactNotOwned",
			"DataVolume %s already exists and is not controlled by this snapshot", dvNN.String())
		log.Info("snapshot denied: the target DataVolume is not owned by this snapshot", "datavolume", dvNN)
		return ctrl.Result{}, nil
	}

	// Check DataVolume status
	switch dv.Status.Phase {
	case cdiv1beta1.Succeeded:
		if err := r.labelArtifactPVC(ctx, &dv); err != nil {
			log.Error(err, "failed to mark snapshot artifact PVC", "pvc", dvNN)
			return ctrl.Result{}, err
		}
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

	return ctrl.Result{}, nil
}
func (r *InstanceSnapshotReconciler) populateMetadata(snapshot *clv1alpha2.InstanceSnapshot, instance *clv1alpha2.Instance, dv *cdiv1beta1.DataVolume) {
	if snapshot.Spec.ImageName != "" {
		dv.Annotations["crownlabs.polito.it/image-name"] = snapshot.Spec.ImageName
	}
	if snapshot.Spec.Description != "" {
		dv.Annotations["crownlabs.polito.it/snapshot-description"] = snapshot.Spec.Description
	}

	tenant := snapshot.Spec.Tenant
	if tenant.Name == "" {
		tenant = instance.Spec.Tenant
	}
	if tenant.Name != "" {
		dv.Annotations["crownlabs.polito.it/snapshot-tenant"] = tenant.Name
	}
}

func (r *InstanceSnapshotReconciler) cleanupDataVolume(ctx context.Context, snapshot *clv1alpha2.InstanceSnapshot) error {
	ref := snapshot.Status.Artifact.DataVolumeRef
	if ref.Name == "" {
		return nil
	}

	// Only a DataVolume controlled by this very snapshot may be deleted: otherwise a tampered reference would
	// make the operator, which holds cluster-wide permissions, delete the disk of somebody else's VM.
	var dv cdiv1beta1.DataVolume
	if err := r.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, &dv); err != nil {
		return client.IgnoreNotFound(err)
	}

	if !metav1.IsControlledBy(&dv, snapshot) {
		r.EventsRecorder.Eventf(snapshot, corev1.EventTypeWarning, "ArtifactNotOwned",
			"DataVolume %s/%s is not controlled by this snapshot: leaving it untouched", ref.Namespace, ref.Name)
		return nil
	}

	return client.IgnoreNotFound(r.Delete(ctx, &dv))
}

// SetupWithManager registers the controller with the manager.
func (r *InstanceSnapshotReconciler) SetupWithManager(mgr ctrl.Manager, _ int) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clv1alpha2.InstanceSnapshot{}).
		Owns(&cdiv1beta1.DataVolume{}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "InstanceSnapshot")).
		Complete(r)
}

func (r *InstanceSnapshotReconciler) labelArtifactPVC(ctx context.Context, dv *cdiv1beta1.DataVolume) error {
	var pvc corev1.PersistentVolumeClaim
	if err := r.Get(ctx, types.NamespacedName{Namespace: dv.Namespace, Name: dv.Name}, &pvc); err != nil {
		return err
	}

	if pvc.Labels[forge.LabelSnapshotArtifactKey] == forge.LabelSnapshotArtifactValue {
		return nil
	}

	if pvc.Labels == nil {
		pvc.Labels = map[string]string{}
	}
	pvc.Labels[forge.LabelSnapshotArtifactKey] = forge.LabelSnapshotArtifactValue

	return r.Update(ctx, &pvc)
}
