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

package imagelist

import (
	"context"
	"time"
)

// StartScheduler starts the periodic image list update scheduler.
func StartScheduler(ctx context.Context) {
	if globalUpdater == nil {
		globalUpdater.log.Error(nil, "image list updater not initialized, cannot start scheduler")
		return
	}

	log := globalUpdater.log

	log.Info("starting image list scheduler", "interval_seconds", globalUpdater.options.Interval)

	ticker := time.NewTicker(time.Duration(globalUpdater.options.Interval) * time.Second)
	defer ticker.Stop()

	// Perform initial update
	if err := Update(ctx); err != nil {
		log.Error(err, "initial image list update failed")
	}

	// Periodic updates
	for {
		select {
		case <-ctx.Done():
			log.Info("image list scheduler stopped")
			return
		case <-ticker.C:
			if err := Update(ctx); err != nil {
				log.Error(err, "periodic image list update failed")
			}
		}
	}
}
