package instance_controller

import (
	"context"
	"fmt"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	instance_creation "github.com/netgroup-polito/CrownLabs/operators/pkg/instance-creation"
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

func buildContainerInstanceDeploymentSpec(
	name string, instance *crownlabsv1alpha2.Instance,
	environment *crownlabsv1alpha2.Environment,
	o *ContainerEnvOpts, httpPort int32,
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

	examMode := false // template.ExamMode (?)

	websockifyPort := int32(8888)
	vncPort := int32(5900)

	noVncProbe := v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Port: intstr.FromString("http-port"),
				Path: "/healthz",
			},
		},
		InitialDelaySeconds: 3,
		PeriodSeconds:       5,
	}

	websockifyProbe := v1.Probe{
		Handler: v1.Handler{
			TCPSocket: &v1.TCPSocketAction{
				Port: intstr.FromString("websockify-port"),
			},
		},
		InitialDelaySeconds: 1,
		PeriodSeconds:       5,
	}

	tigerVncProbe := v1.Probe{
		Handler: v1.Handler{
			TCPSocket: &v1.TCPSocketAction{
				Port: intstr.FromString("vnc-port"),
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
				Name:          "http-port",
			}},
			Env: []v1.EnvVar{{
				Name:  "HIDE_NOVNC_BAR",
				Value: strconv.FormatBool(examMode),
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
				Name:          "websockify-port",
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
				Name:          "vnc-port",
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
			SecurityContext: &contSecCtx,
		},
	}

	template := &instance.Spec.Template
	labels := map[string]string{
		"name":                         name,
		"crownlabs.polito.it/template": template.Namespace + "_" + template.Name,
	}

	return appsv1.DeploymentSpec{
		Replicas: pointer.Int32Ptr(1),
		Selector: &metav1.LabelSelector{
			MatchLabels: labels,
		},
		Template: v1.PodTemplateSpec{

			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: v1.PodSpec{
				Containers:      containers,
				SecurityContext: &podSecCtx,
			},
		},
	}
}

// CreateContainerEnvironment implements the logic to create all the different
// Kubernetes resources required to start a containerized CrownLabs environment.
func (r *InstanceReconciler) CreateContainerEnvironment(
	instance *crownlabsv1alpha2.Instance,
	environment *crownlabsv1alpha2.Environment,
	namespace string,
	name string,
	vmStart time.Time) error {
	ctx := context.TODO()

	service, ingress, err := r.CreateInstanceExpositionEnvironment(ctx, instance, name)
	if err != nil {
		return err
	}

	depl := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, &depl, func() error {
		depl.Spec = buildContainerInstanceDeploymentSpec(name, instance, environment, &r.ContainerEnvOpts, 6080)
		depl.Labels = instance_creation.UpdateLabels(depl.Labels, environment, name)
		return ctrl.SetControllerReference(instance, &depl, r.Scheme)
	}); err != nil {
		r.setInstanceStatus(ctx, "Could not create deployment "+depl.Name+" in namespace "+depl.Namespace+": "+err.Error(), "Error", "VmiNotCreated", instance, "", "")
		return err
	}

	ip := ""
	url := ""
	status := "VmiCreated"
	if depl.Status.ReadyReplicas > 0 {
		ip = service.Spec.ClusterIP
		url = ingress.GetAnnotations()["crownlabs.polito.it/probe-url"]
		status = "VmiReady"
	}
	r.setInstanceStatus(ctx, "Container Deployment "+depl.Name+" in namespace "+depl.Namespace+" status update to "+status, "Normal", status, instance, ip, url)

	return nil
}
