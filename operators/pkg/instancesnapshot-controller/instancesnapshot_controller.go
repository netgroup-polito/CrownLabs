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

// Package instancesnapshot_controller groups the functionalities related to the creation of a persistent VM snapshot.
package instancesnapshot_controller

import (
	"context"
	"fmt"

	batch "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// ContainersSnapshotOpts contains image names and tags of the containers needed for the VM snapshot, along with VM registry access data.
type ContainersSnapshotOpts struct {
	ContainerKaniko    string
	ContainerImgExport string
	VMRegistry         string
	RegistrySecretName string
}

// InstanceSnapshotReconciler reconciles a InstanceSnapshot object.
type InstanceSnapshotReconciler struct {
	client.Client
	EventsRecorder     record.EventRecorder
	Scheme             *runtime.Scheme
	NamespaceWhitelist metav1.LabelSelector
	ContainersSnapshot ContainersSnapshotOpts

	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

// Reconcile reconciles the status of the InstanceSnapshot resource.
func (r *InstanceSnapshotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if r.ReconcileDeferHook != nil {
		defer r.ReconcileDeferHook()
	}

	isnap := &crownlabsv1alpha2.InstanceSnapshot{}

	if err := r.Get(ctx, req.NamespacedName, isnap); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting InstanceSnapshot %s before starting reconcile -> %s", isnap.Name, err)
		return ctrl.Result{}, err
	} else if err != nil {
		klog.Infof("InstanceSnapshot %s already deleted", req.NamespacedName.Name)
		return ctrl.Result{}, nil
	}

	// Check the selector label, in order to know whether to perform or not reconciliation.
	if proceed, err := utils.CheckSelectorLabel(ctx, r.Client, isnap.Namespace, r.NamespaceWhitelist.MatchLabels); !proceed {
		// If there was an error while checking, show the error and try again.
		if err != nil {
			klog.Error(err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	klog.Infof("Start InstanceSnapshot reconciliation of %s in %s namespace", isnap.Name, isnap.Namespace)

	// Check the current status of the InstanceSnapshot by checking
	// the state of its assigned job.
	jobName := types.NamespacedName{
		Namespace: isnap.Namespace,
		Name:      isnap.Name,
	}
	found := &batch.Job{}
	err := r.Get(ctx, jobName, found)

	switch {
	case err != nil && errors.IsNotFound(err):
		if retry, err1 := r.CreateSnapshottingJob(ctx, isnap); err1 != nil {
			klog.Error(err1)
			// Check if we need to try again with the creation of the job
			if retry {
				return ctrl.Result{}, err1
			}
			// Since we don't have to retry, validation failed
			// Add the event and stop reconciliation since the request is not valid.
			r.EventsRecorder.Event(isnap, "Warning", "ValidationError", err1.Error())
			return ctrl.Result{}, nil
		}
		// Job successfully created
		r.EventsRecorder.Event(isnap, "Normal", "Creation", fmt.Sprintf("Job %s for snapshot creation started", isnap.Name))
	case err != nil:
		klog.Errorf("Unable to retrieve the job of InstanceSnapshot %s -> %s", isnap.Name, err)
		return ctrl.Result{}, err
	default:
		// Check the current state of the job and log according to its state
		jstatus, err1 := r.HandleExistingJob(ctx, isnap, found)
		switch {
		case err1 != nil:

			klog.Error(err1)
			return ctrl.Result{}, err1
		case jstatus == batch.JobComplete:

			successMessage := fmt.Sprintf("Image %s created and uploaded", isnap.Spec.ImageName)
			// If we are able to retrieve the execution time, report it
			if found.Status.StartTime != nil && found.Status.CompletionTime != nil {
				extime := found.Status.CompletionTime.Sub(found.Status.StartTime.Time)
				successMessage = fmt.Sprintf("%s in %s", successMessage, extime)
			}
			klog.Info(successMessage)
			r.EventsRecorder.Event(isnap, "Normal", "Created", successMessage)
		case jstatus == batch.JobFailed:

			klog.Infof("Image %s could not be created", isnap.Spec.ImageName)
			r.EventsRecorder.Event(isnap, "Warning", "CreationFailed", "The creation job failed")
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager registers a new controller for InstanceSnapshot resources.
func (r *InstanceSnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// The generation changed predicate allow to avoid updates on the status changes of the InstanceSnapshot
		For(&crownlabsv1alpha2.InstanceSnapshot{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&batch.Job{}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "InstanceSnapshot")).
		Complete(r)
}
