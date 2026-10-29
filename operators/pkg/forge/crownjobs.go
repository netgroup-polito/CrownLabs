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
	"fmt"
	"math"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

const (
	// CrownJobLabelPrefix -> the prefix that replaces "crownlabs.polito.it/" in the labels of the pods of a job,
	// so that the Service of the instance never selects them.
	CrownJobLabelPrefix = "job." + crownLabsLabelPrefix
	// LabelCrownJobKey -> the key of the label identifying the CrownJob of a Job and of its pods.
	LabelCrownJobKey = CrownJobLabelPrefix + "crownjob"
	// LabelCrownJobWorkspaceKey -> the key of the label identifying the Workspace of the pods of a job.
	LabelCrownJobWorkspaceKey = CrownJobLabelPrefix + "workspace"

	// CrownJobRunFolderContainerName -> the name of the init container that creates the run folder of a job.
	CrownJobRunFolderContainerName = "run-folder"
	// CrownJobRunsDirectory -> the folder, in the persistent disk of the instance, holding one folder per run.
	CrownJobRunsDirectory = "runs"

	// CrownJobEnvJob -> the environment variable telling the container it runs as a job (value "true").
	CrownJobEnvJob = "CROWNLABS_JOB"
	// CrownJobEnvJobName -> the environment variable holding the name of the CrownJob.
	CrownJobEnvJobName = "CROWNLABS_JOB_NAME"
	// CrownJobEnvGPUs -> the environment variable holding the number of GPUs requested by the job.
	CrownJobEnvGPUs = "CROWNLABS_JOB_GPUS"
	// CrownJobEnvRunFolder -> the environment variable holding the absolute path of the run folder.
	CrownJobEnvRunFolder = "CROWNLABS_RUN_FOLDER"

	// CrownJobDeadlineMargin -> the time added to maxRunTime to obtain the activeDeadlineSeconds of the Job.
	// It counts from admission, so it also stops a pod that never gets scheduled.
	CrownJobDeadlineMargin = 600 * time.Second
	// CrownJobMaxNameLength -> the maximum length of the name of a CrownJob: the Job has the same name,
	// which Kubernetes copies into a label value.
	CrownJobMaxNameLength = 63

	crownLabsLabelPrefix = "crownlabs.polito.it/"
	crownJobGPUSuffix    = "/gpu"
)

// CannotStartError is an error whose message is a plain sentence, for the user, explaining why a job cannot start.
type CannotStartError struct {
	message string
}

// Error returns the message of the error.
func (e *CannotStartError) Error() string {
	return e.message
}

// NewCannotStartError returns a CannotStartError with the given formatted message.
func NewCannotStartError(format string, args ...interface{}) error {
	return &CannotStartError{message: fmt.Sprintf(format, args...)}
}

// CrownJobJobOpts groups the inputs, gathered by the reconciler, from which the Job of a CrownJob is forged.
type CrownJobJobOpts struct {
	// Workspace is the name of the Workspace whose job queue runs the job.
	Workspace string
	// EnvironmentName is the name of the environment copied to run the job (and of its app container).
	EnvironmentName string
	// PodTemplate is the pod template of the Deployment of the instance. It is copied, not modified.
	PodTemplate *corev1.PodTemplateSpec
	// CommandPolicy is the commandPolicy of the job section of the Template.
	CommandPolicy clv1alpha2.JobCommandPolicy
	// Command is the command chosen by CrownJobCommand (nil if the policy is ImageDefault).
	Command []string
	// OtherResources are the resources requested in addition to those of the environment (job.otherResources).
	OtherResources map[string]resource.Quantity
	// MaxRunTime is the maximum time the job may run, chosen by CrownJobMaxRunTime.
	MaxRunTime time.Duration
	// PriorityClass is the name of the Kueue WorkloadPriorityClass of the job (empty if none).
	PriorityClass string
	// RuntimeClasses maps a resource name to the RuntimeClass that a pod requesting it must use.
	RuntimeClasses map[string]string
	// FinishedJobTTL is how long the Job is kept after it ends, with its pods and its Workload (0: no limit).
	FinishedJobTTL time.Duration
}

// CrownJobBatchJob forges the Job of the given CrownJob, by copying the pod template of the Deployment of the instance:
// only the app container is kept, without probes and ports, with the chosen command, the job environment variables and the
// additional resources, after an init container that creates the run folder. The Job is suspended, so that Kueue admits it.
func CrownJobBatchJob(crownJob *clv1alpha2.CrownJob, opts *CrownJobJobOpts) (*batchv1.Job, error) {
	podTemplate := opts.PodTemplate.DeepCopy()
	podTemplate.Labels = CrownJobPodLabels(podTemplate.Labels, crownJob.Name, opts.Workspace)

	var app *corev1.Container
	for i := range podTemplate.Spec.Containers {
		if podTemplate.Spec.Containers[i].Name == opts.EnvironmentName {
			app = &podTemplate.Spec.Containers[i]
			break
		}
	}
	if app == nil {
		return nil, NewCannotStartError("The instance has no container named %s.", opts.EnvironmentName)
	}

	mountPath := ""
	for i := range app.VolumeMounts {
		if app.VolumeMounts[i].Name == PersistentVolumeName {
			mountPath = app.VolumeMounts[i].MountPath
			break
		}
	}
	if mountPath == "" {
		return nil, NewCannotStartError("The instance has no persistent disk mounted in its container %s.", opts.EnvironmentName)
	}
	runFolder := path.Join(mountPath, CrownJobRunFolder(crownJob.Name))

	if opts.CommandPolicy != clv1alpha2.JobCommandImageDefault {
		app.Command = append([]string(nil), opts.Command...)
		app.Args = nil
		app.WorkingDir = mountPath
	}
	app.ReadinessProbe = nil
	app.LivenessProbe = nil
	app.StartupProbe = nil
	app.Ports = nil

	app.Resources.Requests = addResources(app.Resources.Requests, opts.OtherResources)
	app.Resources.Limits = addResources(app.Resources.Limits, opts.OtherResources)

	setEnvVariable(app, CrownJobEnvJob, strconv.FormatBool(true))
	setEnvVariable(app, CrownJobEnvJobName, crownJob.Name)
	setEnvVariable(app, CrownJobEnvGPUs, strconv.FormatInt(gpuCount(app.Resources.Requests), 10))
	setEnvVariable(app, CrownJobEnvRunFolder, runFolder)

	runtimeClass, err := crownJobRuntimeClass(app.Resources.Requests, opts.RuntimeClasses)
	if err != nil {
		return nil, err
	}
	if runtimeClass != "" {
		podTemplate.Spec.RuntimeClassName = ptr.To(runtimeClass)
	}

	runFolderContainer := GenericContainer(CrownJobRunFolderContainerName, app.Image)
	runFolderContainer.ImagePullPolicy = app.ImagePullPolicy
	runFolderContainer.Command = []string{"mkdir", "-p", runFolder}
	runFolderContainer.SecurityContext = app.SecurityContext.DeepCopy()
	AddContainerVolumeMount(&runFolderContainer, PersistentVolumeName, mountPath, false)
	SetContainerResources(&runFolderContainer, 0.01, 0.1, 16, 64)

	podTemplate.Spec.Containers = []corev1.Container{*app}
	podTemplate.Spec.InitContainers = []corev1.Container{runFolderContainer}
	podTemplate.Spec.RestartPolicy = corev1.RestartPolicyNever
	podTemplate.Spec.ActiveDeadlineSeconds = ptr.To(int64(opts.MaxRunTime / time.Second))

	labels := map[string]string{
		KueueQueueNameLabel: CrownJobQueueName(opts.Workspace),
		LabelCrownJobKey:    crownJob.Name,
	}
	if opts.PriorityClass != "" {
		labels[KueuePriorityClassLabel] = opts.PriorityClass
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            crownJob.Name,
			Namespace:       crownJob.Namespace,
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(crownJob, clv1alpha2.GroupVersion.WithKind("CrownJob"))},
		},
		Spec: batchv1.JobSpec{
			Suspend:                 ptr.To(true),
			BackoffLimit:            ptr.To[int32](0),
			ActiveDeadlineSeconds:   ptr.To(int64((opts.MaxRunTime + CrownJobDeadlineMargin) / time.Second)),
			PodFailurePolicy:        CrownJobPodFailurePolicy(opts.EnvironmentName),
			TTLSecondsAfterFinished: CrownJobFinishedJobTTL(opts.FinishedJobTTL),
			Template:                *podTemplate,
		},
	}, nil
}

// CrownJobFinishedJobTTL returns the ttlSecondsAfterFinished of the Jobs of a CrownJob (nil if the given TTL is 0).
// Once a Job is deleted, its pods no longer keep the disk of the instance in use (pvc-protection counts finished pods);
// the CrownJob is kept, with its terminal status.
func CrownJobFinishedJobTTL(ttl time.Duration) *int32 {
	seconds := int64(ttl / time.Second)
	switch {
	case seconds <= 0:
		return nil
	case seconds > math.MaxInt32:
		return ptr.To[int32](math.MaxInt32)
	default:
		return ptr.To(int32(seconds))
	}
}

// CrownJobPodFailurePolicy forges the pod failure policy of the Job of a CrownJob. The rules are evaluated in order and
// the first match wins: a disrupted pod (e.g., evicted) is ignored and replaced, otherwise a non-zero exit code of the
// app container fails the Job at once. The Ignore rule must come first: a disrupted container is killed, and exits
// with a non-zero code that the FailJob rule would match.
func CrownJobPodFailurePolicy(appContainerName string) *batchv1.PodFailurePolicy {
	return &batchv1.PodFailurePolicy{
		Rules: []batchv1.PodFailurePolicyRule{
			{
				Action: batchv1.PodFailurePolicyActionIgnore,
				OnPodConditions: []batchv1.PodFailurePolicyOnPodConditionsPattern{{
					Type:   corev1.DisruptionTarget,
					Status: corev1.ConditionTrue,
				}},
			},
			{
				Action: batchv1.PodFailurePolicyActionFailJob,
				OnExitCodes: &batchv1.PodFailurePolicyOnExitCodesRequirement{
					ContainerName: ptr.To(appContainerName),
					Operator:      batchv1.PodFailurePolicyOnExitCodesOpNotIn,
					Values:        []int32{0},
				},
			},
		},
	}
}

// CrownJobPodLabels returns the labels of the pods of a job, starting from those of the pods of the instance:
// every key "crownlabs.polito.it/<k>" becomes "job.crownlabs.polito.it/<k>", with the same value, so that the
// Service of the instance never selects a job pod; the labels with the CrownJob and Workspace names are added.
func CrownJobPodLabels(instanceLabels map[string]string, crownJob, workspace string) map[string]string {
	labels := make(map[string]string, len(instanceLabels)+2)
	for key, value := range instanceLabels {
		if suffix, found := strings.CutPrefix(key, crownLabsLabelPrefix); found {
			key = CrownJobLabelPrefix + suffix
		}
		labels[key] = value
	}
	labels[LabelCrownJobKey] = crownJob
	labels[LabelCrownJobWorkspaceKey] = workspace
	return labels
}

// CrownJobRunFolder returns the run folder of the given CrownJob, relative to the persistent disk of the instance.
func CrownJobRunFolder(crownJob string) string {
	return path.Join(CrownJobRunsDirectory, crownJob)
}

// CrownJobCommand returns the command of a job, chosen by the commandPolicy of the job section of the Template:
// Fixed uses the command of the Template; Editable uses the command of the CrownJob, if not empty, or else that of
// the Template; ImageDefault uses no command. The command of the CrownJob is refused unless the policy is Editable.
func CrownJobCommand(templateJob *clv1alpha2.TemplateJob, crownJobCommand []string) ([]string, error) {
	switch templateJob.CommandPolicy {
	case clv1alpha2.JobCommandFixed:
		if len(crownJobCommand) > 0 {
			return nil, NewCannotStartError("The command of this job is fixed by its template: the job cannot set its own command.")
		}
		return templateJob.Command, nil
	case clv1alpha2.JobCommandImageDefault:
		if len(crownJobCommand) > 0 {
			return nil, NewCannotStartError("This job runs the command of its image: the job cannot set a command.")
		}
		return nil, nil
	default:
		// Editable, which is also the default value of the field.
		if len(crownJobCommand) > 0 {
			return crownJobCommand, nil
		}
		if len(templateJob.Command) > 0 {
			return templateJob.Command, nil
		}
		return nil, NewCannotStartError("The job has no command: write the command to run.")
	}
}

// CrownJobMaxRunTime returns the maximum time a job may run, given the maxRunTime of the CrownJob, of the job
// section of the Template and of the queueLimits of the Workspace (each one possibly empty).
// The limit is the smallest of the Template and Workspace values, or defaultMaxRunTime if neither is set.
// The CrownJob value can only lower that limit: a larger value is refused.
func CrownJobMaxRunTime(crownJobValue, templateValue, workspaceValue string, defaultMaxRunTime time.Duration) (time.Duration, error) {
	limit := time.Duration(0)
	for _, value := range []string{templateValue, workspaceValue} {
		if value == "" {
			continue
		}
		parsed, err := utils.ParseDurationWithDays(value)
		if err != nil {
			return 0, NewCannotStartError("The time limit %s set for this job's template or workspace is not valid.", value)
		}
		if limit == 0 || parsed < limit {
			limit = parsed
		}
	}
	if limit == 0 {
		limit = defaultMaxRunTime
	}

	if crownJobValue == "" {
		return limit, nil
	}
	requested, err := utils.ParseDurationWithDays(crownJobValue)
	if err != nil {
		return 0, NewCannotStartError("The time limit %s of the job is not valid.", crownJobValue)
	}
	if requested > limit {
		return 0, NewCannotStartError("The job asks for a time limit of %s, but at most %s is allowed.",
			crownJobValue, utils.FormatDurationWithDays(limit))
	}
	return requested, nil
}

// PodEffectiveRequests returns the resource requests of a pod as the scheduler and Kueue count them: for each
// resource, the largest between the requests of each init container and the sum of the requests of the containers.
func PodEffectiveRequests(spec *corev1.PodSpec) corev1.ResourceList {
	requests := corev1.ResourceList{}
	for i := range spec.Containers {
		requests = addResourceList(requests, spec.Containers[i].Resources.Requests)
	}
	for i := range spec.InitContainers {
		for name, quantity := range spec.InitContainers[i].Resources.Requests {
			if current, found := requests[name]; !found || quantity.Cmp(current) > 0 {
				requests[name] = quantity.DeepCopy()
			}
		}
	}
	return requests
}

// CrownJobCheckQueueLimits checks the requests of a job against the job queue of the given Workspace: every requested
// resource must be covered by the queue (see KueueQueueQuotas), and no request may exceed maxAtOnce, or the job could
// never start.
func CrownJobCheckQueueLimits(workspace string, limits *clv1alpha1.QueueLimits, requests corev1.ResourceList) error {
	quotas := map[corev1.ResourceName]resource.Quantity{}
	for _, quota := range KueueQueueQuotas(limits) {
		quotas[quota.Name] = quota.Quota
	}

	names := make([]string, 0, len(requests))
	for name := range requests {
		names = append(names, name.String())
	}
	sort.Strings(names)

	for _, name := range names {
		requested := requests[corev1.ResourceName(name)]
		quota, found := quotas[corev1.ResourceName(name)]
		if !found {
			return NewCannotStartError("The job queue of workspace %s does not allow %s.", workspace, name)
		}
		if requested.Cmp(quota) > 0 {
			return NewCannotStartError("The job asks for %s of %s, but the job queue of workspace %s allows at most %s at once: it could never start.",
				requested.String(), name, workspace, quota.String())
		}
	}
	return nil
}

// crownJobRuntimeClass returns the RuntimeClass required by the requested resources, according to the given mapping
// from resource names to RuntimeClasses: empty if no requested resource matches, an error if two match different ones.
func crownJobRuntimeClass(requests corev1.ResourceList, runtimeClasses map[string]string) (string, error) {
	matches := map[string]struct{}{}
	for name := range requests {
		if runtimeClass, found := runtimeClasses[name.String()]; found {
			matches[runtimeClass] = struct{}{}
		}
	}

	names := make([]string, 0, len(matches))
	for runtimeClass := range matches {
		names = append(names, runtimeClass)
	}
	sort.Strings(names)

	switch len(names) {
	case 0:
		return "", nil
	case 1:
		return names[0], nil
	default:
		return "", NewCannotStartError("The job asks for resources that need different runtimes (%s): it cannot run.", strings.Join(names, ", "))
	}
}

// gpuCount returns the sum of the requested resources whose name ends with "/gpu".
func gpuCount(requests corev1.ResourceList) int64 {
	total := int64(0)
	for name, quantity := range requests {
		if strings.HasSuffix(name.String(), crownJobGPUSuffix) {
			total += quantity.Value()
		}
	}
	return total
}

// addResources returns the given resource list with the given resources added, summed per name.
func addResources(list corev1.ResourceList, resources map[string]resource.Quantity) corev1.ResourceList {
	other := make(corev1.ResourceList, len(resources))
	for name, quantity := range resources {
		other[corev1.ResourceName(name)] = quantity
	}
	return addResourceList(list, other)
}

// addResourceList returns a copy of the given resource list with the other one added, summed per name.
func addResourceList(list, other corev1.ResourceList) corev1.ResourceList {
	result := make(corev1.ResourceList, len(list)+len(other))
	for name, quantity := range list {
		result[name] = quantity.DeepCopy()
	}
	for name, quantity := range other {
		sum := result[name]
		sum.Add(quantity)
		result[name] = sum
	}
	return result
}

// setEnvVariable sets an environment variable of the given container, replacing its value if it is already present.
func setEnvVariable(c *corev1.Container, name, value string) {
	for i := range c.Env {
		if c.Env[i].Name == name {
			c.Env[i] = corev1.EnvVar{Name: name, Value: value}
			return
		}
	}
	AddEnvVariableToContainer(c, name, value)
}
