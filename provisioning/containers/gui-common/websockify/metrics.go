// Copyright 2020-2023 Politecnico di Torino
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
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	pingLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "novnc_ping_latency_milliseconds",
			Help:    "Latency of ping requests.",
			Buckets: prometheus.ExponentialBuckets(1, 2, 12),
		},
		[]string{"ip", "connection"},
	)
)

func runMetricsServer(addr, metricsEndpoint string) {
	mux := http.NewServeMux()
	mux.Handle(metricsEndpoint, promhttp.Handler())

	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("failed starting metrics server", err)
	}
}

func makeLatencyObserver(ip, connID string) prometheus.Observer {
	return pingLatency.WithLabelValues(ip, connID)
}
