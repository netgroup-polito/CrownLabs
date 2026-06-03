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

package instancesnapshot_controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// ValidateRequest validates the InstanceSnapshot request, returns an error and if there's the need to try again.
func (r *InstanceSnapshotReconciler) ValidateRequest(ctx context.Context, isnap *clv1alpha2.InstanceSnapshot) (bool, error) {
	// First it is needed to check if the instance actually exists.
	instanceName := forge.NamespacedNameFromGenericRef(isnap.Spec.Instance)
	instance := &clv1alpha2.Instance{}

	if err := r.Get(ctx, instanceName, instance); err != nil && kerrors.IsNotFound(err) {
		// The declared instance does not exist so don't try again.
		return false, fmt.Errorf("instance %s not found in namespace %s. It is not possible to complete the InstanceSnapshot %s",
			instanceName.Name, instanceName.Namespace, isnap.Name)
	} else if err != nil {
		return true, fmt.Errorf("error in retrieving the instance for InstanceSnapshot %s -> %w", isnap.Name, err)
	}

	// Get the template of the instance in order to check if it has the requirements to be snapshotted.
	// In order to create a snapshot of the vm, we need first to check that:
	// - the vm is powered off, since it is not possible to steal the DataVolume if it is still running;
	// - the environment is a persistent vm and not a container.

	templateName := forge.NamespacedNameFromGenericRef(instance.Spec.Template)
	template := &clv1alpha2.Template{}

	if err := r.Get(ctx, templateName, template); err != nil && kerrors.IsNotFound(err) {
		// The declared template does not exist set the phase as failed and don't try again.
		return false, fmt.Errorf("template %s not found in namespace %s. It is not possible to complete the InstanceSnapshot %s",
			templateName.Name, templateName.Namespace, isnap.Name)
	} else if err != nil {
		return true, fmt.Errorf("error in retrieving the template for InstanceSnapshot %s -> %w", isnap.Name, err)
	}

	// Retrieve the environment from the template.
	var env *clv1alpha2.Environment

	if isnap.Spec.Environment.Name != "" {
		for i := range template.Spec.EnvironmentList {
			if template.Spec.EnvironmentList[i].Name == isnap.Spec.Environment.Name {
				env = &template.Spec.EnvironmentList[i]
				break
			}
		}

		// Check if the specified environment was found.
		if env == nil {
			return false, fmt.Errorf("environment %s not found in template %s. It is not possible to complete the InstanceSnapshot %s",
				isnap.Spec.Environment.Name, template.Name, isnap.Name)
		}
	} else {
		// If the environment is not explicitly declared, take the first one.
		env = &template.Spec.EnvironmentList[0]
	}

	// Check if the environment is a persistent VM.
	if (env.EnvironmentType != clv1alpha2.ClassVM && env.EnvironmentType != clv1alpha2.ClassCloudVM) || !env.Persistent {
		return false, fmt.Errorf("environment %s is not a persistent VM. It is not possible to complete the InstanceSnapshot %s",
			env.Name, isnap.Name)
	}

	// Check if the VM is running.
	if instance.Spec.Running {
		return false, fmt.Errorf("the vm is running. It is not possible to complete the InstanceSnapshot %s", isnap.Name)
	}

	return false, nil
}

// GetJobStatus sets a Job and returns its status.
func (r *InstanceSnapshotReconciler) GetJobStatus(job *batchv1.Job) (bool, batchv1.JobConditionType) {
	for _, c := range job.Status.Conditions {
		// If the status corresponding to Success or failed is true, it means that the job completed.
		if c.Status == corev1.ConditionTrue && (c.Type == batchv1.JobFailed || c.Type == batchv1.JobComplete) {
			return true, c.Type
		}
	}

	// Job did not complete.
	return false, ""
}

// CreateSnapshottingJobDefinition generates the job to be created.
func (r *InstanceSnapshotReconciler) CreateSnapshottingJobDefinition(ctx context.Context, isnap *clv1alpha2.InstanceSnapshot) (batchv1.Job, error) {
	// Get the tenant name in order to set it as directory of the image
	instanceName := forge.NamespacedNameFromGenericRef(isnap.Spec.Instance)
	instance := &clv1alpha2.Instance{}

	if err := r.Get(ctx, instanceName, instance); err != nil {
		return batchv1.Job{}, fmt.Errorf("error in retrieving the instance for InstanceSnapshot %s -> %w", isnap.Name, err)
	}

	var backoff int32 = 2
	imagetag := time.Now().Format("20060102t150405")
	// Volume name does not accept dots, replace them with dashes
	volumename := strings.ReplaceAll(isnap.Spec.Instance.Name, ".", "-")
	imagedir := utils.ParseDockerDirectory(instance.Spec.Tenant.Name)

	// Define volumes.

	// Define VM VolumeSource.
	vmvolume := corev1.VolumeSource{
		PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
			ClaimName: volumename,
		},
	}

	// Define temp VolumeSource.
	tmpvol := corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{},
	}

	// Define secret VolumeSource.
	secretvol := corev1.VolumeSource{
		Secret: &corev1.SecretVolumeSource{
			SecretName: r.ContainersSnapshot.RegistrySecretName,
			Items: []corev1.KeyToPath{
				{
					Key:  ".dockerconfigjson",
					Path: "config.json",
				},
			},
		},
	}

	volumes := []corev1.Volume{
		{
			Name:         volumename,
			VolumeSource: vmvolume,
		},
		{
			Name:         "tmp-vol",
			VolumeSource: tmpvol,
		},
		{
			Name:         "kaniko-secret",
			VolumeSource: secretvol,
		},
	}

	// Define containers.

	// Define Docker pusher container.
	pushcontainer := corev1.Container{
		Name:  "docker-pusher",
		Image: r.ContainersSnapshot.ContainerKaniko,
		Args: []string{"--dockerfile=/workspace/Dockerfile",
			fmt.Sprintf("--destination=%s/%s/%s:%s", r.ContainersSnapshot.VMRegistry, imagedir, isnap.Spec.ImageName, imagetag)},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "tmp-vol",
				MountPath: "/workspace",
			},
			{
				Name:      "kaniko-secret",
				MountPath: "/kaniko/.docker/",
			},
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				"cpu":    resource.MustParse("1"),
				"memory": resource.MustParse("8Gi"),
			},
			Limits: corev1.ResourceList{
				"cpu":    resource.MustParse("1"),
				"memory": resource.MustParse("32Gi"),
			},
		},
	}

	// Define image exporter container.
	exportcontainer := corev1.Container{
		Name:  "img-generator",
		Image: r.ContainersSnapshot.ContainerImgExport,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      volumename,
				MountPath: "/data",
			},
			{
				Name:      "tmp-vol",
				MountPath: "/img-tmp",
			},
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				"cpu":    resource.MustParse("1"),
				"memory": resource.MustParse("128Mi"),
			},
			Limits: corev1.ResourceList{
				"cpu":    resource.MustParse("1"),
				"memory": resource.MustParse("256Mi"),
			},
		},
	}

	snapjob := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      isnap.Name,
			Namespace: isnap.Namespace,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backoff,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						pushcontainer,
					},
					InitContainers: []corev1.Container{
						exportcontainer,
					},
					Volumes:       volumes,
					RestartPolicy: corev1.RestartPolicyOnFailure,
				},
			},
		},
	}

	return snapjob, nil
}
