package imageList

type ImageListSaver interface {
	UpdateImageList(imageList []map[string]interface{}) error
	IsThisImageYours(imageListSpec map[string]interface{}) bool
}

var RegisteredSavers []ImageListSaver = []ImageListSaver{}
