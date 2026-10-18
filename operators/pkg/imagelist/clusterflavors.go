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

package imagelist

import (
	"context"
	"fmt"
	"sort"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// ClusterFlavorImageListSource retrieves cluster blueprints from a namespace.
type ClusterFlavorImageListSource struct {
	k8sClient client.Client
	namespace string
}

// NewClusterFlavorImageListSource creates a source for a ClusterFlavor catalog.
func NewClusterFlavorImageListSource(k8sClient client.Client, namespace string) (*ClusterFlavorImageListSource, error) {
	if k8sClient == nil {
		return nil, fmt.Errorf("kubernetes client is required")
	}
	if namespace == "" {
		return nil, fmt.Errorf("namespace is required")
	}
	return &ClusterFlavorImageListSource{k8sClient: k8sClient, namespace: namespace}, nil
}

// GetImageList returns exact resource names without registry tag processing.
func (r *ClusterFlavorImageListSource) GetImageList(ctx context.Context) ([]clv1alpha1.ImageListItem, error) {
	flavors := &clv1alpha2.ClusterFlavorList{}
	if err := r.k8sClient.List(ctx, flavors, client.InNamespace(r.namespace)); err != nil {
		return nil, fmt.Errorf("failed to list ClusterFlavors in namespace %s: %w", r.namespace, err)
	}

	images := make([]clv1alpha1.ImageListItem, 0, len(flavors.Items))
	for i := range flavors.Items {
		flavor := &flavors.Items[i]
		if !flavor.DeletionTimestamp.IsZero() {
			continue
		}
		images = append(images, clv1alpha1.ImageListItem{Name: flavor.Name, Versions: []string{}})
	}
	sort.Slice(images, func(i, j int) bool { return images[i].Name < images[j].Name })
	return images, nil
}

func updateClusterFlavorImageList(ctx context.Context, config *RegistryConfig, k8sClient client.Client, log logr.Logger) ([]clv1alpha1.ImageListItem, error) {
	source, err := NewClusterFlavorImageListSource(k8sClient, config.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize ClusterFlavor source: %w", err)
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
		return nil, fmt.Errorf("failed to update ClusterFlavor ImageList: %w", err)
	}
	return images, nil
}
