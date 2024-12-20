// Copyright 2020-2025 Politecnico di Torino
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

package ciregistry

import (
	"mime/multipart"
	"regexp"

	"github.com/swaggest/usecase"
)

// URLRepoPath struct for /{repo} endpoints.
type URLRepoPath struct {
	Repo string `path:"repo" minLength:"3"`
}

func (r *URLRepoPath) isValid() bool {
	return validatePathSegment(r.Repo)
}

// URLRepoImagePath struct for /{repo}/{image} endpoints.
type URLRepoImagePath struct {
	URLRepoPath
	Image string `path:"image" minLength:"3"`
}

func (r *URLRepoImagePath) isValid() bool {
	return r.URLRepoPath.isValid() && validatePathSegment(r.Image)
}

// URLRepoImageTagPath struct for /{repo}/{image}/{tag} endpoints.
type URLRepoImageTagPath struct {
	URLRepoImagePath
	Tag string `path:"tag" minLength:"2"`
}

func (r *URLRepoImageTagPath) isValid() bool {
	return r.URLRepoImagePath.isValid() && validatePathSegment(r.Tag)
}

// UploadFiles struct for uploading image files along with meta annotations.
type UploadFiles struct {
	URLRepoImageTagPath
	MetadataFile multipart.File `formData:"meta"`
	ImageFile    multipart.File `formData:"img"`
}

// GenericJSONReply struct of a basic response.
type GenericJSONReply struct {
	Success bool     `json:"success"`
	Data    []string `json:"data"`
}

// MetaJSONReply struct of a response to meta requests.
type MetaJSONReply struct {
	Success   bool              `json:"success"`
	Data      map[string]string `json:"data"`
	ImagePath string            `json:"imagePath"`
}

// WriterOutput struct for serving files.
type WriterOutput struct {
	ContentType string `header:"Content-Type" description:"MIME type of the file."`
	usecase.OutputWithEmbeddedWriter
}

// validatePathSegment validates a string against a part of the RFC 1123 Label Names rules:
// - Contains only lowercase alphanumeric characters or '-'.
// - Starts and ends with an alphanumeric character.
// - Length is at least 1 character and at most 63 characters.
// The length requirement is ignored.
func validatePathSegment(segment string) bool {
	rfc1123Regex := regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
	return rfc1123Regex.MatchString(segment)
}
