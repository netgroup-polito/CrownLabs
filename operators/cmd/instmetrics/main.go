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

// Package main contains the entrypoint for the instance metrics collector.
package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/klog/v2/textlogger"

	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instmetrics"
)

func main() {
	grpcPort := flag.Int("grpc-port", 9090, "Grpc server listening port")
	runtimeEndpoint := flag.String("runtime-endpoint", "unix:///run/containerd/containerd.sock", "Container runtime endpoint for CRI-API")
	connectionTimeout := flag.Duration("connection-timeout", 5*time.Second, "Timeout of connection to the CRI-API")
	updatePeriod := flag.Duration("update-period", 1*time.Second, "Metrics update period and timeout in seconds for requests to CRI-API")

	klog.InitFlags(nil)
	flag.Parse()

	log := textlogger.NewLogger(textlogger.NewConfig()).WithName("instmetrics")
	ctx := clctx.LoggerIntoContext(context.Background(), log)

	remoteRuntimeClient, err := instmetrics.GetRuntimeService(ctx, *connectionTimeout, *runtimeEndpoint)
	if err != nil {
		log.Error(err, "Error creating remoteRuntimeServiceClient")
		os.Exit(1)
	}

	var statsScraper instmetrics.StatsScraper = instmetrics.CRIMetricsScraper{RuntimeClient: remoteRuntimeClient}

	go func() {
		http.Handle("/ready", &instmetrics.ReadinessProbeHandler{RuntimeClient: remoteRuntimeClient, Log: log.WithName("probeHandler"), Ready: false})
		//nolint:gosec // The server is meant to be accessed only by well behaving clients, hence there are no issues with timeouts.
		if err = http.ListenAndServe(":8081", nil); err != nil {
			log.Error(err, "Error serving readiness probe")
		}
	}()

	err = (&instmetrics.Server{
		MetricsScraperPeriod: *updatePeriod,
		Log:                  log.WithName("gRPCServer"),
		Port:                 *grpcPort,
		RuntimeClient:        remoteRuntimeClient,
		StatsScraper:         &statsScraper,
	}).Start(ctx)
	if err != nil {
		log.Error(err, "Unable to initialize gRPC server")
		os.Exit(1)
	}
}
