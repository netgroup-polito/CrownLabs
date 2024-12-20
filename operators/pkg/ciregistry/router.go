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
	"net/http"

	"github.com/swaggest/openapi-go/openapi3"
	"github.com/swaggest/rest/nethttp"
	"github.com/swaggest/rest/web"
	swgui "github.com/swaggest/swgui/v5cdn"
	"k8s.io/klog/v2"
)

// NewRouter initializes a router for ciregistry service.
func NewRouter() http.Handler {
	log := klog.Background()
	klog.Info("Initializing router for ciregistry service")

	r := openapi3.NewReflector()
	s := web.NewService(r)

	s.OpenAPISchema().SetTitle("Cloud Image Registry")
	s.OpenAPISchema().SetDescription("API for managing cloudimage repositories and metadata.")
	s.OpenAPISchema().SetVersion("1.0.0")

	s.Get("/{repo}", HandleGetImages(klog.LoggerWithName(log, "imagelist")))
	s.Get("/{repo}/{image}", HandleGetImageTags(klog.LoggerWithName(log, "taglist")))
	s.Get("/{repo}/{image}/{tag}", HandleGetImage(klog.LoggerWithName(log, "imagebin")), nethttp.SuccessfulResponseContentType("application/octet-stream"))
	s.Get("/{repo}/{image}/{tag}/meta", HandleGetImageMeta(klog.LoggerWithName(log, "imagemeta")))
	s.Post("/{repo}/{image}/{tag}", HandlePostImage(klog.LoggerWithName(log, "poster")), nethttp.SuccessStatus(http.StatusCreated))
	s.Delete("/{repo}/{image}/{tag}", HandleDeleteTag(klog.LoggerWithName(log, "deleter")))

	s.Docs("/docs", swgui.New)

	klog.Info("Router initialized successfully")
	return s
}
