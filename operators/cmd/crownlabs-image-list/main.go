package main

import (
	"context"
	"flag"
	"os"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/imageList"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/textlogger"
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
	// if advertisedRegistryName == "" || imageListName == "" || registryURL == "" || registryUsername == "" || registryPassword == "" {
	// 	log.Error(nil, `Usage:
	// 		--advertised-registry-name <registry-name>
	// 		--image-list-name <list-name>
	// 		--registry-url <url>
	// 		`)
	// 	os.Exit(1)
	// }
	//Temporal solution until we refactor the imgretriever package
	// imageList.RequestersSharedData["default-requestor-registryurl"] = registryURL
	// imageList.RequestersSharedData["default-requestor-REGISTRY_USERNAME"] = registryUsername
	// imageList.RequestersSharedData["default-requestor-REGISTRY_PASSWORD"] = registryPassword

	// var imageListRequestors []imageList.ImageListRequestor

	// for _, r := range imageList.RegisteredRequestors {
	// 	if r != nil {
	// 		if initResult, err := r.Initialize(); initResult && err == nil {
	// 			imageListRequestors = append(imageListRequestors, r)
	// 		}
	// 	}
	// }
	// if len(imageListRequestors) == 0 {
	// 	log.Error(nil, `No valid Image source is defined`)
	// 	os.Exit(1)
	// }

	// up := imageList.NewImageListUpdater(imageListRequestors, imageListName, advertisedRegistryName, log.WithName("crownlabs-imagelists-updater"))

	// // with the cronjob
	// up.RunUpdateProcess()

	// Create the object reading the list of the images from the registry
	imageListRequestor := imageList.NewDefaultImageListRequestor(log.WithName("crownlabs-imagelists-updater").WithName("defaultRequestor"))
	if initResult, err := imageListRequestor.Initialize(registryUsername, registryPassword, registryURL); !initResult || err != nil {
		log.Error(err, "Failed to initialize the default image list requestor")
		os.Exit(1)
	}

	// Create the object saving the retrieved information as a K8s object
	log.Info("Target ImageList CR name", "name", imageListName)
	ctx := context.Background()
	k8sClient, err := utils.NewK8sClient()
	if err != nil {
		log.Error(err, "Failed to initialize K8s client for ImageList savers")
		os.Exit(1)
	}
	// create the object saving the retrieved information as a K8s object
	imageListSaver, err := imageList.NewDefaultImageListSaver(ctx, imageListName, k8sClient, log.WithName("crownlabs-imagelists-updater").WithName("defaultSaver"))
	if err != nil {
		log.Error(err, "Failed to initialize the default image list saver")
		os.Exit(1)
	}
	// Perform the update process
	imageListUpdater := imageList.NewImageListUpdater([]imageList.ImageListRequestor{imageListRequestor}, imageListName, imageListSaver, advertisedRegistryName, log.WithName("crownlabs-imagelists-updater"))
	if err := imageListUpdater.Update(); err != nil {
		log.Error(err, "Failed to update the ImageList resource")
		os.Exit(1)
	}
	log.Info("ImageList resource updated successfully", "name", imageListName)

}
