// Copyright 2020-2025 Politecnico di Torino
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

import (
	"fmt"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const (
	// WebsockifyName -> name of the websockify sidecar container.
	WebsockifyName = "websockify"
	// XVncName -> name of the x+vnc server sidecar container.
	XVncName = "xvnc"
	// PersistentVolumeName -> name of the persistent volume.
	PersistentVolumeName = "persistent"
	// ContentDownloaderName -> name of the downloader initcontainer.
	ContentDownloaderName = "content-downloader"
	// ContentUploaderName -> name of the uploader initcontainer.
	ContentUploaderName = "content-uploader"
	// PersistentDefaultMountPath -> default path for the container's pvc or persistent storage.
	PersistentDefaultMountPath = "/media/data"
	// HealthzEndpoint -> default endpoint for HTTP probes.
	HealthzEndpoint = "/healthz"
	// CrownLabsUserID -> used as UID and GID for containers security context.
	CrownLabsUserID = int64(1010)
	// SubmissionJobMaxRetries -> max number of retries for submission jobs.
	SubmissionJobMaxRetries = 10
	// SubmissionJobTTLSeconds -> seconds for submission jobs before deletion (either failure or success).
	SubmissionJobTTLSeconds = 300
	// AppCPULimitsEnvName -> name of the env variable containing AppContainer CPU limits.
	AppCPULimitsEnvName = "APP_CPU_LIMITS"
	// AppMEMLimitsEnvName -> name of the env variable containing AppContainer memory limits.
	AppMEMLimitsEnvName = "APP_MEM_LIMITS"
	// PodNameEnvName -> name of the env variable containing the Pod Name.
	PodNameEnvName = "POD_NAME"
	// MyDriveVolumeName -> Name of the NFS volume.
	MyDriveVolumeName = "mydrive"
	// MyDriveVolumeMountPath -> Mount path for the NFS personal volume.
	MyDriveVolumeMountPath = "/media/mydrive"

	containersTerminationGracePeriod = 10
)

var (
	// DefaultDivisor -> "0".
	DefaultDivisor = *resource.NewQuantity(0, "")
	// MilliDivisor -> "1m".
	MilliDivisor = *resource.NewMilliQuantity(1, resource.DecimalSI)
)

// ContainerEnvOpts contains images name and tag for container environment.
type ContainerEnvOpts struct {
	ImagesTag            string
	XVncImg              string
	WebsockifyImg        string
	ContentDownloaderImg string
	ContentUploaderImg   string
	InstMetricsEndpoint  string
}

// PVCSpec forges a PersistentVolumeClaimSpec with the passed arguments.
// The params storageClass and size will be ignored if equal to nil.
func PVCSpec(accessMode corev1.PersistentVolumeAccessMode, storageClass *string, size *resource.Quantity) corev1.PersistentVolumeClaimSpec {
	spec := corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			accessMode,
		},
	}

	if storageClass != nil {
		spec.StorageClassName = storageClass
	}

	if size != nil {
		spec.Resources.Requests = corev1.ResourceList{
			corev1.ResourceStorage: *size,
		}
	}

	return spec
}

// InstancePVCSpec forges a ReadWriteOnce PersistentVolumeClaimSpec
// with requests set as in environment.Resources.Disk.
func InstancePVCSpec(environment *clv1alpha2.Environment) corev1.PersistentVolumeClaimSpec {
	return PVCSpec(corev1.ReadWriteOnce, InstancePVCStorageClassName(environment), &environment.Resources.Disk)
}

// InstancePVCStorageClassName returns the storage class configured as option, or nil if empty.
func InstancePVCStorageClassName(environment *clv1alpha2.Environment) *string {
	if environment.StorageClassName != "" {
		return ptr.To(environment.StorageClassName)
	}
	return nil
}

// SharedVolumePVCSpec forges a ReadWriteMany PersistentVolumeClaimSpec.
func SharedVolumePVCSpec(storageClass *string) corev1.PersistentVolumeClaimSpec {
	return PVCSpec(corev1.ReadWriteMany, storageClass, nil)
}

// PodSecurityContext forges a PodSecurityContext
// with 1010 UID and GID and RunAsNonRoot set.
func PodSecurityContext() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{
		RunAsUser:    ptr.To(CrownLabsUserID),
		RunAsGroup:   ptr.To(CrownLabsUserID),
		FSGroup:      ptr.To(CrownLabsUserID),
		RunAsNonRoot: ptr.To(true),
	}
}

// ReplicasCount returns an int32 pointer to 1 if the instance is not persistent
// or, if persistent, in case environment spec is set as running; 0 otherwise.
func ReplicasCount(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, isNew bool) *int32 {
	if (!isNew && !environment.Persistent) || instance.Spec.Running {
		return ptr.To[int32](1)
	}
	return ptr.To[int32](0)
}

// DeploymentSpec forges the complete DeploymentSpec (without replicas)
// containing the needed sidecars for X-VNC based container instances.
func DeploymentSpec(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, mountInfos []NFSVolumeMountInfo, opts *ContainerEnvOpts) appsv1.DeploymentSpec {
	return appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{MatchLabels: InstanceSelectorLabels(instance)},
		Strategy: appsv1.DeploymentStrategy{
			Type: appsv1.RecreateDeploymentStrategyType,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: InstanceSelectorLabels(instance)},
			Spec:       PodSpec(instance, environment, mountInfos, opts),
		},
	}
}

// PodSpec forges the pod specification for X-VNC based container instance.
func PodSpec(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, mountInfos []NFSVolumeMountInfo, opts *ContainerEnvOpts) corev1.PodSpec {
	return corev1.PodSpec{
		Containers:                    ContainersSpec(instance, environment, mountInfos, opts),
		Volumes:                       ContainerVolumes(instance, environment, mountInfos),
		SecurityContext:               PodSecurityContext(),
		AutomountServiceAccountToken:  ptr.To(false),
		TerminationGracePeriodSeconds: ptr.To[int64](containersTerminationGracePeriod),
		InitContainers:                InitContainers(instance, environment, opts),
		EnableServiceLinks:            ptr.To(false),
		Hostname:                      InstanceHostname(environment),
		NodeSelector:                  NodeSelectorLabels(instance, environment),
	}
}

// SubmissionJobSpec returns the job spec for the submission job.
func SubmissionJobSpec(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, opts *ContainerEnvOpts) batchv1.JobSpec {
	return batchv1.JobSpec{
		BackoffLimit:            ptr.To[int32](SubmissionJobMaxRetries),
		TTLSecondsAfterFinished: ptr.To[int32](SubmissionJobTTLSeconds),
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					ContentUploaderJobContainer(instance.Spec.CustomizationUrls.ContentDestination, instance.Name, opts),
				},
				Volumes:                      ContainerVolumes(instance, environment, nil),
				SecurityContext:              PodSecurityContext(),
				AutomountServiceAccountToken: ptr.To(false),
				RestartPolicy:                corev1.RestartPolicyOnFailure,
			},
		},
	}
}

// ContainersSpec returns the Containers obj based on Environment Type.
func ContainersSpec(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, mountInfos []NFSVolumeMountInfo, opts *ContainerEnvOpts) []corev1.Container {
	var containers []corev1.Container
	volumeMountPath := PersistentMountPath(environment)
	switch environment.EnvironmentType {
	case clv1alpha2.ClassContainer:
		containers = append(containers, WebsockifyContainer(opts, environment, instance), XVncContainer(opts), AppContainer(environment, volumeMountPath, mountInfos))
	case clv1alpha2.ClassStandalone:
		containers = append(containers, StandaloneContainer(instance, environment, volumeMountPath, mountInfos))
	default:
	}
	return containers
}

// WebsockifyContainer forges the sidecar container to proxy requests from websocket
// to the VNC server.
func WebsockifyContainer(opts *ContainerEnvOpts, environment *clv1alpha2.Environment, instance *clv1alpha2.Instance) corev1.Container {
	websockifyContainer := GenericContainer(WebsockifyName, fmt.Sprintf("%s:%s", opts.WebsockifyImg, opts.ImagesTag))
	SetContainerResources(&websockifyContainer, 0.01, 0.1, 30, 100)
	AddEnvVariableFromFieldToContainer(&websockifyContainer, PodNameEnvName, "metadata.name")
	AddEnvVariableFromResourcesToContainer(&websockifyContainer, AppCPULimitsEnvName, environment.Name, corev1.ResourceLimitsCPU, MilliDivisor)
	AddEnvVariableFromResourcesToContainer(&websockifyContainer, AppMEMLimitsEnvName, environment.Name, corev1.ResourceLimitsMemory, DefaultDivisor)
	AddTCPPortToContainer(&websockifyContainer, GUIPortName, GUIPortNumber)
	AddTCPPortToContainer(&websockifyContainer, MetricsPortName, MetricsPortNumber)
	AddContainerArg(&websockifyContainer, "http-addr", fmt.Sprintf(":%d", GUIPortNumber))
	AddContainerArg(&websockifyContainer, "base-path", IngressGUICleanPath(instance))
	AddContainerArg(&websockifyContainer, "metrics-addr", fmt.Sprintf(":%d", MetricsPortNumber))
	AddContainerArg(&websockifyContainer, "show-controls", fmt.Sprint(!environment.DisableControls))
	AddContainerArg(&websockifyContainer, "instmetrics-server-endpoint", opts.InstMetricsEndpoint)
	AddContainerArg(&websockifyContainer, "pod-name", fmt.Sprintf("$(%s)", PodNameEnvName))
	AddContainerArg(&websockifyContainer, "cpu-limit", fmt.Sprintf("$(%s)", AppCPULimitsEnvName))
	AddContainerArg(&websockifyContainer, "memory-limit", fmt.Sprintf("$(%s)", AppMEMLimitsEnvName))
	SetContainerReadinessHTTPProbe(&websockifyContainer, GUIPortName, HealthzEndpoint)
	return websockifyContainer
}

// XVncContainer forges the sidecar container which holds the desktop environment through a X+VNC server.
func XVncContainer(opts *ContainerEnvOpts) corev1.Container {
	xVncContainer := GenericContainer(XVncName, fmt.Sprintf("%s:%s", opts.XVncImg, opts.ImagesTag))
	SetContainerResources(&xVncContainer, 0.05, 0.25, 200, 600)
	AddTCPPortToContainer(&xVncContainer, XVncPortName, XVncPortNumber)
	SetContainerReadinessTCPProbe(&xVncContainer, XVncPortName)
	return xVncContainer
}

// StandaloneContainer forges the Standalone application container of the environment.
func StandaloneContainer(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, volumeMountPath string, mountInfos []NFSVolumeMountInfo) corev1.Container {
	standaloneContainer := AppContainer(environment, volumeMountPath, mountInfos)
	AddTCPPortToContainer(&standaloneContainer, GUIPortName, GUIPortNumber)

	AddEnvVariableToContainer(&standaloneContainer, "CROWNLABS_BASE_PATH", IngressGUICleanPath(instance))
	AddEnvVariableToContainer(&standaloneContainer, "CROWNLABS_LISTEN_PORT", strconv.Itoa(GUIPortNumber))

	if environment.RewriteURL {
		SetContainerReadinessHTTPProbe(&standaloneContainer, GUIPortName, "/")
	} else {
		SetContainerReadinessHTTPProbe(&standaloneContainer, GUIPortName, IngressGUIPath(instance, environment))
	}

	return standaloneContainer
}

// AppContainer forges the main application container of the environment.
func AppContainer(environment *clv1alpha2.Environment, volumeMountPath string, mountInfos []NFSVolumeMountInfo) corev1.Container {
	appContainer := GenericContainer(environment.Name, environment.Image)
	SetContainerResourcesFromEnvironment(&appContainer, environment)
	AddEnvVariableFromResourcesToContainer(&appContainer, "CROWNLABS_CPU_REQUESTS", appContainer.Name, corev1.ResourceRequestsCPU, DefaultDivisor)
	AddEnvVariableFromResourcesToContainer(&appContainer, "CROWNLABS_CPU_LIMITS", appContainer.Name, corev1.ResourceLimitsCPU, DefaultDivisor)
	AddContainerVolumeMount(&appContainer, PersistentVolumeName, volumeMountPath)
	for _, mountInfo := range mountInfos {
		AddContainerVolumeMount(&appContainer, mountInfo.VolumeName, mountInfo.MountPath)
	}
	if environment.ContainerStartupOptions != nil {
		appContainer.Args = environment.ContainerStartupOptions.StartupArgs
		if environment.ContainerStartupOptions.EnforceWorkdir {
			appContainer.WorkingDir = PersistentMountPath(environment)
		}
	}
	return appContainer
}

// InitContainers forges the list of initcontainers for the container based environment.
func InitContainers(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, opts *ContainerEnvOpts) []corev1.Container {
	if check, origin := NeedsInitContainer(instance, environment); check {
		return []corev1.Container{ContentDownloaderInitContainer(origin, opts)}
	}
	return nil
}

// ContentDownloaderInitContainer forges a Container to be used as initContainer for downloading and decompressing an archive file into the <MyDriveName> volume.
func ContentDownloaderInitContainer(contentOrigin string, ceOpts *ContainerEnvOpts) corev1.Container {
	contentDownloader := GenericContainer(ContentDownloaderName, fmt.Sprintf("%s:%s", ceOpts.ContentDownloaderImg, ceOpts.ImagesTag))
	SetContainerResources(&contentDownloader, 0.5, 1, 256, 1024)
	// MyDriveDefaultMountPath as mount point ensures a fixed path just for the download, it will likely be different in the application container
	AddContainerVolumeMount(&contentDownloader, PersistentVolumeName, PersistentDefaultMountPath)
	AddEnvVariableToContainer(&contentDownloader, "SOURCE_ARCHIVE", contentOrigin)
	AddEnvVariableToContainer(&contentDownloader, "DESTINATION_PATH", PersistentDefaultMountPath)
	return contentDownloader
}

// ContentUploaderJobContainer forges a Container to be used within a Job to compress and upload an archive file from the <MyDriveName> volume.
func ContentUploaderJobContainer(contentDestination, filename string, ceOpts *ContainerEnvOpts) corev1.Container {
	contentUploader := GenericContainer(ContentUploaderName, fmt.Sprintf("%s:%s", ceOpts.ContentUploaderImg, ceOpts.ImagesTag))
	SetContainerResources(&contentUploader, 0.5, 1, 256, 1024)
	AddContainerVolumeMount(&contentUploader, PersistentVolumeName, PersistentDefaultMountPath)
	AddEnvVariableToContainer(&contentUploader, "SOURCE_PATH", PersistentDefaultMountPath)
	AddEnvVariableToContainer(&contentUploader, "DESTINATION_URL", contentDestination)
	AddEnvVariableToContainer(&contentUploader, "FILENAME", filename)
	return contentUploader
}

// GenericContainer forges a Container specification with a restrictive security context.
func GenericContainer(name, image string) corev1.Container {
	return corev1.Container{
		Name:            name,
		Image:           image,
		SecurityContext: RestrictiveSecurityContext(),
	}
}

// RestrictiveSecurityContext forges an unprivileged SecurityContext to ensure insulation.
func RestrictiveSecurityContext() *corev1.SecurityContext {
	return &corev1.SecurityContext{
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{
				corev1.Capability("ALL"),
			},
		},
		Privileged:               ptr.To(false),
		AllowPrivilegeEscalation: ptr.To(false),
	}
}

// AddTCPPortToContainer appends a TCP port to the given container's ports.
func AddTCPPortToContainer(c *corev1.Container, name string, port int) {
	c.Ports = append(c.Ports, corev1.ContainerPort{
		Name:          name,
		ContainerPort: int32(port),
		Protocol:      corev1.ProtocolTCP,
	})
}

// AddEnvVariableToContainer appends an environment variable to the given container's env with the specified value.
func AddEnvVariableToContainer(c *corev1.Container, name, value string) {
	c.Env = append(c.Env, corev1.EnvVar{
		Name:  name,
		Value: value,
	})
}

// AddEnvVariableFromFieldToContainer appends an environment variable to the given container's env with a generic field-referenced value.
func AddEnvVariableFromFieldToContainer(c *corev1.Container, name, value string) {
	c.Env = append(c.Env, corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: value,
			},
		},
	})
}

// AddEnvVariableFromResourcesToContainer appends an environment variable to the given container's env with a resource-referenced value from the source container.
func AddEnvVariableFromResourcesToContainer(c *corev1.Container, envVarName, srcContainerName string, resName corev1.ResourceName, divisor resource.Quantity) {
	c.Env = append(c.Env, corev1.EnvVar{
		Name: envVarName,
		ValueFrom: &corev1.EnvVarSource{
			ResourceFieldRef: &corev1.ResourceFieldSelector{
				ContainerName: srcContainerName,
				Resource:      resName.String(),
				Divisor:       divisor,
			},
		},
	})
}

// AddContainerVolumeMount appends a VolumeMount to the given container's volumeMounts.
func AddContainerVolumeMount(c *corev1.Container, name, path string) {
	c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
		Name:      name,
		MountPath: path,
	})
}

// AddContainerArg appends an argument to the given container's args in the format of --name=value.
func AddContainerArg(c *corev1.Container, name, value string) {
	c.Args = append(c.Args, fmt.Sprintf("--%s=%s", name, value))
}

// SetContainerReadinessTCPProbe sets the given container's ReadinessProbe with a TCPSocket handler.
func SetContainerReadinessTCPProbe(c *corev1.Container, portName string) {
	probe := ContainerProbe()
	probe.TCPSocket = &corev1.TCPSocketAction{
		Port: intstr.FromString(portName),
	}
	c.ReadinessProbe = probe
}

// SetContainerReadinessHTTPProbe sets the given container's ReadinessProbe with a HTTPGet handler.
func SetContainerReadinessHTTPProbe(c *corev1.Container, portName, path string) {
	probe := ContainerProbe()
	probe.HTTPGet = &corev1.HTTPGetAction{
		Port: intstr.FromString(portName),
		Path: path,
	}
	c.ReadinessProbe = probe
}

// ContainerProbe forges a Probe with certain preset values and no handler.
func ContainerProbe() *corev1.Probe {
	return &corev1.Probe{
		InitialDelaySeconds: 10,
		PeriodSeconds:       2,
		SuccessThreshold:    2,
		FailureThreshold:    5,
	}
}

// SetContainerResources sets the given container's Resources with the given values:
// cpu values are converted to millis: 1.2 -> 1200m; ram sizes are in MiB.
func SetContainerResources(c *corev1.Container, cpuRequests, cpuLimits float32, memRequestsMi, memLimitsMi int) {
	scaleToMi := 1024 * 1024
	c.Resources = corev1.ResourceRequirements{Requests: corev1.ResourceList{
		corev1.ResourceCPU:    *resource.NewMilliQuantity(int64(cpuRequests*1e3), resource.DecimalSI),
		corev1.ResourceMemory: *resource.NewQuantity(int64(memRequestsMi*scaleToMi), resource.BinarySI),
	}, Limits: corev1.ResourceList{
		corev1.ResourceCPU:    *resource.NewMilliQuantity(int64(cpuLimits*1e3), resource.DecimalSI),
		corev1.ResourceMemory: *resource.NewQuantity(int64(memLimitsMi*scaleToMi), resource.BinarySI),
	}}
}

// SetContainerResourcesFromEnvironment sets the given container's Resources starting from the Resources specified inside the environment.
func SetContainerResourcesFromEnvironment(c *corev1.Container, env *clv1alpha2.Environment) {
	cpuRequestsMillis := env.Resources.CPU * env.Resources.ReservedCPUPercentage * 10

	c.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    *resource.NewMilliQuantity(int64(cpuRequestsMillis), resource.DecimalSI),
			corev1.ResourceMemory: env.Resources.Memory,
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    *resource.NewQuantity(int64(env.Resources.CPU), resource.DecimalSI),
			corev1.ResourceMemory: env.Resources.Memory,
		},
	}
}

// ContainerVolumes forges the list of volumes for the deployment spec, possibly returning an empty
// list in case the environment is not standard and not persistent.
func ContainerVolumes(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, mountInfos []NFSVolumeMountInfo) []corev1.Volume {
	vols := []corev1.Volume{ContainerVolume(PersistentVolumeName, NamespacedName(instance).Name, environment)}

	for _, mountInfo := range mountInfos {
		vols = append(vols, NFSVolume(mountInfo))
	}

	return vols
}

// ContainerVolume forges a Volume containing
// a PVC source in case of persistent envs, an emptydir in case of non persistent envs.
func ContainerVolume(volumeName, claimName string, environment *clv1alpha2.Environment) corev1.Volume {
	if environment.Persistent {
		return corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: claimName,
				},
			},
		}
	}

	return corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}

// NFSVolume receives a specification of a volume and returns the NFS volume.
func NFSVolume(mountInfo NFSVolumeMountInfo) corev1.Volume {
	return corev1.Volume{
		Name: mountInfo.VolumeName,
		VolumeSource: corev1.VolumeSource{
			NFS: &corev1.NFSVolumeSource{
				Server:   mountInfo.ServerAddress,
				Path:     mountInfo.ExportPath,
				ReadOnly: mountInfo.ReadOnly,
			},
		},
	}
}

// NeedsInitContainer returns true if the environment requires an initcontainer in order to be prepopulated.
func NeedsInitContainer(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) (value bool, contentOrigin string) {
	if icu := instance.Spec.CustomizationUrls; icu != nil && icu.ContentOrigin != "" {
		return true, icu.ContentOrigin
	}
	if cso := environment.ContainerStartupOptions; cso != nil && cso.SourceArchiveURL != "" {
		return true, cso.SourceArchiveURL
	}
	return false, ""
}

// PersistentMountPath returns the path on which mounting the persistent volume.
func PersistentMountPath(environment *clv1alpha2.Environment) string {
	cso := environment.ContainerStartupOptions
	if cso != nil && cso.ContentPath != "" {
		return environment.ContainerStartupOptions.ContentPath
	}

	return PersistentDefaultMountPath
}

// InstanceHostname forges the hostname of the instance:
// empty for standard mode (will use pod name) or the lowercase mode otherwise.
func InstanceHostname(environment *clv1alpha2.Environment) string {
	if environment.Mode != clv1alpha2.ModeStandard {
		return strings.ToLower(string(environment.Mode))
	}
	return ""
}

// NodeSelectorLabels returns the node selector labels chosen
// based on the instance and the environment.
func NodeSelectorLabels(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) map[string]string {
	templateLabelSelector := environment.NodeSelector
	instanceLabelSelector := instance.Spec.NodeSelector

	if templateLabelSelector == nil {
		return map[string]string{}
	}
	if len(*templateLabelSelector) == 0 {
		return instanceLabelSelector
	}
	return *templateLabelSelector
}
