package imagesaver

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type DefaultImageListSaver struct {
	Name      string
	Client    dynamic.Interface
	GVR       schema.GroupVersionResource
	Namespace string // Not used for cluster-scoped resources, but kept for extensibility
}

func NewDefaultImageListSaver(name string) (*DefaultImageListSaver, error) {
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

	return &DefaultImageListSaver{
		Name:   name,
		Client: client,
		GVR:    gvr,
	}, nil
}
func (s *DefaultImageListSaver) IsThisImageYours(imageListSpec map[string]interface{}) bool {
	return true
}
func (s *DefaultImageListSaver) UpdateImageList(imageList []map[string]interface{}) error {
	resourceVersion, err := s.getImageListVersion()
	if err != nil {
		return err
	}
	if resourceVersion != "" {
		return s.updateImageList(imageList, resourceVersion)
	}
	return s.createImageList(imageList)
}

func (s *DefaultImageListSaver) getImageListVersion() (string, error) {
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

func (s *DefaultImageListSaver) createImageList(imageList []map[string]interface{}) error {
	obj := s.createImageListObject(imageList, "")
	_, err := s.Client.Resource(s.GVR).Create(context.TODO(), obj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create ImageList: %w", err)
	}
	fmt.Printf("ImageList '%s' created\n", s.Name)
	return nil
}

func (s *DefaultImageListSaver) updateImageList(imageList []map[string]interface{}, resourceVersion string) error {
	obj := s.createImageListObject(imageList, resourceVersion)
	_, err := s.Client.Resource(s.GVR).Update(context.TODO(), obj, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update ImageList: %w", err)
	}
	fmt.Printf("ImageList '%s' updated\n", s.Name)
	return nil
}

func (s *DefaultImageListSaver) createImageListObject(imageList []map[string]interface{}, resourceVersion string) *unstructured.Unstructured {
	specs := map[string]interface{}{
		"registryName": s.Name,
		"images":       imageList,
	}
	obj := map[string]interface{}{
		"apiVersion": "crownlabs.polito.it/v1alpha1",
		"kind":       "ImageList",
		"metadata": map[string]interface{}{
			"name": s.Name,
		},
		"spec": specs,
	}
	if resourceVersion != "" {
		obj["metadata"].(map[string]interface{})["resourceVersion"] = resourceVersion
	}
	return &unstructured.Unstructured{Object: obj}
}

func init() {
	saver, err := NewDefaultImageListSaver("crownlabs-standalone")
	if err == nil {
		RegisteredSavers = append(RegisteredSavers, saver)
	}
	saver, err = NewDefaultImageListSaver("crownlabs-container-envs")
	if err == nil {
		RegisteredSavers = append(RegisteredSavers, saver)
	}
	saver, err = NewDefaultImageListSaver("crownlabs-containerdisks")
	if err == nil {
		RegisteredSavers = append(RegisteredSavers, saver)
	}
}
