package imageList

import (
	"fmt"
	"time"

	"github.com/go-logr/logr"
)

type ImageListUpdater struct {
	Requestors        []ImageListRequestor
	RegistryAdvName   string
	ImageListBaseName string
	Log               logr.Logger
}

func NewImageListUpdater(reqs []ImageListRequestor, imageListBase, registryAdv string, log logr.Logger) *ImageListUpdater {
	return &ImageListUpdater{
		Requestors:        reqs,
		ImageListBaseName: imageListBase,
		RegistryAdvName:   registryAdv,
		Log:               log,
	}
}

func (u *ImageListUpdater) RunUpdateProcess() {
	u.Update()
}

func (u *ImageListUpdater) Update() {
	start := time.Now()
	u.Log.Info("Starting the update process")
	var images []map[string]interface{}
	var err error
	for _, req := range u.Requestors {
		list, reqErr := req.GetImageList()
		if reqErr != nil {
			err = reqErr
			break
		}
		images = append(images, list...)
	}
	if err != nil {
		u.Log.Error(err, "Failed to retrieve data from upstream")
		return
	}
	for _, imgSaver := range RegisteredSavers {
		if imgSaver != nil {
			Saver := imgSaver
			var filteredImages []map[string]interface{}
			for _, image := range images {
				if Saver.IsThisImageYours(image) {
					filteredImages = append(filteredImages, image)
				}
			}
			if err := Saver.UpdateImageList(processImageList(filteredImages)); err != nil {
				u.Log.Error(err, "Failed to save data as ImageList")
				return
			}
		}
	}
	u.Log.Info(fmt.Sprintf("Update process correctly completed in %.2f seconds", time.Since(start).Seconds()))
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
		} else {
			out = append(out, map[string]interface{}{
				"name":     name,
				"versions": []string{"latest"},
			})
		}
	}

	return out
}
