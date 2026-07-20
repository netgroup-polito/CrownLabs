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
	"net/http"
	"net/http/httptest"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	imagelist "github.com/netgroup-polito/CrownLabs/operators/pkg/imagelist"
)

var _ = Describe("ProcessImageList", func() {
	It("skips images that only have 'latest' tag", func() {
		images := []map[string]interface{}{{
			"name": "prova",
			"tags": []string{"latest"},
		}}

		res := imagelist.ProcessImageList(images)
		Expect(res).To(BeEmpty())
	})

	It("keeps non-latest versions and drops latest", func() {
		images := []map[string]interface{}{{
			"name": "prova",
			"tags": []interface{}{"v1.1", "latest"},
		}}

		res := imagelist.ProcessImageList(images)
		Expect(res).To(HaveLen(1))
		Expect(res[0].Name).To(Equal("prova"))
		Expect(res[0].Versions).To(Equal([]string{"v1.1"}))
	})

	It("skips images with no tags field", func() {
		images := []map[string]interface{}{{
			"name": "no-tags",
		}}

		res := imagelist.ProcessImageList(images)
		Expect(res).To(BeEmpty())
	})
})

var _ = Describe("Requestor", func() {

	It("transforms Harbor registry responses into the expected image list structure", func() {
		catalogResponse := []map[string]interface{}{{
			"artifact_count": 1,
			"creation_time":  "2026-05-06T13:33:48.943Z",
			"id":             2,
			"name":           "crownlabs-containerdisks/ubuntu-server-base",
			"project_id":     3,
			"pull_count":     0,
			"update_time":    "2026-05-06T13:33:48.943Z",
		}}
		artifactResponse := []map[string]interface{}{{
			"tags": []interface{}{
				map[string]interface{}{"name": "v1.1"},
				map[string]interface{}{"name": "latest"},
			},
		}}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch r.URL.Path {
			case "/api/v2.0/projects/test-project/repositories":
				if r.URL.RawQuery != "page=1&page_size=100" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				_ = json.NewEncoder(w).Encode(catalogResponse)
			case "/api/v2.0/projects/test-project/repositories/ubuntu-server-base/artifacts":
				_ = json.NewEncoder(w).Encode(artifactResponse)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		oldProjectName, hadProjectName := imagelist.RequestersSharedData["harbor_project_name"]
		imagelist.RequestersSharedData["harbor_project_name"] = "test-project"
		defer func() {
			if hadProjectName {
				imagelist.RequestersSharedData["harbor_project_name"] = oldProjectName
			} else {
				delete(imagelist.RequestersSharedData, "harbor_project_name")
			}
		}()

		requestor := imagelist.NewHarborImageListRequestor(logr.Discard())
		initialized, err := requestor.Initialize("user", "pass", server.URL)
		Expect(err).NotTo(HaveOccurred())
		Expect(initialized).To(BeTrue())

		res, err := requestor.GetImageList(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(Equal([]map[string]interface{}{{
			"name": "ubuntu-server-base",
			"tags": []string{"v1.1"},
		}}))
	})
})

var _ = Describe("DefaultImageListSaver", func() {
	It("creates an ImageList with empty images while keeping registry and project base name", func() {
		scheme := runtime.NewScheme()
		err := clv1alpha1.AddToScheme(scheme)
		Expect(err).NotTo(HaveOccurred())

		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		ctx := context.Background()

		saver, err := imagelist.NewDefaultImageListSaver(ctx, "harbor-containerdisks-crownlabs", fakeClient, logr.Discard())
		Expect(err).NotTo(HaveOccurred())

		err = saver.CreateOrUpdateImageList(
			"harbor.ng.crownlabs.polito.it",
			"crownlabs-containerdisks",
			[]clv1alpha1.ImageListItem{},
		)
		Expect(err).NotTo(HaveOccurred())

		created := &clv1alpha1.ImageList{}
		err = fakeClient.Get(ctx, client.ObjectKey{Name: "harbor-containerdisks-crownlabs"}, created)
		Expect(err).NotTo(HaveOccurred())
		Expect(created.Spec.RegistryName).To(Equal("harbor.ng.crownlabs.polito.it"))
		Expect(created.Spec.ProjectBaseName).To(Equal("crownlabs-containerdisks"))
		Expect(created.Spec.Images).To(BeEmpty())
	})
})

var _ = Describe("InstanceSnapshot ImageList source", func() {
	It("creates an ImageList from completed InstanceSnapshots using the related job destination tag", func() {
		scheme := runtime.NewScheme()
		Expect(clv1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(clv1alpha2.AddToScheme(scheme)).To(Succeed())
		Expect(batchv1.AddToScheme(scheme)).To(Succeed())

		ctx := context.Background()
		namespace := "workspace-a"
		completedSnapshot := &clv1alpha2.InstanceSnapshot{
			ObjectMeta: metav1.ObjectMeta{Name: "snapshot-a", Namespace: namespace},
			Spec: clv1alpha2.InstanceSnapshotSpec{
				ImageName: "snapshot-image",
			},
			Status: clv1alpha2.InstanceSnapshotStatus{Phase: clv1alpha2.Completed},
		}
		processingSnapshot := &clv1alpha2.InstanceSnapshot{
			ObjectMeta: metav1.ObjectMeta{Name: "snapshot-b", Namespace: namespace},
			Spec: clv1alpha2.InstanceSnapshotSpec{
				ImageName: "ignored-image",
			},
			Status: clv1alpha2.InstanceSnapshotStatus{Phase: clv1alpha2.Processing},
		}
		snapshotJob := buildSnapshotJob(namespace, "snapshot-a", "harbor-core.harbor:80/tenant-a/snapshot-image:20260720t101010")

		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(completedSnapshot, processingSnapshot, snapshotJob).
			Build()

		items, err := imagelist.ProcessSingleRegistryConfigWithItems(ctx, &imagelist.RegistryConfig{
			Name:          "snapshot-source",
			Type:          "instancesnapshot",
			Namespace:     namespace,
			RegistryName:  "harbor.ng.crownlabs.polito.it",
			ImageListName: "snapshot-images",
			Project:       "tenant-a",
		}, fakeClient, logr.Discard())
		Expect(err).NotTo(HaveOccurred())
		Expect(items).To(Equal([]clv1alpha1.ImageListItem{{
			Name:     "snapshot-image",
			Versions: []string{"20260720t101010"},
		}}))

		created := &clv1alpha1.ImageList{}
		err = fakeClient.Get(ctx, client.ObjectKey{Name: "snapshot-images"}, created)
		Expect(err).NotTo(HaveOccurred())
		Expect(created.Spec.RegistryName).To(Equal("harbor.ng.crownlabs.polito.it"))
		Expect(created.Spec.ProjectBaseName).To(Equal("tenant-a"))
		Expect(created.Spec.Images).To(Equal(items))
	})

	It("deduplicates versions from completed InstanceSnapshots", func() {
		scheme := runtime.NewScheme()
		Expect(clv1alpha1.AddToScheme(scheme)).To(Succeed())
		Expect(clv1alpha2.AddToScheme(scheme)).To(Succeed())
		Expect(batchv1.AddToScheme(scheme)).To(Succeed())

		ctx := context.Background()
		namespace := "workspace-a"
		snapshotA := &clv1alpha2.InstanceSnapshot{
			ObjectMeta: metav1.ObjectMeta{Name: "snapshot-a", Namespace: namespace},
			Spec: clv1alpha2.InstanceSnapshotSpec{
				ImageName: "snapshot-image",
			},
			Status: clv1alpha2.InstanceSnapshotStatus{Phase: clv1alpha2.Completed},
		}
		snapshotB := &clv1alpha2.InstanceSnapshot{
			ObjectMeta: metav1.ObjectMeta{Name: "snapshot-b", Namespace: namespace},
			Spec: clv1alpha2.InstanceSnapshotSpec{
				ImageName: "snapshot-image",
			},
			Status: clv1alpha2.InstanceSnapshotStatus{Phase: clv1alpha2.Completed},
		}

		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(
				snapshotA,
				snapshotB,
				buildSnapshotJob(namespace, "snapshot-a", "harbor-core.harbor:80/tenant-a/snapshot-image:20260720t101010"),
				buildSnapshotJob(namespace, "snapshot-b", "harbor-core.harbor:80/tenant-a/snapshot-image:20260720t101010"),
			).
			Build()

		items, err := imagelist.ProcessSingleRegistryConfigWithItems(ctx, &imagelist.RegistryConfig{
			Name:          "snapshot-source",
			Type:          "instancesnapshot",
			Namespace:     namespace,
			RegistryName:  "harbor.ng.crownlabs.polito.it",
			ImageListName: "snapshot-images",
			Project:       "tenant-a",
		}, fakeClient, logr.Discard())
		Expect(err).NotTo(HaveOccurred())
		Expect(items).To(Equal([]clv1alpha1.ImageListItem{{
			Name:     "snapshot-image",
			Versions: []string{"20260720t101010"},
		}}))

		created := &clv1alpha1.ImageList{}
		err = fakeClient.Get(ctx, client.ObjectKey{Name: "snapshot-images"}, created)
		Expect(err).NotTo(HaveOccurred())
		Expect(created.Spec.ProjectBaseName).To(Equal("tenant-a"))
	})
})

func buildSnapshotJob(namespace, name, destination string) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name: "docker-pusher",
						Args: []string{"--destination=" + destination},
					}},
				},
			},
		},
	}
}
