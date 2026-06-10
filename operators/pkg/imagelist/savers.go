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

// Package imagelist contains the image list saver logic.
package imagelist

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
)

// Saver defines the interface for objects responsible for saving image lists to Kubernetes resources.
type Saver interface {
	// CreateOrUpdateImageList creates a new ImageList resource or updates an existing one with the provided images from a specific registry.
	CreateOrUpdateImageList(registryName, projectBaseName string, images []clv1alpha1.ImageListItem) error
}

// RegisteredSavers holds the list of all registered image list savers.
var RegisteredSavers = []Saver{}

// DefaultImageListSaver saves the list of images retrieved from a Docker registry as a Kubernetes ImageList resource.
type DefaultImageListSaver struct {
	name   string
	client client.Client
	log    logr.Logger
	ctx    context.Context
}

// NewDefaultImageListSaver creates a new DefaultImageListSaver instance.
func NewDefaultImageListSaver(ctx context.Context, name string, k8sClient client.Client, log logr.Logger) (*DefaultImageListSaver, error) {
	return &DefaultImageListSaver{
		name:   name,
		client: k8sClient,
		log:    log,
		ctx:    ctx,
	}, nil
}

// CreateOrUpdateImageList creates a new ImageList or updates an existing one with the provided images.
func (s *DefaultImageListSaver) CreateOrUpdateImageList(registryName, projectBaseName string, images []clv1alpha1.ImageListItem) error {
	s.log.V(1).Info("creating or updating ImageList", "registryName", registryName, "projectBaseName", projectBaseName, "imageCount", len(images))

	persistedImages := images
	if len(images) == 0 {
		s.log.Info("no images found; persisting empty ImageList", "name", s.name, "registryName", registryName, "projectBaseName", projectBaseName)
		persistedImages = []clv1alpha1.ImageListItem{}
	}

	// Try to get existing ImageList
	imageList := &clv1alpha1.ImageList{}
	key := types.NamespacedName{Name: s.name}
	err := s.client.Get(s.ctx, key, imageList)
	if err != nil {
		if kerrors.IsNotFound(err) {
			// Create new ImageList
			newList := &clv1alpha1.ImageList{
				ObjectMeta: metav1.ObjectMeta{
					Name: s.name,
				},
				Spec: clv1alpha1.ImageListSpec{
					RegistryName:    registryName,
					Images:          persistedImages,
					ProjectBaseName: projectBaseName,
				},
			}

			if err := s.client.Create(s.ctx, newList); err != nil {
				s.log.Error(err, "failed to create ImageList", "name", s.name)
				return fmt.Errorf("failed to create ImageList: %w", err)
			}

			s.log.Info("ImageList created successfully", "name", s.name, "registryName", registryName, "projectBaseName", projectBaseName, "imageCount", len(images))
			return nil
		}

		s.log.Error(err, "failed to get ImageList", "name", s.name)
		return fmt.Errorf("failed to get ImageList: %w", err)
	}

	// Update existing ImageList
	imageList.Spec = clv1alpha1.ImageListSpec{
		RegistryName:    registryName,
		Images:          persistedImages,
		ProjectBaseName: projectBaseName,
	}

	if err := s.client.Update(s.ctx, imageList); err != nil {
		s.log.Error(err, "failed to update ImageList", "name", s.name)
		return fmt.Errorf("failed to update ImageList: %w", err)
	}

	s.log.Info("ImageList updated successfully", "name", s.name, "registryName", registryName, "projectBaseName", projectBaseName, "imageCount", len(images))
	return nil
}
