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

package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/imagelist"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// StartImageListStandalone starts the image list updater as a standalone component
// This function can be used to run image-list independently from the main operator
func StartImageListStandalone() {
	var configFile string
	var updateInterval int

	flag.StringVar(&configFile, "config-file", "/etc/config/registries.yaml", "Path to the registries configuration file")
	flag.IntVar(&updateInterval, "interval", 60, "Update interval in seconds")

	klog.InitFlags(nil)
	flag.Parse()

	ctrl.SetLogger(klog.NewKlogr())
	log := ctrl.Log.WithName("imagelist")

	k8sClient, err := utils.NewK8sClient()
	if err != nil {
		log.Error(err, "unable to prepare k8s client")
		os.Exit(1)
	}

	// Initialize the image list updater
	if err := imagelist.Initialize(k8sClient, log, imagelist.UpdaterOptions{
		ConfigFilePath: configFile,
		Interval:       updateInterval,
	}); err != nil {
		log.Error(err, "failed to initialize image list updater")
		os.Exit(1)
	}

	// Create context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Start the scheduler in a goroutine
	go imagelist.StartScheduler(ctx)

	// Wait for shutdown signal
	sig := <-sigChan
	log.Info("received signal, shutting down", "signal", sig)

	// Give scheduler time to shut down gracefully
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	cancel()
	<-shutdownCtx.Done()

	log.Info("image list updater shut down successfully")
}
