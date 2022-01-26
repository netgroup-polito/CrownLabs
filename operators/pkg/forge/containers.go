// Copyright 2020-2022 Politecnico di Torino
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

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const (
	// WebsockifyName -> name of the websockify sidecar container.
	WebsockifyName = "websockify"
	// XVncName -> name of the x+vnc server sidecar container.
	XVncName = "xvnc"
	// MyDriveName -> name of the filebrowser sidecar container.
	MyDriveName = "mydrive"
	// ContentDownloaderName -> name of the downloader initcontainer.
	ContentDownloaderName = "content-downloader"
	// ContentUploaderName -> name of the uploader initcontainer.
	ContentUploaderName = "content-uploader"
	// MyDriveDefaultMountPath -> default path for the user files accessible through the mydrive.
	MyDriveDefaultMountPath = "/mydrive"
	// MyDriveDBPath -> default path for the filebrowser internal database file.
	MyDriveDBPath = "/tmp/database.db"
	// HealthzEndpoint -> default endpoint for HTTP probes.
	HealthzEndpoint = "/healthz"
	// CrownLabsUserID -> used as UID and GID for containers security context.
	CrownLabsUserID = int64(1010)

	containersTerminationGracePeriod = 10
)

// ContainerEnvOpts contains images name and tag for container environment.
type ContainerEnvOpts struct {
	ImagesTag            string
	XVncImg              string
	WebsockifyImg        string
	MyDriveImgAndTag     string
	ContentDownloaderImg string
	ContentUploaderImg   string
}

// PVCSpec forges a ReadWriteOnce PersistentVolumeClaimSpec
// with requests set as in environment.Resources.Disk.
func PVCSpec(environment *clv1alpha2.Environment) corev1.PersistentVolumeClaimSpec {
	return corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		},
		StorageClassName: PVCStorageClassName(environment),
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: environment.Resources.Disk,
			},
		},
	}
}

// PVCStorageClassName returns the storage class configured as option, or nil if empty.
func PVCStorageClassName(environment *clv1alpha2.Environment) *string {
	if environment.StorageClassName != "" {
		return pointer.String(environment.StorageClassName)
	}
	return nil
}

// PodSecurityContext forges a PodSecurityContext
// with 1010 UID and GID and RunAsNonRoot set.
func PodSecurityContext() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{
		RunAsUser:    pointer.Int64(CrownLabsUserID),
		RunAsGroup:   pointer.Int64(CrownLabsUserID),
		FSGroup:      pointer.Int64(CrownLabsUserID),
		RunAsNonRoot: pointer.Bool(true),
	}
}

// ReplicasCount returns an int32 pointer to 1 if the instance is not persistent
// or, if persistent, in case environment spec is set as running; 0 otherwise.
func ReplicasCount(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, isNew bool) *int32 {
	if (!isNew && !environment.Persistent) || instance.Spec.Running {
		return pointer.Int32(1)
	}
	return pointer.Int32(0)
}

// DeploymentSpec forges the complete DeploymentSpec (without replicas)
// containing the needed sidecars for X-VNC based container instances.
func DeploymentSpec(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, opts *ContainerEnvOpts) appsv1.DeploymentSpec {
	return appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{MatchLabels: InstanceSelectorLabels(instance)},
		Strategy: appsv1.DeploymentStrategy{
			Type: appsv1.RecreateDeploymentStrategyType,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: InstanceSelectorLabels(instance)},
			Spec:       PodSpec(instance, environment, opts),
		},
	}
}

// PodSpec forges the pod specification for X-VNC based container instance,
// conditionally includes the "myDrive" sidecar for standard mode environments.
func PodSpec(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, opts *ContainerEnvOpts) corev1.PodSpec {
	driveMountPath := MyDriveMountPath(environment)

	spec := corev1.PodSpec{
		Containers: []corev1.Container{
			WebsockifyContainer(opts),
			XVncContainer(opts),
			AppContainer(instance, environment, driveMountPath),
		},
		Volumes:                       ContainerVolumes(instance, environment),
		SecurityContext:               PodSecurityContext(),
		AutomountServiceAccountToken:  pointer.Bool(false),
		TerminationGracePeriodSeconds: pointer.Int64(containersTerminationGracePeriod),
		InitContainers:                InitContainers(instance, environment, opts),
	}

	if environment.Mode == clv1alpha2.ModeStandard {
		spec.Containers = append(spec.Containers, MyDriveContainer(instance, opts, driveMountPath))
	}

	return spec
}

// SubmissionJobSpec returns the job spec for the submission job.
func SubmissionJobSpec(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, opts *ContainerEnvOpts) batchv1.JobSpec {
	return batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					ContentUploaderJobContainer(instance.Spec.CustomizationUrls.ContentDestination, instance.Name, opts),
				},
				Volumes:                      ContainerVolumes(instance, environment),
				SecurityContext:              PodSecurityContext(),
				AutomountServiceAccountToken: pointer.Bool(false),
				RestartPolicy:                corev1.RestartPolicyOnFailure,
			},
		},
	}
}

// WebsockifyContainer forges the sidecar container to proxy requests from websocket
// to the VNC server.
func WebsockifyContainer(opts *ContainerEnvOpts) corev1.Container {
	websockifyContainer := GenericContainer(WebsockifyName, fmt.Sprintf("%s:%s", opts.WebsockifyImg, opts.ImagesTag))
	SetContainerResources(&websockifyContainer, 0.01, 0.1, 30, 100)
	AddTCPPortToContainer(&websockifyContainer, GUIPortName, GUIPortNumber)
	AddEnvVariableToContainer(&websockifyContainer, "WS_PORT", fmt.Sprint(GUIPortNumber))
	SetContainerReadinessTCPProbe(&websockifyContainer, GUIPortName)
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

// MyDriveContainer forges the sidecar container which hosts a webservice to browse files in a certain path.
func MyDriveContainer(instance *clv1alpha2.Instance, opts *ContainerEnvOpts, mountPath string) corev1.Container {
	mydriveContainer := GenericContainer(MyDriveName, opts.MyDriveImgAndTag)
	SetContainerResources(&mydriveContainer, 0.01, 0.25, 100, 500)
	AddTCPPortToContainer(&mydriveContainer, MyDriveName, MyDrivePortNumber)
	SetContainerReadinessHTTPProbe(&mydriveContainer, MyDriveName, HealthzEndpoint)
	AddContainerVolumeMount(&mydriveContainer, MyDriveName, mountPath)
	AddContainerArg(&mydriveContainer, "port", fmt.Sprint(MyDrivePortNumber))
	AddContainerArg(&mydriveContainer, "root", mountPath)
	AddContainerArg(&mydriveContainer, "noauth", strconv.FormatBool(true))
	AddContainerArg(&mydriveContainer, "database", MyDriveDBPath)
	AddContainerArg(&mydriveContainer, "baseurl", IngressMyDrivePath(instance))
	return mydriveContainer
}

// AppContainer forges the main application container of the environment.
func AppContainer(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, volumeMountPath string) corev1.Container {
	appContainer := GenericContainer(environment.Name, environment.Image)
	SetContainerResourcesFromEnvironment(&appContainer, environment)
	AddEnvVariableFromResourcesToContainer(&appContainer, "CROWNLABS_CPU_REQUESTS", corev1.ResourceRequestsCPU)
	AddEnvVariableFromResourcesToContainer(&appContainer, "CROWNLABS_CPU_LIMITS", corev1.ResourceLimitsCPU)
	if NeedsContainerVolume(instance, environment) {
		AddContainerVolumeMount(&appContainer, MyDriveName, volumeMountPath)
	}
	if environment.ContainerStartupOptions != nil {
		appContainer.Args = environment.ContainerStartupOptions.StartupArgs
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
	AddContainerVolumeMount(&contentDownloader, MyDriveName, MyDriveDefaultMountPath)
	AddEnvVariableToContainer(&contentDownloader, "SOURCE_ARCHIVE", contentOrigin)
	AddEnvVariableToContainer(&contentDownloader, "DESTINATION_PATH", MyDriveDefaultMountPath)
	return contentDownloader
}

// ContentUploaderJobContainer forges a Container to be used within a Job to compress and upload an archive file from the <MyDriveName> volume.
func ContentUploaderJobContainer(contentDestination, filename string, ceOpts *ContainerEnvOpts) corev1.Container {
	contentUploader := GenericContainer(ContentUploaderName, fmt.Sprintf("%s:%s", ceOpts.ContentUploaderImg, ceOpts.ImagesTag))
	SetContainerResources(&contentUploader, 0.5, 1, 256, 1024)
	AddContainerVolumeMount(&contentUploader, MyDriveName, MyDriveDefaultMountPath)
	AddEnvVariableToContainer(&contentUploader, "SOURCE_PATH", MyDriveDefaultMountPath)
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
		Privileged:               pointer.Bool(false),
		AllowPrivilegeEscalation: pointer.Bool(false),
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

// AddEnvVariableFromResourcesToContainer appends an environment variable to the given container's env with a resource-referenced value.
func AddEnvVariableFromResourcesToContainer(c *corev1.Container, name string, resName corev1.ResourceName) {
	c.Env = append(c.Env, corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			ResourceFieldRef: &corev1.ResourceFieldSelector{
				ContainerName: c.Name,
				Resource:      resName.String(),
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
	probe.Handler.TCPSocket = &corev1.TCPSocketAction{
		Port: intstr.FromString(portName),
	}
	c.ReadinessProbe = probe
}

// SetContainerReadinessHTTPProbe sets the given container's ReadinessProbe with a HTTPGet handler.
func SetContainerReadinessHTTPProbe(c *corev1.Container, portName, path string) {
	probe := ContainerProbe()
	probe.Handler.HTTPGet = &corev1.HTTPGetAction{
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
		Handler:             corev1.Handler{},
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
func ContainerVolumes(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) []corev1.Volume {
	if !NeedsContainerVolume(instance, environment) {
		return nil
	}
	return []corev1.Volume{ContainerVolume(MyDriveName, NamespacedName(instance).Name, environment)}
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

// NeedsContainerVolume returns true in the cases in which a volume mount could be needed.
func NeedsContainerVolume(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) bool {
	needsInit, _ := NeedsInitContainer(instance, environment)
	return environment.Mode == clv1alpha2.ModeStandard || environment.Persistent || needsInit
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

// MyDriveMountPath returns the path on which mounting the myDrive volume.
func MyDriveMountPath(environment *clv1alpha2.Environment) string {
	cso := environment.ContainerStartupOptions
	if cso != nil && cso.ContentPath != "" {
		return environment.ContainerStartupOptions.ContentPath
	}

	return MyDriveDefaultMountPath
}
