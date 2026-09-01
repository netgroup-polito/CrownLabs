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
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2/textlogger"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
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
		client:      &http.Client{Timeout: 10 * time.Second},
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

	if resp.StatusCode != http.StatusOK {
		r.log.Error(nil, "unexpected HTTP status code from registry", "path", path, "status_code", resp.StatusCode)
		return nil, fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}

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
func (r *DockerImageListRequestor) doParallelGets(ctx context.Context, paths []string) ([]map[string]interface{}, error) {
	var wg sync.WaitGroup
	results := make([]map[string]interface{}, 0, len(paths))
	var resultsMutex sync.Mutex
	errors := make([]error, 0)
	var errorsMutex sync.Mutex

	for _, path := range paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			resp, err := r.doSingleGet(ctx, path)
			if err != nil {
				errorsMutex.Lock()
				errors = append(errors, fmt.Errorf("path %s: %w", path, err))
				errorsMutex.Unlock()
				return
			}
			resultsMutex.Lock()
			results = append(results, resp)
			resultsMutex.Unlock()
		}(path)
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
// Returns data in the format expected by processImageList: {"name": "repo", "tags": ["tag1", "tag2"]}.
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
// and returns only the repo part.
func (r *HarborImageListRequestor) extractRepositoryName(repo map[string]interface{}) string {
	if fullName, ok := repo["name"].(string); ok {
		parts := strings.Split(fullName, "/")
		if len(parts) >= 2 {
			repoOnly := parts[len(parts)-1]
			r.log.V(1).Info("extractRepositoryName: split full name", "full_name", fullName, "parts_count", len(parts), "extracted", repoOnly)
			return repoOnly
		}
		r.log.V(1).Info("extractRepositoryName: no slash found", "full_name", fullName)
		return fullName
	}
	r.log.V(1).Info("extractRepositoryName: name field not found or not string", "repo", repo)
	return ""
}

// extractTagsFromArtifacts extracts tag names from Harbor artifacts response
// Harbor artifacts endpoint returns array with objects containing "tags" field.
func (r *HarborImageListRequestor) extractTagsFromArtifacts(repoName string, artifactData map[string]interface{}) []string {
	var tags []string
	r.log.V(1).Info("extractTagsFromArtifacts: starting", "repo_name", repoName, "artifact_data_keys", GetMapKeys(artifactData))

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
		r.log.V(1).Info("extractTagsFromArtifacts: 'artifacts' key not found in artifact data", "repo_name", repoName, "available_keys", GetMapKeys(artifactData))
	}

	// Remove duplicates and "latest" tag
	filtered := r.deduplicateAndFilterTags(tags)
	r.log.V(1).Info("extractTagsFromArtifacts: complete", "repo_name", repoName, "total_tags_before_filter", len(tags), "tags_after_filter", len(filtered))

	return filtered
}

// extractTagsFromSingleArtifact extracts all tag names from a single Harbor artifact object.
func (r *HarborImageListRequestor) extractTagsFromSingleArtifact(repoName string, artifactIdx int, artifact map[string]interface{}) []string {
	var tags []string
	r.log.V(1).Info("extractTagsFromSingleArtifact: processing artifact", "repo_name", repoName, "artifact_index", artifactIdx, "artifact_keys", GetMapKeys(artifact))

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
		r.log.V(1).Info("extractTagsFromSingleArtifact: 'tags' field not found in artifact", "repo_name", repoName, "artifact_index", artifactIdx, "artifact_keys", GetMapKeys(artifact))
	}

	return tags
}

// deduplicateAndFilterTags removes duplicates and "latest" tag.
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

// doSingleGetAsList performs a GET request and expects an array response.
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

	if resp.StatusCode != http.StatusOK {
		r.log.Error(nil, "unexpected HTTP status code from Harbor", "path", path, "status_code", resp.StatusCode)
		return nil, fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.log.Error(err, "failed to read response body", "path", path)
		return nil, err
	}

	// Try to parse as direct array first: [...]
	var directArray []interface{}
	if err := json.Unmarshal(body, &directArray); err == nil {
		r.log.V(1).Info("parsed Harbor response as direct array", "path", path, "item_count", len(directArray))
		// Convert []interface{} to []map[string]interface{}
		result := make([]map[string]interface{}, 0, len(directArray))
		for i, item := range directArray {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				r.log.V(1).Info("skipping non-map item in data array", "path", path, "index", i, "item_type", fmt.Sprintf("%T", item))
				continue
			}
			result = append(result, itemMap)
		}
		return result, nil
	}

	// Try to parse as wrapped format: {"data": [...], "meta": {...}}
	var harborResponse map[string]interface{}
	if err := json.Unmarshal(body, &harborResponse); err != nil {
		r.log.Error(err, "failed to parse JSON response (tried both array and object formats)", "path", path, "body", string(body))
		return nil, err
	}

	// Check if response contains an error
	if errorsIface, ok := harborResponse["errors"]; ok {
		r.log.Error(nil, "Harbor API returned error", "path", path, "errors", errorsIface, "full_response", harborResponse)
		return nil, fmt.Errorf("harbor API error: %v", errorsIface)
	}

	// Extract the data array from Harbor response
	dataIface, ok := harborResponse["data"]
	if !ok {
		r.log.Error(nil, "data field not found in Harbor response", "path", path, "response_keys", GetMapKeys(harborResponse), "full_response", harborResponse)
		return nil, fmt.Errorf("data field not found in harbor response")
	}

	dataArray, ok := dataIface.([]interface{})
	if !ok {
		r.log.Error(nil, "data field is not an array", "path", path, "data_type", fmt.Sprintf("%T", dataIface))
		return nil, fmt.Errorf("data field is not an array: got %T", dataIface)
	}

	// Convert []interface{} to []map[string]interface{}
	result := make([]map[string]interface{}, 0, len(dataArray))
	for i, item := range dataArray {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			r.log.V(1).Info("skipping non-map item in data array", "path", path, "index", i, "item_type", fmt.Sprintf("%T", item))
			continue
		}
		result = append(result, itemMap)
	}

	r.log.V(1).Info("successfully parsed Harbor array response", "path", path, "item_count", len(result))
	return result, nil
}

// doSingleGet performs a single GET request to the target path and returns the parsed JSON result as an object.
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

	if resp.StatusCode != http.StatusOK {
		r.log.Error(nil, "unexpected HTTP status code from Harbor", "path", path, "status_code", resp.StatusCode)
		return nil, fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.log.Error(err, "failed to read response body", "path", path)
		return nil, err
	}

	// Try to parse as array first (for Harbor artifacts endpoint)
	var arrayResult []interface{}
	if err := json.Unmarshal(body, &arrayResult); err == nil {
		// Successfully parsed as array - wrap it in a map
		r.log.V(1).Info("parsed Harbor response as array, wrapping in artifacts", "path", path, "item_count", len(arrayResult))
		return map[string]interface{}{"artifacts": arrayResult}, nil
	}

	// Fall back to parsing as object
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		bodySample := string(body)
		if len(bodySample) > 200 {
			bodySample = bodySample[:200] + "..."
		}
		r.log.Error(err, "failed to parse JSON response", "path", path, "body_sample", bodySample)
		return nil, err
	}

	r.log.V(1).Info("parsed Harbor response as object", "path", path, "response_keys", GetMapKeys(result))
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

// InstanceSnapshotImageListRequestor retrieves image data from completed InstanceSnapshot resources in a Kubernetes namespace.
type InstanceSnapshotImageListRequestor struct {
	k8sClient    client.Client
	namespace    string
	registryName string
	initialized  bool
	log          logr.Logger
}

// NewInstanceSnapshotImageListRequestor creates a new InstanceSnapshotImageListRequestor instance.
func NewInstanceSnapshotImageListRequestor(k8sClient client.Client, namespace, registryName string, log logr.Logger) *InstanceSnapshotImageListRequestor {
	return &InstanceSnapshotImageListRequestor{
		k8sClient:    k8sClient,
		namespace:    namespace,
		registryName: registryName,
		initialized:  false,
		log:          log,
	}
}

// Initialize validates the requestor configuration.
func (r *InstanceSnapshotImageListRequestor) Initialize(_, _, _ string) (bool, error) {
	if r.k8sClient == nil {
		return false, fmt.Errorf("kubernetes client is required")
	}
	if r.namespace == "" {
		return false, fmt.Errorf("namespace is required")
	}

	r.initialized = true
	return true, nil
}

// GetImageList retrieves image names and tags from completed InstanceSnapshots.
func (r *InstanceSnapshotImageListRequestor) GetImageList(ctx context.Context) ([]map[string]interface{}, error) {
	if !r.initialized {
		return nil, fmt.Errorf("InstanceSnapshot requestor is not initialized")
	}

	snapshots := &clv1alpha2.InstanceSnapshotList{}
	if err := r.k8sClient.List(ctx, snapshots, client.InNamespace(r.namespace)); err != nil {
		return nil, fmt.Errorf("failed to list InstanceSnapshots in namespace %s: %w", r.namespace, err)
	}

	versionsByImage := map[string]map[string]struct{}{}

	for i := range snapshots.Items {
		snapshot := &snapshots.Items[i]
		if snapshot.Status.Phase != clv1alpha2.Completed {
			r.log.V(1).Info("skipping non-completed InstanceSnapshot", "name", snapshot.Name, "namespace", snapshot.Namespace, "phase", snapshot.Status.Phase)
			continue
		}

		job := &batchv1.Job{}
		key := types.NamespacedName{Name: snapshot.Name, Namespace: snapshot.Namespace}
		if err := r.k8sClient.Get(ctx, key, job); err != nil {
			if kerrors.IsNotFound(err) {
				r.log.Info("skipping InstanceSnapshot because the snapshot job was not found", "name", snapshot.Name, "namespace", snapshot.Namespace)
				continue
			}
			return nil, fmt.Errorf("failed to get snapshot job %s/%s: %w", snapshot.Namespace, snapshot.Name, err)
		}

		imageName, version, ok := r.imageListEntryFromSnapshotJob(job, snapshot.Spec.ImageName)
		if !ok {
			r.log.Info("skipping InstanceSnapshot because the snapshot job destination is missing or invalid", "name", snapshot.Name, "namespace", snapshot.Namespace)
			continue
		}

		if versionsByImage[imageName] == nil {
			versionsByImage[imageName] = map[string]struct{}{}
		}
		versionsByImage[imageName][version] = struct{}{}
	}

	return r.imageListDataFromVersionsMap(versionsByImage), nil
}

func (r *InstanceSnapshotImageListRequestor) imageListDataFromVersionsMap(versionsByImage map[string]map[string]struct{}) []map[string]interface{} {
	names := make([]string, 0, len(versionsByImage))
	for name := range versionsByImage {
		names = append(names, name)
	}
	sort.Strings(names)

	out := make([]map[string]interface{}, 0, len(names))
	for _, name := range names {
		versions := make([]string, 0, len(versionsByImage[name]))
		for version := range versionsByImage[name] {
			versions = append(versions, version)
		}
		sort.Strings(versions)
		out = append(out, map[string]interface{}{
			"name": name,
			"tags": versions,
		})
	}

	return out
}

func (r *InstanceSnapshotImageListRequestor) imageListEntryFromSnapshotJob(job *batchv1.Job, snapshotImageName string) (imageName, version string, ok bool) {
	for i := range job.Spec.Template.Spec.Containers {
		for _, arg := range job.Spec.Template.Spec.Containers[i].Args {
			destination, found := strings.CutPrefix(arg, "--destination=")
			if !found {
				continue
			}

			refName, tag, hasTag := splitTaggedImageReference(destination)
			if !hasTag {
				return "", "", false
			}

			imagePath := r.stripImageRegistry(refName)
			relativeImageName := r.stripSnapshotImagePath(imagePath, snapshotImageName)
			if relativeImageName == "" {
				return "", "", false
			}

			return relativeImageName, tag, true
		}
	}

	return "", "", false
}

func splitTaggedImageReference(imageRef string) (name, tag string, ok bool) {
	lastSlash := strings.LastIndex(imageRef, "/")
	lastColon := strings.LastIndex(imageRef, ":")
	if lastColon == -1 || lastColon < lastSlash {
		return "", "", false
	}

	name = imageRef[:lastColon]
	tag = imageRef[lastColon+1:]
	return name, tag, name != "" && tag != ""
}

func (r *InstanceSnapshotImageListRequestor) stripImageRegistry(imagePath string) string {
	if r.registryName != "" {
		if stripped, found := strings.CutPrefix(imagePath, r.registryName+"/"); found {
			return stripped
		}
	}

	firstSlash := strings.Index(imagePath, "/")
	if firstSlash == -1 {
		return imagePath
	}

	firstComponent := imagePath[:firstSlash]
	if strings.Contains(firstComponent, ".") || strings.Contains(firstComponent, ":") || firstComponent == "localhost" {
		return imagePath[firstSlash+1:]
	}

	return imagePath
}

func (r *InstanceSnapshotImageListRequestor) stripSnapshotImagePath(imagePath, snapshotImageName string) string {
	if imagePath == "" {
		return ""
	}

	if snapshotImageName == "" {
		return imagePath
	}

	if imagePath == snapshotImageName || strings.HasSuffix(imagePath, "/"+snapshotImageName) {
		return imagePath
	}

	return ""
}

// GetMapKeys returns the keys from the provided map[string]interface{}.
// This is useful for structured logging when the map shape is unknown.
func GetMapKeys(m map[string]interface{}) []string {
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
