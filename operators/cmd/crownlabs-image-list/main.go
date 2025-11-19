package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/klog/v2"

	imageList "github.com/netgroup-polito/CrownLabs/operators/pkg/imageList"
)

func main() {
	var (
		advertisedRegistryName string
		imageListName          string
		registryURL            string
		updateIntervalSec      int
		registryUsername       string
		registryPassword       string
	)

	flag.StringVar(&advertisedRegistryName, "advertised-registry-name", "", "Hostname of the Docker registry as advertised to consumers")
	flag.StringVar(&imageListName, "image-list-name", "", "Base name/prefix for ImageList objects")
	flag.StringVar(&registryURL, "registry-url", "", "Docker registry base URL (e.g. https://registry.example.com)")
	flag.IntVar(&updateIntervalSec, "update-interval", 60, "Interval in seconds between updates")
	flag.StringVar(&registryUsername, "registry-username", "", "Main Resitry User Name")
	flag.StringVar(&registryPassword, "registry-password", "", "Main Registry Password")
	klog.InitFlags(nil)
	flag.Parse()

	if advertisedRegistryName == "" || imageListName == "" || registryURL == "" || registryUsername == "" || registryPassword == "" || updateIntervalSec <= 0 {
		fmt.Printf(`Usage:
			--advertised-registry-name <registry-name>
			--image-list-name <list-name>
			--registry-url <url>
			--update-interval <seconds>
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
		fmt.Printf(`No valid Image source is defined`)
		fmt.Println()
		os.Exit(1)
	}

	up := imageList.NewImageListUpdater(imageListRequestors, imageListName, advertisedRegistryName)

	stop := make(chan struct{})
	go up.RunUpdateProcess(time.Duration(updateIntervalSec)*time.Second, stop)

	// graceful shutdown
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	<-sigc
	close(stop)
}
