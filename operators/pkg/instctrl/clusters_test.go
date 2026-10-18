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
	"context"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	apicommon "github.com/netgroup-polito/CrownLabs/operators/api/common"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/clcontext"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

func TestClusterFlavorNamespaceResolution(t *testing.T) {
	for _, tt := range []struct {
		name       string
		configured string
		want       string
		missing    bool
	}{
		{name: "default", want: forge.DefaultClusterFlavorNamespace},
		{name: "custom", configured: "shared-blueprints", want: "shared-blueprints"},
		{name: "no workspace fallback", want: forge.DefaultClusterFlavorNamespace, missing: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			if err := clv1alpha2.AddToScheme(scheme); err != nil {
				t.Fatal(err)
			}
			objects := []client.Object{&clv1alpha2.ClusterFlavor{
				ObjectMeta: metav1.ObjectMeta{Name: "basic", Namespace: "workspace"},
			}}
			if !tt.missing {
				objects = append(objects, &clv1alpha2.ClusterFlavor{
					ObjectMeta: metav1.ObjectMeta{Name: "basic", Namespace: tt.want},
				})
			}
			stop := errors.New("stop before provisioning cluster resources")
			var lookups []client.ObjectKey
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).
				WithInterceptorFuncs(interceptor.Funcs{
					Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						if _, ok := obj.(*clv1alpha2.ClusterFlavor); ok {
							lookups = append(lookups, key)
							return c.Get(ctx, key, obj, opts...)
						}
						return stop
					},
				}).Build()
			instance := &clv1alpha2.Instance{ObjectMeta: metav1.ObjectMeta{Name: "instance", Namespace: "tenant"}}
			template := &clv1alpha2.Template{ObjectMeta: metav1.ObjectMeta{Name: "template", Namespace: "workspace"}}
			env := &clv1alpha2.Environment{Name: "cluster", Image: "basic", Cluster: &clv1alpha2.ClusterSpec{}}
			ctx, _ := clctx.InstanceInto(context.Background(), instance)
			ctx, _ = clctx.TemplateInto(ctx, template)
			ctx, _ = clctx.EnvironmentInto(ctx, env)
			ctx = clctx.EnvironmentIndexInto(ctx, 0)
			recorder := record.NewFakeRecorder(1)
			r := &InstanceReconciler{
				Client: c, Scheme: scheme, EventsRecorder: recorder, ClusterFlavorNamespace: tt.configured,
			}
			err := r.EnforceClusterEnvironment(ctx)
			if tt.missing {
				if !apierrors.IsNotFound(err) {
					t.Fatalf("expected missing central flavor, got %v", err)
				}
				select {
				case event := <-recorder.Events:
					if !strings.Contains(event, tt.want+"/basic") {
						t.Fatalf("event does not identify the central flavor: %s", event)
					}
				default:
					t.Fatal("expected a missing flavor event")
				}
			} else if !errors.Is(err, stop) {
				t.Fatalf("expected central flavor lookup to succeed before provisioning, got %v", err)
			}
			want := []client.ObjectKey{{Namespace: tt.want, Name: "basic"}}
			if !reflect.DeepEqual(lookups, want) {
				t.Fatalf("flavor lookups: got %v, want %v", lookups, want)
			}
		})
	}
}

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
