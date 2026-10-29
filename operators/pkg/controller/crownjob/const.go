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

package crownjob

import "time"

// The reasons of the status of a CrownJob (also used as the reasons of its events).
const (
	// ReasonInQueue -> the job is waiting in the queue of the Workspace.
	ReasonInQueue = "InQueue"
	// ReasonQuotaReached -> the job is waiting because the queue of the Workspace is using all its resources.
	ReasonQuotaReached = "QuotaReached"
	// ReasonInstanceNotReady -> the job is waiting for the Deployment of the Instance to exist.
	ReasonInstanceNotReady = "InstanceNotReady"
	// ReasonDiskInUse -> the job is waiting for the Instance to stop, since its disk can be used by one node at a time.
	ReasonDiskInUse = "DiskInUse"
	// ReasonQueueNotReady -> the job is waiting for the ClusterQueue of the Workspace to exist.
	ReasonQueueNotReady = "QueueNotReady"
	// ReasonAdmitted -> the job has been admitted by the queue, and its pod is not running yet.
	ReasonAdmitted = "Admitted"
	// ReasonRunning -> the pod of the job is running.
	ReasonRunning = "Running"
	// ReasonCompleted -> the job has completed successfully.
	ReasonCompleted = "Completed"
	// ReasonTimeLimit -> the job was stopped because it reached its maximum run time.
	ReasonTimeLimit = "TimeLimit"
	// ReasonExitCode -> the command of the job ended with a non-zero exit code.
	ReasonExitCode = "ExitCode"
	// ReasonCannotStart -> the job cannot start (refused before the Job exists, or its pod was rejected).
	ReasonCannotStart = "CannotStart"
	// ReasonError -> the job failed for another reason.
	ReasonError = "Error"
	// ReasonCanceled -> the Instance was deleted before the job was admitted.
	ReasonCanceled = "Canceled"

	// ReasonLogsNotSaved -> the reason of the event recorded when the log of a job pod could not be saved.
	// It never changes the phase of the CrownJob.
	ReasonLogsNotSaved = "LogsNotSaved"
)

// The messages of the status of a CrownJob: plain sentences, for a student.
const (
	msgInQueue          = "Your job is waiting in the queue of workspace %s."
	msgQuotaReached     = "The workspace is using all the resources its job queue allows: your job will start when some are free."
	msgInstanceNotReady = "The instance %s is not ready yet: the job will be created when it is."
	msgInstanceDeleting = "The instance %s is being deleted."
	msgInstanceGone     = "The instance %s does not exist."
	msgDiskInUse        = "Your job will start when you stop the instance %s: its disk can be used by one machine at a time."
	msgQueueNotReady    = "The job queue of workspace %s is not ready yet: your job will start when it is."
	msgCanceled         = "The instance %s was deleted before your job started."
	msgLogSaved         = " The output is saved in %s."
	msgLogsNotSaved     = "The output of the job could not be saved in %s: %s."
	msgStarting         = "Your job is starting."
	msgStartingDownload = "Your job is starting (downloading the image)."
	msgStartingPullFail = "Your job is starting (the image cannot be downloaded: retrying)."
	msgRunning          = "Your job is running."
	msgCompleted        = "Your job finished successfully."
	msgTimeLimit        = "Your job was stopped because it reached its time limit (%s)."
	msgExitCode         = "Your command ended with exit code %d."
	msgRunFolder        = "The folder %s could not be created on the disk of the instance (exit code %d)."
	msgPodRejected      = "The job could not start: %s."
	msgJobFailed        = "The job failed: %s."
	msgJobFailedEmpty   = "The job failed."
	msgJobDeleted       = "The job was deleted before it ended."
	msgJobNameTaken     = "Another job named %s already exists in this namespace."
	msgJobInvalid       = "The job could not be created: %s."
)

const (
	// requeueWaiting -> how often a job that is Waiting or Starting is checked again.
	requeueWaiting = 30 * time.Second
	// requeueInstanceNotReady -> how often a job waiting for the Deployment of its Instance is checked again.
	requeueInstanceNotReady = 15 * time.Second
	// requeueDiskInUse -> how often a job waiting for its Instance to stop is checked again.
	requeueDiskInUse = 30 * time.Second
	// requeueQueueNotReady -> how often a job waiting for the ClusterQueue of its Workspace is checked again.
	requeueQueueNotReady = 30 * time.Second
	// requeueConflict -> how soon a CrownJob or Instance is reconciled again after a conflict on its write.
	requeueConflict = time.Second
	// maxMessageLength -> the maximum length of a message taken from another object (e.g., the reason of a pod).
	maxMessageLength = 200

	// podReasonDeadlineExceeded -> the reason set by the kubelet on a pod that exceeded its activeDeadlineSeconds.
	podReasonDeadlineExceeded = "DeadlineExceeded"
	// containerReasonErrImagePull and containerReasonImagePullBackOff -> the reasons of a container waiting for an image
	// whose pull failed: the kubelet retries it.
	containerReasonErrImagePull     = "ErrImagePull"
	containerReasonImagePullBackOff = "ImagePullBackOff"
	// eventReasonPulling, eventReasonPulled, eventReasonFailed and eventReasonBackOff -> the reasons of the events of
	// the kubelet about the images of a pod (Failed and BackOff also for other failures, e.g., "Back-off restarting").
	eventReasonPulling = "Pulling"
	eventReasonPulled  = "Pulled"
	eventReasonFailed  = "Failed"
	eventReasonBackOff = "BackOff"
	// eventInvolvedObjectUIDField -> the field selector of the events about a given object.
	eventInvolvedObjectUIDField = "involvedObject.uid"
	// kueueInsufficientQuota -> the text of the message of the QuotaReserved condition of a Workload when the
	// ClusterQueue does not have enough unused quota for it (e.g., "insufficient unused quota for cpu in flavor ...").
	kueueInsufficientQuota = "insufficient"
)
