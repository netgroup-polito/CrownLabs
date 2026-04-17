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

// Package main contains the entry point for the crownlabs-image-list updater.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/textlogger"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	imagelist "github.com/netgroup-polito/CrownLabs/operators/pkg/imagelist"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// RegistryConfig contains the configuration for a single registry endpoint
type RegistryConfig struct {
	Name          string `yaml:"name" json:"name"`
	Type          string `yaml:"type" json:"type"`
	URL           string `yaml:"url" json:"url"`
	Advertised    string `yaml:"advertised" json:"advertised"`
	Username      string `yaml:"username" json:"username"`
	Password      string `yaml:"password" json:"password"`
	ImageListName string `yaml:"imageListName" json:"imageListName"`
	Project       string `yaml:"project" json:"project"`
}

func main() {
	var (
		configMapName      string
		configMapNamespace string
	)

	flag.StringVar(&configMapName, "configmap-name", "registry-configuration", "Name of the ConfigMap containing registry configurations")
	flag.StringVar(&configMapNamespace, "configmap-namespace", "crownlabs-system", "Namespace of the ConfigMap containing registry configurations")
	klog.InitFlags(nil)
	flag.Parse()

	log := textlogger.NewLogger(textlogger.NewConfig()).WithName("imageList")
	ctx := context.Background()

	k8sClient, err := utils.NewK8sClient()
	if err != nil {
		log.Error(err, "Failed to initialize K8s client for ImageList savers")
		os.Exit(1)
	}

	if err := processRegistriesConfigFromConfigMap(ctx, configMapName, configMapNamespace, k8sClient, log); err != nil {
		log.Error(err, "Failed to process registries configuration from ConfigMap")
		os.Exit(1)
	}

	log.Info("ImageList resources updated successfully")
}

// processSingleRegistry handles the legacy single registry mode
func processSingleRegistry(ctx context.Context, registryType, registryURL, username, password,
	advertisedRegistryName, imageListName, harborProject string,
	k8sClient client.Client, logger logr.Logger) error {

	if registryURL == "" || imageListName == "" || username == "" || password == "" {
		return fmt.Errorf("missing required parameters: registry-url, image-list-name, registry-username, registry-password")
	}

	var requestor imagelist.Requestor

	switch registryType {
	case "docker":
		requestor = imagelist.NewDockerImageListRequestor(logger.WithName("crownlabs-imagelists-updater").WithName("dockerRequestor"))
	case "harbor":
		if harborProject == "" {
			return fmt.Errorf("harbor-project is required for Harbor registries")
		}
		imagelist.RequestersSharedData["harbor_project_name"] = harborProject
		requestor = imagelist.NewHarborImageListRequestor(logger.WithName("crownlabs-imagelists-updater").WithName("harborRequestor"))
	default:
		return fmt.Errorf("unsupported registry type: %s", registryType)
	}

	if initResult, err := requestor.Initialize(username, password, registryURL); !initResult || err != nil {
		return fmt.Errorf("failed to initialize %s image list requestor: %w", registryType, err)
	}

	logger.Info("Target ImageList CR name", "name", imageListName, "registry_type", registryType)

	imageListSaver, err := imagelist.NewDefaultImageListSaver(ctx, imageListName, k8sClient, logger.WithName("crownlabs-imagelists-updater").WithName("defaultSaver"))
	if err != nil {
		return fmt.Errorf("failed to initialize the default image list saver: %w", err)
	}

	imageListUpdater := imagelist.NewUpdater([]imagelist.Requestor{requestor}, imageListName, imageListSaver, advertisedRegistryName, logger.WithName("crownlabs-imagelists-updater"))
	if err := imageListUpdater.Update(ctx); err != nil {
		return fmt.Errorf("failed to update the ImageList resource: %w", err)
	}

	return nil
}

// processRegistriesConfigFromConfigMap reads the configuration from a Kubernetes ConfigMap
func processRegistriesConfigFromConfigMap(ctx context.Context, cmName, cmNamespace string, k8sClient client.Client, logger logr.Logger) error {
	// Read configuration from ConfigMap
	configMap := &corev1.ConfigMap{}
	key := types.NamespacedName{Name: cmName, Namespace: cmNamespace}

	if err := k8sClient.Get(ctx, key, configMap); err != nil {
		return fmt.Errorf("failed to read ConfigMap %s/%s: %w", cmNamespace, cmName, err)
	}

	// Get the configuration data from the ConfigMap
	configData, ok := configMap.Data["registries"]
	if !ok {
		return fmt.Errorf("ConfigMap %s/%s does not contain 'registries' key", cmNamespace, cmName)
	}

	logger.Info("Raw registries data from ConfigMap", "data", configData)

	var config []RegistryConfig
	if err := yaml.Unmarshal([]byte(configData), &config); err != nil {
		return fmt.Errorf("failed to parse registries configuration from ConfigMap: %w", err)
	}

	if len(config) == 0 {
		return fmt.Errorf("no registries configured in ConfigMap %s/%s", cmNamespace, cmName)
	}

	logger.Info("Processing registries from ConfigMap", "configmap", fmt.Sprintf("%s/%s", cmNamespace, cmName), "registry_count", len(config))
	for i, reg := range config {
		logger.Info("Parsed registry", "index", i, "name", reg.Name, "type", reg.Type, "imageListName", reg.ImageListName)
	}

	// Process each registry
	for _, regConfig := range config {
		if err := processSingleRegistryConfig(ctx, regConfig, k8sClient, logger); err != nil {
			logger.Error(err, "Failed to process registry", "registry_name", regConfig.Name)
			// Continue processing other registries even if one fails
			continue
		}
		logger.Info("Successfully processed registry", "registry_name", regConfig.Name, "imageListName", regConfig.ImageListName)
	}

	return nil
}

// processSingleRegistryConfig handles a single registry from the configuration file
func processSingleRegistryConfig(ctx context.Context, regConfig RegistryConfig, k8sClient client.Client, log logr.Logger) error {
	var requestor imagelist.Requestor

	switch regConfig.Type {
	case "docker":
		requestor = imagelist.NewDockerImageListRequestor(log.WithName(regConfig.Name).WithName("dockerRequestor"))
	case "harbor":
		if regConfig.Project == "" {
			return fmt.Errorf("project is required for Harbor registry")
		}
		imagelist.RequestersSharedData["harbor_project_name"] = regConfig.Project
		requestor = imagelist.NewHarborImageListRequestor(log.WithName(regConfig.Name).WithName("harborRequestor"))
	default:
		return fmt.Errorf("unsupported registry type: %s", regConfig.Type)
	}

	if initResult, err := requestor.Initialize(regConfig.Username, regConfig.Password, regConfig.URL); !initResult || err != nil {
		return fmt.Errorf("failed to initialize %s image list requestor: %w", regConfig.Type, err)
	}

	log.Info("Updating ImageList CR", "name", regConfig.ImageListName, "registry", regConfig.Name)

	imageListSaver, err := imagelist.NewDefaultImageListSaver(ctx, regConfig.ImageListName, k8sClient, log.WithName(regConfig.Name).WithName("saver"))
	if err != nil {
		return fmt.Errorf("failed to initialize the image list saver: %w", err)
	}

	imageListUpdater := imagelist.NewUpdater([]imagelist.Requestor{requestor}, regConfig.ImageListName, imageListSaver, regConfig.Advertised, log.WithName(regConfig.Name).WithName("updater"))
	if err := imageListUpdater.Update(ctx); err != nil {
		return fmt.Errorf("failed to update the ImageList resource: %w", err)
	}

	return nil
}
