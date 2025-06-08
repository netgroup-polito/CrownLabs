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

// Package main contains the entrypoint for the bastion ssh tracker.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	tracker "github.com/netgroup-polito/CrownLabs/operators/pkg/bastion-ssh-tracker"
)

var (
	trackerRunning atomic.Bool
	trackerError   error
	healthMutex    sync.RWMutex
)

func main() {
	// Ensure to correctly handle tracker graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	iface := flag.String("ssh-tracker-interface", "any", "The network interface on which the SSH tracker will listen for connections.")
	port := flag.Int("ssh-tracker-port", 22, "The port on which the SSH tracker will listen for connections.")
	snaplen := flag.Int("ssh-tracker-snaplen", 1600, "The snaplen for the SSH tracker.")
	metricsAddr := flag.String("ssh-tracker-metrics-addr", ":8082", "The address the metric endpoint binds to.")

	flag.Parse()

	metricsHandler := http.NewServeMux()
	metricsHandler.Handle("/metrics", promhttp.Handler())
	metricsServer := &http.Server{
		Addr:         *metricsAddr,
		Handler:      metricsHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		log.Println("Starting metrics server on", *metricsAddr)
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	healthHandler := http.NewServeMux()
	healthHandler.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		healthMutex.RLock()
		defer healthMutex.RUnlock()
		if trackerError != nil {
			http.Error(w, trackerError.Error(), http.StatusInternalServerError)
			return
		}
		if !trackerRunning.Load() {
			http.Error(w, "Tracker is not running", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	healthHandler.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
		healthMutex.RLock()
		defer healthMutex.RUnlock()
		if trackerError != nil {
			http.Error(w, trackerError.Error(), http.StatusInternalServerError)
			return
		}
		if !trackerRunning.Load() {
			http.Error(w, "Tracker is not running", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	healthServer := &http.Server{
		Addr:         ":8083",
		Handler:      healthHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		log.Println("Starting health check server on :8083")
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Health check server error: %v", err)
		}
	}()

	sshTracker := tracker.NewSSHTracker()
	go func() {
		trackerRunning.Store(true)
		log.Printf("Starting SSH tracker on interface %s, port %d, snaplen %d", *iface, *port, *snaplen)
		if err := sshTracker.Start(*iface, *port, *snaplen); err != nil {
			healthMutex.Lock()
			trackerError = err
			healthMutex.Unlock()
			trackerRunning.Store(false)
			log.Printf("Tracker stopped with error: %v", err)
		}
	}()

	<-signalCh
	log.Println("Signal received, shutting down...")
	trackerRunning.Store(false)
	sshTracker.Stop()

	// Graceful shutdown for HTTP servers
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = metricsServer.Shutdown(ctx)
	_ = healthServer.Shutdown(ctx)

	log.Println("Shutdown complete")
}
