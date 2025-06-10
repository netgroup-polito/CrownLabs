package forge

import (
	"fmt"

	controlplanekamajiv1 "github.com/clastix/cluster-api-control-plane-provider-kamaji/api/v1alpha1"
	"github.com/clastix/kamaji/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	virtv1 "kubevirt.io/api/core/v1"
	capiv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
)

// KamajiControlPlaneSpec forges the specification of a Kamaji controlplane object
func KamajiControlPlaneSpec(environment *clv1alpha2.Environment, host string) controlplanekamajiv1.KamajiControlPlaneSpec {
	return controlplanekamajiv1.KamajiControlPlaneSpec{
		KamajiControlPlaneFields: KamajiControlPlaneFields(environment, host),
		Replicas:                 ptr.To(int32(environment.Cluster.ControlPlane.Replicas)),
		Version:                  environment.Cluster.Version,
	}
}

// KamajiControlPlaneFields forges the specification of a Kamaji controlplane spec
func KamajiControlPlaneFields(environment *clv1alpha2.Environment, host string) controlplanekamajiv1.KamajiControlPlaneFields {
	return controlplanekamajiv1.KamajiControlPlaneFields{
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
			CertSANs: []string{
				host,
				"ingress.local",
				environment.Cluster.ClusterNet.CertSAN,
			},
		},
		Deployment: controlplanekamajiv1.DeploymentComponent{},
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

// ClusterControlPlaneSepc forges the specification of a cluster controlplane spec
func ClusterControlPlaneSepc(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, host string) controlplanev1.KubeadmControlPlaneSpec {
	return controlplanev1.KubeadmControlPlaneSpec{
		MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
			InfrastructureRef: MachineInfrastructureRef(instance, environment, fmt.Sprintf("%s-control-plane-machine", environment.Cluster.Name)),
		},
		KubeadmConfigSpec: ControlPlaneKubeadmConfigSpec(instance, environment, host),
		Version:           environment.Cluster.Version,
	}
}

// ControlPlaneKubeadmConfigSpec  forges the specification of a kubeadm controlplane spec
func ControlPlaneKubeadmConfigSpec(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, host string) bootstrapv1.KubeadmConfigSpec {
	return bootstrapv1.KubeadmConfigSpec{
		ClusterConfiguration: ptr.To(ControlPlaneClusterConfiguration(instance, environment, host)),
		InitConfiguration: ptr.To(bootstrapv1.InitConfiguration{
			NodeRegistration: bootstrapv1.NodeRegistrationOptions{CRISocket: "/var/run/containerd/containerd.sock"},
		}),
		JoinConfiguration: ptr.To(bootstrapv1.JoinConfiguration{
			NodeRegistration: bootstrapv1.NodeRegistrationOptions{CRISocket: "/var/run/containerd/containerd.sock"},
		}),
	}
}

// ControlPlaneClusterConfiguration  forges the specification of a cluster controlplane configuration
func ControlPlaneClusterConfiguration(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, host string) bootstrapv1.ClusterConfiguration {
	return bootstrapv1.ClusterConfiguration{
		Networking: ControlPlaneNetworking(instance, environment),
		APIServer: bootstrapv1.APIServer{
			CertSANs: []string{
				host,
				"ingress.local",
				environment.Cluster.ClusterNet.CertSAN,
			},
		},
	}
}

// ControlPlaneNetworking forges the spcification of controlplane network configuration
func ControlPlaneNetworking(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) bootstrapv1.Networking {
	return bootstrapv1.Networking{
		DNSDomain:     fmt.Sprintf("%s.%s.local", environment.Cluster.Name, instance.Namespace),
		PodSubnet:     environment.Cluster.ClusterNet.Pods,
		ServiceSubnet: environment.Cluster.ClusterNet.Services,
	}
}

// ClusterSpec forges the specification of a cluster object
func ClusterSpec(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) capiv1.ClusterSpec {
	Provider := environment.Cluster.ControlPlane.Provider
	if Provider == clv1alpha2.ProviderKubeadm {
		return capiv1.ClusterSpec{
			ClusterNetwork: ptr.To(ClusterNetworking(environment)),
			InfrastructureRef: ptr.To(corev1.ObjectReference{
				APIVersion: "infrastructure.cluster.x-k8s.io/v1alpha1",
				Kind:       "KubevirtCluster",
				Name:       fmt.Sprintf("%s-infra", environment.Cluster.Name),
				Namespace:  instance.Namespace,
			}),
			ControlPlaneRef: ptr.To(corev1.ObjectReference{
				APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
				Kind:       "KubeadmControlPlane",
				Name:       fmt.Sprintf("%s-control-plane", environment.Cluster.Name),
				Namespace:  instance.Namespace,
			}),
		}
	} else {
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

}

// ClusterNetworking forges the spcification of cluster network
func ClusterNetworking(environment *clv1alpha2.Environment) capiv1.ClusterNetwork {
	return capiv1.ClusterNetwork{
		Pods: ptr.To(capiv1.NetworkRanges{
			CIDRBlocks: []string{environment.Cluster.ClusterNet.Pods},
		}),
		Services: ptr.To(capiv1.NetworkRanges{
			CIDRBlocks: []string{environment.Cluster.ClusterNet.Services},
		}),
	}
}
