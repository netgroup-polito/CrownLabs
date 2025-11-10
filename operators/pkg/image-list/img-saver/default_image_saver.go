package imagesaver

import (
	"context"
	"fmt"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/restcfg"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

type DefaultImageListSaver struct {
	Name      string
	Client    client.Client
	GVR       schema.GroupVersionResource
	Namespace string // Not used for cluster-scoped resources, but kept for extensibility
}

func NewDefaultImageListSaver(name string) (*DefaultImageListSaver, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(clv1alpha2.AddToScheme(scheme))
	kubeconfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("k8s config error: %w", err)
	}
	client, err := client.New(restcfg.SetRateLimiter(kubeconfig), client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	gvr := schema.GroupVersionResource{
		Group:    "crownlabs.polito.it",
		Version:  "v1alpha2",
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
	obj := &unstructured.Unstructured{}

	obj.SetGroupVersionKind(clv1alpha2.GroupVersion.WithKind("ImageList"))
	key := client.ObjectKey{
		Name:      s.Name,
		Namespace: s.Namespace,
	}
	err := s.Client.Get(context.TODO(), key, obj)

	if err != nil {
		if errors.IsNotFound(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to get object: %w", err)
	}

	rv := obj.GetResourceVersion()
	if rv == "" {
		return "", nil
	}
	return rv, nil
}

func (s *DefaultImageListSaver) createImageList(imageList []map[string]interface{}) error {
	obj := s.createImageListObject(imageList, "")
	err := s.Client.Create(context.TODO(), obj)
	if err != nil {
		return fmt.Errorf("failed to create ImageList: %w", err)
	}

	fmt.Printf("ImageList '%s' created\n", s.Name)
	return nil
}
func (s *DefaultImageListSaver) updateImageList(imageList []map[string]interface{}, resourceVersion string) error {
	obj := s.createImageListObject(imageList, resourceVersion)

	err := s.Client.Update(context.TODO(), obj)
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
		"apiVersion": "crownlabs.polito.it/v1alpha2",
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
