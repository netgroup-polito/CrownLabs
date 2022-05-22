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
	"context"
	"embed"
	"flag"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	//go:embed novnc/*
	novncFS embed.FS
)

func main() {
	httpAddr := flag.String("http-addr", "0.0.0.0:8080", "websockify server listen address:port")
	metricsAddr := flag.String("metrics-addr", "0.0.0.0:9090", "prometheus metrics server listen address:port")
	targetAddr := flag.String("target", "127.0.0.1:5900", "vnc service address:port")
	basePath := flag.String("base-path", "", "base path on which listening for connections")
	pingInterval := flag.Int("ping-interval", 2, "ping interval in seconds")
	showBar := flag.Bool("show-controls", false, "show novnc control bar")
	instMetricsServerEndpoint := flag.String("instmetrics-server-endpoint", "localhost:9090", "endpoint for connection to instance metrics server")
	instMetricsConnectionTimeout := flag.Duration("connection-timeout", 5*time.Second, "timeout of connection to instmetrics server")
	podName := flag.String("pod-name", "", "instance podName")
	cpuLimit := flag.String("cpu-limit", "", "application container resources.limits.cpu")
	memLimit := flag.String("memory-limit", "", "application container resources.limits.memory")

	log.SetFlags(0)
	flag.Parse()

	if _, err := strconv.ParseFloat(*cpuLimit, 64); err != nil {
		log.Fatal("error parsing cpuLimit to float", err)
	}

	if !strings.HasPrefix(*basePath, "/") {
		*basePath = "/" + *basePath
	}
	*basePath = strings.TrimSuffix(*basePath, "/")

	ctx := context.Background()
	instMetricsClient, err := GetInstanceMetricsClient(ctx, *instMetricsConnectionTimeout, *instMetricsServerEndpoint)
	if err != nil {
		log.Println("Failed connecting to instmetrics server", err)
		log.Println("Instance metrics will not be available")
	}

	go runMetricsServer(*metricsAddr, "/metrics")

	log.Printf("Websockify listening on %s%s", *httpAddr, *basePath)

	// SyncMap shared between NoVncHandler and InstanceMetricsHandler
	var connectionsTracking sync.Map

	http.Handle("/", &NoVncHandler{
		BasePath:            *basePath,
		NoVncFS:             http.FileServer(http.FS(novncFS)),
		ShowNoVncBar:        *showBar,
		TargetSocket:        *targetAddr,
		PingInterval:        time.Second * time.Duration(*pingInterval),
		connectionsTracking: &connectionsTracking,
		MetricsHandler: &InstanceMetricsHandler{
			cpuLimit:              *cpuLimit,
			memoryLimit:           *memLimit,
			podName:               *podName,
			connectionsTracking:   &connectionsTracking,
			cachedResourcesMutex:  sync.RWMutex{},
			instanceMetricsClient: instMetricsClient,
		},
	})

	if err := http.ListenAndServe(*httpAddr, nil); err != nil {
		log.Fatal("failed starting websockify server", err)
	}
}
