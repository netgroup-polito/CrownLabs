// Copyright 2020-2021 Politecnico di Torino
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

package instance_controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

func buildResRequirements(
	cpu float32, reservedCPUPerc uint32,
	mem resource.Quantity,
) v1.ResourceRequirements {
	return v1.ResourceRequirements{
		Requests: v1.ResourceList{
			"cpu":    resource.MustParse(fmt.Sprintf("%f", cpu/100*float32(reservedCPUPerc))),
			"memory": mem,
		},
		Limits: v1.ResourceList{
			"cpu":    resource.MustParse(fmt.Sprintf("%f", cpu)),
			"memory": mem,
		},
	}
}

func buildContainerVolume(
	volumeName, claimName string,
	environment *clv1alpha2.Environment,
) v1.Volume {
	volume := v1.Volume{
		Name:         volumeName,
		VolumeSource: v1.VolumeSource{},
	}

	if environment.Resources.Disk.IsZero() {
		volume.VolumeSource.EmptyDir = &v1.EmptyDirVolumeSource{}
	} else {
		volume.VolumeSource.PersistentVolumeClaim = &v1.PersistentVolumeClaimVolumeSource{
			ClaimName: claimName,
		}
	}

	return volume
}

func buildContainerInstanceDeploymentSpec(
	name string, instance *clv1alpha2.Instance,
	environment *clv1alpha2.Environment,
	o *ContainerEnvOpts, httpPort int32,
	fileBrowserPort int32, mountPath, basePath string,
) appsv1.DeploymentSpec {
	userID := int64(1010)
	yes := true
	no := false

	podSecCtx := v1.PodSecurityContext{
		RunAsUser:    &userID,
		RunAsGroup:   &userID,
		RunAsNonRoot: &yes,
	}

	contSecCtx := v1.SecurityContext{
		Capabilities: &v1.Capabilities{
			Drop: []v1.Capability{
				v1.Capability("ALL"),
			},
		},
		Privileged:               &no,
		AllowPrivilegeEscalation: &no,
	}

	noVncPortName := "http-port"
	noVncProbe := v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Port: intstr.FromString(noVncPortName),
				Path: "/healthz",
			},
		},
		InitialDelaySeconds: 3,
		PeriodSeconds:       5,
	}

	websockifyPort := int32(8888)
	websockifyPortName := "websockify-port"
	websockifyProbe := v1.Probe{
		Handler: v1.Handler{
			TCPSocket: &v1.TCPSocketAction{
				Port: intstr.FromString(websockifyPortName),
			},
		},
		InitialDelaySeconds: 1,
		PeriodSeconds:       2,
		SuccessThreshold:    3,
	}

	vncPort := int32(5900)
	vncPortName := "vnc-port"
	tigerVncProbe := v1.Probe{
		Handler: v1.Handler{
			TCPSocket: &v1.TCPSocketAction{
				Port: intstr.FromString(vncPortName),
			},
		},
		InitialDelaySeconds: 3,
		PeriodSeconds:       2,
		SuccessThreshold:    3,
	}

	fileBrowserPortName := "browser-port"
	fileBrowserProbe := v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Port: intstr.FromString(fileBrowserPortName),
				Path: "/health",
			},
		},
		InitialDelaySeconds: 3,
		PeriodSeconds:       5,
	}

	containers := []v1.Container{
		{
			Name:      "novnc",
			Image:     o.NovncImg + ":" + o.ImagesTag,
			Resources: buildResRequirements(0.02, 50, resource.MustParse("100Mi")), // actual: ~25MiB
			Ports: []v1.ContainerPort{{
				ContainerPort: httpPort,
				Name:          noVncPortName,
			}},
			Env: []v1.EnvVar{{
				Name:  "HIDE_NOVNC_BAR",
				Value: strconv.FormatBool(environment.Mode == clv1alpha2.ModeExam || environment.Mode == clv1alpha2.ModeExercise),
			}, {
				Name:  "HTTP_PORT",
				Value: fmt.Sprintf("%d", httpPort),
			}, {
				Name:  "WEBSOCKIFY_PORT",
				Value: fmt.Sprintf("%d", websockifyPort),
			}},
			SecurityContext: &contSecCtx,
			// LivenessProbe:   &noVncProbe,
			ReadinessProbe: &noVncProbe,
		},
		{
			Name:      "websockify",
			Image:     o.WebsockifyImg + ":" + o.ImagesTag,
			Resources: buildResRequirements(0.02, 50, resource.MustParse("50Mi")), // actual: ~2MiB
			Env: []v1.EnvVar{{
				Name:  "WS_PORT",
				Value: fmt.Sprintf("%d", websockifyPort),
			}},
			Ports: []v1.ContainerPort{{
				ContainerPort: websockifyPort,
				Name:          websockifyPortName,
			}},
			SecurityContext: &contSecCtx,
			// LivenessProbe: &websockifyProbe,
			ReadinessProbe: &websockifyProbe,
		},
		{
			Name:            "tigervnc",
			Image:           o.VncImg + ":" + o.ImagesTag,
			Resources:       buildResRequirements(0.5, 50, resource.MustParse("500Mi")), // Mem depends on screen resolution, should easily manage up to 2k virtual screens
			SecurityContext: &contSecCtx,
			Ports: []v1.ContainerPort{{
				ContainerPort: vncPort,
				Name:          vncPortName,
			}},
			// LivenessProbe:   &tigerVncProbe,
			ReadinessProbe: &tigerVncProbe,
		},
		{
			Name:  name,
			Image: environment.Image,
			Resources: buildResRequirements(
				float32(environment.Resources.CPU),
				environment.Resources.ReservedCPUPercentage,
				environment.Resources.Memory,
			),
			Env: []v1.EnvVar{{
				ValueFrom: &v1.EnvVarSource{
					ResourceFieldRef: &v1.ResourceFieldSelector{
						ContainerName: name,
						Resource:      "requests.cpu",
					},
				},
				Name: "CROWNLABS_CPU_REQUESTS",
			}, {
				ValueFrom: &v1.EnvVarSource{
					ResourceFieldRef: &v1.ResourceFieldSelector{
						ContainerName: name,
						Resource:      "limits.cpu",
					},
				},
				Name: "CROWNLABS_CPU_LIMITS",
			}},
			SecurityContext: &contSecCtx,
			VolumeMounts: []v1.VolumeMount{{
				Name:      "shared",
				MountPath: mountPath, // Same as filebrowser for simplicity
			}},
		},
	}

	if environment.Mode == clv1alpha2.ModeStandard {
		containers = append(containers, v1.Container{
			Name:  "filebrowser",
			Image: o.FileBrowserImg + ":" + o.FileBrowserImgTag,
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					"cpu":    resource.MustParse(fmt.Sprintf("%f", 0.01)),
					"memory": resource.MustParse("100Mi"),
				},
				Limits: v1.ResourceList{
					"cpu":    resource.MustParse(fmt.Sprintf("%f", 0.25)),
					"memory": resource.MustParse("500Mi"),
				},
			},
			Args: []string{
				"--port=" + fmt.Sprintf("%d", fileBrowserPort),
				"--root=" + mountPath,
				"--baseurl=" + basePath,
				"--database=/tmp/database.db",
				"--noauth=true",
			},
			SecurityContext: &contSecCtx,
			Ports: []v1.ContainerPort{{
				ContainerPort: fileBrowserPort,
				Name:          fileBrowserPortName,
			}},
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      "shared",
					MountPath: mountPath,
				},
			},
			ReadinessProbe: &fileBrowserProbe,
		})
	}

	return appsv1.DeploymentSpec{
		Replicas: pointer.Int32Ptr(1),
		Selector: &metav1.LabelSelector{
			MatchLabels: forge.InstanceSelectorLabels(instance),
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: forge.InstanceSelectorLabels(instance),
			},
			Spec: v1.PodSpec{
				Containers:                   containers,
				SecurityContext:              &podSecCtx,
				AutomountServiceAccountToken: &no,
				Volumes: []v1.Volume{
					buildContainerVolume("shared", name, environment),
				},
			},
		},
	}
}

func buildContainerInstancePVCSpec(
	environment *clv1alpha2.Environment,
) v1.PersistentVolumeClaimSpec {
	return v1.PersistentVolumeClaimSpec{
		AccessModes: []v1.PersistentVolumeAccessMode{
			v1.ReadWriteOnce,
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceStorage: environment.Resources.Disk,
			},
		},
	}
}

// EnforceContainerEnvironment implements the logic to create all the different
// Kubernetes resources required to start a containerized CrownLabs environment.
func (r *InstanceReconciler) EnforceContainerEnvironment(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)
	environment := clctx.EnvironmentFrom(ctx)

	name := strings.ReplaceAll(instance.GetName(), ".", "-")
	namespace := instance.GetNamespace()

	// Enforce the service and the ingress to expose the environment.
	err := r.EnforceInstanceExposition(ctx)
	if err != nil {
		log.Error(err, "failed to enforce the instance exposition objects")
		return err
	}

	if !environment.Resources.Disk.IsZero() {
		pvc := v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}

		res, err := ctrl.CreateOrUpdate(ctx, r.Client, &pvc, func() error {
			// PVC's spec is immutable, it has to be set at creation
			if pvc.ObjectMeta.CreationTimestamp.IsZero() {
				pvc.Spec = buildContainerInstancePVCSpec(environment)
			}
			return ctrl.SetControllerReference(instance, &pvc, r.Scheme)
		})
		if err != nil {
			log.Error(err, "failed to enforce PVC existence", "pvc", klog.KObj(&pvc))
			return err
		}
		klog.Infof("PVC for instance %s/%s %s", instance.GetNamespace(), instance.GetName(), res)
	}

	depl := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, &depl, func() error {
		depl.Spec = buildContainerInstanceDeploymentSpec(name, instance, environment, &r.ContainerEnvOpts, 6080, 8080, "/mydrive", forge.IngressMyDrivePath(instance))
		depl.SetLabels(forge.InstanceObjectLabels(depl.GetLabels(), instance))
		return ctrl.SetControllerReference(instance, &depl, r.Scheme)
	}); err != nil {
		log.Error(err, "failed to enforce deployment existence", "deployment", klog.KObj(&depl))
		return err
	}

	phase := r.RetrievePhaseFromDeployment(&depl)
	if phase != instance.Status.Phase {
		log.Info("phase changed", "deployment", klog.KObj(&depl),
			"previous", string(instance.Status.Phase), "current", string(phase))
		instance.Status.Phase = phase
	}

	return nil
}
