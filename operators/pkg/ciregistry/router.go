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
	"net/http"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi3"
	"github.com/swaggest/rest/nethttp"
	"github.com/swaggest/rest/web"
	swgui "github.com/swaggest/swgui/v5cdn"
	"k8s.io/klog/v2"
)

// CIRServer holds cloud-img registry server parameters.
type CIRServer struct {
	serverPrefix   string
	registryPrefix string
	log            klog.Logger
}

func (r *CIRServer) makePath(subpath string) string {
	return path.Join(r.serverPrefix, r.registryPrefix, subpath)
}

// MakeCIRServer initializes a cloud-img registry server and a router.
func MakeCIRServer(serverPrefix, registryPrefix, listenerAddr string, readTimeoutSeconds int) *http.Server {
	r := openapi3.NewReflector()
	s := web.NewService(r)

	registryPrefix = strings.Trim(registryPrefix, "/")
	if registryPrefix == "" {
		registryPrefix = "reg"
	}
	serverPrefix = "/" + strings.Trim(serverPrefix, "/")

	cs := CIRServer{
		serverPrefix:   serverPrefix,
		registryPrefix: registryPrefix,
		log:            klog.Background(),
	}

	cs.log.Info("Initializing router for ciregistry service")

	s.OpenAPISchema().SetTitle("Cloud Image Registry")
	s.OpenAPISchema().SetDescription("API for managing cloudimage repositories and metadata.")
	s.OpenAPISchema().SetVersion("1.0.0")

	refl := s.OpenAPIReflector().JSONSchemaReflector()
	refl.DefaultOptions = append(
		refl.DefaultOptions,
		func(rc *jsonschema.ReflectContext) {
			rc.DefName = func(t reflect.Type, _ string) string {
				return t.Name()
			}
		})

	s.Get("/healthz", HealthzHandler())

	s.Get(cs.makePath(""), cs.GetRepos())

	s.Get(cs.makePath("{repo}"), cs.GetImages())

	s.Get(cs.makePath("{repo}/{image}"), cs.GetImageTags())

	s.Get(cs.makePath("{repo}/{image}/{tag}"), cs.GetImage(), nethttp.SuccessfulResponseContentType("application/octet-stream"))
	s.Post(cs.makePath("{repo}/{image}/{tag}"), cs.PostImage(), nethttp.SuccessStatus(http.StatusCreated))
	s.Delete(cs.makePath("{repo}/{image}/{tag}"), cs.DeleteTag())

	s.Get(cs.makePath("{repo}/{image}/{tag}/meta"), cs.GetImageMeta())

	s.Docs("/docs", swgui.New)

	cs.log.Info("Router initialized successfully")

	server := &http.Server{
		Addr:              listenerAddr,
		Handler:           s,
		ReadHeaderTimeout: time.Duration(readTimeoutSeconds) * time.Second,
	}

	return server
}
