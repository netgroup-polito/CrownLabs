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
	"os"

	"k8s.io/klog/v2"
	"k8s.io/klog/v2/textlogger"

	imagelist "github.com/netgroup-polito/CrownLabs/operators/pkg/imagelist"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

func main() {
	var (
		advertisedRegistryName string
		imageListName          string
		registryURL            string
		registryUsername       string
		registryPassword       string
	)

	flag.StringVar(&advertisedRegistryName, "advertised-registry-name", "", "Hostname of the Docker registry as advertised to consumers")
	flag.StringVar(&imageListName, "image-list-name", "", "Base name/prefix for ImageList objects")
	flag.StringVar(&registryURL, "registry-url", "", "Docker registry base URL (e.g. https://registry.example.com)")
	flag.StringVar(&registryUsername, "registry-username", "", "Main Resitry User Name")
	flag.StringVar(&registryPassword, "registry-password", "", "Main Registry Password")
	klog.InitFlags(nil)
	flag.Parse()

	log := textlogger.NewLogger(textlogger.NewConfig()).WithName("imageList")
	ctx := context.Background()

	// Create the object reading the list of the images from the registry
	imageListRequestor := imagelist.NewDefaultImageListRequestor(log.WithName("crownlabs-imagelists-updater").WithName("defaultRequestor"))
	if initResult, err := imageListRequestor.Initialize(registryUsername, registryPassword, registryURL); !initResult || err != nil {
		log.Error(err, "Failed to initialize the default image list requestor")
		os.Exit(1)
	}

	// Create the object saving the retrieved information as a K8s object
	log.Info("Target ImageList CR name", "name", imageListName)

	k8sClient, err := utils.NewK8sClient()
	if err != nil {
		log.Error(err, "Failed to initialize K8s client for ImageList savers")
		os.Exit(1)
	}
	// create the object saving the retrieved information as a K8s object
	imageListSaver, err := imagelist.NewDefaultImageListSaver(ctx, imageListName, k8sClient, log.WithName("crownlabs-imagelists-updater").WithName("defaultSaver"))
	if err != nil {
		log.Error(err, "Failed to initialize the default image list saver")
		os.Exit(1)
	}
	// Perform the update process
	imageListUpdater := imagelist.NewUpdater([]imagelist.Requestor{imageListRequestor}, imageListName, imageListSaver, advertisedRegistryName, log.WithName("crownlabs-imagelists-updater"))
	if err := imageListUpdater.Update(ctx); err != nil {
		log.Error(err, "Failed to update the ImageList resource")
		os.Exit(1)
	}
	log.Info("ImageList resource updated successfully", "name", imageListName)
}
