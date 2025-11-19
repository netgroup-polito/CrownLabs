package imageList

import (
	"context"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2/textlogger"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/examagent"
)

type DefaultImageListSaver struct {
	Name      string
	Client    client.Client
	GVR       schema.GroupVersionResource
	Namespace string
	Log       logr.Logger
}

func NewDefaultImageListSaver(name string, client client.Client, log logr.Logger) (*DefaultImageListSaver, error) {
	gvr := schema.GroupVersionResource{
		Group:    "crownlabs.polito.it",
		Version:  "v1alpha1",
		Resource: "imagelists",
	}

	return &DefaultImageListSaver{
		Name:   name,
		Client: client,
		GVR:    gvr,
		Log:    log,
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

	obj.SetGroupVersionKind(clv1alpha1.GroupVersion.WithKind("ImageList"))
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
func (s *DefaultImageListSaver) createImageListObject(imageList []map[string]interface{}, resourceVersion string) *clv1alpha1.ImageList {
	// 1. Convert the input []map[string]interface{} into []ImageListItem
	imageItems := make([]clv1alpha1.ImageListItem, len(imageList))
	for i, imgMap := range imageList {
		name, _ := imgMap["name"].(string)
		versionsInterface, _ := imgMap["versions"].([]interface{})
		versions := make([]string, len(versionsInterface))
		if len(versions) < 1 {
			versions = append(versions, "latest")
		}
		for j, v := range versionsInterface {
			if versionStr, ok := v.(string); ok {
				versions[j] = versionStr
			}
		}

		imageItems[i] = clv1alpha1.ImageListItem{
			Name:     name,
			Versions: versions,
		}
	}

	// 2. Create the ImageListSpec
	spec := clv1alpha1.ImageListSpec{
		RegistryName: s.Name,
		Images:       imageItems,
	}

	// 3. Create the ImageList object
	imgList := &clv1alpha1.ImageList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "crownlabs.polito.it/v1alpha1", // Matches the original unstructured object
			Kind:       "ImageList",                    // Matches the original unstructured object
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: s.Name,
		},
		Spec: spec,
	}

	// 4. Set the ResourceVersion if provided
	if resourceVersion != "" {
		imgList.ObjectMeta.ResourceVersion = resourceVersion
	}

	return imgList
}

func init() {

	log := textlogger.NewLogger(textlogger.NewConfig()).WithName("examagent")
	client, err := examagent.NewK8sClient()
	if err != nil {
		log.Error(err, "unable to prepare k8s client")
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error initializing K8s client for ImageList savers: %v\n", err)
		return
	}
	saver, err := NewDefaultImageListSaver("crownlabs-standalone", client, log)
	if err == nil {
		RegisteredSavers = append(RegisteredSavers, saver)
	}
	saver, err = NewDefaultImageListSaver("crownlabs-container-envs", client, log)
	if err == nil {
		RegisteredSavers = append(RegisteredSavers, saver)
	}
	saver, err = NewDefaultImageListSaver("crownlabs-containerdisks", client, log)
	if err == nil {
		RegisteredSavers = append(RegisteredSavers, saver)
	}
}
