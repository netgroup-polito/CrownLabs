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

// Package imagelist contains the image list requestor logic.
package imagelist

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// PublicSnapshotImageListSource retrieves local images from completed snapshots.
type PublicSnapshotImageListSource struct {
	k8sClient client.Client
	namespace string
	log       logr.Logger
}

// NewPublicSnapshotImageListSource creates a source for a public snapshot catalog.
func NewPublicSnapshotImageListSource(k8sClient client.Client, namespace string, log logr.Logger) (*PublicSnapshotImageListSource, error) {
	if k8sClient == nil {
		return nil, fmt.Errorf("kubernetes client is required")
	}
	if namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}
	return &PublicSnapshotImageListSource{k8sClient: k8sClient, namespace: namespace, log: log}, nil
}

// GetImageList retrieves DataVolume references from completed InstanceSnapshots.
func (r *PublicSnapshotImageListSource) GetImageList(ctx context.Context) ([]clv1alpha1.ImageListItem, error) {
	snapshots := &clv1alpha2.InstanceSnapshotList{}
	if err := r.k8sClient.List(ctx, snapshots, client.InNamespace(r.namespace)); err != nil {
		return nil, fmt.Errorf("failed to list InstanceSnapshots in namespace %s: %w", r.namespace, err)
	}

	versionsByName := make(map[string][]string)
	for i := range snapshots.Items {
		snapshot := &snapshots.Items[i]
		if snapshot.Status.Phase != clv1alpha2.SnapshotPhaseCompleted || !snapshot.DeletionTimestamp.IsZero() {
			continue
		}

		ref := snapshot.Status.Artifact.DataVolumeRef
		if ref.Name == "" || ref.Namespace != r.namespace {
			r.log.V(1).Info("skipping snapshot without an artifact in the catalog namespace", "name", snapshot.Name)
			continue
		}
		name, version := publicSnapshotNameAndVersion(snapshot.Name, ref.Name)
		versionsByName[name] = append(versionsByName[name], version)
	}

	images := make([]clv1alpha1.ImageListItem, 0, len(versionsByName))
	for name, versions := range versionsByName {
		// Fixed-width timestamps sort newest first without changing the time zone
		// or value encoded by the snapshot creator. An unversioned choice sorts last.
		sort.Sort(sort.Reverse(sort.StringSlice(versions)))
		versions = slices.Compact(versions)
		if len(versions) == 1 && versions[0] == "" {
			versions = []string{}
		}
		images = append(images, clv1alpha1.ImageListItem{Name: name, Versions: versions})
	}
	sort.Slice(images, func(i, j int) bool { return images[i].Name < images[j].Name })
	return images, nil
}

// publicSnapshotNameAndVersion recognizes <image-name>-YYYYMMDD-HHmmss from PR #1200.
// Only split when the result reconstructs the actual artifact reference from PR #1179.
// Older or custom names remain usable as unversioned artifacts.
func publicSnapshotNameAndVersion(snapshotName, artifactName string) (name, version string) {
	const timestampLayout = "20060102-150405"
	separator := len(snapshotName) - len(timestampLayout) - 1
	if snapshotName != artifactName || separator <= 0 || snapshotName[separator] != '-' {
		return artifactName, ""
	}
	version = snapshotName[separator+1:]
	if _, err := time.Parse(timestampLayout, version); err != nil {
		return artifactName, ""
	}
	return snapshotName[:separator], version
}

// updatePublicSnapshotImageList keeps the local artifact pipeline separate from registry tag processing.
func updatePublicSnapshotImageList(ctx context.Context, config *RegistryConfig, k8sClient client.Client, log logr.Logger) ([]clv1alpha1.ImageListItem, error) {
	source, err := NewPublicSnapshotImageListSource(k8sClient, config.Namespace, log.WithName(config.Name))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize public snapshot source: %w", err)
	}
	images, err := source.GetImageList(ctx)
	if err != nil {
		return nil, err
	}
	saver, err := NewDefaultImageListSaver(ctx, config.ImageListName, k8sClient, log.WithName(config.Name).WithName("saver"))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the image list saver: %w", err)
	}
	if err := saver.CreateOrUpdateImageList(config.Namespace, "", images); err != nil {
		return nil, fmt.Errorf("failed to update public snapshot ImageList: %w", err)
	}
	return images, nil
}
