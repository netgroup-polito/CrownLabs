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
	"strconv"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

// The log of every job pod is saved into the run folder by a helper Job, created when the app container of the pod is
// first seen started (running, or already ended for a short job). The helper follows the log from the start of the
// container, so that nothing is lost when the kubelet rotates its files. A helper never changes the phase of the
// CrownJob: its failure is recorded as a LogsNotSaved event. A helper is deleted some time after it ends (see
// FinishedJobTTL), possibly while the main Job still runs a replacement pod: the number of helpers created is kept on
// the main Job, which numbers the log files and tells whether the log of the first pod was saved.

// logHelpers returns the log helper Jobs of the given CrownJob: those with its name and the log-helper component labels,
// and controlled by it. The main Job carries the CrownJob label too, but not the component label.
func (r *Reconciler) logHelpers(ctx context.Context, crownJob *clv1alpha2.CrownJob) ([]batchv1.Job, error) {
	var jobs batchv1.JobList
	if err := r.List(ctx, &jobs, client.InNamespace(crownJob.Namespace), client.MatchingLabels{
		forge.LabelCrownJobKey:          crownJob.Name,
		forge.LabelCrownJobComponentKey: forge.CrownJobComponentLogHelper,
	}); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "failed retrieving the log helpers of the crownjob")
		return nil, err
	}

	helpers := make([]batchv1.Job, 0, len(jobs.Items))
	for i := range jobs.Items {
		if metav1.IsControlledBy(&jobs.Items[i], crownJob) {
			helpers = append(helpers, jobs.Items[i])
		}
	}
	return helpers, nil
}

// ensureLogHelper creates the log helper of the given job pod, if its app container has started and it has none yet,
// and returns the log helpers of the CrownJob. A helper that cannot be created is reported with a LogsNotSaved event;
// only an error that may be transient is returned.
func (r *Reconciler) ensureLogHelper(ctx context.Context, crownJob *clv1alpha2.CrownJob, job *batchv1.Job, pod *corev1.Pod) ([]batchv1.Job, error) {
	log := ctrl.LoggerFrom(ctx)

	helpers, err := r.logHelpers(ctx, crownJob)
	if err != nil || pod == nil || pod.Spec.NodeName == "" || !appContainerStarted(job, pod) {
		return helpers, err
	}
	created := logHelpersCreated(job, helpers)
	for i := range helpers {
		if helpers[i].Annotations[forge.AnnotationCrownJobPodKey] == pod.Name {
			return helpers, r.recordLogHelpersCreated(ctx, job, created)
		}
	}

	index := created + 1
	logFile := forge.CrownJobLogFile(crownJob.Name, index)
	notSaved := func(reason string) ([]batchv1.Job, error) {
		r.EventsRecorder.Event(crownJob, corev1.EventTypeWarning, ReasonLogsNotSaved, fmt.Sprintf(msgLogsNotSaved, logFile, strings.TrimSuffix(reason, ".")))
		return helpers, nil
	}

	// The helper runs on the node of the job pod, unless the disk is ReadWriteMany.
	readWriteMany := false
	if claimName := persistentClaimName(&job.Spec.Template.Spec); claimName != "" {
		var pvc corev1.PersistentVolumeClaim
		if found, err := r.read(ctx, types.NamespacedName{Namespace: job.Namespace, Name: claimName}, &pvc); err != nil {
			return helpers, err
		} else if found {
			readWriteMany = forge.IsReadWriteMany(pvc.Spec.AccessModes)
		}
	}

	helper, err := forge.CrownJobLogHelperJob(crownJob, job, pod, index, &forge.CrownJobLogHelperOpts{
		Image: r.LogHelperImage, ReadWriteMany: readWriteMany, FinishedJobTTL: r.FinishedJobTTL,
	})
	if err != nil {
		return notSaved(err.Error())
	}

	// The ServiceAccount of the helpers, which can only read the pods and their logs, is shared by the namespace.
	for _, obj := range []client.Object{
		forge.CrownJobLogReaderServiceAccount(job.Namespace),
		forge.CrownJobLogReaderRole(job.Namespace),
		forge.CrownJobLogReaderRoleBinding(job.Namespace),
	} {
		if err := r.Patch(ctx, obj, client.Apply, client.FieldOwner(forge.CrownJobFieldManager), client.ForceOwnership); err != nil {
			log.Error(err, "failed to apply the permissions of the log helpers", "kind", obj.GetObjectKind().GroupVersionKind().Kind)
			if kerrors.IsForbidden(err) || kerrors.IsInvalid(err) {
				return notSaved(shorten(err.Error()))
			}
			return helpers, err
		}
	}

	if err := r.Create(ctx, helper); err != nil {
		if !kerrors.IsAlreadyExists(err) {
			log.Error(err, "failed to create the log helper", "job", klog.KObj(helper))
			if kerrors.IsForbidden(err) || kerrors.IsInvalid(err) {
				return notSaved(shorten(err.Error()))
			}
			return helpers, err
		}
		// Created by a previous reconcile that the cache has not seen yet, or a Job with the same name that is not ours.
		var existing batchv1.Job
		if err := r.APIReader.Get(ctx, client.ObjectKeyFromObject(helper), &existing); err != nil {
			return helpers, err
		}
		if !metav1.IsControlledBy(&existing, crownJob) || existing.Labels[forge.LabelCrownJobComponentKey] != forge.CrownJobComponentLogHelper {
			return notSaved(fmt.Sprintf(msgJobNameTaken, existing.Name))
		}
		return append(helpers, existing), r.recordLogHelpersCreated(ctx, job, index)
	}
	log.Info("log helper created", "job", klog.KObj(helper), "pod", pod.Name, "file", logFile)
	return append(helpers, *helper), r.recordLogHelpersCreated(ctx, job, index)
}

// logHelpersCreated returns the number of log helpers created for the pods of the given main Job: the count recorded on
// the Job, or the number of helpers that still exist if larger (e.g., the count could not be recorded yet).
func logHelpersCreated(job *batchv1.Job, helpers []batchv1.Job) int {
	recorded, err := strconv.Atoi(job.Annotations[forge.AnnotationCrownJobLogFilesKey])
	if err != nil {
		recorded = 0
	}
	return max(recorded, len(helpers))
}

// recordLogHelpersCreated records on the main Job that the given number of log helpers has been created, if it is more
// than the recorded one. The Job lives at least until its last pod has ended, unlike the helpers.
func (r *Reconciler) recordLogHelpersCreated(ctx context.Context, job *batchv1.Job, created int) error {
	if created <= logHelpersCreated(job, nil) {
		return nil
	}
	original := job.DeepCopy()
	if job.Annotations == nil {
		job.Annotations = map[string]string{}
	}
	job.Annotations[forge.AnnotationCrownJobLogFilesKey] = strconv.Itoa(created)
	if err := r.Patch(ctx, job, client.MergeFrom(original)); err != nil && !kerrors.IsNotFound(err) {
		ctrl.LoggerFrom(ctx).Error(err, "failed to record the log helpers on the job", "job", klog.KObj(job))
		return err
	}
	return nil
}

// reportFailedLogHelpers records a LogsNotSaved event for each failed log helper not reported yet, and marks it.
func (r *Reconciler) reportFailedLogHelpers(ctx context.Context, crownJob *clv1alpha2.CrownJob, helpers []batchv1.Job) error {
	for i := range helpers {
		helper := &helpers[i]
		failed := jobConditionTrue(helper, batchv1.JobFailed)
		if failed == nil || helper.Annotations[forge.AnnotationCrownJobLogsNotSavedKey] != "" {
			continue
		}

		reason := strings.TrimSuffix(failed.Message, ".")
		if reason == "" {
			reason = failed.Reason
		}
		r.EventsRecorder.Event(crownJob, corev1.EventTypeWarning, ReasonLogsNotSaved,
			fmt.Sprintf(msgLogsNotSaved, helper.Annotations[forge.AnnotationCrownJobLogFileKey], shorten(reason)))

		original := helper.DeepCopy()
		if helper.Annotations == nil {
			helper.Annotations = map[string]string{}
		}
		helper.Annotations[forge.AnnotationCrownJobLogsNotSavedKey] = "true"
		if err := r.Patch(ctx, helper, client.MergeFrom(original)); err != nil && !kerrors.IsNotFound(err) {
			ctrl.LoggerFrom(ctx).Error(err, "failed to mark the log helper as reported", "job", klog.KObj(helper))
			return err
		}
	}
	return nil
}

// logSavedSuffix returns the sentence ending the terminal messages when a log helper was created for the first pod
// of the job (the controller cannot read the disk, so it relies on the creation of the helper): as recorded on the
// main Job (nil if it is gone), or as shown by the helper itself.
func logSavedSuffix(crownJob *clv1alpha2.CrownJob, job *batchv1.Job, helpers []batchv1.Job) string {
	first := forge.CrownJobLogFile(crownJob.Name, 1)
	if job != nil && logHelpersCreated(job, nil) > 0 {
		return fmt.Sprintf(msgLogSaved, first)
	}
	for i := range helpers {
		if helpers[i].Annotations[forge.AnnotationCrownJobLogFileKey] == first {
			return fmt.Sprintf(msgLogSaved, first)
		}
	}
	return ""
}

// appContainerStarted returns whether the app container of the given job pod is running, or has ended.
func appContainerStarted(job *batchv1.Job, pod *corev1.Pod) bool {
	if len(job.Spec.Template.Spec.Containers) == 0 {
		return false
	}
	name := job.Spec.Template.Spec.Containers[0].Name
	for i := range pod.Status.ContainerStatuses {
		if status := &pod.Status.ContainerStatuses[i]; status.Name == name {
			return status.State.Running != nil || status.State.Terminated != nil
		}
	}
	return false
}
