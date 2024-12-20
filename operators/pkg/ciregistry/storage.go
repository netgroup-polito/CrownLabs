// Copyright 2020-2024 Politecnico di Torino
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
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	"k8s.io/klog/v2"
)

var (
	// DataRoot is the global data root position.
	DataRoot = "/data"
)

// ListRepoDirs lists all directories inside a repository.
func ListRepoDirs(repo string, log logr.Logger) ([]string, error) {
	p := filepath.Join(DataRoot, repo)

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
func DeleteImageTag(repo, image, tag string, log logr.Logger) error {
	tagDir := filepath.Join(DataRoot, repo, image, tag)
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

// ServeFile serves image.bin and meta.json files.
func ServeFile(fileName, contentType string, log klog.Logger) usecase.Interactor {
	u := usecase.NewInteractor(func(_ context.Context, in URLRepoImageTagPath, out *WriterOutput) error {
		log := log.WithValues("repo", in.Repo, "img", in.Image, "tag", in.Tag)

		if !in.isValid() {
			log.Error(nil, "Invalid path parameters")
			return status.Wrap(errors.New("invalid path parameters"), status.InvalidArgument)
		}

		if fileName == "image.bin" {
			log.Info("started: Handling GetImage")
		} else {
			log.Info("started: Handling GetImageMeta")
		}

		var (
			err  error
			file *os.File
		)

		filePath := filepath.Join(DataRoot, in.Repo, in.Image, in.Tag, fileName)
		safeFilePath := filepath.Clean(filePath)

		file, err = os.Open(safeFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				log.Error(err, "File not found")
				return status.Wrap(errors.New("file not found"), status.NotFound)
			}
			log.Error(err, "Failed to open file")
			return status.Wrap(err, status.Internal)
		}
		defer func() {
			if clErr := file.Close(); clErr != nil {
				log.Error(clErr, "Failed to close file")
				err = clErr
			}
			if err == nil {
				log.Info("success")
			}
		}()

		out.ContentType = contentType
		_, err = io.Copy(out, file)
		if err != nil {
			log.Error(err, "Failed to copy file content")
			return status.Wrap(err, status.Internal)
		}

		return err
	})

	u.SetExpectedErrors(status.NotFound, status.Internal, status.InvalidArgument)

	return u
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
