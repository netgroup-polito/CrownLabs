package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/klog/v2"

	imgupdate "github.com/netgroup-polito/CrownLabs/operators/pkg/image-list/img-updater"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/image-list/imgretriever"
)

func main() {
	var (
		advertisedRegistryName string
		imageListName          string
		registryURL            string
		updateIntervalSec      int
	)

	flag.StringVar(&advertisedRegistryName, "advertised-registry-name", "", "Hostname of the Docker registry as advertised to consumers")
	flag.StringVar(&imageListName, "image-list-name", "", "Base name/prefix for ImageList objects")
	flag.StringVar(&registryURL, "registry-url", "", "Docker registry base URL (e.g. https://registry.example.com)")
	flag.IntVar(&updateIntervalSec, "update-interval", 60, "Interval in seconds between updates")
	klog.InitFlags(nil)
	flag.Parse()

	if advertisedRegistryName == "" || imageListName == "" || registryURL == "" || updateIntervalSec <= 0 {
		fmt.Printf(`Usage:
			--advertised-registry-name <registry-name>
			--image-list-name <list-name>
			--registry-url <url>
			--update-interval <seconds>
			`)
		os.Exit(1)
	}
	//Temporal solution until we refactor the imgretriever package
	defaultRequestorUsername := os.Getenv("REGISTRY_USERNAME")
	defaultRequestorPassword := os.Getenv("REGISTRY_PASSWORD")
	imgretriever.RequestersSharedData["default-requestor-registryurl"] = registryURL
	imgretriever.RequestersSharedData["default-requestor-REGISTRY_USERNAME"] = defaultRequestorUsername
	imgretriever.RequestersSharedData["default-requestor-REGISTRY_PASSWORD"] = defaultRequestorPassword

	var imageListRequestors []imgretriever.ImageListRequestor

	for _, r := range imgretriever.RegisteredRequestors {
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

	up := imgupdate.NewImageListUpdater(imageListRequestors, imageListName, advertisedRegistryName)

	stop := make(chan struct{})
	go up.RunUpdateProcess(time.Duration(updateIntervalSec)*time.Second, stop)

	// graceful shutdown
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	<-sigc
	close(stop)
}
