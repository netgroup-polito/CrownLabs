package instancesnapshot_controller

import (
	"context"
	"fmt"

	batch "k8s.io/api/batch/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// CreateSnapshottingJob creates the job in charge of creating the snapshot.
func (r *InstanceSnapshotReconciler) CreateSnapshottingJob(ctx context.Context, isnap *crownlabsv1alpha2.InstanceSnapshot) (bool, error) {
	r.EventsRecorder.Event(isnap, "Normal", "Validating", "Start validation of the request")

	isnap.Status.Phase = crownlabsv1alpha2.Pending
	if err := r.Status().Update(ctx, isnap); err != nil {
		return true, fmt.Errorf("error when updating status of InstanceSnapshot %s -> %w", isnap.Name, err)
	}

	if retry, err := r.ValidateRequest(ctx, isnap); err != nil {
		// Print the validation error in the log and check if there is the need to
		// set the operation as failed, or to try again.
		if retry {
			return true, err
		}

		// Set the status as failed
		isnap.Status.Phase = crownlabsv1alpha2.Failed
		if uerr := r.Status().Update(ctx, isnap); uerr != nil {
			return true, fmt.Errorf("error when updating status of InstanceSnapshot %s -> %w", isnap.Name, uerr)
		}

		return false, err
	}

	// Get the job to be created
	snapjob, err1 := r.CreateSnapshottingJobDefinition(ctx, isnap)
	if err1 != nil {
		return true, err1
	}

	// Set the owner reference in order to delete the job when the InstanceSnapshot is deleted.
	if err := ctrl.SetControllerReference(isnap, &snapjob, r.Scheme); err != nil {
		return true, err
	}

	if err := r.Create(ctx, &snapjob); err != nil {
		// It was not possible to create the job
		return true, fmt.Errorf("error when creating the job for %s -> %w", isnap.Name, err)
	}

	isnap.Status.Phase = crownlabsv1alpha2.Processing
	if err := r.Status().Update(ctx, isnap); err != nil {
		return true, fmt.Errorf("error when updating status of InstanceSnapshot %s -> %w", isnap.Name, err)
	}

	return false, nil
}

// HandleExistingJob checks the status of the existing job and updates the status of the InstanceSnapshot accordingly.
func (r *InstanceSnapshotReconciler) HandleExistingJob(ctx context.Context, isnap *crownlabsv1alpha2.InstanceSnapshot, snapjob *batch.Job) (batch.JobConditionType, error) {
	completed, jstatus := r.GetJobStatus(snapjob)
	if completed {
		if jstatus == batch.JobComplete {
			// The job is completed and the image has been uploaded to the registry
			isnap.Status.Phase = crownlabsv1alpha2.Completed
			if err := r.Status().Update(ctx, isnap); err != nil {
				return "", fmt.Errorf("error when updating status of InstanceSnapshot %s -> %w", isnap.Name, err)
			}
		} else {
			// The creation of the snapshot failed since the job failed
			isnap.Status.Phase = crownlabsv1alpha2.Failed
			if err := r.Status().Update(ctx, isnap); err != nil {
				return "", fmt.Errorf("error when updating status of InstanceSnapshot %s -> %w", isnap.Name, err)
			}
		}
	}
	return jstatus, nil
}
