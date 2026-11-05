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
	"errors"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var (
	// errInstanceNotReady is returned when the Deployment of the Instance, or its disk, does not exist yet.
	errInstanceNotReady = errors.New("the deployment of the instance does not exist yet")
	// errDiskInUse is returned when the disk of the Instance is not ReadWriteMany and its container is running, or
	// still stopping.
	errDiskInUse = errors.New("the disk of the instance is in use")
	// errInstanceStopping is returned when the container of the Instance has been stopped, but one of its pods still
	// exists: the disk is still in use, for a short time.
	errInstanceStopping = fmt.Errorf("%w: the container of the instance is stopping", errDiskInUse)
	// errQueueNotReady is returned when the ClusterQueue of the Workspace does not exist yet.
	errQueueNotReady = errors.New("the job queue of the workspace does not exist yet")
)

// jobInputs groups what the Job of a CrownJob is forged from, once the CrownJob has been validated.
type jobInputs struct {
	instance  *clv1alpha2.Instance
	workspace *clv1alpha1.Workspace
	opts      forge.CrownJobJobOpts
}

// gatherJobInputs reads the objects a CrownJob refers to, and checks the rules a job must respect before its Job
// is created. It returns a forge.CannotStartError if the job can never start; errInstanceNotReady, errDiskInUse (or
// errInstanceStopping, which wraps it) or errQueueNotReady if it has to wait (with the Instance and, for the queue, the Workspace in the inputs); and any
// other error if an object could not be read.
// The objects are read with the API reader (without cache), since they are needed only once per CrownJob.
func (r *Reconciler) gatherJobInputs(ctx context.Context, crownJob *clv1alpha2.CrownJob) (*jobInputs, error) {
	cannotStart := forge.NewCannotStartError

	// 1. The Instance must be in the namespace of the CrownJob: a tenant must not copy the disk of another tenant.
	if ns := crownJob.Spec.Instance.Namespace; ns != "" && ns != crownJob.Namespace {
		return nil, cannotStart("The job must use an instance in its own namespace (%s), not in %s.", crownJob.Namespace, ns)
	}

	// 2. The Job has the same name as the CrownJob, and Kubernetes copies it into a label value.
	if len(crownJob.Name) > forge.CrownJobMaxNameLength {
		return nil, cannotStart("The job name is too long: it can have at most %d characters.", forge.CrownJobMaxNameLength)
	}

	// 3. The Instance, its Template with a job section, and the job environment.
	var instance clv1alpha2.Instance
	if found, err := r.read(ctx, types.NamespacedName{Namespace: crownJob.Namespace, Name: crownJob.Spec.Instance.Name}, &instance); err != nil {
		return nil, err
	} else if !found {
		return nil, cannotStart(msgInstanceGone, crownJob.Spec.Instance.Name)
	}

	var template clv1alpha2.Template
	templateName := forge.NamespacedNameFromGenericRef(instance.Spec.Template)
	if found, err := r.read(ctx, templateName, &template); err != nil {
		return nil, err
	} else if !found {
		return nil, cannotStart("The template %s of the instance %s does not exist.", templateName.Name, instance.Name)
	}
	templateJob := template.Spec.Job
	if templateJob == nil {
		return nil, cannotStart("Jobs cannot be started from the template %s.", template.Name)
	}

	environment, err := jobEnvironment(&template)
	if err != nil {
		return nil, err
	}

	// 4. The Workspace, with a job queue.
	workspaceName := template.Spec.WorkspaceRef.Name
	if workspaceName == "" {
		return nil, cannotStart("The template %s does not belong to a workspace.", template.Name)
	}
	if msgs := validation.IsValidLabelValue(forge.CrownJobQueueName(workspaceName)); len(msgs) > 0 {
		return nil, cannotStart("The name of the workspace %s is too long for a job queue.", workspaceName)
	}
	var workspace clv1alpha1.Workspace
	if found, err := r.read(ctx, types.NamespacedName{Name: workspaceName}, &workspace); err != nil {
		return nil, err
	} else if !found {
		return nil, cannotStart("The workspace %s does not exist.", workspaceName)
	}
	if workspace.Spec.QueueLimits == nil {
		return nil, cannotStart("The workspace %s has no job queue.", workspaceName)
	}

	// 5. The Tenant must be a member (user or manager) of the Workspace.
	membership, err := r.workspaceMembership(ctx, instance.Spec.Tenant.Name, workspaceName)
	if err != nil {
		return nil, err
	}

	// 6. The command, by commandPolicy.
	command, err := forge.CrownJobCommand(templateJob, crownJob.Spec.Command)
	if err != nil {
		return nil, err
	}

	// 7. The maximum run time.
	maxRunTime, err := forge.CrownJobMaxRunTime(crownJob.Spec.MaxRunTime, templateJob.MaxRunTime, workspace.Spec.QueueLimits.MaxRunTime, r.DefaultMaxRunTime)
	if err != nil {
		return nil, err
	}

	// 9. The priority class: that of the Tenant in the Workspace, or else that of the Workspace. It must exist.
	priorityClass, err := r.jobPriorityClass(ctx, membership, &workspace)
	if err != nil {
		return nil, err
	}

	// The Instance must not be being deleted: its finalizer would keep it, and its disk, for the whole run of the job.
	if !instance.DeletionTimestamp.IsZero() {
		return nil, cannotStart(msgInstanceDeleting, instance.Name)
	}

	// 10. The Deployment of the Instance, whose pod template the Job copies. It may not exist yet (not terminal).
	var deployment appsv1.Deployment
	if found, err := r.read(ctx, forge.NamespacedNameWithSuffix(&instance, environment.Name), &deployment); err != nil {
		return nil, err
	} else if !found {
		return &jobInputs{instance: &instance}, errInstanceNotReady
	}

	// The disk: one that is not ReadWriteMany can be used by one node at a time, so the job waits while the container
	// of the instance runs, or still stops. It is checked only before the Job exists: an instance started later waits
	// for the disk.
	if err := r.checkDisk(ctx, &deployment); errors.Is(err, errInstanceNotReady) || errors.Is(err, errDiskInUse) {
		return &jobInputs{instance: &instance}, err
	} else if err != nil {
		return nil, err
	}

	// The job queue of the Workspace: without its ClusterQueue, Kueue would keep the Job waiting with no explanation.
	clusterQueue := &unstructured.Unstructured{}
	clusterQueue.SetGroupVersionKind(forge.KueueClusterQueueGVK)
	if found, err := r.read(ctx, types.NamespacedName{Name: forge.CrownJobQueueName(workspaceName)}, clusterQueue); err != nil {
		return nil, err
	} else if !found {
		return &jobInputs{instance: &instance, workspace: &workspace}, errQueueNotReady
	}

	// Rule 8 (resources) is checked on the forged Job, by the caller: it needs the pod template of the Deployment.
	return &jobInputs{
		instance:  &instance,
		workspace: &workspace,
		opts: forge.CrownJobJobOpts{
			Workspace:       workspaceName,
			EnvironmentName: environment.Name,
			PodTemplate:     &deployment.Spec.Template,
			CommandPolicy:   templateJob.CommandPolicy,
			Command:         command,
			OtherResources:  templateJob.OtherResources,
			MaxRunTime:      maxRunTime,
			PriorityClass:   priorityClass,
			RuntimeClasses:  r.RuntimeClasses,
			FinishedJobTTL:  r.FinishedJobTTL,
		},
	}, nil
}

// workspaceMembership returns the entry of the given Tenant for the given Workspace, which must have the role
// user or manager (not candidate).
func (r *Reconciler) workspaceMembership(ctx context.Context, tenantName, workspaceName string) (*clv1alpha2.TenantWorkspaceEntry, error) {
	var tenant clv1alpha2.Tenant
	if found, err := r.read(ctx, types.NamespacedName{Name: tenantName}, &tenant); err != nil {
		return nil, err
	} else if !found {
		return nil, forge.NewCannotStartError("The tenant %s does not exist.", tenantName)
	}
	for i := range tenant.Spec.Workspaces {
		entry := &tenant.Spec.Workspaces[i]
		if entry.Name == workspaceName && (entry.Role == clv1alpha2.User || entry.Role == clv1alpha2.Manager) {
			return entry, nil
		}
	}
	return nil, forge.NewCannotStartError("The tenant %s is not a member of the workspace %s.", tenantName, workspaceName)
}

// jobPriorityClass returns the priority class of a job: the jobPriorityClass of the Tenant in the Workspace, or else
// the priorityClass of the jobPolicy of the Workspace (empty if neither is set). The Kueue WorkloadPriorityClass must exist.
func (r *Reconciler) jobPriorityClass(ctx context.Context, membership *clv1alpha2.TenantWorkspaceEntry, workspace *clv1alpha1.Workspace) (string, error) {
	priorityClass := membership.JobPriorityClass
	if priorityClass == "" && workspace.Spec.JobPolicy != nil {
		priorityClass = workspace.Spec.JobPolicy.PriorityClass
	}
	if priorityClass == "" {
		return "", nil
	}

	wpc := &unstructured.Unstructured{}
	wpc.SetGroupVersionKind(forge.KueueWorkloadPriorityClassGVK)
	if found, err := r.read(ctx, types.NamespacedName{Name: priorityClass}, wpc); err != nil {
		return "", err
	} else if !found {
		return "", forge.NewCannotStartError("The job priority class %s does not exist.", priorityClass)
	}
	return priorityClass, nil
}

// jobEnvironment returns the environment of the Template named by its job section, which must be persistent
// and of type Container or Standalone.
func jobEnvironment(template *clv1alpha2.Template) (*clv1alpha2.Environment, error) {
	name := template.Spec.Job.EnvironmentName
	for i := range template.Spec.EnvironmentList {
		environment := &template.Spec.EnvironmentList[i]
		if environment.Name != name {
			continue
		}
		if !environment.Persistent {
			return nil, forge.NewCannotStartError("Jobs need a persistent environment, and the environment %s of the template %s is not.", name, template.Name)
		}
		if environment.EnvironmentType != clv1alpha2.ClassContainer && environment.EnvironmentType != clv1alpha2.ClassStandalone {
			return nil, forge.NewCannotStartError("Jobs need an environment of type Container or Standalone, and the environment %s of the template %s is of type %s.",
				name, template.Name, environment.EnvironmentType)
		}
		return environment, nil
	}
	return nil, forge.NewCannotStartError("The template %s has no environment named %s.", template.Name, name)
}

// persistentClaimName returns the name of the claim of the persistent volume of the given pod spec (empty if none).
func persistentClaimName(spec *corev1.PodSpec) string {
	for i := range spec.Volumes {
		if spec.Volumes[i].Name == forge.PersistentVolumeName && spec.Volumes[i].PersistentVolumeClaim != nil {
			return spec.Volumes[i].PersistentVolumeClaim.ClaimName
		}
	}
	return ""
}

// checkDisk returns errInstanceNotReady if the disk of the instance of the given Deployment does not exist yet, and
// errDiskInUse (or errInstanceStopping, which wraps it) if the disk is not ReadWriteMany and the container of the
// instance runs, or still stops.
func (r *Reconciler) checkDisk(ctx context.Context, deployment *appsv1.Deployment) error {
	claimName := persistentClaimName(&deployment.Spec.Template.Spec)
	if claimName == "" {
		return nil
	}
	var pvc corev1.PersistentVolumeClaim
	if found, err := r.read(ctx, types.NamespacedName{Namespace: deployment.Namespace, Name: claimName}, &pvc); err != nil {
		return err
	} else if !found {
		return errInstanceNotReady
	} else if forge.IsReadWriteMany(pvc.Spec.AccessModes) {
		return nil
	}

	switch state, err := r.instanceContainer(ctx, deployment); {
	case err != nil:
		return err
	case state == containerRunning:
		return errDiskInUse
	case state == containerStopping:
		return errInstanceStopping
	}
	return nil
}

// containerState tells whether the container of an instance uses its disk.
type containerState int

const (
	// containerStopped -> the Deployment wants no replica, and each pod it selects is gone, or has ended without being deleted.
	containerStopped containerState = iota
	// containerRunning -> the Deployment wants a replica: the instance is running, or starting.
	containerRunning
	// containerStopping -> the Deployment wants no replica, but one of its pods is being deleted, or has not ended yet.
	containerStopping
)

// instanceContainer returns the state of the container of the instance of the given Deployment. A pod being deleted
// still uses the disk, but status.replicas does not count it, and status.terminatingReplicas is set only where the
// feature gate DeploymentReplicaSetTerminatingReplicas is enabled: when the counters say "stopped", the pods selected
// by the Deployment are listed with the API reader, and one that is being deleted, or has neither succeeded nor
// failed, keeps the disk in use. A pod being deleted counts until it is gone, whatever its phase: the kubelet sets it
// to Failed when its container stops, and deletes it once its volumes are unmounted.
func (r *Reconciler) instanceContainer(ctx context.Context, deployment *appsv1.Deployment) (containerState, error) {
	if ptr.Deref(deployment.Spec.Replicas, 1) > 0 {
		return containerRunning, nil
	}
	if deployment.Status.Replicas > 0 || ptr.Deref(deployment.Status.TerminatingReplicas, 0) > 0 {
		return containerStopping, nil
	}

	selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return containerStopped, fmt.Errorf("invalid selector of deployment %s/%s: %w", deployment.Namespace, deployment.Name, err)
	}
	var pods corev1.PodList
	if err := r.APIReader.List(ctx, &pods, client.InNamespace(deployment.Namespace), client.MatchingLabelsSelector{Selector: selector}); err != nil {
		return containerStopped, fmt.Errorf("failed retrieving the pods of deployment %s/%s: %w", deployment.Namespace, deployment.Name, err)
	}
	for i := range pods.Items {
		pod := &pods.Items[i]
		if phase := pod.Status.Phase; !pod.DeletionTimestamp.IsZero() || (phase != corev1.PodSucceeded && phase != corev1.PodFailed) {
			return containerStopping, nil
		}
	}
	return containerStopped, nil
}

// read gets the given object with the API reader, and returns whether it exists.
func (r *Reconciler) read(ctx context.Context, key types.NamespacedName, obj client.Object) (bool, error) {
	if err := r.APIReader.Get(ctx, key, obj); err != nil {
		if kerrors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed retrieving %T %s: %w", obj, key, err)
	}
	return true, nil
}
