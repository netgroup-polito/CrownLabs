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

import (
	"context"
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// observeJob sets the status of the CrownJob from the state of its Job, reading the pod of the Job and its Kueue
// Workload only when needed. Once the pod of the Job is read, its log helper is created if needed.
func (r *Reconciler) observeJob(ctx context.Context, crownJob *clv1alpha2.CrownJob, job *batchv1.Job) (ctrl.Result, error) {
	workspace := strings.TrimPrefix(job.Labels[forge.KueueQueueNameLabel], forge.CrownJobQueuePrefix)

	if jobConditionTrue(job, batchv1.JobComplete) != nil {
		helpers, _, err := r.observePod(ctx, crownJob, job)
		if err != nil {
			return ctrl.Result{}, err
		}
		setStatus(crownJob, clv1alpha2.CrownJobPhaseSucceeded, ReasonCompleted, msgCompleted+logSavedSuffix(crownJob, job, helpers))
		return ctrl.Result{}, nil
	}

	if failed := jobConditionTrue(job, batchv1.JobFailed); failed != nil {
		helpers, pod, err := r.observePod(ctx, crownJob, job)
		if err != nil {
			return ctrl.Result{}, err
		}
		reason, message := failureStatus(job, failed, pod)
		setStatus(crownJob, clv1alpha2.CrownJobPhaseFailed, reason, message+logSavedSuffix(crownJob, job, helpers))
		return ctrl.Result{}, nil
	}

	if ptr.Deref(job.Spec.Suspend, false) {
		reason, message := ReasonInQueue, fmt.Sprintf(msgInQueue, workspace)
		if r.quotaReached(ctx, job) {
			reason, message = ReasonQuotaReached, msgQuotaReached
		}
		setStatus(crownJob, clv1alpha2.CrownJobPhaseWaiting, reason, message)
		return ctrl.Result{RequeueAfter: requeueWaiting}, nil
	}

	if ptr.Deref(job.Status.Ready, 0) > 0 {
		wasRunning := crownJob.Status.Phase == clv1alpha2.CrownJobPhaseRunning
		setStatus(crownJob, clv1alpha2.CrownJobPhaseRunning, ReasonRunning, msgRunning)
		// The pod is read at every reconcile while running (they happen only on changes of the Job or the CrownJob),
		// so that the log helper of a pod is created even if a previous attempt failed.
		_, pod, err := r.observePod(ctx, crownJob, job)
		if err != nil {
			return ctrl.Result{}, err
		}
		if wasRunning && crownJob.Status.Deadline != nil {
			return ctrl.Result{}, nil
		}
		// The deadline is that of the running pod: its start time plus the maximum run time.
		if pod == nil || pod.Status.StartTime == nil || job.Spec.Template.Spec.ActiveDeadlineSeconds == nil {
			return ctrl.Result{RequeueAfter: requeueWaiting}, nil
		}
		deadline := pod.Status.StartTime.Add(time.Duration(*job.Spec.Template.Spec.ActiveDeadlineSeconds) * time.Second)
		crownJob.Status.Deadline = ptr.To(metav1.NewTime(deadline))
		return ctrl.Result{}, nil
	}

	// Admitted, and no pod is ready.
	_, pod, err := r.observePod(ctx, crownJob, job)
	if err != nil {
		return ctrl.Result{}, err
	}
	if crownJob.Status.Phase == clv1alpha2.CrownJobPhaseRunning && isRunningPod(crownJob, job, pod) {
		// The pod that was running is being stopped (e.g., at its time limit), and the Job reports the end soon:
		// the job is not starting again. A new pod that replaces it (e.g., after an eviction) is starting instead.
		return ctrl.Result{RequeueAfter: requeueWaiting}, nil
	}
	message := msgStarting
	switch r.imagePull(ctx, pod) {
	case imagePullInProgress:
		message = msgStartingDownload
	case imagePullFailing:
		message = msgStartingPullFail
	case imagePullNone:
	}
	setStatus(crownJob, clv1alpha2.CrownJobPhaseStarting, ReasonAdmitted, message)
	crownJob.Status.Deadline = nil
	return ctrl.Result{RequeueAfter: requeueWaiting}, nil
}

// observePod reads the newest pod of the Job, creates its log helper if needed, reports the failed log helpers, and
// returns the log helpers of the CrownJob and the pod.
func (r *Reconciler) observePod(ctx context.Context, crownJob *clv1alpha2.CrownJob, job *batchv1.Job) ([]batchv1.Job, *corev1.Pod, error) {
	pod, err := r.jobPod(ctx, job)
	if err != nil {
		return nil, nil, err
	}
	helpers, err := r.ensureLogHelper(ctx, crownJob, job, pod)
	if err != nil {
		return nil, nil, err
	}
	return helpers, pod, r.reportFailedLogHelpers(ctx, crownJob, helpers)
}

// isRunningPod returns whether the given pod is the one whose start defined the deadline of the CrownJob.
func isRunningPod(crownJob *clv1alpha2.CrownJob, job *batchv1.Job, pod *corev1.Pod) bool {
	if pod == nil || pod.Status.StartTime == nil || crownJob.Status.Deadline == nil || job.Spec.Template.Spec.ActiveDeadlineSeconds == nil {
		return false
	}
	deadline := pod.Status.StartTime.Add(time.Duration(*job.Spec.Template.Spec.ActiveDeadlineSeconds) * time.Second)
	return deadline.Equal(crownJob.Status.Deadline.Time)
}

// failureStatus returns the reason and the message of a failed job. The checks are ordered so that a job stopped
// by the platform is never reported as a failure of the command: the time limit first, then a pod rejected or
// evicted (whose container may have been killed with a non-zero exit code), then the exit code of the command.
func failureStatus(job *batchv1.Job, failed *batchv1.JobCondition, pod *corev1.Pod) (reason, message string) {
	if failed.Reason == batchv1.JobReasonDeadlineExceeded || (pod != nil && pod.Status.Reason == podReasonDeadlineExceeded) {
		maxRunTime := time.Duration(ptr.Deref(job.Spec.Template.Spec.ActiveDeadlineSeconds, 0)) * time.Second
		return ReasonTimeLimit, fmt.Sprintf(msgTimeLimit, utils.FormatDurationWithDays(maxRunTime))
	}

	if pod != nil && pod.Status.Reason != "" {
		detail := pod.Status.Reason
		if pod.Status.Message != "" {
			detail += ": " + pod.Status.Message
		}
		return ReasonCannotStart, fmt.Sprintf(msgPodRejected, shorten(strings.TrimSuffix(detail, ".")))
	}

	if pod != nil {
		if code := terminatedExitCode(pod.Status.InitContainerStatuses, forge.CrownJobRunFolderContainerName); code != nil && *code != 0 {
			return ReasonCannotStart, fmt.Sprintf(msgRunFolder, forge.CrownJobRunFolder(job.Name), *code)
		}
		if len(job.Spec.Template.Spec.Containers) > 0 {
			if code := terminatedExitCode(pod.Status.ContainerStatuses, job.Spec.Template.Spec.Containers[0].Name); code != nil && *code != 0 {
				return ReasonExitCode, fmt.Sprintf(msgExitCode, *code)
			}
		}
	}

	if failed.Message == "" {
		return ReasonError, msgJobFailedEmpty
	}
	return ReasonError, fmt.Sprintf(msgJobFailed, shorten(strings.TrimSuffix(failed.Message, ".")))
}

// imagePullState tells whether the kubelet is pulling an image of a pod.
type imagePullState int

const (
	// imagePullNone -> no image pull is in progress: none was needed (the image is on the node), or it is done.
	imagePullNone imagePullState = iota
	// imagePullInProgress -> the kubelet reported that it is pulling an image (Pulling), and not yet that it pulled it.
	imagePullInProgress
	// imagePullFailing -> a pull failed and the kubelet retries it (ErrImagePull, ImagePullBackOff): the image may
	// never arrive, so the job must not read as "downloading".
	imagePullFailing
)

// imagePull returns the image pull state of the given pod: failing if a container waits for a failed pull, otherwise
// as told by the events of the kubelet about the pod. A container waiting to be created is not a download by itself:
// it waits so also while the volumes are mounted, and for about a second when the image is already on the node.
func (r *Reconciler) imagePull(ctx context.Context, pod *corev1.Pod) imagePullState {
	if pod == nil {
		return imagePullNone
	}
	// The containers already created (with an image), by the field path of their events: they no longer pull.
	created := map[string]bool{}
	for _, group := range []struct {
		fieldPath string
		statuses  []corev1.ContainerStatus
	}{{"spec.initContainers{%s}", pod.Status.InitContainerStatuses}, {"spec.containers{%s}", pod.Status.ContainerStatuses}} {
		for i := range group.statuses {
			status := &group.statuses[i]
			if waiting := status.State.Waiting; waiting != nil && (waiting.Reason == containerReasonErrImagePull || waiting.Reason == containerReasonImagePullBackOff) {
				return imagePullFailing
			}
			if status.ImageID != "" {
				created[fmt.Sprintf(group.fieldPath, status.Name)] = true
			}
		}
	}

	// Events expire (after one hour by default): this matters only in the first minutes of a pod.
	var events corev1.EventList
	if err := r.APIReader.List(ctx, &events, client.InNamespace(pod.Namespace),
		client.MatchingFields{eventInvolvedObjectUIDField: string(pod.UID)}); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "failed retrieving the events of the job pod")
		return imagePullNone
	}
	return imagePullFromEvents(events.Items, created)
}

// imagePullFromEvents returns the image pull state told by the given events of a pod. For each container (the field
// path of the event), the newest pull event decides: Pulling means a download in progress, unless the container has
// been created meanwhile (the kubelet may drop an event); Pulled, that it is done (or that the image was already
// present); a failed pull, retries. The time of the kubelet events has a precision of a second: an event at the same
// time as Pulling counts as the newer one, so that a pull shorter than a second is done.
func imagePullFromEvents(events []corev1.Event, created map[string]bool) imagePullState {
	type pullEvent struct {
		state imagePullState
		at    time.Time
	}
	newest := map[string]pullEvent{}
	for i := range events {
		state, ok := pullEventState(&events[i])
		if !ok {
			continue
		}
		at, container := eventTime(&events[i]), events[i].InvolvedObject.FieldPath
		current, found := newest[container]
		if !found || at.After(current.at) || (at.Equal(current.at) && current.state == imagePullInProgress) {
			newest[container] = pullEvent{state: state, at: at}
		}
	}

	result := imagePullNone
	for container, event := range newest {
		switch event.state {
		case imagePullFailing:
			return imagePullFailing
		case imagePullInProgress:
			if !created[container] {
				result = imagePullInProgress
			}
		case imagePullNone:
		}
	}
	return result
}

// pullEventState returns the image pull state that the given event tells, and false if it is not about pulling an image.
// The reasons Failed and BackOff are used by the kubelet for other failures too: only those about a pull count.
func pullEventState(event *corev1.Event) (imagePullState, bool) {
	switch event.Reason {
	case eventReasonPulling:
		return imagePullInProgress, true
	case eventReasonPulled:
		return imagePullNone, true
	case eventReasonFailed, eventReasonBackOff:
		if strings.Contains(strings.ToLower(event.Message), "pull") {
			return imagePullFailing, true
		}
	}
	return imagePullNone, false
}

// eventTime returns the last time the given event happened.
func eventTime(event *corev1.Event) time.Time {
	switch {
	case !event.LastTimestamp.IsZero():
		return event.LastTimestamp.Time
	case event.Series != nil && !event.Series.LastObservedTime.IsZero():
		return event.Series.LastObservedTime.Time
	case !event.EventTime.IsZero():
		return event.EventTime.Time
	case !event.FirstTimestamp.IsZero():
		return event.FirstTimestamp.Time
	default:
		return event.CreationTimestamp.Time
	}
}

// quotaReached returns whether the Kueue Workload of the given Job is waiting because its ClusterQueue does not
// have enough unused quota. A missing Workload kind (e.g., in tests) or a read error count as "no".
func (r *Reconciler) quotaReached(ctx context.Context, job *batchv1.Job) bool {
	workloads := &unstructured.UnstructuredList{}
	workloads.SetGroupVersionKind(forge.KueueWorkloadGVK.GroupVersion().WithKind(forge.KueueWorkloadGVK.Kind + "List"))
	if err := r.APIReader.List(ctx, workloads, client.InNamespace(job.Namespace),
		client.MatchingLabels{forge.KueueJobUIDLabel: string(job.UID)}); err != nil {
		if !meta.IsNoMatchError(err) {
			ctrl.LoggerFrom(ctx).Error(err, "failed retrieving the workload of the job")
		}
		return false
	}

	for i := range workloads.Items {
		if !metav1.IsControlledBy(&workloads.Items[i], job) {
			continue
		}
		conditions, _, _ := unstructured.NestedSlice(workloads.Items[i].Object, "status", "conditions")
		for _, c := range conditions {
			condition, ok := c.(map[string]interface{})
			if !ok || condition["type"] != forge.KueueConditionQuotaReserved || condition["status"] != string(metav1.ConditionFalse) {
				continue
			}
			message, _ := condition["message"].(string)
			return strings.Contains(strings.ToLower(message), kueueInsufficientQuota)
		}
	}
	return false
}

// jobPod returns the most recent pod of the given Job, or nil if it has none.
func (r *Reconciler) jobPod(ctx context.Context, job *batchv1.Job) (*corev1.Pod, error) {
	var pods corev1.PodList
	if err := r.APIReader.List(ctx, &pods, client.InNamespace(job.Namespace),
		client.MatchingLabels{batchv1.ControllerUidLabel: string(job.UID)}); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "failed retrieving the pods of the job")
		return nil, err
	}

	var newest *corev1.Pod
	for i := range pods.Items {
		pod := &pods.Items[i]
		if newest == nil || newest.CreationTimestamp.Before(&pod.CreationTimestamp) ||
			(newest.CreationTimestamp.Equal(&pod.CreationTimestamp) && newest.Name < pod.Name) {
			newest = pod
		}
	}
	return newest, nil
}

// jobConditionTrue returns the condition of the given type of the Job, if it is true.
func jobConditionTrue(job *batchv1.Job, conditionType batchv1.JobConditionType) *batchv1.JobCondition {
	for i := range job.Status.Conditions {
		if job.Status.Conditions[i].Type == conditionType && job.Status.Conditions[i].Status == corev1.ConditionTrue {
			return &job.Status.Conditions[i]
		}
	}
	return nil
}

// terminatedExitCode returns the exit code of the named container, if it has terminated.
func terminatedExitCode(statuses []corev1.ContainerStatus, name string) *int32 {
	for i := range statuses {
		if statuses[i].Name == name && statuses[i].State.Terminated != nil {
			return &statuses[i].State.Terminated.ExitCode
		}
	}
	return nil
}
