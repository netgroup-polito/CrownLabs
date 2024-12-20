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

// Package ciregistry contains the main server logic
// and http handlers for the CrownLabs cloud image registry
package ciregistry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
)

// GetRepos lists all repositories currently in the registry.
func (r *CIRServer) getRepos() usecase.Interactor {
	u := usecase.NewInteractor(func(_ context.Context, _ struct{}, out *GenericJSONReply) error {
		log := r.log.WithName("repolist")
		log.Info("started: Handling GetRepos")

		repos, err := ListRepoDirs(r.dataRoot, "", log.V(1))
		if err != nil {
			log.Error(err, "Failed to list registry repositories")
			return err
		}

		out.Success = true
		out.Data = repos
		log.Info("success")

		return nil
	})

	u.SetTitle("GetRepos")
	u.SetName("GetRepos")
	u.SetExpectedErrors(status.Internal, status.NotFound)

	return u
}

// GetImages lists all images present in a directory.
func (r *CIRServer) getImages() usecase.Interactor {
	u := usecase.NewInteractor(func(_ context.Context, in URLRepoPath, out *GenericJSONReply) error {
		log := r.log.WithName("imagelist").WithValues("repo", in.Repo)

		if !in.isValid() {
			log.Error(nil, "Invalid path parameters")
			return status.Wrap(errors.New("invalid path parameters"), status.InvalidArgument)
		}

		log.Info("started: Handling GetImages")

		images, err := ListRepoDirs(r.dataRoot, in.Repo, log.V(1))
		if err != nil {
			log.Error(err, "Failed to list repository images")
			return err
		}

		out.Success = true
		out.Data = images
		log.Info("success")

		return nil
	})

	u.SetTitle("GetImages")
	u.SetName("GetImages")
	u.SetExpectedErrors(status.Internal, status.NotFound, status.InvalidArgument)

	return u
}

// GetImageTags lists all tags available for an image.
func (r *CIRServer) getImageTags() usecase.Interactor {
	u := usecase.NewInteractor(func(_ context.Context, in URLRepoImagePath, out *GenericJSONReply) error {
		log := r.log.WithName("taglist").WithValues("path", in)

		if !in.isValid() {
			log.Error(nil, "Invalid path parameters")
			return status.Wrap(errors.New("invalid path parameters"), status.InvalidArgument)
		}

		log.Info("started: Handling GetImageTags")

		imagePath := filepath.Join(in.Repo, in.Image)

		tags, err := ListRepoDirs(r.dataRoot, imagePath, log.V(1))
		if err != nil {
			log.Error(err, "Failed to list image tags")
			return err
		}

		out.Success = true
		out.Data = tags
		log.Info("success")

		return nil
	})

	u.SetTitle("GetImageTags")
	u.SetName("GetImageTags")
	u.SetExpectedErrors(status.Internal, status.NotFound, status.InvalidArgument)

	return u
}

// GetImage serves an image file.
func (r *CIRServer) getImage() usecase.Interactor {
	u := usecase.NewInteractor(func(_ context.Context, in URLRepoImageTagPath, out *WriterOutput) error {
		log := r.log.WithName("imagebin").WithValues("path", in)

		if !in.isValid() {
			log.Error(nil, "Invalid path parameters")
			return status.Wrap(errors.New("invalid path parameters"), status.InvalidArgument)
		}

		log.Info("started: Handling GetImage")

		var (
			err  error
			file *os.File
		)

		filePath := filepath.Join(r.dataRoot, in.Repo, in.Image, in.Tag, "image.bin")
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

		out.ContentType = "application/octet-stream"
		_, err = io.Copy(out, file)
		if err != nil {
			log.Error(err, "Failed to copy file content")
			return nil // we do not return anything to avoid sending headers in case of a connection reset
		}

		return err
	})

	u.SetTitle("GetImage")
	u.SetName("GetImage")
	u.SetExpectedErrors(status.NotFound, status.Internal, status.InvalidArgument)

	return u
}

// GetImageMeta serves meta annotations pertaining to a version of an image.
func (r *CIRServer) getImageMeta() usecase.Interactor {
	u := usecase.NewInteractor(func(_ context.Context, in URLRepoImageTagPath, out *MetaJSONReply) error {
		log := r.log.WithName("imagemeta").WithValues("path", in)

		if !in.isValid() {
			log.Error(nil, "Invalid path parameters")
			return status.Wrap(errors.New("invalid path parameters"), status.InvalidArgument)
		}

		log.Info("started: Handling GetImageMeta")

		var (
			err  error
			file *os.File
		)

		filePath := filepath.Join(r.dataRoot, in.Repo, in.Image, in.Tag, "meta.json")
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
				out.Success = true
				out.ImagePath = r.makePath(fmt.Sprintf("%s/%s/%s", in.Repo, in.Image, in.Tag))
				log.Info("success")
			}
		}()

		raw, err := io.ReadAll(file)
		if err != nil {
			log.Error(err, "Failed to read metadata file")
			return status.Wrap(err, status.Internal)
		}

		err = json.Unmarshal(raw, &out.Data)
		if err != nil {
			log.Error(err, "Failed to copy file content")
			return status.Wrap(err, status.Internal)
		}

		return err
	})

	u.SetTitle("GetImageMeta")
	u.SetName("GetImageMeta")
	u.SetExpectedErrors(status.NotFound, status.Internal, status.InvalidArgument)

	return u
}

// PostImage uploads an image file and related meta annotations to a directory.
func (r *CIRServer) postImage() usecase.Interactor {
	u := usecase.NewInteractor(func(_ context.Context, in UploadFiles, out *GenericJSONReply) (err error) {
		log := r.log.WithName("poster").WithValues("path", in.URLRepoImageTagPath)

		if !in.isValid() {
			log.Error(nil, "Invalid path parameters")
			return status.Wrap(errors.New("invalid path parameters"), status.InvalidArgument)
		}

		log.Info("started: Handling PostImage")

		var (
			raw     []byte
			imgFile *os.File
		)

		imageDir := filepath.Join(r.dataRoot, in.Repo, in.Image, in.Tag)
		err = os.MkdirAll(imageDir, os.ModePerm)
		if err != nil {
			log.Error(err, "Failed to create directory")
			return status.Wrap(err, status.Internal)
		}

		defer func() {
			clErr := in.MetadataFile.Close()
			if clErr != nil && err == nil {
				log.Error(clErr, "Failed to close metadata file")
				err = clErr
			}

			clErr = in.ImageFile.Close()
			if clErr != nil && err == nil {
				log.Error(clErr, "Failed to close image file")
				err = clErr
			}

			if err == nil {
				out.Success = true
				log.Info("success")
			}
		}()

		raw, err = io.ReadAll(in.MetadataFile)
		if err != nil {
			log.Error(err, "Failed to read metadata file")
			return status.Wrap(err, status.Internal)
		}

		var data map[string]string
		err = json.Unmarshal(raw, &data)
		if err != nil {
			log.Error(err, "Failed to parse metadata file")
			return status.Wrap(err, status.Internal)
		}
		raw, err = json.Marshal(data)
		if err != nil {
			log.Error(err, "Failed to parse metadata file")
			return status.Wrap(err, status.Internal)
		}

		metaFilePath := filepath.Join(imageDir, "meta.json")
		err = os.WriteFile(metaFilePath, raw, 0o600)
		if err != nil {
			log.Error(err, "Failed to write metadata file")
			return status.Wrap(err, status.Internal)
		}

		imgFilePath := filepath.Join(imageDir, "image.bin")
		safeFilePath := filepath.Clean(imgFilePath)
		imgFile, err = os.Create(safeFilePath)
		if err != nil {
			log.Error(err, "Failed to create image file")
			return status.Wrap(err, status.Internal)
		}
		defer imgFile.Close()

		_, err = io.Copy(imgFile, in.ImageFile)
		if err != nil {
			log.Error(err, "Failed to copy image file")
			return status.Wrap(err, status.Internal)
		}

		return err
	})

	u.SetTitle("PostImage")
	u.SetName("PostImage")
	u.SetExpectedErrors(status.Internal, status.InvalidArgument)

	return u
}

// DeleteTag deletes a tag of an image.
func (r *CIRServer) deleteTag() usecase.Interactor {
	u := usecase.NewInteractor(func(_ context.Context, in URLRepoImageTagPath, out *GenericJSONReply) error {
		log := r.log.WithName("deleter").WithValues("path", in)

		if !in.isValid() {
			log.Error(nil, "Invalid path parameters")
			return status.Wrap(errors.New("invalid path parameters"), status.InvalidArgument)
		}

		log.Info("started: Handling DeleteTag")

		err := DeleteImageTag(r.dataRoot, in.Repo, in.Image, in.Tag, log.V(1))
		if err != nil {
			return err
		}

		out.Success = true
		log.Info("success")

		return nil
	})

	u.SetTitle("DeleteTag")
	u.SetName("DeleteTag")
	u.SetExpectedErrors(status.Internal, status.NotFound, status.InvalidArgument)

	return u
}

// HealthzHandler is used for performing readiness probes.
func (r *CIRServer) healthzHandler() usecase.Interactor {
	u := usecase.NewInteractor(func(_ context.Context, _ struct{}, out *GenericJSONReply) error {
		out.Success = true
		return nil
	})

	u.SetTitle("ReadinessProbe")
	u.SetName("ReadinessProbe")
	u.SetExpectedErrors(status.Unavailable)

	return u
}
