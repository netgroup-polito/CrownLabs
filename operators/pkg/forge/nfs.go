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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
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

	//XXX: Remove this
	// NFSSecretName -> NFS secret name for MyDrive.
	NFSSecretName = "mydrive-info"
	// NFSSecretServerNameKey -> NFS Server key in NFS secret.
	NFSSecretServerNameKey = "server-name"
	// NFSSecretPathKey -> NFS path key in NFS secret.
	NFSSecretPathKey = "path"
)

var (
	DefaultMirrorCapacity = resource.MustParse("1")
)

// NFSVolumeMountInfo contains information about a volume that has to be mounted through NFS.
type NFSVolumeMountInfo struct { //XXX: Remove this
	VolumeName    string
	ServerAddress string
	ExportPath    string
	MountPath     string
	ReadOnly      bool
}

// NFSVolumeMount forges the mount string array for a generic NFS volume. //XXX: Remove this
func NFSVolumeMount(nfsServer, exportPath, mountPath string, readOnly bool) []string {
	rwPermission := "rw"

	if readOnly {
		rwPermission = "ro"
	}

	return []string{
		fmt.Sprintf("%s:%s", nfsServer, exportPath),
		mountPath,
		"nfs",
		fmt.Sprintf("%s,tcp,hard,intr,rsize=8192,wsize=8192,timeo=14,_netdev,user", rwPermission),
		"0",
		"0",
	}
}

// MyDriveVolumeMount forges the mount string array for the MyDrive volume. //XXX: Remove this
func MyDriveVolumeMount(nfsServer, exportPath string) []string {
	return NFSVolumeMount(nfsServer, exportPath, MyDriveVolumeMountPath, false)
}

// SharedVolumeMount forges the mount string array for a SharedVolume. //XXX: Remove this
func SharedVolumeMount(shvol *clv1alpha2.SharedVolume, mountInfo clv1alpha2.SharedVolumeMountInfo) []string {
	// if shvol.Status.ServerAddress == "" || shvol.Status.ExportPath == "" {
	// 	return CommentMount("Here lies an invalid SharedVolume mount")
	// }

	// return NFSVolumeMount(shvol.Status.ServerAddress, shvol.Status.ExportPath, mountInfo.MountPath, mountInfo.ReadOnly)
	return CommentMount("deprecated")
}

// CommentMount forges the mount string array for a comment. //XXX: Remove this
func CommentMount(comment string) []string {
	return []string{
		"# " + comment,
		"",
		"",
		"",
		"",
		"",
	}
}

// MyDriveNFSVolumeMountInfo forges the NFSVolumeMountInfo for the MyDrive volume. //XXX: Remove this
func MyDriveNFSVolumeMountInfo(serverAddress, exportPath string) NFSVolumeMountInfo {
	return NFSVolumeMountInfo{
		VolumeName:    MyDriveVolumeName,
		ServerAddress: serverAddress,
		ExportPath:    exportPath,
		MountPath:     MyDriveVolumeMountPath,
		ReadOnly:      false,
	}
}

// ShVolNFSVolumeMountInfo forges the NFSVolumeMountInfo given a SharedVolume and SharedVolumeMountInfo, its name will be nfs{i}. //XXX: Remove this
func ShVolNFSVolumeMountInfo(i int, shvol *clv1alpha2.SharedVolume, mount clv1alpha2.SharedVolumeMountInfo) NFSVolumeMountInfo {
	return NFSVolumeMountInfo{
		VolumeName:    fmt.Sprintf("nfs%d", i),
		ServerAddress: "", // shvol.Status.ServerAddress,
		ExportPath:    "", // shvol.Status.ExportPath,
		MountPath:     mount.MountPath,
		ReadOnly:      mount.ReadOnly,
	}
}

// NFSShVolSpec obtains the NFS server address and the export path from the passed Persistent Volume. //XXX: Remove this
func NFSShVolSpec(pv *corev1.PersistentVolume) (serverAddress, exportPath string) {
	serverAddress = ""
	exportPath = ""

	if pv.Spec.CSI != nil && pv.Spec.CSI.VolumeAttributes != nil {
		serverAddress = fmt.Sprintf("%s.%s", pv.Spec.CSI.VolumeAttributes["server"], pv.Spec.CSI.VolumeAttributes["clusterID"])
		exportPath = pv.Spec.CSI.VolumeAttributes["share"]
	}

	return
}

// GetNFSSpecs extracts the NFS server name and path for the tenant's personal NFS volume,
// required to mount the MyDrive disk of a given tenant from the associated secret. //XXX: Remove this
func GetNFSSpecs(ctx context.Context, c client.Client) (nfsServerName, nfsPath string, err error) {
	var serverNameBytes, serverPathBytes []byte
	instance := clctx.InstanceFrom(ctx)
	secretName := types.NamespacedName{Namespace: instance.Namespace, Name: NFSSecretName}

	secret := corev1.Secret{}
	if err = c.Get(ctx, secretName, &secret); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "failed to retrieve secret", "secret", secretName)
		return
	}

	serverNameBytes, ok := secret.Data[NFSSecretServerNameKey]
	if !ok {
		err = fmt.Errorf("cannot find %v key in secret", NFSSecretServerNameKey)
		ctrl.LoggerFrom(ctx).Error(err, "failed to retrieve NFS spec from secret", "secret", secretName)
		return
	}

	serverPathBytes, ok = secret.Data[NFSSecretPathKey]
	if !ok {
		err = fmt.Errorf("cannot find %v key in secret", NFSSecretPathKey)
		ctrl.LoggerFrom(ctx).Error(err, "failed to retrieve NFS spec from secret", "secret", secretName)
		return
	}

	return string(serverNameBytes), string(serverPathBytes), nil
}

// NFSVolumeMountInfosFromEnvironment extracts the array of NFSVolumeMountInfo from the passed environment
// adding the MyDrive volume if needed, and setting RW permissions in case the Tenant is manager of the Workspace.
// In case of error, the first value returned is nil, followed by error reason (string) and error. //XXX: Remove this
func NFSVolumeMountInfosFromEnvironment(ctx context.Context, c client.Client, env *clv1alpha2.Environment) ([]NFSVolumeMountInfo, string, error) {
	mountInfos := []NFSVolumeMountInfo{}

	// Check and mount MyDrive
	if env.MountMyDriveVolume {
		nfsServerName, nfsPath, err := GetNFSSpecs(ctx, c)
		if err != nil {
			return nil, "unable to retrieve NFS specs", err
		}

		mountInfos = append(mountInfos, MyDriveNFSVolumeMountInfo(nfsServerName, nfsPath))
	}

	// Check if tenant is manager of workspace
	tenant := clctx.TenantFrom(ctx)
	template := clctx.TemplateFrom(ctx)
	workspaceName := template.Spec.WorkspaceRef.Name
	labelSelector := map[string]string{clv1alpha2.WorkspaceLabelPrefix + workspaceName: string(clv1alpha2.Manager)}

	var managers clv1alpha2.TenantList
	if err := c.List(ctx, &managers, client.MatchingLabels(labelSelector)); err != nil {
		return nil, "failed to retrieve managers for workspace " + workspaceName, err
	}

	isManager := false
	for i := range managers.Items {
		if managers.Items[i].Name == tenant.Name {
			isManager = true
			break
		}
	}

	// Check and mount SharedVolumes
	for i, mount := range env.SharedVolumeMounts {
		var shvol clv1alpha2.SharedVolume
		if err := c.Get(ctx, NamespacedNameFromMount(mount), &shvol); err != nil {
			return nil, "unable to retrieve shvol to mount", err
		}

		if isManager {
			mount.ReadOnly = false
		}

		mountInfos = append(mountInfos, ShVolNFSVolumeMountInfo(i, &shvol, mount))
	}

	return mountInfos, "", nil
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
		// Check existance before mounting
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
	// The maximum name length for a Kubernetes resource is 253 characters, (253 - len("--mirror") - 1) = 122.
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

// ConfigureMyDriveSecret configures a Secret for tenant's MyDrive access. //XXX: Remove this
func ConfigureMyDriveSecret(secret *corev1.Secret, serverName, path string, labels map[string]string) {
	// Set the labels
	if secret.Labels == nil {
		secret.Labels = make(map[string]string)
	}

	// Copy the provided labels
	maps.Copy(secret.Labels, labels)

	// Configure the Secret data
	secret.Type = corev1.SecretTypeOpaque
	secret.Data = make(map[string][]byte, 2)
	secret.Data[NFSSecretServerNameKey] = []byte(serverName)
	secret.Data[NFSSecretPathKey] = []byte(path)
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
			Kind:      origin.Kind,
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
