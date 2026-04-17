// Copyright 2020-2026 Politecnico di Torino
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package imagelist contains the image list requestor logic.
package imagelist

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

// DockerImageListRequestor interacts with a Docker registry to retrieve the list of images currently available.
type DockerImageListRequestor struct {
	url         string
	username    string
	password    string
	client      *http.Client
	initialized bool
	log         logr.Logger
}

// NewDockerImageListRequestor creates a new DockerImageListRequestor instance.
func NewDockerImageListRequestor(log logr.Logger) *DockerImageListRequestor {
	return &DockerImageListRequestor{
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
func (r *DockerImageListRequestor) Initialize(username, password, registryURL string) (bool, error) {
	r.url = registryURL
	r.username = username
	r.password = password
	r.initialized = true
	return true, nil
}

// GetImageList retrieves the list of images from the upstream registry.
// It fetches the catalog first, then retrieves the tags for each repository in parallel.
func (r *DockerImageListRequestor) GetImageList(ctx context.Context) ([]map[string]interface{}, error) {
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
func (r *DockerImageListRequestor) doSingleGet(ctx context.Context, path string) (map[string]interface{}, error) {
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
// Skips repositories that return 404 (not found) and logs them as warnings.
func (r *DockerImageListRequestor) doParallelGets(ctx context.Context, paths []string) ([]map[string]interface{}, error) {
	var wg sync.WaitGroup
	results := make([]map[string]interface{}, 0, len(paths))
	var resultsMutex sync.Mutex
	errors := make([]error, 0)
	var errorsMutex sync.Mutex

	for i, path := range paths {
		wg.Add(1)
		go func(i int, path string) {
			defer wg.Done()
			resp, err := r.doSingleGet(ctx, path)
			if err != nil {
				// Check if it's a 404 error (repository not found)
				if strings.Contains(err.Error(), "404") {
					r.log.V(1).Info("repository not found (404), skipping", "path", path)
					return
				}
				// For other errors, record them
				errorsMutex.Lock()
				errors = append(errors, fmt.Errorf("path %s: %w", path, err))
				errorsMutex.Unlock()
				return
			}
			resultsMutex.Lock()
			results = append(results, resp)
			resultsMutex.Unlock()
		}(i, path)
	}

	wg.Wait()

	// Check if any critical errors occurred (non-404)
	if len(errors) > 0 {
		return nil, errors[0]
	}

	return results, nil
}

// getCatalogPath returns the URL path corresponding to the Docker registry catalog.
func (r *DockerImageListRequestor) getCatalogPath() string {
	return "/v2/_catalog"
}

// mapRepositoriesToPaths converts a list of repository names to their corresponding registry API paths.
func (r *DockerImageListRequestor) mapRepositoriesToPaths(repositories []interface{}) []string {
	paths := make([]string, len(repositories))
	for i, repo := range repositories {
		paths[i] = fmt.Sprintf("/v2/%v/tags/list", repo)
	}
	return paths
}

// HarborImageListRequestor interacts with a Harbor registry to retrieve the list of images currently available.
// Harbor uses different API endpoints compared to Docker registry V2.
type HarborImageListRequestor struct {
	url         string
	username    string
	password    string
	projectName string
	client      *http.Client
	initialized bool
	log         logr.Logger
}

// NewHarborImageListRequestor creates a new HarborImageListRequestor instance.
func NewHarborImageListRequestor(log logr.Logger) *HarborImageListRequestor {
	return &HarborImageListRequestor{
		url:         "",
		username:    "",
		password:    "",
		projectName: "",
		client:      &http.Client{},
		initialized: false,
		log:         log,
	}
}

// Initialize initializes the requestor with configuration from shared data.
// For Harbor, the projectName should be provided in RequestersSharedData["harbor_project_name"]
// Returns true if initialization was successful, false otherwise.
func (r *HarborImageListRequestor) Initialize(username, password, registryURL string) (bool, error) {
	r.url = registryURL
	r.username = username
	r.password = password

	// Try to get project name from shared data
	projectName, ok := RequestersSharedData["harbor_project_name"]
	if !ok {
		err := fmt.Errorf("harbor_project_name not found in RequestersSharedData")
		r.log.Error(err, "failed to initialize Harbor requestor: missing project name")
		return false, err
	}

	r.projectName = projectName
	r.initialized = true
	return true, nil
}

// GetImageList retrieves the list of images from the Harbor registry.
// It fetches the repositories first, then retrieves the artifacts for each repository in parallel.
// Returns data in the format expected by processImageList: {"name": "repo", "tags": ["tag1", "tag2"]}
func (r *HarborImageListRequestor) GetImageList(ctx context.Context) ([]map[string]interface{}, error) {
	r.log.V(1).Info("requesting Harbor repositories", "project", r.projectName)
	repositories, err := r.doSingleGetAsList(ctx, r.getCatalogPath())
	if err != nil {
		r.log.Error(err, "failed to retrieve repositories from Harbor")
		return nil, err
	}

	r.log.V(1).Info("requesting artifact details from Harbor", "repository_count", len(repositories))
	paths := r.mapRepositoriesToPaths(repositories)

	// Fetch artifacts for all repositories in parallel
	artifactResults, err := r.doParallelGets(ctx, paths)
	if err != nil {
		return nil, err
	}

	// Transform Harbor artifacts format to match Docker registry format
	// Harbor returns array of artifacts, we need to convert to {name, tags} format
	var result []map[string]interface{}
	r.log.V(1).Info("starting transformation of Harbor data", "repositories_count", len(repositories), "artifact_results_count", len(artifactResults))

	for i, repoData := range repositories {
		if i >= len(artifactResults) {
			r.log.V(1).Info("artifact result missing for repository", "index", i)
			break
		}

		repoName := r.extractRepositoryName(repoData)
		r.log.V(1).Info("extracted repository name", "index", i, "repo_data", repoData, "extracted_name", repoName)

		if repoName == "" {
			r.log.V(1).Info("empty repository name, skipping", "index", i, "repo_data", repoData)
			continue
		}

		artifactsData := artifactResults[i]
		tags := r.extractTagsFromArtifacts(repoName, artifactsData)
		r.log.V(1).Info("extracted tags from artifacts", "repo_name", repoName, "tag_count", len(tags), "tags", tags)

		if len(tags) == 0 {
			r.log.V(1).Info("no tags extracted, artifacts data", "repo_name", repoName, "artifacts_data", artifactsData)
		}

		result = append(result, map[string]interface{}{
			"name": repoName,
			"tags": tags,
		})
	}

	r.log.V(1).Info("transformation complete", "final_result_count", len(result))
	return result, nil
}

// extractRepositoryName extracts the repository name from a Harbor repository object (format: "project/repo")
// and returns only the repo part
func (r *HarborImageListRequestor) extractRepositoryName(repo map[string]interface{}) string {
	if fullName, ok := repo["name"].(string); ok {
		parts := strings.Split(fullName, "/")
		if len(parts) >= 2 {
			repo_only := parts[len(parts)-1]
			r.log.V(1).Info("extractRepositoryName: split full name", "full_name", fullName, "parts_count", len(parts), "extracted", repo_only)
			return repo_only
		}
		r.log.V(1).Info("extractRepositoryName: no slash found", "full_name", fullName)
		return fullName
	}
	r.log.V(1).Info("extractRepositoryName: name field not found or not string", "repo", repo)
	return ""
}

// extractTagsFromArtifacts extracts tag names from Harbor artifacts response
// Harbor artifacts endpoint returns array with objects containing "tags" field
func (r *HarborImageListRequestor) extractTagsFromArtifacts(repoName string, artifactData map[string]interface{}) []string {
	var tags []string
	r.log.V(1).Info("extractTagsFromArtifacts: starting", "repo_name", repoName, "artifact_data_keys", getMapKeys(artifactData))

	// Handle the case where artifacts are wrapped in "artifacts" key (from our wrapper)
	if artifactsIface, ok := artifactData["artifacts"]; ok {
		r.log.V(1).Info("extractTagsFromArtifacts: found 'artifacts' key", "repo_name", repoName, "type", fmt.Sprintf("%T", artifactsIface))

		if artifactsArray, ok := artifactsIface.([]interface{}); ok {
			r.log.V(1).Info("extractTagsFromArtifacts: artifacts is array", "repo_name", repoName, "artifact_count", len(artifactsArray))

			for idx, artifact := range artifactsArray {
				if artifactObj, ok := artifact.(map[string]interface{}); ok {
					// Extract tags from this artifact
					tagsFromArtifact := r.extractTagsFromSingleArtifact(repoName, idx, artifactObj)
					tags = append(tags, tagsFromArtifact...)
				} else {
					r.log.V(1).Info("extractTagsFromArtifacts: artifact is not a map", "repo_name", repoName, "index", idx, "type", fmt.Sprintf("%T", artifact))
				}
			}
		} else {
			r.log.V(1).Info("extractTagsFromArtifacts: artifacts field is not array", "repo_name", repoName, "type", fmt.Sprintf("%T", artifactsIface))
		}
	} else {
		r.log.V(1).Info("extractTagsFromArtifacts: 'artifacts' key not found in artifact data", "repo_name", repoName, "available_keys", getMapKeys(artifactData))
	}

	// Remove duplicates and "latest" tag
	filtered := r.deduplicateAndFilterTags(tags)
	r.log.V(1).Info("extractTagsFromArtifacts: complete", "repo_name", repoName, "total_tags_before_filter", len(tags), "tags_after_filter", len(filtered))

	return filtered
}

// extractTagsFromSingleArtifact extracts all tag names from a single Harbor artifact object
func (r *HarborImageListRequestor) extractTagsFromSingleArtifact(repoName string, artifactIdx int, artifact map[string]interface{}) []string {
	var tags []string
	r.log.V(1).Info("extractTagsFromSingleArtifact: processing artifact", "repo_name", repoName, "artifact_index", artifactIdx, "artifact_keys", getMapKeys(artifact))

	if tagsIface, ok := artifact["tags"]; ok {
		r.log.V(1).Info("extractTagsFromSingleArtifact: found 'tags' field", "repo_name", repoName, "artifact_index", artifactIdx, "type", fmt.Sprintf("%T", tagsIface))

		if tagsArray, ok := tagsIface.([]interface{}); ok {
			r.log.V(1).Info("extractTagsFromSingleArtifact: tags is array", "repo_name", repoName, "artifact_index", artifactIdx, "tag_count", len(tagsArray))

			for tagIdx, tagObj := range tagsArray {
				if tagMap, ok := tagObj.(map[string]interface{}); ok {
					if tagName, ok := tagMap["name"].(string); ok {
						r.log.V(1).Info("extractTagsFromSingleArtifact: extracted tag", "repo_name", repoName, "artifact_index", artifactIdx, "tag_index", tagIdx, "tag_name", tagName)
						tags = append(tags, tagName)
					} else {
						r.log.V(1).Info("extractTagsFromSingleArtifact: tag name not found", "repo_name", repoName, "artifact_index", artifactIdx, "tag_index", tagIdx, "tag_map", tagMap)
					}
				} else {
					r.log.V(1).Info("extractTagsFromSingleArtifact: tag object not a map", "repo_name", repoName, "artifact_index", artifactIdx, "tag_index", tagIdx, "type", fmt.Sprintf("%T", tagObj))
				}
			}
		} else {
			r.log.V(1).Info("extractTagsFromSingleArtifact: tags field is not array", "repo_name", repoName, "artifact_index", artifactIdx, "type", fmt.Sprintf("%T", tagsIface))
		}
	} else {
		r.log.V(1).Info("extractTagsFromSingleArtifact: 'tags' field not found in artifact", "repo_name", repoName, "artifact_index", artifactIdx, "artifact_keys", getMapKeys(artifact))
	}

	return tags
}

// deduplicateAndFilterTags removes duplicates and "latest" tag
func (r *HarborImageListRequestor) deduplicateAndFilterTags(tags []string) []string {
	r.log.V(1).Info("deduplicateAndFilterTags: processing tags", "input_count", len(tags), "tags", tags)

	seen := make(map[string]bool)
	var result []string

	for _, tag := range tags {
		if tag == "latest" {
			r.log.V(1).Info("deduplicateAndFilterTags: filtering out 'latest' tag")
			continue
		}

		if seen[tag] {
			r.log.V(1).Info("deduplicateAndFilterTags: skipping duplicate tag", "tag", tag)
			continue
		}

		seen[tag] = true
		result = append(result, tag)
	}

	r.log.V(1).Info("deduplicateAndFilterTags: complete", "output_count", len(result), "tags", result)
	return result
}

// doSingleGetAsList performs a GET request and expects an array response
func (r *HarborImageListRequestor) doSingleGetAsList(ctx context.Context, path string) ([]map[string]interface{}, error) {
	url := r.url + path
	r.log.V(1).Info("performing GET request to Harbor (expecting array)", "url", url)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
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

	var result []map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		r.log.Error(err, "failed to parse JSON array response", "path", path)
		return nil, err
	}

	return result, nil
}

// doSingleGet performs a single GET request to the target path and returns the parsed JSON result as an object
func (r *HarborImageListRequestor) doSingleGet(ctx context.Context, path string) (map[string]interface{}, error) {
	url := r.url + path
	r.log.V(1).Info("performing GET request to Harbor", "url", url)
	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
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

	// Try to parse as array first (for Harbor artifacts endpoint)
	var arrayResult []interface{}
	if err := json.Unmarshal(body, &arrayResult); err == nil {
		// Successfully parsed as array - wrap it in a map
		return map[string]interface{}{"artifacts": arrayResult}, nil
	}

	// Fall back to parsing as object
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		r.log.Error(err, "failed to parse JSON response", "path", path)
		return nil, err
	}

	return result, nil
}

// doParallelGets performs concurrent GET requests to multiple paths and returns all results.
func (r *HarborImageListRequestor) doParallelGets(ctx context.Context, paths []string) ([]map[string]interface{}, error) {
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

// getCatalogPath returns the URL path corresponding to the Harbor repositories catalog.
func (r *HarborImageListRequestor) getCatalogPath() string {
	return fmt.Sprintf("/api/v2.0/projects/%s/repositories?page=1&page_size=100", r.projectName)
}

// mapRepositoriesToPaths converts a list of repository objects to their corresponding Harbor API paths.
// Harbor repository objects contain a "name" field in format "project/repository".
// We extract only the repository name part (after the slash) for the API path.
func (r *HarborImageListRequestor) mapRepositoriesToPaths(repositories []map[string]interface{}) []string {
	paths := make([]string, len(repositories))
	for i, repo := range repositories {
		var repoName string
		// Harbor responses contain repository objects with a "name" field (format: "project/repository")
		if name, ok := repo["name"].(string); ok {
			// Extract only the repository name part (after the last slash)
			parts := strings.Split(name, "/")
			if len(parts) >= 2 {
				repoName = parts[len(parts)-1]
			} else {
				repoName = name
			}
		} else {
			r.log.V(1).Info("Could not extract repository name from Harbor response", "repository", repo)
			repoName = fmt.Sprintf("%v", repo)
		}
		paths[i] = fmt.Sprintf("/api/v2.0/projects/%s/repositories/%s/artifacts", r.projectName, repoName)
	}
	return paths
}

// Helper function to get keys from a map for logging
func getMapKeys(m map[string]interface{}) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func init() {
	dockerLog := textlogger.NewLogger(textlogger.NewConfig()).WithName("imageList").WithName("dockerRequestor")
	RegisteredRequestors = append(RegisteredRequestors, NewDockerImageListRequestor(dockerLog))
	harborLog := textlogger.NewLogger(textlogger.NewConfig()).WithName("imageList").WithName("harborRequestor")
	RegisteredRequestors = append(RegisteredRequestors, NewHarborImageListRequestor(harborLog))
}
