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
	"os"
	"sync"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
)

// RegistryConfig contains the configuration for a single registry endpoint.
type RegistryConfig struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	URL           string `json:"url"`
	Advertised    string `json:"advertised"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	ImageListName string `json:"imageListName"`
	Project       string `json:"project,omitempty"` // Only for Harbor
}

// UpdateResult represents the result of updating a single image list.
type UpdateResult struct {
	ImageListName string
	Items         []clv1alpha1.ImageListItem
	Error         error
}

// UpdaterOptions holds configuration for the image list updater.
type UpdaterOptions struct {
	ConfigFilePath string
	Interval       int // interval in seconds
}

// updater manages periodic image list updates.
type updater struct {
	k8sClient client.Client
	log       logr.Logger
	mu        sync.Mutex
	updating  bool
	options   UpdaterOptions
}

var globalUpdater *updater

// LoadRegistriesConfig loads registry configuration from file.
func LoadRegistriesConfig(filePath string) ([]RegistryConfig, error) {
	configData, err := os.ReadFile(filePath) // #nosec G304: path is from controlled configuration
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file %s: %w", filePath, err)
	}

	var config []RegistryConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse registries configuration from file: %w", err)
	}

	if len(config) == 0 {
		return nil, fmt.Errorf("no registries configured in file %s", filePath)
	}

	return config, nil
}

// Initialize initializes the image list updater with the given configuration.
func Initialize(k8sClient client.Client, log logr.Logger, options UpdaterOptions) error {
	if globalUpdater != nil {
		return fmt.Errorf("image list updater already initialized")
	}

	globalUpdater = &updater{
		k8sClient: k8sClient,
		log:       log.WithName("imagelist-updater"),
		options:   options,
	}

	return nil
}

// Update performs a periodic update of all configured image lists.
func Update(ctx context.Context) error {
	if globalUpdater == nil {
		return fmt.Errorf("image list updater not initialized")
	}

	return globalUpdater.update(ctx)
}

// update executes the update logic with mutex protection.
func (u *updater) update(ctx context.Context) error {
	log := u.log

	// Prevent concurrent updates
	if !u.mu.TryLock() {
		log.Info("update already in progress, skipping")
		return nil
	}
	defer u.mu.Unlock()

	u.updating = true
	defer func() { u.updating = false }()

	log.Info("starting image list update")

	config, err := LoadRegistriesConfig(u.options.ConfigFilePath)
	if err != nil {
		log.Error(err, "failed to load registries configuration")
		return err
	}

	successCount := 0
	errorCount := 0

	for i := range config {
		result, err := ProcessSingleRegistryConfigWithItems(ctx, &config[i], u.k8sClient, log)
		if err != nil {
			log.Error(err, "failed to process registry", "registry_name", config[i].Name, "imagelist_name", config[i].ImageListName)
			errorCount++
		} else {
			log.Info("successfully updated image list", "registry_name", config[i].Name, "imagelist_name", config[i].ImageListName, "items_count", len(result))
			successCount++
		}
	}

	log.Info("image list update completed", "success_count", successCount, "error_count", errorCount)

	if errorCount > 0 {
		return fmt.Errorf("update completed with %d errors out of %d registries", errorCount, len(config))
	}

	return nil
}

// ProcessSingleRegistryConfig processes a single registry configuration.
func ProcessSingleRegistryConfig(ctx context.Context, regConfig *RegistryConfig, k8sClient client.Client, log logr.Logger) error {
	var requestor Requestor

	switch regConfig.Type {
	case "docker":
		requestor = NewDockerImageListRequestor(log.WithName(regConfig.Name).WithName("dockerRequestor"))
	case "harbor":
		if regConfig.Project == "" {
			return fmt.Errorf("project is required for Harbor registry")
		}
		RequestersSharedData["harbor_project_name"] = regConfig.Project
		requestor = NewHarborImageListRequestor(log.WithName(regConfig.Name).WithName("harborRequestor"))
	default:
		return fmt.Errorf("unsupported registry type: %s", regConfig.Type)
	}

	if initResult, err := requestor.Initialize(regConfig.Username, regConfig.Password, regConfig.URL); !initResult || err != nil {
		return fmt.Errorf("failed to initialize %s image list requestor: %w", regConfig.Type, err)
	}

	log.Info("updating ImageList CR", "name", regConfig.ImageListName, "registry", regConfig.Name)

	imageListSaver, err := NewDefaultImageListSaver(ctx, regConfig.ImageListName, k8sClient, log.WithName(regConfig.Name).WithName("saver"))
	if err != nil {
		return fmt.Errorf("failed to initialize the image list saver: %w", err)
	}

	imageListUpdater := NewUpdater([]Requestor{requestor}, regConfig.ImageListName, imageListSaver, regConfig.Advertised, log.WithName(regConfig.Name).WithName("updater"))
	if err := imageListUpdater.Update(ctx); err != nil {
		return fmt.Errorf("failed to update the ImageList resource: %w", err)
	}

	return nil
}

// ProcessSingleRegistryConfigWithItems processes a single registry configuration and returns the updated items.
func ProcessSingleRegistryConfigWithItems(ctx context.Context, regConfig *RegistryConfig, k8sClient client.Client, log logr.Logger) ([]clv1alpha1.ImageListItem, error) {
	var requestor Requestor

	switch regConfig.Type {
	case "docker":
		requestor = NewDockerImageListRequestor(log.WithName(regConfig.Name).WithName("dockerRequestor"))
	case "harbor":
		if regConfig.Project == "" {
			return nil, fmt.Errorf("project is required for Harbor registry")
		}
		RequestersSharedData["harbor_project_name"] = regConfig.Project
		requestor = NewHarborImageListRequestor(log.WithName(regConfig.Name).WithName("harborRequestor"))
	default:
		return nil, fmt.Errorf("unsupported registry type: %s", regConfig.Type)
	}

	if initResult, err := requestor.Initialize(regConfig.Username, regConfig.Password, regConfig.URL); !initResult || err != nil {
		return nil, fmt.Errorf("failed to initialize %s image list requestor: %w", regConfig.Type, err)
	}

	log.Info("updating ImageList CR", "name", regConfig.ImageListName, "registry", regConfig.Name)

	imageListSaver, err := NewDefaultImageListSaver(ctx, regConfig.ImageListName, k8sClient, log.WithName(regConfig.Name).WithName("saver"))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the image list saver: %w", err)
	}

	imageListUpdater := NewUpdater([]Requestor{requestor}, regConfig.ImageListName, imageListSaver, regConfig.Advertised, log.WithName(regConfig.Name).WithName("updater"))
	if err := imageListUpdater.Update(ctx); err != nil {
		return nil, fmt.Errorf("failed to update the ImageList resource: %w", err)
	}

	// Retrieve the updated items from the requestor
	images, err := requestor.GetImageList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated images: %w", err)
	}

	items := ProcessImageList(images)
	return items, nil
}
