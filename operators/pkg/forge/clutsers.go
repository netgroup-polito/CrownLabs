package forge

import (
	"fmt"

	controlplanekamajiv1 "github.com/clastix/cluster-api-control-plane-provider-kamaji/api/v1alpha1"
	"github.com/clastix/kamaji/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	virtv1 "kubevirt.io/api/core/v1"
	capiv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
)

const (
	PodSubnet     = "10.243.0.0/16"
	ServiceSubnet = "10.95.0.0/16"
)

// KamajiControlPlaneSpec forges the specification of a Kamaji controlplane object
func KamajiControlPlaneSpec(environment *clv1alpha2.Environment, host string) controlplanekamajiv1.KamajiControlPlaneSpec {
	certSANs := []string{host, "ingress.local"}
	if environment.Cluster.ClusterNet != nil {
		certSANs = append(certSANs, environment.Cluster.ClusterNet.CertSAN)
	}
	return controlplanekamajiv1.KamajiControlPlaneSpec{
		DataStoreName: "default",
		Addons: controlplanekamajiv1.AddonsSpec{
			AddonsSpec: v1alpha1.AddonsSpec{
				CoreDNS:   ptr.To(v1alpha1.AddonSpec{}),
				KubeProxy: ptr.To(v1alpha1.AddonSpec{}),
			},
		},
		Kubelet: v1alpha1.KubeletSpec{
			CGroupFS: v1alpha1.CGroupDriver("systemd"),
			PreferredAddressTypes: []v1alpha1.KubeletPreferredAddressType{
				"InternalIP",
				"ExternalIP",
			},
		},
		Network: controlplanekamajiv1.NetworkComponent{
			ServiceType: v1alpha1.ServiceType(environment.Cluster.ServiceType),
			CertSANs:    certSANs,
		},
		Deployment: controlplanekamajiv1.DeploymentComponent{},
		Replicas:   ptr.To(int32(environment.Cluster.ControlPlane.Replicas)),
		Version:    environment.Cluster.Version,
	}
}

// ClusterVMSpec forges the specification of a cluster virtual machine spec
func ClusterVMSpec(environment *clv1alpha2.Environment) virtv1.VirtualMachineSpec {

	return virtv1.VirtualMachineSpec{
		RunStrategy: ptr.To(virtv1.RunStrategyAlways),
		Template: &virtv1.VirtualMachineInstanceTemplateSpec{
			Spec: ClusterVMISpec(environment),
		},
	}
}

// ClusterVMISpec forges the specification of a cluster virtual machine instance spec
func ClusterVMISpec(environment *clv1alpha2.Environment) virtv1.VirtualMachineInstanceSpec {
	return virtv1.VirtualMachineInstanceSpec{
		Domain: ClusterVMDomain(environment),
		Volumes: []virtv1.Volume{
			{
				Name: "containervolume",
				VolumeSource: virtv1.VolumeSource{
					ContainerDisk: &virtv1.ContainerDiskSource{
						Image: environment.Image,
					},
				},
			},
		},
		EvictionStrategy: ptr.To(virtv1.EvictionStrategyExternal),
	}
}

// ClusterVMDomain forges  VirtualMachineDomain forges the specification of the domain of a Kubevirt VirtualMachineInstance in cluster
func ClusterVMDomain(environment *clv1alpha2.Environment) virtv1.DomainSpec {
	return virtv1.DomainSpec{
		CPU:       &virtv1.CPU{Cores: environment.Resources.CPU},
		Memory:    &virtv1.Memory{Guest: &environment.Resources.Memory},
		Resources: VirtualMachineResources(environment),
		Devices: virtv1.Devices{
			NetworkInterfaceMultiQueue: ptr.To(true),
			Disks:                      []virtv1.Disk{VolumeDiskTarget("containervolume")},
		},
	}
}

// MachineDeploymentSepc forges the specification of a machine deployment object
func MachineDeploymentSepc(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) capiv1.MachineSpec {
	return capiv1.MachineSpec{
		ClusterName: fmt.Sprintf("%s-cluster", environment.Cluster.Name),
		Version:     ptr.To(environment.Cluster.Version),
		Bootstrap: capiv1.Bootstrap{
			ConfigRef: ptr.To(BootstrapConfigRef(instance, environment)),
		},
		InfrastructureRef: MachineInfrastructureRef(instance, environment, fmt.Sprintf("%s-md-worker", environment.Cluster.Name)),
	}
}

// BootstrapConfigRef forges the specification of a Bootstrap configuration
func BootstrapConfigRef(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) corev1.ObjectReference {
	return corev1.ObjectReference{
		Name:       fmt.Sprintf("%s-md-bootstrap", environment.Cluster.Name),
		Namespace:  instance.Namespace,
		APIVersion: "bootstrap.cluster.x-k8s.io/v1beta1",
		Kind:       "KubeadmConfigTemplate",
	}
}

// MachineInfrastructureRef forges the specification of a Macine infrastructure reference
func MachineInfrastructureRef(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, Name string) corev1.ObjectReference {
	return corev1.ObjectReference{
		Name:       Name,
		Namespace:  instance.Namespace,
		APIVersion: "infrastructure.cluster.x-k8s.io/v1alpha1",
		Kind:       "KubevirtMachineTemplate",
	}
}

// ControlPlaneNetworking forges the spcification of controlplane network configuration
func ControlPlaneNetworking(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) bootstrapv1.Networking {
	return bootstrapv1.Networking{
		DNSDomain:     fmt.Sprintf("%s.%s.local", environment.Cluster.Name, instance.Namespace),
		PodSubnet:     PodSubnet,
		ServiceSubnet: ServiceSubnet,
	}
}

// ClusterSpec forges the specification of a cluster object
func ClusterSpec(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) capiv1.ClusterSpec {
	return capiv1.ClusterSpec{
		ClusterNetwork: ptr.To(ClusterNetworking(environment)),
		InfrastructureRef: ptr.To(corev1.ObjectReference{
			APIVersion: "infrastructure.cluster.x-k8s.io/v1alpha1",
			Kind:       "KubevirtCluster",
			Name:       fmt.Sprintf("%s-infra", environment.Cluster.Name),
			Namespace:  instance.Namespace,
		}),
		ControlPlaneRef: ptr.To(corev1.ObjectReference{
			APIVersion: "controlplane.cluster.x-k8s.io/v1alpha1",
			Kind:       "KamajiControlPlane",
			Name:       fmt.Sprintf("%s-control-plane", environment.Cluster.Name),
			Namespace:  instance.Namespace,
		}),
	}
}

// ClusterNetworking forges the spcification of cluster network
func ClusterNetworking(environment *clv1alpha2.Environment) capiv1.ClusterNetwork {
	return capiv1.ClusterNetwork{
		Pods: ptr.To(capiv1.NetworkRanges{
			CIDRBlocks: []string{PodSubnet},
		}),
		Services: ptr.To(capiv1.NetworkRanges{
			CIDRBlocks: []string{ServiceSubnet},
		}),
	}
}

func GuiDeploymentSpec(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) appsv1.DeploymentSpec {
	return appsv1.DeploymentSpec{
		Replicas:             ptr.To(int32(1)),
		RevisionHistoryLimit: ptr.To(int32(10)),
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": "capi-visualizer",
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "capi-visualizer",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:            "capi-visualizer",
						Image:           "docker.io/kuohandong/cluster_gui:latest",
						ImagePullPolicy: corev1.PullAlways,
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 8082,
							},
						},
						Args: []string{
							fmt.Sprintf("-cluster-name=%s-cluster", environment.Cluster.Name),
							fmt.Sprintf("-namespace=%s", instance.Namespace),
						},
					},
				},
				ServiceAccountName: "capi-visualizer",
			},
		},
	}
}
func GuiServiceSpec(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) corev1.ServiceSpec {
	return corev1.ServiceSpec{
		Type:     corev1.ServiceTypeClusterIP,
		Selector: map[string]string{"app": "capi-visualizer"},
		Ports: []corev1.ServicePort{
			{
				Name:       "https",
				Port:       8081,
				TargetPort: intstr.FromInt(8082),
				Protocol:   corev1.ProtocolTCP,
			},
		},
	}
}
