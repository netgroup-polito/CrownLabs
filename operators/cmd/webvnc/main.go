// Copyright 2020-2026 Politecnico di Torino
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

// Package main contains the entrypoint for webVNC, a WebSocket VNC bridge for CrownLabs.
package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/go-logr/stdr"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/webvnc"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = crownlabsv1alpha1.AddToScheme(scheme)
	_ = crownlabsv1alpha2.AddToScheme(scheme)
}

func main() {
	webvncmaxconncountFlag := flag.String("webvncmaxconncount", "1000", "The maximum number of concurrent VNC connections.")
	webvncwebsocketportFlag := flag.String("webvncwebsocketport", "8086", "The port on which the WebSocket server listens.")

	flag.Parse()

	maxConn64, err := strconv.ParseInt(*webvncmaxconncountFlag, 10, 32)
	if err != nil {
		maxConn64 = 1000
	}

	stdLogger := log.New(os.Stderr, "", log.LstdFlags)
	baseLogger := stdr.New(stdLogger)

	webVNCCtx := &webvnc.ServerContext{}
	webVNCCtx.BaseLogger = baseLogger
	webVNCCtx.MaxConnectionCount = int32(maxConn64)
	webVNCCtx.WebsocketPort = *webvncwebsocketportFlag
	webVNCCtx.BaseConfig, err = utils.GetRestConfig()

	if err != nil {
		webVNCCtx.BaseLogger.Error(err, "Failed to get REST config")
		return
	}

	webVNCCtx.BaseLogger.Info("Config loaded",
		"MaxConnectionCount", webVNCCtx.MaxConnectionCount,
		"WebsocketPort", webVNCCtx.WebsocketPort)

	webVNCCtx.StartWebVNC()
}
