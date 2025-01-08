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

// Package instmetrics contains the main logic and helpers
// for the crownlabs custom-metrics-server component.
package instmetrics

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-logr/logr"
	"google.golang.org/grpc"
)

// Server interface.
type Server struct {
	// MetricsScraperPeriod indicates the update interval for CustomMetrics update.
	MetricsScraperPeriod time.Duration
	Port                 int
	Log                  logr.Logger
	RuntimeClient        *RemoteRuntimeServiceClient
	StatsScraper         *StatsScraper
	metricsScraper       *MetricsScraper
	grpcServer           *grpc.Server
	UnimplementedInstanceMetricsServer
}

// Start metricsScraper and RPC server.
func (instmetrics *Server) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s%d", "0.0.0.0:", instmetrics.Port))
	if err != nil {
		return err
	}

	instmetrics.grpcServer = grpc.NewServer()
	RegisterInstanceMetricsServer(instmetrics.grpcServer, instmetrics)

	instmetrics.Log.Info("Custom Metrics gRPC Server started", "listenerAddress", lis.Addr())

	instmetrics.metricsScraper = &MetricsScraper{
		Log:           instmetrics.Log.WithName("metrics-scraper"),
		UpdatePeriod:  instmetrics.MetricsScraperPeriod,
		RuntimeClient: instmetrics.RuntimeClient,
		cachedMetrics: make(map[string]*CustomMetrics),
		scraper:       *instmetrics.StatsScraper,
	}
	go instmetrics.metricsScraper.Start(ctx)

	if err := instmetrics.grpcServer.Serve(lis); err != nil {
		instmetrics.Log.Error(err, "Unable to start gRPC server")
		return err
	}

	return nil
}

// ContainerMetrics receives the PodName and returns a ContainerMetricsResponse with resource utilization.
func (instmetrics *Server) ContainerMetrics(_ context.Context, in *ContainerMetricsRequest) (*ContainerMetricsResponse, error) {
	if in == nil || in.PodName == "" {
		return nil, fmt.Errorf("wrong request: valid CustomMetricsRequest required")
	}

	// Retrieve metrics from metricsScraper cache.
	instmetrics.metricsScraper.cachedMetricsMutex.RLock()
	metrics, ok := instmetrics.metricsScraper.cachedMetrics[in.GetPodName()]
	instmetrics.metricsScraper.cachedMetricsMutex.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no existing application container for requested PodName %v", in.GetPodName())
	}

	response := &ContainerMetricsResponse{
		CpuPerc:   metrics.CPUPerc,
		MemBytes:  metrics.MemBytes,
		DiskBytes: metrics.DiskBytes,
	}

	return response, nil
}
