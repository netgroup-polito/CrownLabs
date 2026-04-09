// Package imagelist contains the image list update logic.
package imagelist

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
)

// ImageListUpdater manages the process of updating ImageList resources with image data from requestors.
type ImageListUpdater struct {
	Requestor         []Requestor
	RegistryAdvName   string
	ImageListBaseName string
	ImageListSaver    Saver
	Log               logr.Logger
}

// NewImageListUpdater creates a new ImageListUpdater instance.
func NewImageListUpdater(requestor []Requestor, imageListBase string, imageListSaver Saver, registryAdv string, log logr.Logger) *ImageListUpdater {
	return &ImageListUpdater{
		Requestor:         requestor,
		ImageListBaseName: imageListBase,
		RegistryAdvName:   registryAdv,
		ImageListSaver:    imageListSaver,
		Log:               log,
	}
}

// Update performs the update process for the ImageList resource.
func (u *ImageListUpdater) Update(ctx context.Context) error {
	start := time.Now()
	u.Log.Info("Starting the update process")
	images := []map[string]interface{}{}
	for _, r := range u.Requestor {
		list, err := r.GetImageList(ctx)
		if err != nil {
			u.Log.Error(err, "failed to retrieve data from upstream")
			return err
		}
		images = append(images, list...)
	}

	// Process and convert images to CRD format
	imageListItems := processImageList(images)
	u.Log.V(1).Info("processed images", "imageCount", len(imageListItems))

	// Save images using the configured saver
	if u.ImageListSaver != nil {
		if err := u.ImageListSaver.UpdateImageList(u.RegistryAdvName, imageListItems); err != nil {
			u.Log.Error(err, "failed to save data as ImageList", "registry", u.RegistryAdvName)
			return err
		}
	}

	u.Log.Info("update process completed successfully", "duration_seconds", time.Since(start).Seconds())
	return nil
}

// processImageList converts raw image data from the registry into CRD ImageListItem objects.
// It removes "latest" tags and ensures all images have at least one version tag.
func processImageList(images []map[string]interface{}) []clv1alpha1.ImageListItem {
	var out []clv1alpha1.ImageListItem

	for _, image := range images {
		name, _ := image["name"].(string)
		var versions []string

		// Extract versions/tags from the image data
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

		// Ensure at least one version exists
		if len(versions) == 0 {
			versions = []string{"latest"}
		}

		out = append(out, clv1alpha1.ImageListItem{
			Name:     name,
			Versions: versions,
		})
	}

	return out
}
