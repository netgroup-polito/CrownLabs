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

// Package main contains the entrypoint for the cloud image registry.
package main

import (
	"flag"

	"k8s.io/klog/v2"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/ciregistry"
)

var (
	dataRoot                 = flag.String("data-root", "/data", "Root data path for the server")
	serverPrefix             = flag.String("www-prefix", "/", "Starting path where the web server should run from")
	registryPrefix           = flag.String("repo-prefix", "reg", "Subpath of data location on web")
	listenerAddr             = flag.String("listener-addr", ":8080", "Address for the server to listen on")
	readHeaderTimeoutSeconds = flag.Int("read-header-timeout-secs", 2, "Number of seconds allowed to read request headers")
)

func main() {
	// Initialize klog
	klog.InitFlags(nil)
	defer klog.Flush()

	// Parse flags
	flag.Parse()

	// Start the server
	ciregistry.Start(*dataRoot, *serverPrefix, *registryPrefix, *listenerAddr, *readHeaderTimeoutSeconds)
}
