package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/imgretriever"
)

type ImageListSaver struct {
	Name      string
	Client    dynamic.Interface
	GVR       schema.GroupVersionResource
	Namespace string // Not used for cluster-scoped resources, but kept for extensibility
}

func NewImageListSaver(name string) (*ImageListSaver, error) {
	var config *rest.Config
	var err error

	// Try kubeconfig first, then fallback to in-cluster config
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			// fallback to default kubeconfig location
			config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load kube config: %w", err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    "crownlabs.polito.it",
		Version:  "v1alpha1",
		Resource: "imagelists",
	}

	return &ImageListSaver{
		Name:   name,
		Client: client,
		GVR:    gvr,
	}, nil
}

func (s *ImageListSaver) UpdateImageList(imageListSpec map[string]interface{}) error {
	resourceVersion, err := s.getImageListVersion()
	if err != nil {
		return err
	}
	if resourceVersion != "" {
		return s.updateImageList(imageListSpec, resourceVersion)
	}
	return s.createImageList(imageListSpec)
}

func (s *ImageListSaver) getImageListVersion() (string, error) {
	obj, err := s.Client.Resource(s.GVR).Get(context.TODO(), s.Name, metav1.GetOptions{})
	if err != nil {
		// Not found is not an error, just means it doesn't exist yet
		return "", nil
	}
	metadata, found, _ := unstructured.NestedMap(obj.Object, "metadata")
	if !found {
		return "", nil
	}
	if rv, ok := metadata["resourceVersion"].(string); ok {
		return rv, nil
	}
	return "", nil
}

func (s *ImageListSaver) createImageList(imageListSpec map[string]interface{}) error {
	obj := s.createImageListObject(imageListSpec, "")
	_, err := s.Client.Resource(s.GVR).Create(context.TODO(), obj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create ImageList: %w", err)
	}
	fmt.Printf("ImageList '%s' created\n", s.Name)
	return nil
}

func (s *ImageListSaver) updateImageList(imageListSpec map[string]interface{}, resourceVersion string) error {
	obj := s.createImageListObject(imageListSpec, resourceVersion)
	_, err := s.Client.Resource(s.GVR).Update(context.TODO(), obj, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ImageList: %w", err)
	}
	fmt.Printf("ImageList '%s' updated\n", s.Name)
	return nil
}

func (s *ImageListSaver) createImageListObject(imageListSpec map[string]interface{}, resourceVersion string) *unstructured.Unstructured {
	obj := map[string]interface{}{
		"apiVersion": "crownlabs.polito.it/v1alpha1",
		"kind":       "ImageList",
		"metadata": map[string]interface{}{
			"name": s.Name,
		},
		"spec": imageListSpec,
	}
	if resourceVersion != "" {
		obj["metadata"].(map[string]interface{})["resourceVersion"] = resourceVersion
	}
	return &unstructured.Unstructured{Object: obj}
}

type ImageListUpdater struct {
	Requestors        []imgretriever.ImageListRequestor
	Saver             *ImageListSaver
	RegistryAdvName   string
	ImageListBaseName string
}

func NewImageListUpdater(reqs []imgretriever.ImageListRequestor, sv *ImageListSaver, imageListBase, registryAdv string) *ImageListUpdater {
	return &ImageListUpdater{
		Requestors:        reqs,
		Saver:             sv,
		ImageListBaseName: imageListBase,
		RegistryAdvName:   registryAdv,
	}
}

func (u *ImageListUpdater) RunUpdateProcess(interval time.Duration, stop <-chan struct{}) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			u.Update()
		case <-stop:
			return
		}
	}
}

func (u *ImageListUpdater) Update() {
	start := time.Now()
	log.Println("Starting update process")

	var all []map[string]interface{}
	for _, r := range u.Requestors {
		list, err := r.GetImageList()
		if err != nil {
			log.Printf("requestor error: %v", err)
			return
		}
		all = append(all, list...)
	}

	blocks := []struct {
		name         string
		registryName string
	}{
		{"crownlabs-containerdisks", "crownlabs-containerdisks"},
		{"crownlabs-container-envs", "crownlabs-container-envs"},
		{"crownlabs-standalone", "crownlabs-standalone"},
	}
	for _, b := range blocks {
		spec := map[string]interface{}{
			"registryName": b.registryName,
			"images":       processImageList(all),
		}
		sv, err := NewImageListSaver(b.name)
		if err != nil {
			log.Printf("saver(%s) error: %v", b.name, err)
			return
		}
		if err := sv.UpdateImageList(spec); err != nil {
			log.Printf("update(%s) error: %v", b.name, err)
			return
		}
	}

	log.Printf("Update completed in %.2fs", time.Since(start).Seconds())
}

func processImageList(images []map[string]interface{}) []map[string]interface{} {
	var out []map[string]interface{}
	for _, image := range images {
		name, _ := image["name"].(string)
		var versions []string

		if tagsIface, ok := image["tags"]; ok {
			switch tags := tagsIface.(type) {
			case []interface{}:
				for _, t := range tags {
					if s, ok := t.(string); ok && s != "latest" {
						versions = append(versions, s)
					}
				}
			case []string:
				for _, s := range tags {
					if s != "latest" {
						versions = append(versions, s)
					}
				}
			}
		}

		if len(versions) > 0 {
			out = append(out, map[string]interface{}{
				"name":     name,
				"versions": versions,
			})
		}
	}
	return out
}

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

	var imageListRequestors []imgretriever.ImageListRequestor

	for _, r := range imgretriever.RegisteredRequestors {
		if r != nil {
			imageListRequestors = append(imageListRequestors, r)
		}
	}
	if len(imageListRequestors) == 0 {
		fmt.Printf(`No valid Image source is defined`)
		fmt.Println()
		os.Exit(1)
	}

	sv, err := NewImageListSaver(imageListName)
	if err != nil {
		log.Fatalf("creating saver: %v", err)
	}

	up := NewImageListUpdater(imageListRequestors, sv, imageListName, advertisedRegistryName)

	stop := make(chan struct{})
	go up.RunUpdateProcess(time.Duration(updateIntervalSec)*time.Second, stop)

	// graceful shutdown
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	<-sigc
	close(stop)
}
