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

	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/imagelist"
)

var (
	imageListConfigFile     string
	imageListUpdateInterval int
)

func init() {
	flag.StringVar(&imageListConfigFile, "image-list-config-file", "/etc/config/registries.yaml", "Path to the image list registries configuration file")
	flag.IntVar(&imageListUpdateInterval, "image-list-update-interval", 300, "Image list update interval in seconds")
}

func setupImageList(mgr manager.Manager, log klog.Logger) error {
	// Initialize the image list updater
	if err := imagelist.Initialize(mgr.GetClient(), log.WithName("imagelist"), imagelist.UpdaterOptions{
		ConfigFilePath: imageListConfigFile,
		Interval:       imageListUpdateInterval,
	}); err != nil {
		return err
	}

	// Add the image list scheduler as a runnable to the manager
	return mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		imagelist.StartScheduler(ctx)
		return nil
	}))
}
