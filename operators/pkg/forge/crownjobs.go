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
	// Command is the command chosen by CrownJobCommand (empty if the container runs as in the instance).
	Command []string
	// Resources are the resources of the job section of the Template (job.resources): each one replaces the request and
	// the limit of the instance.
	Resources map[string]resource.Quantity
	// MaxRunTime is the maximum time the job may run, chosen by CrownJobMaxRunTime.
	MaxRunTime time.Duration
	// RuntimeClasses maps a resource name to the RuntimeClass that a pod requesting it must use.
	RuntimeClasses map[string]string
	// FinishedJobTTL is how long the Job is kept after it ends, with its pods and its Workload (0: no limit).
	FinishedJobTTL time.Duration
}

// CrownJobBatchJob forges the Job of the given CrownJob, by copying the pod template of the Deployment of the instance:
// only the app container is kept, without probes and ports, with the chosen command, the job environment variables and the
// resources of the job, after an init container that creates the run folder. The Job is suspended, so that Kueue admits it.
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

	setJobCommand(app, opts.Command, mountPath)
	app.ReadinessProbe = nil
	app.LivenessProbe = nil
	app.StartupProbe = nil
	app.Ports = nil

	setJobResources(&app.Resources, opts.Resources)
	// A zero request asks for nothing: it is left out of the Job (the Template editor of the dashboard writes 0 GPUs in
	// the environment, and the job section of a Template lists 0 of a resource that its jobs must not ask for).
	dropZeroExtendedResources(&app.Resources)

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

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      crownJob.Name,
			Namespace: crownJob.Namespace,
			Labels: map[string]string{
				KueueQueueNameLabel:     CrownJobQueueName(opts.Workspace),
				KueuePriorityClassLabel: CrownJobPriorityClassName(opts.Workspace),
				LabelCrownJobKey:        crownJob.Name,
			},
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

// CrownJobCommand returns the command of a job, given the job section of its Template and the command of the CrownJob,
// where absent (nil) and empty are different. If the job section is not editable, the CrownJob cannot set a command
// (the empty one included), and the job runs the command of the Template. If it is editable, the job runs the command
// of the CrownJob, or that of the Template if the CrownJob sets none. An empty result means that the container runs as
// in the instance: this is the case when the chosen command is empty (no command in the Template, or [] in the CrownJob).
func CrownJobCommand(templateJob *clv1alpha2.TemplateJob, crownJobCommand []string) ([]string, error) {
	if crownJobCommand == nil {
		return templateJob.Command, nil
	}
	if !templateJob.Editable {
		return nil, NewCannotStartError("The command of this job is fixed by its template: the job cannot set its own command.")
	}
	return crownJobCommand, nil
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

// CrownJobCheckTemplateResources checks the resources of the job section of the given Template (job.resources), which
// the API accepts whatever their quantity: a negative quantity is refused, and so is 0 of a standard resource (a name
// without a "/", e.g., cpu and memory), which no job can do without. 0 of an extended resource (e.g., nvidia.com/gpu)
// is accepted: it means that the job does not ask for it.
func CrownJobCheckTemplateResources(template string, resources map[string]resource.Quantity) error {
	for _, name := range sortedResourceNames(resources) {
		quantity := resources[name]
		switch {
		case quantity.Sign() < 0:
			return NewCannotStartError("The template %s sets a negative quantity of %s for its jobs (%s).", template, name, quantity.String())
		case quantity.IsZero() && !strings.Contains(name, "/"):
			return NewCannotStartError("The template %s sets 0 of %s for its jobs: only an extended resource (e.g., nvidia.com/gpu) can be 0.",
				template, name)
		}
	}
	return nil
}

// CrownJobCheckJobResources checks the pod of a job against the resources of the queue limits of the given Workspace
// (queueLimits.resources), which are the most that one job may ask for. For each resource they list, neither the
// request of the pod (as PodEffectiveRequests counts it) nor the limit of its containers (a job pod has one: the app
// container) may exceed it. A resource they do not list has no bound for one job. A negative bound is refused.
func CrownJobCheckJobResources(workspace string, limits *clv1alpha1.QueueLimits, spec *corev1.PodSpec) error {
	if limits == nil {
		return nil
	}
	requests := PodEffectiveRequests(spec)
	containerLimits := corev1.ResourceList{}
	for i := range spec.Containers {
		containerLimits = addResourceList(containerLimits, spec.Containers[i].Resources.Limits)
	}

	for _, name := range sortedResourceNames(limits.Resources) {
		bound := limits.Resources[name]
		if bound.Sign() < 0 {
			return NewCannotStartError("The job queue of workspace %s sets a negative quantity of %s for one job (%s).", workspace, name, bound.String())
		}
		// What the job asks for: the larger of its request and its limit.
		asked := requests[corev1.ResourceName(name)]
		if limit := containerLimits[corev1.ResourceName(name)]; limit.Cmp(asked) > 0 {
			asked = limit
		}
		if asked.Cmp(bound) > 0 {
			return NewCannotStartError("The job asks for %s of %s, but one job of workspace %s may ask for at most %s.",
				asked.String(), name, workspace, bound.String())
		}
	}
	return nil
}

// CrownJobCheckQueueLimits checks the requests of a job against the job queue of the given Workspace: every requested
// resource must be covered by the queue (see KueueQueueQuotas), and no request may exceed maxAtOnce, or the job could
// never start. A zero request asks for nothing, and is ignored.
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
		if requested.IsZero() {
			continue
		}
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
// A zero request matches nothing.
func crownJobRuntimeClass(requests corev1.ResourceList, runtimeClasses map[string]string) (string, error) {
	matches := map[string]struct{}{}
	for name, quantity := range requests {
		if quantity.IsZero() {
			continue
		}
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

// dropZeroExtendedResources removes every extended resource (a name with a "/", e.g., nvidia.com/gpu) that is zero,
// or absent, in both the requests and the limits. Standard resources (cpu, memory, ...) are kept, even at zero.
func dropZeroExtendedResources(resources *corev1.ResourceRequirements) {
	for _, list := range []corev1.ResourceList{resources.Requests, resources.Limits} {
		for name := range list {
			request, limit := resources.Requests[name], resources.Limits[name]
			if strings.Contains(name.String(), "/") && request.IsZero() && limit.IsZero() {
				delete(resources.Requests, name)
				delete(resources.Limits, name)
			}
		}
	}
}

// setJobCommand sets the command of a job in its app container, whose disk is mounted at the given path. An empty
// command leaves the container as it is in the instance (command, args and working directory), so that the image runs
// as it does there. Any other command replaces the command of the container, clears its args, and runs in the folder
// of the disk. How a command is applied is decided here only: the alternative is to keep the entrypoint of the image
// and to set only its args.
func setJobCommand(app *corev1.Container, command []string, mountPath string) {
	if len(command) == 0 {
		return
	}
	app.Command = append([]string(nil), command...)
	app.Args = nil
	app.WorkingDir = mountPath
}

// setJobResources sets the given resources of a job in the resources of its app container: each one replaces both the
// request and the limit of the instance, so that the request equals the limit. The other resources are left untouched.
func setJobResources(requirements *corev1.ResourceRequirements, resources map[string]resource.Quantity) {
	if len(resources) == 0 {
		return
	}
	if requirements.Requests == nil {
		requirements.Requests = corev1.ResourceList{}
	}
	if requirements.Limits == nil {
		requirements.Limits = corev1.ResourceList{}
	}
	for name, quantity := range resources {
		requirements.Requests[corev1.ResourceName(name)] = quantity.DeepCopy()
		requirements.Limits[corev1.ResourceName(name)] = quantity.DeepCopy()
	}
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

// sortedResourceNames returns the names of the given resources, sorted, so that the first refusal does not depend on
// the order of a map.
func sortedResourceNames(resources map[string]resource.Quantity) []string {
	names := make([]string, 0, len(resources))
	for name := range resources {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
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
