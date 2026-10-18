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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"

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
		Expect(spec).To(HaveKey("template"))
		Expect(spec["template"]).To(HaveKey("spec"))
	})

	It("supports an existing DataVolume", func() {
		spec, err := forge.KubevirtMachineTemplateSpec("tenant", &clv1alpha2.WorkerSpec{
			Disk: clv1alpha2.WorkerDiskSpec{
				Type: clv1alpha2.WorkerDiskDataVolume,
				Name: "ubuntu-worker",
			},
		}, resources)
		Expect(err).NotTo(HaveOccurred())
		vmSpec := spec["template"].(map[string]interface{})["spec"].(map[string]interface{})["virtualMachineTemplate"].(map[string]interface{})["spec"].(map[string]interface{})
		volumes := vmSpec["template"].(map[string]interface{})["spec"].(map[string]interface{})["volumes"].([]interface{})
		Expect(volumes[0]).To(Equal(map[string]interface{}{
			"name":       "containervolume",
			"dataVolume": map[string]interface{}{"name": "ubuntu-worker"},
		}))
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
		vmSpec := spec["template"].(map[string]interface{})["spec"].(map[string]interface{})["virtualMachineTemplate"].(map[string]interface{})["spec"].(map[string]interface{})
		Expect(vmSpec).To(HaveKey("dataVolumeTemplates"))
		Expect(vmSpec["template"].(map[string]interface{})["spec"].(map[string]interface{})["volumes"]).To(ContainElement(
			map[string]interface{}{
				"name":       "containervolume",
				"dataVolume": map[string]interface{}{"name": "workervolume"},
			},
		))
	})

	It("rejects incomplete disk configurations", func() {
		_, err := forge.KubevirtMachineTemplateSpec("tenant", &clv1alpha2.WorkerSpec{
			Disk: clv1alpha2.WorkerDiskSpec{Type: clv1alpha2.WorkerDiskDataVolume},
		}, resources)
		Expect(err).To(MatchError(`workers.disk.name is required for type "dataVolume"`))
	})
})
