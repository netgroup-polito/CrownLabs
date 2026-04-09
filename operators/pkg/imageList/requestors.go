// Package imagelist contains the image list requestor logic.
package imagelist

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2/textlogger"
)

// Requestor defines the interface for objects responsible to retrieve the list of images from upstream sources.
// Each registry implementation must satisfy this interface, and the updater will use it to retrieve the data to be saved in the ImageList objects.
type Requestor interface {
	// GetImageList retrieves the list of images from the upstream registry.
	GetImageList(ctx context.Context) ([]map[string]interface{}, error)
	// Initialize initializes the requestor with configuration data.
	Initialize(username, password, registryURL string) (bool, error)
}

// RegisteredRequestors holds the list of all registered image list requestors.
// RegisteredRequestors holds the list of all registered image list requestors.
var RegisteredRequestors = []Requestor{}

// RequestersSharedData stores configuration data shared across requestors.
var RequestersSharedData = map[string]string{}

// DefaultImageListRequestor interacts with a Docker registry to retrieve the list of images currently available.
type DefaultImageListRequestor struct {
	url         string
	username    string
	password    string
	client      *http.Client
	initialized bool
	log         logr.Logger
}

// NewDefaultImageListRequestor creates a new DefaultImageListRequestor instance.
func NewDefaultImageListRequestor(log logr.Logger) *DefaultImageListRequestor {
	return &DefaultImageListRequestor{
		url:         "",
		username:    "",
		password:    "",
		client:      &http.Client{},
		initialized: false,
		log:         log,
	}
}

// Initialize initializes the requestor with configuration from shared data.
// Returns true if initialization was successful, false otherwise.
func (r *DefaultImageListRequestor) Initialize(username, password, registryURL string) (bool, error) {
	r.url = registryURL
	r.username = username
	r.password = password
	r.initialized = true
	return true, nil
}

// GetImageList retrieves the list of images from the upstream registry.
// It fetches the catalog first, then retrieves the tags for each repository in parallel.
func (r *DefaultImageListRequestor) GetImageList(ctx context.Context) ([]map[string]interface{}, error) {
	r.log.V(1).Info("requesting registry catalog upstream")
	repositories, err := r.doSingleGet(ctx, r.getCatalogPath())
	if err != nil {
		r.log.Error(err, "failed to retrieve catalog")
		return nil, err
	}

	// Type assert to convert interface{} to []interface{}
	reposInterface, ok := repositories["repositories"].([]interface{})
	if !ok {
		err := fmt.Errorf("unexpected catalog format: repositories not found or invalid type")
		r.log.Error(err, "invalid catalog response")
		return nil, err
	}

	r.log.V(1).Info("requesting image details upstream", "repository_count", len(reposInterface))
	paths := r.mapRepositoriesToPaths(reposInterface)
	return r.doParallelGets(ctx, paths)
}

// doSingleGet performs a single GET request to the target path and returns the parsed JSON result.
func (r *DefaultImageListRequestor) doSingleGet(ctx context.Context, path string) (map[string]interface{}, error) {
	r.log.V(1).Info("performing GET request to registry", "url", r.url+path)
	req, err := http.NewRequestWithContext(ctx, "GET", r.url+path, http.NoBody)
	if err != nil {
		r.log.Error(err, "failed to create HTTP request", "path", path)
		return nil, err
	}

	req.SetBasicAuth(r.username, r.password)

	resp, err := r.client.Do(req)
	if err != nil {
		r.log.Error(err, "failed to perform HTTP request", "path", path)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.log.Error(err, "failed to read response body", "path", path)
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		r.log.Error(err, "failed to parse JSON response", "path", path)
		return nil, err
	}

	return result, nil
}

// doParallelGets performs concurrent GET requests to multiple paths and returns all results.
func (r *DefaultImageListRequestor) doParallelGets(ctx context.Context, paths []string) ([]map[string]interface{}, error) {
	var wg sync.WaitGroup
	results := make([]map[string]interface{}, len(paths))
	errors := make([]error, len(paths))

	for i, path := range paths {
		wg.Add(1)
		go func(i int, path string) {
			defer wg.Done()
			resp, err := r.doSingleGet(ctx, path)
			if err != nil {
				errors[i] = err
				return
			}
			results[i] = resp
		}(i, path)
	}

	wg.Wait()

	// Check if any errors occurred
	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

// getCatalogPath returns the URL path corresponding to the Docker registry catalog.
func (r *DefaultImageListRequestor) getCatalogPath() string {
	return "/v2/_catalog"
}

// mapRepositoriesToPaths converts a list of repository names to their corresponding registry API paths.
func (r *DefaultImageListRequestor) mapRepositoriesToPaths(repositories []interface{}) []string {
	paths := make([]string, len(repositories))
	for i, repo := range repositories {
		paths[i] = fmt.Sprintf("/v2/%v/tags/list", repo)
	}
	return paths
}

func init() {
	log := textlogger.NewLogger(textlogger.NewConfig()).WithName("imageList").WithName("defaultRequestor")
	RegisteredRequestors = append(RegisteredRequestors, NewDefaultImageListRequestor(log))
}
