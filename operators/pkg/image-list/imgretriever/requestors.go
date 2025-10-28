package imgretriever

type ImageListRequestor interface {
	GetImageList() ([]map[string]interface{}, error)
	Initialize() (bool, error)
}

var RegisteredRequestors []ImageListRequestor = []ImageListRequestor{}
var RequestersSharedData map[string]string = map[string]string{}
