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

// Package main contains the entrypoint for the cloud image registry.
package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/klog/v2"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/ciregistry"
)

var (
	dataRoot                 = flag.String("data-root", "/data", "Root data path for the server")
	listenerAddr             = flag.String("listener-addr", ":8080", "Address for the server to listen on")
	readHeaderTimeoutSeconds = flag.Int("read-header-timeout-secs", 2, "Number of seconds allowed to read request headers")
)

func main() {
	// Initialize klog
	klog.InitFlags(nil)
	defer klog.Flush()

	// Parse flags
	flag.Parse()

	// Update ciregistry configuration
	ciregistry.DataRoot = *dataRoot

	// Start the server
	server := initializeServer(*listenerAddr, *readHeaderTimeoutSeconds)

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

func initializeServer(addr string, readTimeoutSeconds int) *http.Server {
	handler := ciregistry.NewRouter()
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: time.Duration(readTimeoutSeconds) * time.Second,
	}

	return server
}
