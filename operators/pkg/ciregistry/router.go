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
	"context"
	"net/http"
	"os"
	"os/signal"
	"path"
	"reflect"
	"strings"
	"syscall"
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
	dataRoot       string
	serverPrefix   string
	registryPrefix string
	log            klog.Logger
}

func (r *CIRServer) makePath(subpath string) string {
	return path.Join(r.serverPrefix, r.registryPrefix, subpath)
}

// MakeCIRServer initializes a cloud-img registry server and a router.
func MakeCIRServer(dataRoot, serverPrefix, registryPrefix, listenerAddr string, readHeaderTimeoutSeconds int) *http.Server {
	r := openapi3.NewReflector()
	s := web.NewService(r)

	registryPrefix = strings.Trim(registryPrefix, "/")
	if registryPrefix == "" || registryPrefix == "healthz" || registryPrefix == "docs" {
		registryPrefix = "reg"
	}
	serverPrefix = "/" + strings.Trim(serverPrefix, "/")

	cs := CIRServer{
		dataRoot:       dataRoot,
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

	s.Get("/healthz", cs.healthzHandler())

	s.Get(cs.makePath(""), cs.getRepos())

	s.Get(cs.makePath("{repo}"), cs.getImages())

	s.Get(cs.makePath("{repo}/{image}"), cs.getImageTags())

	s.Get(cs.makePath("{repo}/{image}/{tag}"), cs.getImage(), nethttp.SuccessfulResponseContentType("application/octet-stream"))
	s.Post(cs.makePath("{repo}/{image}/{tag}"), cs.postImage(), nethttp.SuccessStatus(http.StatusCreated))
	s.Delete(cs.makePath("{repo}/{image}/{tag}"), cs.deleteTag())

	s.Get(cs.makePath("{repo}/{image}/{tag}/meta"), cs.getImageMeta())

	s.Docs("/docs", swgui.New)

	cs.log.Info("Router initialized successfully")

	server := &http.Server{
		Addr:              listenerAddr,
		Handler:           s,
		ReadHeaderTimeout: time.Duration(readHeaderTimeoutSeconds) * time.Second,
	}

	return server
}

// Start is a function to start a listening server.
func Start(dataRoot, serverPrefix, registryPrefix, listenerAddr string, readHeaderTimeoutSeconds int) {
	// Start the server
	server := MakeCIRServer(dataRoot, serverPrefix, registryPrefix, listenerAddr, readHeaderTimeoutSeconds)

	// Graceful shutdown setup
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		klog.Infof("Starting server on %s", server.Addr)
		klog.Infof("API documentation available at http://localhost%s/docs", server.Addr)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			klog.Fatalf("Server failed: %v", err)
		}
	}()

	<-stop
	klog.Info("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		klog.Fatalf("Server forced to shutdown: %v", err)
	}

	klog.Info("Server gracefully stopped")
}
