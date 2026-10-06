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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

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
	GVKCluster = schema.GroupVersionKind{Group: "cluster.x-k8s.io", Version: "v1beta2", Kind: "Cluster"}
	// GVKKubevirtCluster is the GVK of the KubeVirt infrastructure provider Cluster resource.
	GVKKubevirtCluster = schema.GroupVersionKind{Group: "infrastructure.cluster.x-k8s.io", Version: "v1alpha1", Kind: "KubevirtCluster"}
	// GVKKamajiControlPlane is the GVK of the Kamaji control plane provider resource.
	GVKKamajiControlPlane = schema.GroupVersionKind{Group: "controlplane.cluster.x-k8s.io", Version: "v1alpha1", Kind: "KamajiControlPlane"}
	// GVKMachineDeployment is the GVK of the Cluster API MachineDeployment resource.
	GVKMachineDeployment = schema.GroupVersionKind{Group: "cluster.x-k8s.io", Version: "v1beta2", Kind: "MachineDeployment"}
	// GVKKubevirtMachineTemplate is the GVK of the KubeVirt infrastructure provider machine template resource.
	GVKKubevirtMachineTemplate = schema.GroupVersionKind{Group: "infrastructure.cluster.x-k8s.io", Version: "v1alpha1", Kind: "KubevirtMachineTemplate"}
	// GVKKubeadmConfigTemplate is the GVK of the kubeadm bootstrap provider config template resource.
	GVKKubeadmConfigTemplate = schema.GroupVersionKind{Group: "bootstrap.cluster.x-k8s.io", Version: "v1beta2", Kind: "KubeadmConfigTemplate"}
	// GVKHelmChartProxy is the GVK of the Cluster API Add-on Provider for Helm chart proxy resource.
	GVKHelmChartProxy = schema.GroupVersionKind{Group: "addons.cluster.x-k8s.io", Version: "v1alpha1", Kind: "HelmChartProxy"}
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
func ClusterSpec(names ClusterResourceNames) map[string]interface{} {
	return map[string]interface{}{
		"controlPlaneRef": map[string]interface{}{
			"apiGroup": GVKKamajiControlPlane.Group,
			"kind":     GVKKamajiControlPlane.Kind,
			"name":     names.ControlPlane,
		},
		"infrastructureRef": map[string]interface{}{
			"apiGroup": GVKKubevirtCluster.Group,
			"kind":     GVKKubevirtCluster.Kind,
			"name":     names.Infrastructure,
		},
		"clusterNetwork": map[string]interface{}{
			"services": map[string]interface{}{
				"cidrBlocks": []interface{}{ClusterServiceCIDR},
			},
			"pods": map[string]interface{}{
				"cidrBlocks": []interface{}{ClusterPodCIDR},
			},
		},
	}
}

// KubevirtClusterAnnotations returns the annotations to be set on the KubevirtCluster resource, marking it
// as externally managed by Kamaji.
func KubevirtClusterAnnotations() map[string]string {
	return map[string]string{ClusterKamajiManagedByAnnotation: ClusterKamajiManagedByValue}
}

// componentResourcesSpec returns the resources.limits stanza for a Kamaji control plane component.
func componentResourcesSpec(res *clv1alpha2.ComponentResources) map[string]interface{} {
	return map[string]interface{}{
		"resources": map[string]interface{}{
			"limits": map[string]interface{}{
				"cpu":    res.CPU.String(),
				"memory": res.Memory.String(),
			},
		},
	}
}

// KamajiControlPlaneSpec returns the spec of the KamajiControlPlane resource, forged from the ClusterFlavor
// referenced by the Cluster environment.
func KamajiControlPlaneSpec(flavor *clv1alpha2.ClusterFlavor) map[string]interface{} {
	controlPlane := flavor.Spec.ControlPlane
	return map[string]interface{}{
		"replicas":        int64(controlPlane.Replicas),
		"version":         flavor.Spec.Version,
		"dataStoreName":   controlPlane.DataStoreName,
		"dataStoreSchema": ClusterKamajiDataStoreSchema,
		"network": map[string]interface{}{
			"serviceType": "ClusterIP",
		},
		"apiServer":         componentResourcesSpec(&controlPlane.APIServerResources),
		"scheduler":         componentResourcesSpec(&controlPlane.SchedulerResources),
		"controllerManager": componentResourcesSpec(&controlPlane.ControllerManagerResources),
		"addons": map[string]interface{}{
			"coreDNS":   map[string]interface{}{},
			"kubeProxy": map[string]interface{}{},
			"konnectivity": map[string]interface{}{
				"server": map[string]interface{}{
					"port": 8132,
					"resources": map[string]interface{}{
						"limits": map[string]interface{}{
							"cpu":    controlPlane.KonnectivityResources.CPU.String(),
							"memory": controlPlane.KonnectivityResources.Memory.String(),
						},
					},
				},
			},
		},
	}
}

// MachineDeploymentSpec returns the spec of the MachineDeployment resource hosting the worker nodes of the
// Cluster environment.
func MachineDeploymentSpec(clusterName string, names ClusterResourceNames, version string, workersCount int) map[string]interface{} {
	return map[string]interface{}{
		"clusterName": clusterName,
		"replicas":    int64(workersCount),
		"selector": map[string]interface{}{
			"matchLabels": map[string]interface{}{},
		},
		"template": map[string]interface{}{
			"spec": map[string]interface{}{
				"clusterName": clusterName,
				"version":     version,
				"infrastructureRef": map[string]interface{}{
					"apiGroup": GVKKubevirtMachineTemplate.Group,
					"kind":     GVKKubevirtMachineTemplate.Kind,
					"name":     names.MachineDeployment,
				},
				"bootstrap": map[string]interface{}{
					"configRef": map[string]interface{}{
						"apiGroup": GVKKubeadmConfigTemplate.Group,
						"kind":     GVKKubeadmConfigTemplate.Kind,
						"name":     names.MachineDeployment,
					},
				},
			},
		},
	}
}

// KubevirtMachineTemplateSpec returns the spec of the KubevirtMachineTemplate resource describing the VM
// hosting each worker node of the Cluster environment. The worker image comes from the ClusterFlavor, while
// the amount of CPU/Memory assigned to each worker node comes from the Environment resources.
func KubevirtMachineTemplateSpec(namespace, workerImage string, resources *clv1alpha2.EnvironmentResources) map[string]interface{} {
	return map[string]interface{}{
		"template": map[string]interface{}{
			"spec": map[string]interface{}{
				"virtualMachineBootstrapCheck": map[string]interface{}{
					"checkStrategy": "ssh",
				},
				"virtualMachineTemplate": map[string]interface{}{
					"metadata": map[string]interface{}{
						"namespace": namespace,
					},
					"spec": map[string]interface{}{
						"runStrategy": "Always",
						"template": map[string]interface{}{
							"spec": map[string]interface{}{
								"domain": map[string]interface{}{
									"cpu": map[string]interface{}{
										"cores": resources.CPU,
									},
									"memory": map[string]interface{}{
										"guest": resources.Memory.String(),
									},
									"devices": map[string]interface{}{
										"disks": []interface{}{
											map[string]interface{}{
												"name": ClusterContainerVolumeName,
												"disk": map[string]interface{}{
													"bus": ClusterVirtioBus,
												},
											},
										},
									},
								},
								"volumes": []interface{}{
									map[string]interface{}{
										"name": ClusterContainerVolumeName,
										"containerDisk": map[string]interface{}{
											"image": workerImage,
										},
									},
								},
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
func KubeadmConfigTemplateSpec() map[string]interface{} {
	return map[string]interface{}{
		"template": map[string]interface{}{
			"spec": map[string]interface{}{
				"joinConfiguration": map[string]interface{}{
					"nodeRegistration": map[string]interface{}{
						"criSocket": ClusterCRISocket,
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
func HelmChartProxySpec(addon *clv1alpha2.AddonSpec, selectorLabels map[string]string) map[string]interface{} {
	matchLabels := make(map[string]interface{}, len(selectorLabels))
	for k, v := range selectorLabels {
		matchLabels[k] = v
	}

	spec := map[string]interface{}{
		"clusterSelector": map[string]interface{}{
			"matchLabels": matchLabels,
		},
		"releaseName": addon.Name,
		"repoURL":     addon.RepoURL,
		"chartName":   addon.ChartName,
		"version":     addon.Version,
		"namespace":   addon.Namespace,
	}
	if addon.ValuesTemplate != "" {
		spec["valuesTemplate"] = addon.ValuesTemplate
	}
	return spec
}
