package main

import (
	"flag"
	"os"

	"k8s.io/klog/v2"
	"k8s.io/klog/v2/textlogger"

	imageList "github.com/netgroup-polito/CrownLabs/operators/pkg/imageList"
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
	if advertisedRegistryName == "" || imageListName == "" || registryURL == "" || registryUsername == "" || registryPassword == "" {
		log.Error(nil, `Usage:
			--advertised-registry-name <registry-name>
			--image-list-name <list-name>
			--registry-url <url>
			`)
		os.Exit(1)
	}
	//Temporal solution until we refactor the imgretriever package
	imageList.RequestersSharedData["default-requestor-registryurl"] = registryURL
	imageList.RequestersSharedData["default-requestor-REGISTRY_USERNAME"] = registryUsername
	imageList.RequestersSharedData["default-requestor-REGISTRY_PASSWORD"] = registryPassword

	var imageListRequestors []imageList.ImageListRequestor

	for _, r := range imageList.RegisteredRequestors {
		if r != nil {
			if initResult, err := r.Initialize(); initResult && err == nil {
				imageListRequestors = append(imageListRequestors, r)
			}
		}
	}
	if len(imageListRequestors) == 0 {
		log.Error(nil, `No valid Image source is defined`)
		os.Exit(1)
	}

	up := imageList.NewImageListUpdater(imageListRequestors, imageListName, advertisedRegistryName, log.WithName("crownlabs-imagelists-updater"))

	// with the cronjob
	up.RunUpdateProcess()

}
