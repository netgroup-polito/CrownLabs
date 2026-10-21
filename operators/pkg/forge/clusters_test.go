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

package forge_test

import (
	"encoding/json"
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"

	apicommon "github.com/netgroup-polito/CrownLabs/operators/api/common"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("KubeVirt worker machine templates", func() {
	resources := &clv1alpha2.EnvironmentResources{
		ResourceSpec: apicommon.ResourceSpec{
			CPU:    2,
			Memory: resource.MustParse("2Gi"),
		},
	}

	It("supports a container disk", func() {
		spec, err := forge.KubevirtMachineTemplateSpec("tenant", &clv1alpha2.WorkerSpec{
			Disk: clv1alpha2.WorkerDiskSpec{
				Type:  clv1alpha2.WorkerDiskContainerDisk,
				Image: "registry.example/worker:latest",
			},
		}, resources)
		Expect(err).NotTo(HaveOccurred())
		vmSpec := spec.Template.Spec.VirtualMachineTemplate.Spec
		Expect(vmSpec.DataVolumeTemplates).To(BeEmpty())
		Expect(vmSpec.Template.Spec.Volumes[0].ContainerDisk.Image).To(Equal("registry.example/worker:latest"))
	})

	It("supports an existing DataVolume", func() {
		spec, err := forge.KubevirtMachineTemplateSpec("tenant", &clv1alpha2.WorkerSpec{
			Disk: clv1alpha2.WorkerDiskSpec{
				Type: clv1alpha2.WorkerDiskDataVolume,
				Name: "ubuntu-worker",
			},
		}, resources)
		Expect(err).NotTo(HaveOccurred())
		vmSpec := spec.Template.Spec.VirtualMachineTemplate.Spec
		Expect(vmSpec.DataVolumeTemplates).To(BeEmpty())
		Expect(vmSpec.Template.Spec.Volumes[0].Name).To(Equal("containervolume"))
		Expect(vmSpec.Template.Spec.Volumes[0].DataVolume.Name).To(Equal("ubuntu-worker"))
	})

	It("supports importing a cloud image through CDI", func() {
		spec, err := forge.KubevirtMachineTemplateSpec("tenant", &clv1alpha2.WorkerSpec{
			Disk: clv1alpha2.WorkerDiskSpec{
				Type:             clv1alpha2.WorkerDiskCloudImage,
				URL:              "https://example.org/ubuntu.img",
				StorageClassName: "ceph-block",
				StorageSize:      resource.MustParse("20Gi"),
			},
		}, resources)
		Expect(err).NotTo(HaveOccurred())
		vmSpec := spec.Template.Spec.VirtualMachineTemplate.Spec
		Expect(vmSpec.DataVolumeTemplates).To(HaveLen(1))
		dv := vmSpec.DataVolumeTemplates[0]
		Expect(dv.Name).To(Equal("workervolume"))
		Expect(dv.Spec.Source.HTTP.URL).To(Equal("https://example.org/ubuntu.img"))
		Expect(*dv.Spec.Storage.StorageClassName).To(Equal("ceph-block"))
		Expect(dv.Spec.Storage.Resources.Requests.Storage().String()).To(Equal("20Gi"))
		Expect(vmSpec.Template.Spec.Volumes[0].DataVolume.Name).To(Equal("workervolume"))
	})

	It("rejects incomplete disk configurations", func() {
		_, err := forge.KubevirtMachineTemplateSpec("tenant", &clv1alpha2.WorkerSpec{
			Disk: clv1alpha2.WorkerDiskSpec{Type: clv1alpha2.WorkerDiskDataVolume},
		}, resources)
		Expect(err).To(MatchError(`workers.disk.name is required for type "dataVolume"`))
	})
})

func TestTypedClusterSpecs(t *testing.T) {
	names := forge.ClusterResourceNames{
		Cluster: "cluster", ControlPlane: "control-plane",
		Infrastructure: "infrastructure", MachineDeployment: "workers",
	}
	component := clv1alpha2.ComponentResources{
		CPU: resource.MustParse("250m"), Memory: resource.MustParse("256Mi"),
	}
	flavor := &clv1alpha2.ClusterFlavor{
		Spec: clv1alpha2.ClusterFlavorSpec{
			Version: "v1.35.0",
			ControlPlane: clv1alpha2.ControlPlaneSpec{
				Replicas: 3, DataStoreName: "etcd",
				APIServerResources: component, SchedulerResources: component,
				ControllerManagerResources: component, KonnectivityResources: component,
			},
		},
	}
	resources := &clv1alpha2.EnvironmentResources{
		ResourceSpec: apicommon.ResourceSpec{CPU: 2, Memory: resource.MustParse("2Gi")},
	}
	addon := &clv1alpha2.AddonSpec{
		Name: "cni", RepoURL: "https://charts.example.org", ChartName: "cni",
		Version: "1.2.3", Namespace: "kube-system", ValuesTemplate: "enabled: true",
	}
	workerSpec, err := forge.KubevirtMachineTemplateSpec("tenant", &clv1alpha2.WorkerSpec{
		Disk: clv1alpha2.WorkerDiskSpec{Type: clv1alpha2.WorkerDiskContainerDisk, Image: "registry.example.org/worker:v1"},
	}, resources)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		spec interface{}
		want string
	}{
		{
			name: "cluster", spec: forge.ClusterSpec(names),
			want: `{
				"controlPlaneRef":{"apiGroup":"controlplane.cluster.x-k8s.io","kind":"KamajiControlPlane","name":"control-plane"},
				"infrastructureRef":{"apiGroup":"infrastructure.cluster.x-k8s.io","kind":"KubevirtCluster","name":"infrastructure"},
				"clusterNetwork":{"services":{"cidrBlocks":["10.96.0.0/16"]},"pods":{"cidrBlocks":["10.244.0.0/16"]}}
			}`,
		},
		{
			name: "control plane", spec: forge.KamajiControlPlaneSpec(flavor),
			want: `{
				"replicas":3,"version":"v1.35.0","dataStoreName":"etcd","dataStoreSchema":"etcd",
				"network":{"serviceType":"ClusterIP"},
				"apiServer":{"resources":{"limits":{"cpu":"250m","memory":"256Mi"}}},
				"scheduler":{"resources":{"limits":{"cpu":"250m","memory":"256Mi"}}},
				"controllerManager":{"resources":{"limits":{"cpu":"250m","memory":"256Mi"}}},
				"addons":{"coreDNS":{},"kubeProxy":{},"konnectivity":{"server":{"port":8132,"resources":{"limits":{"cpu":"250m","memory":"256Mi"}}}}}
			}`,
		},
		{
			name: "workers", spec: forge.MachineDeploymentSpec(names.Cluster, names, flavor.Spec.Version, 2),
			want: `{
				"clusterName":"cluster","replicas":2,"selector":{},
				"template":{"spec":{"clusterName":"cluster","version":"v1.35.0",
					"infrastructureRef":{"apiGroup":"infrastructure.cluster.x-k8s.io","kind":"KubevirtMachineTemplate","name":"workers"},
					"bootstrap":{"configRef":{"apiGroup":"bootstrap.cluster.x-k8s.io","kind":"KubeadmConfigTemplate","name":"workers"}}
				}}
			}`,
		},
		{
			name: "worker VM", spec: workerSpec,
			want: `{"template":{"spec":{
				"virtualMachineBootstrapCheck":{"checkStrategy":"ssh"},
				"virtualMachineTemplate":{"metadata":{"namespace":"tenant"},"spec":{
					"runStrategy":"Always","template":{"spec":{
						"domain":{"cpu":{"cores":2},"memory":{"guest":"2Gi"},"devices":{"disks":[{"name":"containervolume","disk":{"bus":"virtio"}}]}},
						"volumes":[{"name":"containervolume","containerDisk":{"image":"registry.example.org/worker:v1"}}]
					}}
				}}
			}}}`,
		},
		{
			name: "bootstrap", spec: forge.KubeadmConfigTemplateSpec(),
			want: `{"template":{"spec":{"joinConfiguration":{"nodeRegistration":{"criSocket":"unix:///var/run/containerd/containerd.sock"}}}}}`,
		},
		{
			name: "addon", spec: forge.HelmChartProxySpec(addon, map[string]string{"cluster": "cluster"}),
			want: `{"clusterSelector":{"matchLabels":{"cluster":"cluster"}},"releaseName":"cni","repoURL":"https://charts.example.org","chartName":"cni","version":"1.2.3","namespace":"kube-system","valuesTemplate":"enabled: true"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tt.spec)
			if err != nil {
				t.Fatal(err)
			}
			// Normalize integer representations before comparing the JSON fields.
			data, err := json.Marshal(spec)
			if err != nil {
				t.Fatal(err)
			}
			var got, want interface{}
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Fatal(err)
			}
			assertClusterJSONFields(t, "", got, want)
		})
	}
}

func assertClusterJSONFields(t *testing.T, path string, got, want interface{}) {
	t.Helper()
	switch expected := want.(type) {
	case map[string]interface{}:
		actual, ok := got.(map[string]interface{})
		if !ok {
			t.Fatalf("%s: expected object, got %T", path, got)
		}
		for key, value := range expected {
			field, found := actual[key]
			if !found {
				t.Errorf("%s.%s: missing field", path, key)
				continue
			}
			assertClusterJSONFields(t, path+"."+key, field, value)
		}
	case []interface{}:
		actual, ok := got.([]interface{})
		if !ok || len(actual) != len(expected) {
			t.Fatalf("%s: expected %v, got %v", path, want, got)
		}
		for i := range expected {
			assertClusterJSONFields(t, path+"[]", actual[i], expected[i])
		}
	default:
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: expected %v, got %v", path, want, got)
		}
	}
}

func TestClusterSpecOptionalFields(t *testing.T) {
	spec, err := runtime.DefaultUnstructuredConverter.ToUnstructured(forge.KubeadmConfigTemplateSpec())
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"template":{"spec":{"joinConfiguration":{"nodeRegistration":{"criSocket":"unix:///var/run/containerd/containerd.sock"}}}}}`
	if string(data) != want {
		t.Fatalf("unexpected bootstrap fields: %s", data)
	}
	labels := map[string]string{"cluster": "one"}
	addon := forge.HelmChartProxySpec(&clv1alpha2.AddonSpec{Name: "cni"}, labels)
	addon.ClusterSelector.MatchLabels["cluster"] = "two"
	if labels["cluster"] != "one" {
		t.Fatal("addon selector aliases the input labels")
	}
	addonMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(addon)
	if err != nil {
		t.Fatal(err)
	}
	if _, found := addonMap["valuesTemplate"]; found {
		t.Fatal("empty addon values must be omitted")
	}
}
