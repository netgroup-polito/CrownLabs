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

// Package forge groups the methods used to forge the Kubernetes object definitions
// required by the different controllers.
package forge

//TODO: Rinominare nfs.go in storage.go o simili

import (
	"context"
	"fmt"
	"maps"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
)

const (
	// ProvisionJobBaseImage -> Base container image for Drive provision job.
	ProvisionJobBaseImage = "busybox"
	// ProvisionJobMaxRetries -> Maximum number of retries for Provision jobs.
	ProvisionJobMaxRetries = 3
	// ProvisionJobTTLSeconds -> Seconds for Provision jobs before deletion (either failure or success).
	ProvisionJobTTLSeconds = 3600 * 24 * 7
)

var (
	// DefaultMirrorCapacity is the default size for mirror PVs and PVCs (which does not match the real size).
	DefaultMirrorCapacity = resource.MustParse("1")
)

// NFSShVolSpec obtains the NFS server address and the export path from the passed Persistent Volume. //XXX: Remove this.
func NFSShVolSpec(pv *corev1.PersistentVolume) (serverAddress, exportPath string) {
	serverAddress = ""
	exportPath = ""

	if pv.Spec.CSI != nil && pv.Spec.CSI.VolumeAttributes != nil {
		serverAddress = fmt.Sprintf("%s.%s", pv.Spec.CSI.VolumeAttributes["server"], pv.Spec.CSI.VolumeAttributes["clusterID"])
		exportPath = pv.Spec.CSI.VolumeAttributes["share"]
	}

	return
}

// MyDriveMountInfo forges the VolumeMount for the MyDrive volume.
func MyDriveMountInfo(tnName string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      GetMyDrivePVCMirrorName(tnName),
		MountPath: MyDriveVolumeMountPath,
		ReadOnly:  false,
	}
}

// ShVolMountInfo forges the VolumeMount given the SharedVolumeMountInfo, its name will be shvol{i}.
func ShVolMountInfo(mount clv1alpha2.SharedVolumeMountInfo, instanceName string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      GetShVolPVCMirrorName(mount.SharedVolumeRef.Name, instanceName),
		MountPath: mount.MountPath,
		ReadOnly:  mount.ReadOnly,
	}
}

// PVCMountInfosFromEnvironment extracts the array of VolumeMount from the environment in the ctx
// adding the MyDrive volume if needed, and setting RW permissions in case the Tenant is manager of the Workspace.
// While calculating the array, it checks if the wanted SharedVolume exists.
// In case of error, the first value returned is nil, followed by error reason (string) and error.
func PVCMountInfosFromEnvironment(ctx context.Context, c client.Client) ([]corev1.VolumeMount, string, error) {
	tenant := clctx.TenantFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	template := clctx.TemplateFrom(ctx)
	env := clctx.EnvironmentFrom(ctx)

	mountInfos := []corev1.VolumeMount{}

	// Check and mount MyDrive if needed
	if env.MountMyDriveVolume {
		mountInfos = append(mountInfos, MyDriveMountInfo(tenant.Name))
	}

	// Check if tenant is manager of workspace
	isManager := false
	workspaceLabelKey := fmt.Sprintf("%s%s", clv1alpha2.WorkspaceLabelPrefix, template.Spec.WorkspaceRef.Name)
	if val, found := tenant.Labels[workspaceLabelKey]; found && val == string(clv1alpha2.Manager) {
		isManager = true
	}

	// Check and mount SharedVolumes
	for _, mount := range env.SharedVolumeMounts {
		// Check existence before mounting
		var shvol clv1alpha2.SharedVolume
		if err := c.Get(ctx, NamespacedNameFromMount(mount), &shvol); err != nil {
			return nil, "unable to retrieve shvol to mount", err
		}

		if isManager {
			mount.ReadOnly = false
		}
		mountInfos = append(mountInfos, ShVolMountInfo(mount, instance.Name))
	}

	return mountInfos, "", nil
}

// PVCProvisioningJobSpec forges the spec for the PVC Provisioning job.
func PVCProvisioningJobSpec(pvc *corev1.PersistentVolumeClaim) batchv1.JobSpec {
	return batchv1.JobSpec{
		BackoffLimit:            ptr.To[int32](ProvisionJobMaxRetries),
		TTLSecondsAfterFinished: ptr.To[int32](ProvisionJobTTLSeconds),
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyOnFailure,
				Containers: []corev1.Container{{
					Name:    "chown-container",
					Image:   ProvisionJobBaseImage,
					Command: []string{"chown", "-R", fmt.Sprintf("%d:%d", CrownLabsUserID, CrownLabsUserID), MyDriveVolumeMountPath},
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "drive",
						MountPath: MyDriveVolumeMountPath,
					}},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"cpu":    resource.MustParse("100m"),
							"memory": resource.MustParse("128Mi"),
						},
						Limits: corev1.ResourceList{
							"cpu":    resource.MustParse("100m"),
							"memory": resource.MustParse("128Mi"),
						},
					},
				}},
				Volumes: []corev1.Volume{{
					Name: "drive",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvc.Name,
						},
					},
				}},
			},
		},
	}
}

// GetMyDrivePVCName returns the name for a tenant's MyDrive PVC.
func GetMyDrivePVCName(tenantName string) string {
	return fmt.Sprintf("%s-drive", strings.ReplaceAll(tenantName, ".", "-"))
}

// GetMyDrivePVCMirrorName returns the name for the mirror of the tenant's MyDrive PVC.
func GetMyDrivePVCMirrorName(tenantName string) string {
	return fmt.Sprintf("%s-mirror", GetMyDrivePVCName(tenantName))
}

// GetShVolPVCMirrorName returns the name for the mirror of the SharedVolume PVC for the specified Instance.
func GetShVolPVCMirrorName(shvolName, instanceName string) string {
	// The maximum name length for a Kubernetes resource is 253 characters, (253 - len("--mirror") - 1)/2 = 122.
	return fmt.Sprintf("%s-%s-mirror", LastCharsOf(shvolName, 122), LastCharsOf(instanceName, 122))
}

// ConfigureMyDrivePVC configures a PVC for tenant's MyDrive storage.
func ConfigureMyDrivePVC(pvc *corev1.PersistentVolumeClaim, storageClassName string, storageSize resource.Quantity, labels, annotations map[string]string) {
	// Set the labels and annotations
	if pvc.Labels == nil {
		pvc.Labels = make(map[string]string)
	}
	if pvc.Annotations == nil {
		pvc.Annotations = make(map[string]string)
	}

	// Copy the provided labels and annotations
	maps.Copy(pvc.Labels, labels)
	maps.Copy(pvc.Annotations, annotations)

	// Configure the PVC spec
	pvc.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}
	pvc.Spec.StorageClassName = &storageClassName

	// Set resources if the current ones are missing or smaller than the requested ones
	if pvc.Spec.Resources.Requests == nil {
		pvc.Spec.Resources.Requests = corev1.ResourceList{corev1.ResourceStorage: storageSize}
	} else if oldSize := *pvc.Spec.Resources.Requests.Storage(); storageSize.Cmp(oldSize) > 0 || oldSize.IsZero() {
		pvc.Spec.Resources.Requests = corev1.ResourceList{corev1.ResourceStorage: storageSize}
	}
}

// UpdatePVCProvisioningJobLabel updates the provisioning job label for a PVC.
func UpdatePVCProvisioningJobLabel(pvc *corev1.PersistentVolumeClaim, value string) {
	if pvc.Labels == nil {
		pvc.Labels = make(map[string]string)
	}
	pvc.Labels[ProvisionJobLabel] = value
}

// ConfigureMyDriveProvisioningJob configures a Job for provisioning a PVC.
func ConfigureMyDriveProvisioningJob(job *batchv1.Job, pvc *corev1.PersistentVolumeClaim) {
	// Set job spec using PVCProvisioningJobSpec
	jobSpec := PVCProvisioningJobSpec(pvc)
	job.Spec = jobSpec
}

// MirrorPVCSpec forges the spec for the PVC that will mirror the "origin" one.
func MirrorPVCSpec(origin *corev1.PersistentVolumeClaim, mirrorStorageClassName string) corev1.PersistentVolumeClaimSpec {
	return corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
		Resources: corev1.VolumeResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: DefaultMirrorCapacity,
			},
		},
		StorageClassName: &mirrorStorageClassName,
		DataSourceRef: &corev1.TypedObjectReference{
			APIGroup:  nil, // Kind PVC is in the core API group, which is nil
			Kind:      "PersistentVolumeClaim",
			Namespace: &origin.Namespace,
			Name:      origin.Name,
		},
	}
}

// UpdateMyDriveMirrorPVCLabels updates the labels for a Mirror PVC of a MyDrive.
func UpdateMyDriveMirrorPVCLabels(labels map[string]string, targetLabel common.KVLabel) map[string]string {
	labels = UpdateTenantResourceCommonLabels(labels, targetLabel)
	labels[LabelVolumeTypeKey] = VolumeTypeValueMirror

	return labels
}

// UpdateShVolMirrorPVCLabels updates the labels for a Mirror PVC of a SharedVolume.
func UpdateShVolMirrorPVCLabels(labels map[string]string) map[string]string {
	if labels == nil {
		labels = make(map[string]string, 1)
	}
	labels[LabelManagedByKey] = labelManagedByInstanceValue
	labels[LabelVolumeTypeKey] = VolumeTypeValueMirror

	return labels
}
