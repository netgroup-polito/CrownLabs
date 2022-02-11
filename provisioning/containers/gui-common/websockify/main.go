// Copyright 2020-2022 Politecnico di Torino
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

package main

import (
	"embed"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	//go:embed novnc/*
	novncFS embed.FS
)

func main() {
	httpAddr := flag.String("http-addr", "0.0.0.0:8080", "websockify server listen address:port")
	metricsAddr := flag.String("metrics-addr", "0.0.0.0:9090", "metrics server listen address:port")
	targetAddr := flag.String("target", "127.0.0.1:5900", "vnc service address:port")
	basePath := flag.String("base-path", "", "base path on which listening for connections")
	pingInterval := flag.Int("ping-interval", 5, "ping interval in seconds")
	showBar := flag.Bool("show-controls", false, "show novnc control bar")

	flag.Parse()
	log.SetFlags(0)

	if !strings.HasPrefix(*basePath, "/") {
		*basePath = "/" + *basePath
	}
	*basePath = strings.TrimSuffix(*basePath, "/")

	go runMetricsServer(*metricsAddr, "/metrics")

	log.Printf("Websockify listening on %s%s", *httpAddr, *basePath)

	if err := http.ListenAndServe(*httpAddr, &NoVncHandler{
		BasePath:     *basePath,
		NoVncFS:      http.FileServer(http.FS(novncFS)),
		ShowNoVncBar: *showBar,
		TargetSocket: *targetAddr,
		PingInterval: time.Second * time.Duration(*pingInterval),
	}); err != nil {
		log.Fatal("failed starting websockify server", err)
	}
}
