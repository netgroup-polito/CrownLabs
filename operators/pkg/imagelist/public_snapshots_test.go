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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	imagelist "github.com/netgroup-polito/CrownLabs/operators/pkg/imagelist"
)

var _ = Describe("Public snapshot ImageList source", func() {
	const namespace = "public-images"
	var scheme *runtime.Scheme
	var config *imagelist.RegistryConfig

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(clv1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(clv1alpha2.AddToScheme(scheme)).To(Succeed())
		config = &imagelist.RegistryConfig{
			Name: "snapshots", Type: "public-snapshots", Namespace: namespace, ImageListName: "public-snapshots",
		}
	})

	It("publishes the artifact reference without Docker tags or access to snapshot jobs", func() {
		const resourceName = "desktop-20260925-103000"
		snapshot := buildPublicSnapshot(namespace, resourceName, "desktop", clv1alpha2.SnapshotPhaseCompleted, namespace, resourceName)
		// The version comes from the resource name, not the API server's creation time.
		snapshot.SetCreationTimestamp(metav1.Date(2026, 9, 25, 8, 30, 5, 0, time.UTC))
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(snapshot).Build()
		// Old registry settings must not turn local PVC references into Docker images.
		config.RegistryName = "harbor.example.org"
		config.Project = "old-project"
		items, err := imagelist.ProcessSingleRegistryConfigWithItems(context.Background(), config, fakeClient, logr.Discard())
		Expect(err).NotTo(HaveOccurred())
		Expect(items).To(Equal([]clv1alpha1.ImageListItem{{
			Name: "desktop", Versions: []string{"20260925-103000"},
		}}))
		created := &clv1alpha1.ImageList{}
		Expect(fakeClient.Get(context.Background(), client.ObjectKey{Name: config.ImageListName}, created)).To(Succeed())
		Expect(created.Spec.RegistryName).To(Equal(namespace))
		Expect(created.Spec.ProjectBaseName).To(BeEmpty())
		Expect(created.Spec.Images).To(Equal(items))
	})

	It("groups resource name prefixes and sorts images alphabetically and versions newest first", func() {
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			buildPublicSnapshot(namespace, "ubuntu-lab-20260924-103000", "ubuntu-lab", clv1alpha2.SnapshotPhaseCompleted, namespace, "ubuntu-lab-20260924-103000"),
			buildPublicSnapshot(namespace, "ubuntu-lab-20260925-090001", "ubuntu-lab", clv1alpha2.SnapshotPhaseCompleted, namespace, "ubuntu-lab-20260925-090001"),
			buildPublicSnapshot(namespace, "ubuntu-lab-20260924-235959", "", clv1alpha2.SnapshotPhaseCompleted, namespace, "ubuntu-lab-20260924-235959"),
			buildPublicSnapshot(namespace, "alpine-20260925-090001", "ubuntu-lab", clv1alpha2.SnapshotPhaseCompleted, namespace, "alpine-20260925-090001"),
		).Build()
		items, err := imagelist.ProcessSingleRegistryConfigWithItems(context.Background(), config, fakeClient, logr.Discard())
		Expect(err).NotTo(HaveOccurred())
		Expect(items).To(Equal([]clv1alpha1.ImageListItem{
			{Name: "alpine", Versions: []string{"20260925-090001"}},
			{Name: "ubuntu-lab", Versions: []string{"20260925-090001", "20260924-235959", "20260924-103000"}},
		}))
	})

	DescribeTable("extracts only the final valid date and time", func(resourceName, imageName, version string) {
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			buildPublicSnapshot(namespace, resourceName, "", clv1alpha2.SnapshotPhaseCompleted, namespace, resourceName),
		).Build()
		items, err := imagelist.ProcessSingleRegistryConfigWithItems(context.Background(), config, fakeClient, logr.Discard())
		Expect(err).NotTo(HaveOccurred())
		Expect(items).To(Equal([]clv1alpha1.ImageListItem{{Name: imageName, Versions: []string{version}}}))
	},
		Entry("one-character image name and midnight", "a-20260101-000000", "a", "20260101-000000"),
		Entry("leap day", "desktop-20240229-235959", "desktop", "20240229-235959"),
		Entry("date-like image name", "desktop-20240101-120000-20260925-103000", "desktop-20240101-120000", "20260925-103000"),
	)

	DescribeTable("preserves an exact unversioned artifact reference when the name cannot be split", func(resourceName, artifactName string) {
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			buildPublicSnapshot(namespace, resourceName, "desktop", clv1alpha2.SnapshotPhaseCompleted, namespace, artifactName),
		).Build()
		items, err := imagelist.ProcessSingleRegistryConfigWithItems(context.Background(), config, fakeClient, logr.Discard())
		Expect(err).NotTo(HaveOccurred())
		Expect(items).To(Equal([]clv1alpha1.ImageListItem{{Name: artifactName, Versions: []string{}}}))
	},
		Entry("legacy name", "desktop", "desktop"),
		Entry("different artifact name", "desktop-20260925-103000", "actual-volume"),
		Entry("invalid leap day", "desktop-20260229-103000", "desktop-20260229-103000"),
		Entry("invalid month", "desktop-20261301-103000", "desktop-20261301-103000"),
		Entry("invalid time", "desktop-20260925-240000", "desktop-20260925-240000"),
		Entry("incomplete time", "desktop-20260925-1030", "desktop-20260925-1030"),
		Entry("non-numeric time", "desktop-20260925-10ab00", "desktop-20260925-10ab00"),
		Entry("no image prefix", "20260925-103000", "20260925-103000"),
		Entry("custom suffix", "desktop-20260925-103000-abcde", "desktop-20260925-103000-abcde"),
	)

	It("keeps an unversioned artifact alongside dated versions without duplicating choices", func() {
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			buildPublicSnapshot(namespace, "desktop", "desktop", clv1alpha2.SnapshotPhaseCompleted, namespace, "desktop"),
			buildPublicSnapshot(namespace, "alias", "desktop", clv1alpha2.SnapshotPhaseCompleted, namespace, "desktop"),
			buildPublicSnapshot(namespace, "desktop-20260925-103000", "desktop", clv1alpha2.SnapshotPhaseCompleted, namespace, "desktop-20260925-103000"),
		).Build()
		items, err := imagelist.ProcessSingleRegistryConfigWithItems(context.Background(), config, fakeClient, logr.Discard())
		Expect(err).NotTo(HaveOccurred())
		Expect(items).To(Equal([]clv1alpha1.ImageListItem{{Name: "desktop", Versions: []string{"20260925-103000", ""}}}))
	})

	It("reads typed snapshot artifacts and phases over HTTP", func() {
		const resourceName = "desktop-20260925-103000"
		snapshot := buildPublicSnapshot(namespace, resourceName, "desktop", clv1alpha2.SnapshotPhaseCompleted, namespace, resourceName)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet || r.URL.Path != "/apis/crownlabs.polito.it/v1alpha2/namespaces/"+namespace+"/instancesnapshots" {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"apiVersion": clv1alpha2.GroupVersion.String(), "kind": "InstanceSnapshotList",
				"items": []interface{}{snapshot},
			})
		}))
		defer server.Close()
		apiScheme := runtime.NewScheme()
		Expect(clv1alpha2.AddToScheme(apiScheme)).To(Succeed())
		mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{clv1alpha2.GroupVersion})
		mapper.Add(clv1alpha2.GroupVersion.WithKind("InstanceSnapshot"), meta.RESTScopeNamespace)
		k8sClient, err := client.New(&rest.Config{Host: server.URL}, client.Options{Scheme: apiScheme, Mapper: mapper})
		Expect(err).NotTo(HaveOccurred())
		source, err := imagelist.NewPublicSnapshotImageListSource(k8sClient, namespace, logr.Discard())
		Expect(err).NotTo(HaveOccurred())
		images, err := source.GetImageList(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(images).To(Equal([]clv1alpha1.ImageListItem{{
			Name: "desktop", Versions: []string{"20260925-103000"},
		}}))
	})

	It("excludes unfinished, deleting, incomplete and out-of-namespace artifacts", func() {
		deleting := buildPublicSnapshot(namespace, "deleting", "", clv1alpha2.SnapshotPhaseCompleted, namespace, "deleting")
		now := metav1.Now()
		deleting.SetDeletionTimestamp(&now)
		deleting.SetFinalizers([]string{"snapshot-cleanup"})
		legacy := buildPublicSnapshot(namespace, "legacy", "", clv1alpha2.SnapshotPhaseCompleted, namespace, "legacy")
		legacy.Status.Artifact = clv1alpha2.SnapshotArtifact{}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			deleting, legacy,
			buildPublicSnapshot(namespace, "pending", "", clv1alpha2.SnapshotPhasePending, namespace, "pending"),
			buildPublicSnapshot(namespace, "processing", "", clv1alpha2.SnapshotPhaseProcessing, namespace, "processing"),
			buildPublicSnapshot(namespace, "failed", "", clv1alpha2.SnapshotPhaseFailed, namespace, "failed"),
			buildPublicSnapshot(namespace, "uninitialized", "", "", namespace, "uninitialized"),
			buildPublicSnapshot(namespace, "missing-name", "", clv1alpha2.SnapshotPhaseCompleted, namespace, ""),
			buildPublicSnapshot(namespace, "missing-namespace", "", clv1alpha2.SnapshotPhaseCompleted, "", "artifact"),
			buildPublicSnapshot(namespace, "foreign-artifact", "", clv1alpha2.SnapshotPhaseCompleted, "tenant-private", "artifact"),
			buildPublicSnapshot("tenant-private", "private", "", clv1alpha2.SnapshotPhaseCompleted, "tenant-private", "private"),
		).Build()
		items, err := imagelist.ProcessSingleRegistryConfigWithItems(context.Background(), config, fakeClient, logr.Discard())
		Expect(err).NotTo(HaveOccurred())
		Expect(items).To(BeEmpty())
	})

	It("removes deleted versions and clears the catalog when the last snapshot is deleted", func() {
		older := buildPublicSnapshot(namespace, "desktop-20260924-103000", "desktop", clv1alpha2.SnapshotPhaseCompleted, namespace, "desktop-20260924-103000")
		newer := buildPublicSnapshot(namespace, "desktop-20260925-103000", "desktop", clv1alpha2.SnapshotPhaseCompleted, namespace, "desktop-20260925-103000")
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(older, newer).Build()
		Expect(imagelist.ProcessSingleRegistryConfig(context.Background(), config, fakeClient, logr.Discard())).To(Succeed())
		created := &clv1alpha1.ImageList{}
		Expect(fakeClient.Get(context.Background(), client.ObjectKey{Name: config.ImageListName}, created)).To(Succeed())
		Expect(created.Spec.Images).To(Equal([]clv1alpha1.ImageListItem{{Name: "desktop", Versions: []string{"20260925-103000", "20260924-103000"}}}))
		Expect(fakeClient.Delete(context.Background(), newer)).To(Succeed())
		items, err := imagelist.ProcessSingleRegistryConfigWithItems(context.Background(), config, fakeClient, logr.Discard())
		Expect(err).NotTo(HaveOccurred())
		Expect(items).To(Equal([]clv1alpha1.ImageListItem{{Name: "desktop", Versions: []string{"20260924-103000"}}}))
		Expect(fakeClient.Delete(context.Background(), older)).To(Succeed())
		items, err = imagelist.ProcessSingleRegistryConfigWithItems(context.Background(), config, fakeClient, logr.Discard())
		Expect(err).NotTo(HaveOccurred())
		Expect(items).To(BeEmpty())
		Expect(fakeClient.Get(context.Background(), client.ObjectKey{Name: config.ImageListName}, created)).To(Succeed())
		Expect(created.Spec.Images).To(Equal([]clv1alpha1.ImageListItem{}))
		Expect(created.Spec.RegistryName).To(Equal(namespace))
	})

	It("preserves the existing catalog when snapshots cannot be listed", func() {
		original := &clv1alpha1.ImageList{
			ObjectMeta: metav1.ObjectMeta{Name: config.ImageListName},
			Spec:       clv1alpha1.ImageListSpec{Images: []clv1alpha1.ImageListItem{{Name: "existing", Versions: []string{"v1"}}}},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(original).WithInterceptorFuncs(interceptor.Funcs{
			List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
				return fmt.Errorf("snapshot access denied")
			},
		}).Build()
		Expect(imagelist.ProcessSingleRegistryConfig(context.Background(), config, fakeClient, logr.Discard())).To(MatchError(ContainSubstring("snapshot access denied")))
		stored := &clv1alpha1.ImageList{}
		Expect(fakeClient.Get(context.Background(), client.ObjectKey{Name: config.ImageListName}, stored)).To(Succeed())
		Expect(stored.Spec).To(Equal(original.Spec))
	})

	DescribeTable("preserves registry catalogs when adding public snapshots", func(registryType string) {
		responses := map[string]interface{}{
			"/v2/_catalog":                            map[string]interface{}{"repositories": []string{"desktop"}},
			"/v2/desktop/tags/list":                   map[string]interface{}{"name": "desktop", "tags": []string{"v1", "latest"}},
			"/api/v2.0/projects/project/repositories": []interface{}{map[string]interface{}{"name": "project/desktop"}},
			"/api/v2.0/projects/project/repositories/desktop/artifacts": []interface{}{
				map[string]interface{}{"tags": []interface{}{map[string]interface{}{"name": "v1"}, map[string]interface{}{"name": "latest"}}},
			},
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response, ok := responses[r.URL.Path]
			if !ok {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()
		oldProject, hadProject := imagelist.RequestersSharedData["harbor_project_name"]
		defer func() {
			if hadProject {
				imagelist.RequestersSharedData["harbor_project_name"] = oldProject
			} else {
				delete(imagelist.RequestersSharedData, "harbor_project_name")
			}
		}()
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			buildPublicSnapshot(namespace, "snapshot", "desktop", clv1alpha2.SnapshotPhaseCompleted, namespace, "artifact"),
		).Build()
		registryConfig := &imagelist.RegistryConfig{
			Name: "registry", Type: registryType, URL: server.URL, RegistryName: "registry.example.org",
			Project: "project", ImageListName: "registry-images",
		}
		items, err := imagelist.ProcessSingleRegistryConfigWithItems(context.Background(), registryConfig, fakeClient, logr.Discard())
		Expect(err).NotTo(HaveOccurred())
		Expect(items).To(Equal([]clv1alpha1.ImageListItem{{Name: "desktop", Versions: []string{"v1"}}}))
		Expect(imagelist.ProcessSingleRegistryConfig(context.Background(), config, fakeClient, logr.Discard())).To(Succeed())
		stored := &clv1alpha1.ImageList{}
		Expect(fakeClient.Get(context.Background(), client.ObjectKey{Name: registryConfig.ImageListName}, stored)).To(Succeed())
		Expect(stored.Spec).To(Equal(clv1alpha1.ImageListSpec{
			RegistryName: "registry.example.org", ProjectBaseName: "project", Images: items,
		}))
	}, Entry("Docker", "docker"), Entry("Harbor", "harbor"))

	It("requires an explicit namespace and a Kubernetes client", func() {
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		config.Namespace = ""
		Expect(imagelist.ProcessSingleRegistryConfig(context.Background(), config, fakeClient, logr.Discard())).To(MatchError(ContainSubstring("namespace is required")))
		source, err := imagelist.NewPublicSnapshotImageListSource(nil, namespace, logr.Discard())
		Expect(source).To(BeNil())
		Expect(err).To(MatchError("kubernetes client is required"))
	})
})

// buildPublicSnapshot uses the InstanceSnapshot types from the snapshot controller API.
func buildPublicSnapshot(namespace, name, imageName string, phase clv1alpha2.SnapshotPhase, artifactNamespace, artifactName string) *clv1alpha2.InstanceSnapshot {
	return &clv1alpha2.InstanceSnapshot{
		TypeMeta:   metav1.TypeMeta{APIVersion: clv1alpha2.GroupVersion.String(), Kind: "InstanceSnapshot"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: clv1alpha2.InstanceSnapshotSpec{
			ImageName: imageName, Environment: "desktop",
			Instance:    clv1alpha2.GenericRef{Name: "source-vm", Namespace: "tenant-author"},
			Description: "A reusable image", Tenant: clv1alpha2.GenericRef{Name: "author"},
		},
		Status: clv1alpha2.InstanceSnapshotStatus{
			Phase: phase,
			Artifact: clv1alpha2.SnapshotArtifact{
				DataVolumeRef: clv1alpha2.GenericRef{Name: artifactName, Namespace: artifactNamespace},
				VolumeSize:    resource.MustParse("10Gi"),
			},
		},
	}
}
