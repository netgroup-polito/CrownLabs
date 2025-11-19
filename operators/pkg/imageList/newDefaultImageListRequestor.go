package imageList

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

type DefaultImageListRequestor struct {
	URL         string
	Username    string
	Password    string
	Client      *http.Client
	Initialized bool
}

func NewDefaultImageListRequestor() *DefaultImageListRequestor {
	return &DefaultImageListRequestor{
		URL:         "",
		Username:    "",
		Password:    "",
		Client:      &http.Client{},
		Initialized: false,
	}
}

func (r *DefaultImageListRequestor) Initialize() (bool, error) {
	var username, password, registryURL string

	username = RequestersSharedData["default-requestor-REGISTRY_USERNAME"]
	password = RequestersSharedData["default-requestor-REGISTRY_PASSWORD"]
	registryURL = RequestersSharedData["default-requestor-registryurl"]
	if username == "" || password == "" || registryURL == "" {
		fmt.Printf("Is Not a valid Image source definition (DefaultImageListRequestor ). Skipping...\n")
		return false, fmt.Errorf("Is Not a valid Image source definition (DefaultImageListRequestor ). Skipping...\n")
	}
	r.URL = registryURL
	r.Username = username
	r.Password = password
	r.Initialized = true
	return true, nil
}
func (r *DefaultImageListRequestor) GetImageList() ([]map[string]interface{}, error) {
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

func (r *DefaultImageListRequestor) doSingleGet(path string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", r.URL+path, nil)
	if err != nil {
		return nil, err
	}
	if r.Username != "" && r.Password != "" {
		req.SetBasicAuth(r.Username, r.Password)
	}
	resp, err := r.Client.Do(req)
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

func (r *DefaultImageListRequestor) doParallelGets(paths []string) ([]map[string]interface{}, error) {
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

func (r *DefaultImageListRequestor) getCatalogPath() string {
	return "/v2/_catalog"
}

func (r *DefaultImageListRequestor) mapRepositoriesToPaths(repositories []interface{}) []string {
	paths := make([]string, len(repositories))
	for i, repo := range repositories {
		paths[i] = fmt.Sprintf("/v2/%v/tags/list", repo)
	}
	return paths
}

func init() {
	RegisteredRequestors = append(RegisteredRequestors, NewDefaultImageListRequestor())
}
