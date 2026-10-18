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

package instctrl

import (
	"math"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"

	apicommon "github.com/netgroup-polito/CrownLabs/operators/api/common"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

func TestClusterSpecToUnstructured(t *testing.T) {
	names := forge.ClusterResourceNames{
		Cluster: "cluster", ControlPlane: "control-plane",
		Infrastructure: "infrastructure", MachineDeployment: "workers",
	}
	resources := &clv1alpha2.EnvironmentResources{
		ResourceSpec: apicommon.ResourceSpec{CPU: 2, Memory: resource.MustParse("2Gi")},
	}
	tests := []struct {
		name string
		spec interface{}
	}{
		{"cluster", forge.ClusterSpec(names)},
		{"control plane", forge.KamajiControlPlaneSpec(&clv1alpha2.ClusterFlavor{})},
		{"workers", forge.MachineDeploymentSpec(names.Cluster, names, "v1.35.0", 2)},
		{"worker VM", forge.KubevirtMachineTemplateSpec("tenant", "worker:v1", resources)},
		{"bootstrap", forge.KubeadmConfigTemplateSpec()},
		{"addon", forge.HelmChartProxySpec(&clv1alpha2.AddonSpec{Name: "cni"}, map[string]string{"cluster": "cluster"})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := clusterSpecToUnstructured(tt.spec)
			if err != nil {
				t.Fatal(err)
			}
			copied := runtime.DeepCopyJSON(got)
			data, err := json.Marshal(tt.spec)
			if err != nil {
				t.Fatal(err)
			}
			var want map[string]interface{}
			if err := json.Unmarshal(data, &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(copied, want) {
				t.Fatalf("converted spec differs from its JSON representation:\ngot: %#v\nwant: %#v", copied, want)
			}
			if tt.name == "worker VM" {
				cores, found, err := unstructured.NestedInt64(got,
					"template", "spec", "virtualMachineTemplate", "spec", "template", "spec", "domain", "cpu", "cores")
				if err != nil || !found || cores != 2 {
					t.Fatalf("expected CPU cores as int64(2), got %d (found=%t, err=%v)", cores, found, err)
				}
			}
		})
	}
}

func TestClusterSpecToUnstructuredIntegers(t *testing.T) {
	spec := &struct {
		Unsigned uint64 `json:"unsigned"`
		Signed   int64  `json:"signed"`
	}{Unsigned: math.MaxInt64, Signed: math.MaxInt64}
	got, err := clusterSpecToUnstructured(spec)
	if err != nil {
		t.Fatal(err)
	}
	runtime.DeepCopyJSON(got)
	for _, field := range []string{"unsigned", "signed"} {
		value, found, err := unstructured.NestedInt64(got, field)
		if err != nil || !found || value != math.MaxInt64 {
			t.Fatalf("%s: integer precision lost: %d (found=%t, err=%v)", field, value, found, err)
		}
	}
}

func TestClusterSpecToUnstructuredErrors(t *testing.T) {
	for _, spec := range []interface{}{
		&struct{ Unsupported chan int }{},
		&struct{ Invalid float64 }{Invalid: math.NaN()},
		[]string{"not", "an", "object"},
	} {
		if _, err := clusterSpecToUnstructured(spec); err == nil {
			t.Fatalf("expected conversion error for %#v", spec)
		}
	}
}
