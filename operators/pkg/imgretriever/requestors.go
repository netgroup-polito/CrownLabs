package imgretriever

type ImageListRequestor interface {
	GetImageList() ([]map[string]interface{}, error)
}

var RegisteredRequestors []ImageListRequestor = []ImageListRequestor{}
