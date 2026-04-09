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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

// Saver defines the interface for objects responsible for saving image lists to Kubernetes resources.
type Saver interface {
	// UpdateImageList updates the ImageList resource with the provided images from a specific registry.
	UpdateImageList(registryName string, images []clv1alpha1.ImageListItem) error
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

// UpdateImageList updates or creates the ImageList resource with the provided images.
func (s *DefaultImageListSaver) UpdateImageList(registryName string, images []clv1alpha1.ImageListItem) error {
	s.log.V(1).Info("updating ImageList", "registryName", registryName, "imageCount", len(images))

	resourceVersion, err := s.getImageListResourceVersion()
	if err != nil {
		return err
	}

	if resourceVersion != "" {
		return s.updateImageList(registryName, images, resourceVersion)
	}

	return s.createImageList(registryName, images)
}

// getImageListResourceVersion retrieves the resource version of the existing ImageList resource, or empty string if not found.
func (s *DefaultImageListSaver) getImageListResourceVersion() (string, error) {
	imageList := &clv1alpha1.ImageList{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.name,
		},
	}

	key := forge.NamespacedNameImageList(imageList)
	err := s.client.Get(s.ctx, key, imageList)

	if err != nil {
		if errors.IsNotFound(err) {
			s.log.V(1).Info("ImageList not found, will create", "name", s.name)
			return "", nil
		}
		s.log.Error(err, "failed to get ImageList", "name", s.name)
		return "", fmt.Errorf("failed to get ImageList: %w", err)
	}

	return imageList.GetResourceVersion(), nil
}

// createImageList creates a new ImageList resource in the cluster.
func (s *DefaultImageListSaver) createImageList(registryName string, images []clv1alpha1.ImageListItem) error {
	imageList := &clv1alpha1.ImageList{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.name,
		},
		Spec: clv1alpha1.ImageListSpec{
			RegistryName: registryName,
			Images:       images,
		},
	}

	err := s.client.Create(s.ctx, imageList)
	if err != nil {
		s.log.Error(err, "failed to create ImageList", "name", s.name)
		return fmt.Errorf("failed to create ImageList: %w", err)
	}

	s.log.Info("ImageList created successfully", "name", s.name, "registryName", registryName, "imageCount", len(images))
	return nil
}

// updateImageList updates an existing ImageList resource in the cluster.
func (s *DefaultImageListSaver) updateImageList(registryName string, images []clv1alpha1.ImageListItem, resourceVersion string) error {
	imageList := &clv1alpha1.ImageList{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.name,
		},
		Spec: clv1alpha1.ImageListSpec{
			RegistryName: registryName,
			Images:       images,
		},
	}
	imageList.SetResourceVersion(resourceVersion)

	err := s.client.Update(s.ctx, imageList)
	if err != nil {
		s.log.Error(err, "failed to update ImageList", "name", s.name)
		return fmt.Errorf("failed to update ImageList: %w", err)
	}

	s.log.Info("ImageList updated successfully", "name", s.name, "registryName", registryName, "imageCount", len(images))
	return nil
}
