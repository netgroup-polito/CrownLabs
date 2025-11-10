package imageupdater

import (
	"log"
	"time"

	imagesaver "github.com/netgroup-polito/CrownLabs/operators/pkg/image-list/img-saver"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/image-list/imgretriever"
)

type ImageListUpdater struct {
	Requestors        []imgretriever.ImageListRequestor
	RegistryAdvName   string
	ImageListBaseName string
}

func NewImageListUpdater(reqs []imgretriever.ImageListRequestor, imageListBase, registryAdv string) *ImageListUpdater {
	return &ImageListUpdater{
		Requestors:        reqs,
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

// Update performs the actual update process.
func (u *ImageListUpdater) Update() {
	start := time.Now()
	log.Println("Starting the update process")

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
		log.Printf("Failed to retrieve data from upstream: %v", err)
		return
	}
	for _, img_saver := range imagesaver.RegisteredSavers {
		if img_saver != nil {
			Saver := img_saver
			var filteredImages []map[string]interface{}
			for _, image := range images {
				if Saver.IsThisImageYours(image) {
					filteredImages = append(filteredImages, image)
				}
			}
			if err := Saver.UpdateImageList(processImageList(filteredImages)); err != nil {
				log.Printf("Failed to save data as ImageList: %v", err)
				return
			}
		}
	}
	log.Printf("Update process correctly completed in %.2f seconds", time.Since(start).Seconds())
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
