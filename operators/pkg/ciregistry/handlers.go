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

// Package ciregistry contains the main server logic
// and http handlers for the CrownLabs cloud image registry
package ciregistry

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	"k8s.io/klog/v2"
)

// HandleGetImages lists all images present in a directory.
func HandleGetImages(log klog.Logger) usecase.Interactor {
	u := usecase.NewInteractor(func(_ context.Context, in URLRepoPath, out *BasicJSONReply) error {
		log := log.WithValues("repo", in.Repo)

		if !in.isValid() {
			log.Error(nil, "Invalid path parameters")
			return status.Wrap(errors.New("invalid path parameters"), status.InvalidArgument)
		}

		log.Info("started: Handling GetImages")

		images, err := ListRepoDirs(in.Repo, log.V(1))
		if err != nil {
			log.Error(err, "Failed to list repository images")
			return err
		}

		out.Success = true
		out.Data = images
		log.Info("success")

		return nil
	})

	u.SetExpectedErrors(status.Internal, status.NotFound, status.InvalidArgument)

	return u
}

// HandleGetImageTags lists all tags available for an image.
func HandleGetImageTags(log klog.Logger) usecase.Interactor {
	u := usecase.NewInteractor(func(_ context.Context, in URLRepoImagePath, out *BasicJSONReply) error {
		log := log.WithValues("repo", in.Repo, "image", in.Image)

		if !in.isValid() {
			log.Error(nil, "Invalid path parameters")
			return status.Wrap(errors.New("invalid path parameters"), status.InvalidArgument)
		}

		log.Info("started: Handling GetImageTags")

		imagePath := filepath.Join(in.Repo, in.Image)

		tags, err := ListRepoDirs(imagePath, log.V(1))
		if err != nil {
			log.Error(err, "Failed to list image tags")
			return err
		}

		out.Success = true
		out.Data = tags
		log.Info("success")

		return nil
	})

	u.SetExpectedErrors(status.Internal, status.NotFound, status.InvalidArgument)

	return u
}

// HandleGetImage serves an image file.
func HandleGetImage(log klog.Logger) usecase.Interactor {
	return ServeFile("image.bin", "application/octet-stream", log)
}

// HandleGetImageMeta serves annotations pertaining to a version of an image.
func HandleGetImageMeta(log klog.Logger) usecase.Interactor {
	return ServeFile("meta.json", "application/json", log)
}

// HandlePostImage uploads an image file and related annotations to a directory.
func HandlePostImage(log klog.Logger) usecase.Interactor {
	u := usecase.NewInteractor(func(_ context.Context, in UploadFiles, out *BasicJSONReply) (err error) {
		log := log.WithValues("repo", in.Repo, "image", in.Image, "tag", in.Tag)

		if !in.isValid() {
			log.Error(nil, "Invalid path parameters")
			return status.Wrap(errors.New("invalid path parameters"), status.InvalidArgument)
		}

		log.Info("started: Handling PostImage")

		var (
			raw     []byte
			imgFile *os.File
		)

		imageDir := filepath.Join(DataRoot, in.Repo, in.Image, in.Tag)
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

	u.SetExpectedErrors(status.Internal, status.InvalidArgument)

	return u
}

// HandleDeleteTag deletes a tag of an image.
func HandleDeleteTag(log klog.Logger) usecase.Interactor {
	u := usecase.NewInteractor(func(_ context.Context, in URLRepoImageTagPath, out *BasicJSONReply) error {
		log := log.WithValues("repo", in.Repo, "img", in.Image, "tag", in.Tag)

		if !in.isValid() {
			log.Error(nil, "Invalid path parameters")
			return status.Wrap(errors.New("invalid path parameters"), status.InvalidArgument)
		}

		log.Info("started: Handling DeleteTag")

		err := DeleteImageTag(in.Repo, in.Image, in.Tag, log.V(1))
		if err != nil {
			return err
		}

		out.Success = true
		log.Info("success")

		return nil
	})

	u.SetExpectedErrors(status.Internal, status.NotFound, status.InvalidArgument)

	return u
}
