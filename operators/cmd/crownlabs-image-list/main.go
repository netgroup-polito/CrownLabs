package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"log"
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
)

// "k8s.io/client-go/kubernetes"
type ImageListRequestor struct {
	URL      string
	Username string
	Password string
	client   *http.Client
}

func NewImageListRequestor(url, username, password string) *ImageListRequestor {
	return &ImageListRequestor{
		URL:      url,
		Username: username,
		Password: password,
		client:   &http.Client{},
	}
}

func (r *ImageListRequestor) getImageList() ([]map[string]interface{}, error) {
	repositories, err := r.doSingleGet(r.getCatalogPath())
	if err != nil {
		return nil, err
	}
	repos, ok := repositories["repositories"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected catalog format")
	}

	paths := r.mapRepositoriesToPaths(repos)
	return r.doParallelGets(paths)
}

func (r *ImageListRequestor) doSingleGet(path string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", r.URL+path, nil)
	if err != nil {
		return nil, err
	}
	if r.Username != "" && r.Password != "" {
		req.SetBasicAuth(r.Username, r.Password)
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *ImageListRequestor) doParallelGets(paths []string) ([]map[string]interface{}, error) {
	var wg sync.WaitGroup
	results := make([]map[string]interface{}, len(paths))
	errors := make([]error, len(paths))

	for i, path := range paths {
		wg.Add(1)
		go func(i int, path string) {
			defer wg.Done()
			resp, err := r.doSingleGet(path)
			if err != nil {
				errors[i] = err
				return
			}
			results[i] = resp
		}(i, path)
	}
	wg.Wait()

	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

func (r *ImageListRequestor) getCatalogPath() string {
	return "/v2/_catalog"
}

func (r *ImageListRequestor) mapRepositoriesToPaths(repositories []interface{}) []string {
	paths := make([]string, len(repositories))
	for i, repo := range repositories {
		paths[i] = fmt.Sprintf("/v2/%v/tags/list", repo)
	}
	return paths
}

/////////////////////Image Saver start //////////////////////

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

// ImageListUpdater periodically requests the list of images from the Docker registry
// and saves the obtained information as a Kubernetes object.
type ImageListUpdater struct {
	ImageListRequestor *ImageListRequestor
	ImageListSaver     *ImageListSaver
	RegistryAdvName    string
	ImageListPrefix    string // Prefix for the ImageList name, if needed
	// Scheduler removed; using goroutine and time.Ticker for scheduling
}

// NewImageListUpdater initializes the object.
func NewImageListUpdater(imageListRequestor *ImageListRequestor, imageListSaver *ImageListSaver, imageListPrefix string, registryAdvName string) *ImageListUpdater {
	return &ImageListUpdater{
		ImageListPrefix:    imageListPrefix,
		ImageListRequestor: imageListRequestor,
		ImageListSaver:     imageListSaver,
		RegistryAdvName:    registryAdvName,
	}
}

// RunUpdateProcess starts the scheduler loop to request and save the image list.
func (u *ImageListUpdater) RunUpdateProcess(interval time.Duration, stopCh <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			u.Update()
		case <-stopCh:
			return
		}
	}
}

// Update performs the actual update process.
func (u *ImageListUpdater) Update() {
	start := time.Now()
	log.Println("Starting the update process")

	images, err := u.ImageListRequestor.getImageList()
	if err != nil {
		log.Printf("Failed to retrieve data from upstream: %v", err)
		return
	}

	// spec := map[string]interface{}{
	// 	"registryName": u.RegistryAdvName,
	// 	"images":       processImageList(images),
	// }

	// if err := u.ImageListSaver.UpdateImageList(spec); err != nil {
	// 	log.Printf("Failed to save data as ImageList: %v", err)
	// 	return
	// }

	speccrownlabs_containerdisks := map[string]interface{}{
		"registryName": "crownlabs-containerdisks",
		"images":       processImageList(images),
	}
	u.ImageListSaver, err = NewImageListSaver("crownlabs-containerdisks")
	if err != nil {
		fmt.Printf("Error creating ImageListSaver: %v\n", err)
		os.Exit(1)
	}
	if err := u.ImageListSaver.UpdateImageList(speccrownlabs_containerdisks); err != nil {
		log.Printf("Failed to save data as ImageList: %v", err)
		return
	}

	spec_crownlabs_container_envs := map[string]interface{}{
		"registryName": "crownlabs-container-envs",
		"images":       processImageList(images),
	}

	u.ImageListSaver, err = NewImageListSaver("crownlabs-container-envs")
	if err != nil {
		fmt.Printf("Error creating ImageListSaver: %v\n", err)
		os.Exit(1)
	}
	if err := u.ImageListSaver.UpdateImageList(spec_crownlabs_container_envs); err != nil {
		log.Printf("Failed to save data as ImageList: %v", err)
		return
	}

	specCloadVm := map[string]interface{}{
		"registryName": "crownlabs-standalone",
		"images":       processImageList(images),
	}
	u.ImageListSaver, err = NewImageListSaver("crownlabs-standalone")
	if err != nil {
		fmt.Printf("Error creating ImageListSaver: %v\n", err)
		os.Exit(1)
	}

	if err := u.ImageListSaver.UpdateImageList(specCloadVm); err != nil {
		log.Printf("Failed to save data as ImageList: %v", err)
		return
	}
	log.Printf("Update process correctly completed in %.2f seconds", time.Since(start).Seconds())
}

// processImageList processes the list of images returned from upstream to remove the "latest" tags
// and converts it to the correct format expected by Kubernetes.
func processImageList(images []map[string]interface{}) []map[string]interface{} {
	var convertedImages []map[string]interface{}
	for _, image := range images {
		name, _ := image["name"].(string)
		tagsIface, ok := image["tags"]
		var versions []string
		if ok {
			switch tags := tagsIface.(type) {
			case []interface{}:
				for _, t := range tags {
					if tagStr, ok := t.(string); ok && tagStr != "latest" {
						versions = append(versions, tagStr)
					}
				}
			case []string:
				for _, tagStr := range tags {
					if tagStr != "latest" {
						versions = append(versions, tagStr)
					}
				}
			}
		}
		if len(versions) > 0 {
			convertedImages = append(convertedImages, map[string]interface{}{
				"name":     name,
				"versions": versions,
			})
		}
	}
	return convertedImages
}
func main() {
	var advertisedRegistryName, imageListName, registryURL string
	var updateInterval int

	flag.StringVar(&advertisedRegistryName, "advertised-registry-name", "", "The host name of the Docker registry where the images can be retrieved")
	flag.StringVar(&imageListName, "image-list-name", "", "The name assigned to the resulting ImageList object ")
	flag.StringVar(&registryURL, "registry-url", "", "The URL used to contact the Docker registry")
	flag.IntVar(&updateInterval, "update-interval", -1, "the interval (in seconds) between one update and the following")

	klog.InitFlags(nil)
	flag.Parse()

	if advertisedRegistryName == "" || imageListName == "" || registryURL == "" || updateInterval == -1 {
		fmt.Printf(`Usage: %s
		--advertised-registry-name <registry-name>
		--image-list-name <list-name>
		--registry-url <url>
		--update-interval <seconds>`, os.Args[0])
		fmt.Println()
		os.Exit(1)
	}

	// Example: create ImageListRequestor using parsed arguments
	imageListRequestor := NewImageListRequestor(registryURL, os.Getenv("REGISTRY_USERNAME"), os.Getenv("REGISTRY_PASSWORD"))

	// Example usage: get image list
	imageList, err := imageListRequestor.getImageList()
	if err != nil {
		fmt.Printf("Error retrieving image list: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Retrieved image list: %+v\n", imageList)

	fmt.Printf("Target ImageList object: '%s'\n", imageListName)
	imageListSaver, err := NewImageListSaver(imageListName)
	if err != nil {
		fmt.Printf("Error creating ImageListSaver: %v\n", err)
		os.Exit(1)
	}
	imageListUpdater := NewImageListUpdater(imageListRequestor, imageListSaver, imageListName, advertisedRegistryName)

	log.Println("Starting the update process")
	stopCh := make(chan struct{})
	go imageListUpdater.RunUpdateProcess(time.Duration(updateInterval)*time.Second, stopCh)

	// Wait for interrupt signal to gracefully shutdown
	c := make(chan os.Signal, 1)
	// import "os/signal" and "syscall" at the top
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Received stop signal. Exiting")
	close(stopCh)
}
