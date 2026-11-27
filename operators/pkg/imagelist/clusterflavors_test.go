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

package imagelist_test

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/imagelist"
)

var _ = Describe("ClusterFlavor ImageList source", func() {
	const namespace = "crownlabs-clusterflavors"
	var scheme *runtime.Scheme
	var config *imagelist.RegistryConfig
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(clv1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(clv1alpha2.AddToScheme(scheme)).To(Succeed())
		config = &imagelist.RegistryConfig{
			Name: "clusterflavors", Type: "clusterflavors",
			Namespace: namespace, ImageListName: "local-clusterflavors",
		}
	})

	It("publishes sorted exact names without tags or registry settings", func() {
		deleting := buildClusterFlavor(namespace, "deleting", "v1.35.0")
		now := metav1.Now()
		deleting.DeletionTimestamp = &now
		deleting.Finalizers = []string{"test.crownlabs.polito.it/cleanup"}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			buildClusterFlavor(namespace, "production", "v1.35.0"),
			buildClusterFlavor(namespace, "basic-20261018-120000", "v1.34.0"),
			buildClusterFlavor("private", "private-flavor", "v1.35.0"),
			deleting,
		).Build()
		config.RegistryName = "ignored.example.com"
		config.Project = "ignored-project"
		items, err := imagelist.ProcessSingleRegistryConfigWithItems(ctx, config, fakeClient, logr.Discard())
		Expect(err).NotTo(HaveOccurred())
		Expect(items).To(Equal([]clv1alpha1.ImageListItem{
			{Name: "basic-20261018-120000", Versions: []string{}},
			{Name: "production", Versions: []string{}},
		}))
		stored := &clv1alpha1.ImageList{}
		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: config.ImageListName}, stored)).To(Succeed())
		Expect(stored.Namespace).To(BeEmpty())
		Expect(stored.Spec).To(Equal(clv1alpha1.ImageListSpec{
			RegistryName: namespace, Images: items,
		}))
	})

	It("updates the catalog and clears stale entries after all flavors are deleted", func() {
		flavor := buildClusterFlavor(namespace, "basic", "v1.35.0")
		existing := &clv1alpha1.ImageList{
			ObjectMeta: metav1.ObjectMeta{Name: config.ImageListName},
			Spec: clv1alpha1.ImageListSpec{
				RegistryName: "old-registry", ProjectBaseName: "old-project",
				Images: []clv1alpha1.ImageListItem{{Name: "stale", Versions: []string{"v1"}}},
			},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(flavor, existing).Build()
		Expect(imagelist.ProcessSingleRegistryConfig(ctx, config, fakeClient, logr.Discard())).To(Succeed())
		stored := &clv1alpha1.ImageList{}
		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: config.ImageListName}, stored)).To(Succeed())
		Expect(stored.Spec).To(Equal(clv1alpha1.ImageListSpec{
			RegistryName: namespace,
			Images:       []clv1alpha1.ImageListItem{{Name: "basic", Versions: []string{}}},
		}))
		Expect(fakeClient.Delete(ctx, flavor)).To(Succeed())
		items, err := imagelist.ProcessSingleRegistryConfigWithItems(ctx, config, fakeClient, logr.Discard())
		Expect(err).NotTo(HaveOccurred())
		Expect(items).To(Equal([]clv1alpha1.ImageListItem{}))
		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: config.ImageListName}, stored)).To(Succeed())
		Expect(stored.Spec.Images).To(Equal([]clv1alpha1.ImageListItem{}))
	})

	It("creates an empty catalog when the namespace has no flavors", func() {
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		Expect(imagelist.ProcessSingleRegistryConfig(ctx, config, fakeClient, logr.Discard())).To(Succeed())
		stored := &clv1alpha1.ImageList{}
		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: config.ImageListName}, stored)).To(Succeed())
		Expect(stored.Spec.RegistryName).To(Equal(namespace))
		Expect(stored.Spec.Images).To(Equal([]clv1alpha1.ImageListItem{}))
	})

	It("preserves the existing catalog when listing flavors fails", func() {
		existing := &clv1alpha1.ImageList{
			ObjectMeta: metav1.ObjectMeta{Name: config.ImageListName},
			Spec: clv1alpha1.ImageListSpec{
				RegistryName: namespace, Images: []clv1alpha1.ImageListItem{{Name: "existing", Versions: []string{}}},
			},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existing).
			WithInterceptorFuncs(interceptor.Funcs{
				List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
					return fmt.Errorf("flavor access denied")
				},
			}).Build()
		items, err := imagelist.ProcessSingleRegistryConfigWithItems(ctx, config, fakeClient, logr.Discard())
		Expect(err).To(MatchError(ContainSubstring("flavor access denied")))
		Expect(items).To(BeNil())
		stored := &clv1alpha1.ImageList{}
		Expect(fakeClient.Get(ctx, client.ObjectKey{Name: config.ImageListName}, stored)).To(Succeed())
		Expect(stored.Spec).To(Equal(existing.Spec))
	})

	DescribeTable("propagates catalog persistence failures", func(operation string) {
		builder := fake.NewClientBuilder().WithScheme(scheme)
		if operation == "create" {
			builder = builder.WithInterceptorFuncs(interceptor.Funcs{
				Create: func(context.Context, client.WithWatch, client.Object, ...client.CreateOption) error {
					return fmt.Errorf("catalog write denied")
				},
			})
		} else {
			builder = builder.WithObjects(&clv1alpha1.ImageList{
				ObjectMeta: metav1.ObjectMeta{Name: config.ImageListName},
			}).WithInterceptorFuncs(interceptor.Funcs{
				Update: func(context.Context, client.WithWatch, client.Object, ...client.UpdateOption) error {
					return fmt.Errorf("catalog write denied")
				},
			})
		}
		Expect(imagelist.ProcessSingleRegistryConfig(ctx, config, builder.Build(), logr.Discard())).
			To(MatchError(ContainSubstring("catalog write denied")))
	}, Entry("create", "create"), Entry("update", "update"))

	It("requires an explicit namespace and a Kubernetes client", func() {
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		config.Namespace = ""
		Expect(imagelist.ProcessSingleRegistryConfig(ctx, config, fakeClient, logr.Discard())).
			To(MatchError(ContainSubstring("namespace is required")))
		items, err := imagelist.ProcessSingleRegistryConfigWithItems(ctx, config, fakeClient, logr.Discard())
		Expect(items).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring("namespace is required")))
		source, err := imagelist.NewClusterFlavorImageListSource(nil, namespace)
		Expect(source).To(BeNil())
		Expect(err).To(MatchError("kubernetes client is required"))
	})
})

func buildClusterFlavor(namespace, name, version string) *clv1alpha2.ClusterFlavor {
	return &clv1alpha2.ClusterFlavor{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec:       clv1alpha2.ClusterFlavorSpec{Version: version},
	}
}
