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
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/swaggest/usecase/status"
)

// ListRepoDirs lists all directories inside a repository.
func ListRepoDirs(dataRoot, repo string, log logr.Logger) ([]string, error) {
	p := filepath.Join(dataRoot, repo)

	if _, err := os.Stat(p); os.IsNotExist(err) || err != nil {
		log.Error(err, "Repository directory does not exist")
		return nil, status.Wrap(err, status.NotFound)
	}

	dirs, err := os.ReadDir(p)
	if err != nil {
		log.Error(err, "Failed to list directories")
		return nil, status.Wrap(err, status.Internal)
	}

	var list []string
	for _, dir := range dirs {
		if dir.IsDir() {
			list = append(list, dir.Name())
		}
	}

	return list, nil
}

// DeleteImageTag deletes a tag of an image and
// performs clean-up if necessary.
func DeleteImageTag(dataRoot, repo, image, tag string, log logr.Logger) error {
	tagDir := filepath.Join(dataRoot, repo, image, tag)
	log.Info("Deleting tag")

	if _, err := os.Stat(tagDir); os.IsNotExist(err) {
		log.Error(err, "Tag path does not exist")
		return status.Wrap(err, status.NotFound)
	}

	err := os.RemoveAll(tagDir)
	if err != nil {
		log.Error(err, "Failed to delete tag")
		return status.Wrap(err, status.Internal)
	}
	log.Info("Tag deletion completed")

	// Clean up parent directory if empty
	imageDir := filepath.Dir(tagDir)
	proceed, err := deleteDir(imageDir, log)
	if err != nil {
		return err
	}

	if proceed {
		repoDir := filepath.Dir(imageDir)
		_, err = deleteDir(repoDir, log)
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteDir(dir string, log logr.Logger) (bool, error) {
	remaining, err := os.ReadDir(dir)
	if err != nil {
		log.Error(err, "Failed to read parent directory during cleanup")
		return false, status.Wrap(err, status.Internal)
	}

	if len(remaining) > 0 {
		return false, nil
	}

	log.Info("Current directory is empty, starting cleanup")
	if err = os.Remove(dir); err != nil {
		log.Error(err, "Failed to delete directory")
		return false, status.Wrap(err, status.Internal)
	}

	log.Info("Directory deletion completed")
	return true, nil
}
