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

package forge

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const (
	// CrownJobInstanceFinalizer -> the finalizer that keeps an Instance while it has CrownJobs that are not terminal.
	CrownJobInstanceFinalizer = "crownlabs.polito.it/crownjobs"

	// LabelCrownJobComponentKey -> the key of the label telling the role of a Job owned by a CrownJob that is not
	// its main Job (the main Job has the same name as the CrownJob).
	LabelCrownJobComponentKey = CrownJobLabelPrefix + "component"
	// CrownJobComponentLogHelper -> the value of the component label of the Jobs that save the log of a job pod.
	CrownJobComponentLogHelper = "log-helper"
	// AnnotationCrownJobPodKey -> the annotation of a log helper Job naming the job pod whose log it saves.
	AnnotationCrownJobPodKey = CrownJobLabelPrefix + "pod"
	// AnnotationCrownJobLogFileKey -> the annotation of a log helper Job with the file it writes, relative to the disk.
	AnnotationCrownJobLogFileKey = CrownJobLabelPrefix + "log-file"
	// AnnotationCrownJobLogsNotSavedKey -> the annotation set on a failed log helper Job once its failure is reported.
	AnnotationCrownJobLogsNotSavedKey = CrownJobLabelPrefix + "logs-not-saved"
	// AnnotationCrownJobLogFilesKey -> the annotation of the main Job of a CrownJob with the number of log helpers
	// created for its pods. The helpers are deleted some time after they end, while the main Job lives longer.
	AnnotationCrownJobLogFilesKey = CrownJobLabelPrefix + "log-files"

	// CrownJobLogReaderName -> the name of the ServiceAccount, Role and RoleBinding of the log helpers of a namespace.
	CrownJobLogReaderName = "crownjob-log-reader"
	// CrownJobLogHelperContainerName -> the name of the container of a log helper.
	CrownJobLogHelperContainerName = "save-log"

	crownJobLogHelperBackoffLimit = 2
	crownJobLogHelperHashLength   = 5
	crownJobLogFileName           = "log"
)

// CrownJobLogHelperOpts groups the inputs of the log helper of a job pod that do not come from the Job or the pod.
type CrownJobLogHelperOpts struct {
	// Image is the image of the helper: it must contain /go-runner and /bin/kubectl (e.g., registry.k8s.io/kubectl).
	Image string
	// ReadWriteMany tells whether the disk of the job can be mounted on another node: if not, the helper runs on the
	// node of the job pod.
	ReadWriteMany bool
	// FinishedJobTTL is how long the helper Job is kept after it ends (0: no limit).
	FinishedJobTTL time.Duration
}

// CrownJobLogFile returns the file, relative to the persistent disk of the instance, where the log of the index-th pod of
// a job (counting from 1) is saved: runs/<name>/log.txt for the first pod, log-2.txt, log-3.txt, ... for its replacements.
func CrownJobLogFile(crownJob string, index int) string {
	name := crownJobLogFileName + ".txt"
	if index > 1 {
		name = fmt.Sprintf("%s-%d.txt", crownJobLogFileName, index)
	}
	return path.Join(CrownJobRunFolder(crownJob), name)
}

// CrownJobLogHelperName returns the name of the index-th log helper Job of a CrownJob: the name of the CrownJob, cut to
// fit in 63 characters, followed by "-log<index>-" and a hash of the CrownJob UID, so that it never equals the name
// of the main Job of another CrownJob of the namespace by chance.
func CrownJobLogHelperName(crownJob *clv1alpha2.CrownJob, index int) string {
	hash := sha256.Sum256([]byte(crownJob.UID))
	suffix := fmt.Sprintf("-log%d-%s", index, hex.EncodeToString(hash[:])[:crownJobLogHelperHashLength])
	base := crownJob.Name
	if maxLength := CrownJobMaxNameLength - len(suffix); len(base) > maxLength {
		base = base[:maxLength]
	}
	return strings.TrimRight(base, "-.") + suffix
}

// CrownJobLogHelperJob forges the Job that saves into the run folder the log of the given pod of the Job of a CrownJob.
// Its pod follows the log of the app container from its start (kubectl logs -f through go-runner, which writes it to the
// file), with the security context and the persistent disk of the job pod. It carries a pod-level deadline (the maximum
// run time of the job plus a margin), so that the namespace quota of the instances does not count it, and no Kueue label.
func CrownJobLogHelperJob(crownJob *clv1alpha2.CrownJob, job *batchv1.Job, pod *corev1.Pod, index int, opts *CrownJobLogHelperOpts) (*batchv1.Job, error) {
	podSpec := &job.Spec.Template.Spec
	if len(podSpec.Containers) == 0 {
		return nil, fmt.Errorf("the job %s has no container", job.Name)
	}
	app := &podSpec.Containers[0]

	var mount *corev1.VolumeMount
	for i := range app.VolumeMounts {
		if app.VolumeMounts[i].Name == PersistentVolumeName {
			mount = &app.VolumeMounts[i]
			break
		}
	}
	var volume *corev1.Volume
	for i := range podSpec.Volumes {
		if podSpec.Volumes[i].Name == PersistentVolumeName {
			volume = podSpec.Volumes[i].DeepCopy()
			break
		}
	}
	if mount == nil || volume == nil {
		return nil, fmt.Errorf("the job %s has no persistent disk", job.Name)
	}
	logFile := CrownJobLogFile(crownJob.Name, index)

	container := GenericContainer(CrownJobLogHelperContainerName, opts.Image)
	container.ImagePullPolicy = corev1.PullIfNotPresent
	container.Command = []string{
		"/go-runner", "-log-file=" + path.Join(mount.MountPath, logFile), "-redirect-stderr=false",
		"/bin/kubectl", "logs", "-f", "-n", pod.Namespace, pod.Name, "-c", app.Name,
	}
	AddContainerVolumeMount(&container, PersistentVolumeName, mount.MountPath, false)
	SetContainerResources(&container, 0.01, 0.1, 32, 64)

	maxRunTime := time.Duration(ptr.Deref(podSpec.ActiveDeadlineSeconds, 0)) * time.Second
	labels := map[string]string{
		LabelCrownJobKey:          crownJob.Name,
		LabelCrownJobComponentKey: CrownJobComponentLogHelper,
	}

	helper := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CrownJobLogHelperName(crownJob, index),
			Namespace: crownJob.Namespace,
			Labels:    labels,
			Annotations: map[string]string{
				AnnotationCrownJobPodKey:     pod.Name,
				AnnotationCrownJobLogFileKey: logFile,
			},
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(crownJob, clv1alpha2.GroupVersion.WithKind("CrownJob"))},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            ptr.To[int32](crownJobLogHelperBackoffLimit),
			TTLSecondsAfterFinished: CrownJobFinishedJobTTL(opts.FinishedJobTTL),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					RestartPolicy:                corev1.RestartPolicyNever,
					ServiceAccountName:           CrownJobLogReaderName,
					AutomountServiceAccountToken: ptr.To(true),
					SecurityContext:              podSpec.SecurityContext.DeepCopy(),
					ActiveDeadlineSeconds:        ptr.To(int64((maxRunTime + CrownJobDeadlineMargin) / time.Second)),
					Containers:                   []corev1.Container{container},
					Volumes:                      []corev1.Volume{*volume},
				},
			},
		},
	}

	// A disk that is not ReadWriteMany can be mounted only on the node of the job pod. The node is selected by name,
	// so that the helper is scheduled there also when the job pod has already ended.
	if !opts.ReadWriteMany && pod.Spec.NodeName != "" {
		helper.Spec.Template.Spec.Affinity = &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchFields: []corev1.NodeSelectorRequirement{{
						Key:      metav1.ObjectNameField,
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{pod.Spec.NodeName},
					}},
				}},
			},
		}}
	}

	return helper, nil
}

// CrownJobLogReaderServiceAccount forges the ServiceAccount of the log helpers of a namespace.
func CrownJobLogReaderServiceAccount(namespace string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta:   metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: "ServiceAccount"},
		ObjectMeta: crownJobLogReaderMeta(namespace),
	}
}

// CrownJobLogReaderRole forges the Role of the log helpers of a namespace: it can only read the pods and their logs.
func CrownJobLogReaderRole(namespace string) *rbacv1.Role {
	return &rbacv1.Role{
		TypeMeta:   metav1.TypeMeta{APIVersion: rbacv1.SchemeGroupVersion.String(), Kind: "Role"},
		ObjectMeta: crownJobLogReaderMeta(namespace),
		Rules: []rbacv1.PolicyRule{
			{APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"get", "list"}},
			{APIGroups: []string{""}, Resources: []string{"pods/log"}, Verbs: []string{"get"}},
		},
	}
}

// CrownJobLogReaderRoleBinding forges the RoleBinding granting the Role of the log helpers to their ServiceAccount.
func CrownJobLogReaderRoleBinding(namespace string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		TypeMeta:   metav1.TypeMeta{APIVersion: rbacv1.SchemeGroupVersion.String(), Kind: "RoleBinding"},
		ObjectMeta: crownJobLogReaderMeta(namespace),
		RoleRef:    rbacv1.RoleRef{APIGroup: rbacv1.GroupName, Kind: "Role", Name: CrownJobLogReaderName},
		Subjects:   []rbacv1.Subject{{Kind: rbacv1.ServiceAccountKind, Name: CrownJobLogReaderName, Namespace: namespace}},
	}
}

// IsReadWriteMany returns whether the given access modes include ReadWriteMany.
func IsReadWriteMany(accessModes []corev1.PersistentVolumeAccessMode) bool {
	for _, mode := range accessModes {
		if mode == corev1.ReadWriteMany {
			return true
		}
	}
	return false
}

// crownJobLogReaderMeta returns the metadata of the objects of the log helpers of a namespace.
func crownJobLogReaderMeta(namespace string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      CrownJobLogReaderName,
		Namespace: namespace,
		Labels:    map[string]string{LabelManagedByKey: labelManagedByCrownJobValue},
	}
}
