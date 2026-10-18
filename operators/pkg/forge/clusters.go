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

package forge

import (
	kamajicpv1alpha1 "github.com/clastix/cluster-api-control-plane-provider-kamaji/api/v1alpha1"
	kamajiv1alpha1 "github.com/clastix/kamaji/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	virtv1 "kubevirt.io/api/core/v1"
	helmv1alpha1 "sigs.k8s.io/cluster-api-addon-provider-helm/api/v1alpha1"
	infrav1alpha1 "sigs.k8s.io/cluster-api-provider-kubevirt/api/v1alpha1"
	kubeadmv1beta2 "sigs.k8s.io/cluster-api/api/bootstrap/kubeadm/v1beta2"
	clusterv1beta2 "sigs.k8s.io/cluster-api/api/core/v1beta2"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const (
	// ClusterServiceCIDR is the CIDR block assigned to Services of the workload clusters created by CrownLabs.
	ClusterServiceCIDR = "10.96.0.0/16"
	// ClusterPodCIDR is the CIDR block assigned to Pods of the workload clusters created by CrownLabs.
	ClusterPodCIDR = "10.244.0.0/16"

	// ClusterKamajiDataStoreSchema is the datastore schema configured for the Kamaji control planes created by CrownLabs.
	ClusterKamajiDataStoreSchema = "etcd"
	// ClusterKamajiManagedByAnnotation is the annotation marking a KubevirtCluster as managed by Kamaji.
	ClusterKamajiManagedByAnnotation = "cluster.x-k8s.io/managed-by"
	// ClusterKamajiManagedByValue is the value of the ClusterKamajiManagedByAnnotation annotation.
	ClusterKamajiManagedByValue = "kamaji"

	// ClusterVirtioBus is the disk bus used for the container disk of the workload cluster worker nodes.
	ClusterVirtioBus = "virtio"
	// ClusterContainerVolumeName is the name of the volume/disk hosting the worker node container image.
	ClusterContainerVolumeName = "containervolume"
	// ClusterCRISocket is the CRI socket configured on the workload cluster worker nodes.
	ClusterCRISocket = "unix:///var/run/containerd/containerd.sock"
)

// Group/Version/Kind of the Cluster API (and related providers) resources composing a CrownLabs Cluster environment.
var (
	// GVKCluster is the GVK of the Cluster API Cluster resource.
	GVKCluster = clusterv1beta2.GroupVersion.WithKind("Cluster")
	// GVKKubevirtCluster is the GVK of the KubeVirt infrastructure provider Cluster resource.
	GVKKubevirtCluster = infrav1alpha1.GroupVersion.WithKind("KubevirtCluster")
	// GVKKamajiControlPlane is the GVK of the Kamaji control plane provider resource.
	GVKKamajiControlPlane = kamajicpv1alpha1.GroupVersion.WithKind("KamajiControlPlane")
	// GVKMachineDeployment is the GVK of the Cluster API MachineDeployment resource.
	GVKMachineDeployment = clusterv1beta2.GroupVersion.WithKind("MachineDeployment")
	// GVKKubevirtMachineTemplate is the GVK of the KubeVirt infrastructure provider machine template resource.
	GVKKubevirtMachineTemplate = infrav1alpha1.GroupVersion.WithKind("KubevirtMachineTemplate")
	// GVKKubeadmConfigTemplate is the GVK of the kubeadm bootstrap provider config template resource.
	GVKKubeadmConfigTemplate = kubeadmv1beta2.GroupVersion.WithKind("KubeadmConfigTemplate")
	// GVKHelmChartProxy is the GVK of the Cluster API Add-on Provider for Helm chart proxy resource.
	GVKHelmChartProxy = helmv1alpha1.GroupVersion.WithKind("HelmChartProxy")
)

// ClusterResourceNames contains the names of the Cluster API resources composing a CrownLabs Cluster environment.
type ClusterResourceNames struct {
	// Cluster is the name shared by the Cluster and KubevirtCluster resources.
	Cluster string
	// ControlPlane is the name of the KamajiControlPlane resource.
	ControlPlane string

	// Infrastructure is the name of the KubevirtCluster resource.
	Infrastructure string

	// MachineDeployment is the name shared by the MachineDeployment, KubevirtMachineTemplate and KubeadmConfigTemplate resources.
	MachineDeployment string
}

// ClusterNamesFromInstance computes the names of the Cluster API resources associated with a Cluster environment.
func ClusterNamesFromInstance(instance metav1.Object, environment *clv1alpha2.Environment) ClusterResourceNames {
	base := ObjectMetaWithSuffix(instance, environment.Name).Name
	return ClusterResourceNames{
		Cluster:           base,
		ControlPlane:      base + StringSeparator + "control-plane",
		Infrastructure:    base + StringSeparator + "infrastructure",
		MachineDeployment: base + StringSeparator + "md-0",
	}
}

// ClusterSelectorLabels returns the labels uniquely identifying the Cluster API resources of a Cluster
// environment, used to let the HelmChartProxy addons target only the corresponding workload cluster.
func ClusterSelectorLabels(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) map[string]string {
	return map[string]string{
		LabelInstanceKey:    instance.Name,
		LabelEnvironmentKey: environment.Name,
	}
}

// NewClusterAPIObject returns an empty unstructured object with the given GVK, name and namespace set.
func NewClusterAPIObject(gvk schema.GroupVersionKind, name, namespace string) *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(gvk)
	u.SetName(name)
	u.SetNamespace(namespace)
	return u
}

// ClusterSpec returns the spec of the Cluster resource representing a CrownLabs Cluster environment.
func ClusterSpec(names ClusterResourceNames) *clusterv1beta2.ClusterSpec {
	return &clusterv1beta2.ClusterSpec{
		ControlPlaneRef: clusterv1beta2.ContractVersionedObjectReference{
			APIGroup: GVKKamajiControlPlane.Group,
			Kind:     GVKKamajiControlPlane.Kind,
			Name:     names.ControlPlane,
		},
		InfrastructureRef: clusterv1beta2.ContractVersionedObjectReference{
			APIGroup: GVKKubevirtCluster.Group,
			Kind:     GVKKubevirtCluster.Kind,
			Name:     names.Infrastructure,
		},
		ClusterNetwork: clusterv1beta2.ClusterNetwork{
			Services: clusterv1beta2.NetworkRanges{CIDRBlocks: []string{ClusterServiceCIDR}},
			Pods:     clusterv1beta2.NetworkRanges{CIDRBlocks: []string{ClusterPodCIDR}},
		},
	}
}

// KubevirtClusterAnnotations returns the annotations to be set on the KubevirtCluster resource, marking it
// as externally managed by Kamaji.
func KubevirtClusterAnnotations() map[string]string {
	return map[string]string{ClusterKamajiManagedByAnnotation: ClusterKamajiManagedByValue}
}

// componentResourcesSpec returns the resource limits for a Kamaji control plane component.
func componentResourcesSpec(res *clv1alpha2.ComponentResources) corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    res.CPU.DeepCopy(),
			corev1.ResourceMemory: res.Memory.DeepCopy(),
		},
	}
}

// KamajiControlPlaneSpec returns the spec of the KamajiControlPlane resource, forged from the ClusterFlavor
// referenced by the Cluster environment.
func KamajiControlPlaneSpec(flavor *clv1alpha2.ClusterFlavor) *kamajicpv1alpha1.KamajiControlPlaneSpec {
	controlPlane := flavor.Spec.ControlPlane
	return &kamajicpv1alpha1.KamajiControlPlaneSpec{
		Replicas: ptr.To(controlPlane.Replicas),
		Version:  flavor.Spec.Version,
		KamajiControlPlaneFields: kamajicpv1alpha1.KamajiControlPlaneFields{
			DataStoreName:   controlPlane.DataStoreName,
			DataStoreSchema: ClusterKamajiDataStoreSchema,
			Network: kamajicpv1alpha1.NetworkComponent{
				ServiceType: "ClusterIP",
			},
			ApiServer:         kamajicpv1alpha1.ControlPlaneComponent{Resources: componentResourcesSpec(&controlPlane.APIServerResources)},
			Scheduler:         kamajicpv1alpha1.ControlPlaneComponent{Resources: componentResourcesSpec(&controlPlane.SchedulerResources)},
			ControllerManager: kamajicpv1alpha1.ControlPlaneComponent{Resources: componentResourcesSpec(&controlPlane.ControllerManagerResources)},
			Addons: kamajicpv1alpha1.AddonsSpec{
				CoreDNS: &kamajicpv1alpha1.CoreDNSAddonSpec{},
				AddonsSpec: kamajiv1alpha1.AddonsSpec{
					KubeProxy: &kamajiv1alpha1.AddonSpec{},
					Konnectivity: &kamajiv1alpha1.KonnectivitySpec{
						KonnectivityServerSpec: kamajiv1alpha1.KonnectivityServerSpec{
							Port:      8132,
							Resources: ptr.To(componentResourcesSpec(&controlPlane.KonnectivityResources)),
						},
					},
				},
			},
		},
	}
}

// MachineDeploymentSpec returns the spec of the MachineDeployment resource hosting the worker nodes of the
// Cluster environment.
func MachineDeploymentSpec(clusterName string, names ClusterResourceNames, version string, workersCount int) *clusterv1beta2.MachineDeploymentSpec {
	return &clusterv1beta2.MachineDeploymentSpec{
		ClusterName: clusterName,
		Replicas:    ptr.To(int32(workersCount)), //nolint:gosec // The Environment CRD limits workersCount to 1..10.
		Selector:    metav1.LabelSelector{MatchLabels: map[string]string{}},
		Template: clusterv1beta2.MachineTemplateSpec{
			Spec: clusterv1beta2.MachineSpec{
				ClusterName: clusterName,
				Version:     version,
				InfrastructureRef: clusterv1beta2.ContractVersionedObjectReference{
					APIGroup: GVKKubevirtMachineTemplate.Group,
					Kind:     GVKKubevirtMachineTemplate.Kind,
					Name:     names.MachineDeployment,
				},
				Bootstrap: clusterv1beta2.Bootstrap{
					ConfigRef: clusterv1beta2.ContractVersionedObjectReference{
						APIGroup: GVKKubeadmConfigTemplate.Group,
						Kind:     GVKKubeadmConfigTemplate.Kind,
						Name:     names.MachineDeployment,
					},
				},
			},
		},
	}
}

// KubevirtMachineTemplateSpec returns the spec of the KubevirtMachineTemplate resource describing the VM
// hosting each worker node of the Cluster environment. The worker image comes from the ClusterFlavor, while
// the amount of CPU/Memory assigned to each worker node comes from the Environment resources.
func KubevirtMachineTemplateSpec(namespace, workerImage string, resources *clv1alpha2.EnvironmentResources) *infrav1alpha1.KubevirtMachineTemplateSpec {
	return &infrav1alpha1.KubevirtMachineTemplateSpec{
		Template: infrav1alpha1.KubevirtMachineTemplateResource{
			Spec: infrav1alpha1.KubevirtMachineSpec{
				BootstrapCheckSpec: infrav1alpha1.VirtualMachineBootstrapCheckSpec{CheckStrategy: "ssh"},
				VirtualMachineTemplate: infrav1alpha1.VirtualMachineTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Namespace: namespace},
					Spec: virtv1.VirtualMachineSpec{
						RunStrategy: ptr.To(virtv1.RunStrategyAlways),
						Template: &virtv1.VirtualMachineInstanceTemplateSpec{
							Spec: virtv1.VirtualMachineInstanceSpec{
								Domain: virtv1.DomainSpec{
									CPU:    &virtv1.CPU{Cores: Int64ToUint32(resources.CPU)},
									Memory: &virtv1.Memory{Guest: ptr.To(resources.Memory.DeepCopy())},
									Devices: virtv1.Devices{
										Disks: []virtv1.Disk{{
											Name:       ClusterContainerVolumeName,
											DiskDevice: virtv1.DiskDevice{Disk: &virtv1.DiskTarget{Bus: ClusterVirtioBus}},
										}},
									},
								},
								Volumes: []virtv1.Volume{{
									Name: ClusterContainerVolumeName,
									VolumeSource: virtv1.VolumeSource{
										ContainerDisk: &virtv1.ContainerDiskSource{Image: workerImage},
									},
								}},
							},
						},
					},
				},
			},
		},
	}
}

// KubeadmConfigTemplateSpec returns the spec of the KubeadmConfigTemplate resource used to join the worker
// nodes of the Cluster environment.
func KubeadmConfigTemplateSpec() *kubeadmv1beta2.KubeadmConfigTemplateSpec {
	return &kubeadmv1beta2.KubeadmConfigTemplateSpec{
		Template: kubeadmv1beta2.KubeadmConfigTemplateResource{
			Spec: kubeadmv1beta2.KubeadmConfigSpec{
				JoinConfiguration: kubeadmv1beta2.JoinConfiguration{
					NodeRegistration: kubeadmv1beta2.NodeRegistrationOptions{
						CRISocket: ClusterCRISocket,
					},
				},
			},
		},
	}
}

// HelmChartProxyName returns the name of the HelmChartProxy resource installing the given addon on the
// Cluster environment identified by names.
func HelmChartProxyName(names ClusterResourceNames, addon *clv1alpha2.AddonSpec) string {
	return names.Cluster + StringSeparator + CanonicalName(addon.Name)
}

// HelmChartProxySpec returns the spec of the HelmChartProxy resource installing the given addon only on the
// workload cluster identified by selectorLabels.
func HelmChartProxySpec(addon *clv1alpha2.AddonSpec, selectorLabels map[string]string) *helmv1alpha1.HelmChartProxySpec {
	matchLabels := make(map[string]string, len(selectorLabels))
	for k, v := range selectorLabels {
		matchLabels[k] = v
	}

	return &helmv1alpha1.HelmChartProxySpec{
		ClusterSelector:  metav1.LabelSelector{MatchLabels: matchLabels},
		ReleaseName:      addon.Name,
		RepoURL:          addon.RepoURL,
		ChartName:        addon.ChartName,
		Version:          addon.Version,
		ReleaseNamespace: addon.Namespace,
		ValuesTemplate:   addon.ValuesTemplate,
	}
}
