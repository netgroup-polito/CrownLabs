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
		return
	}

	log := globalUpdater.log

	log.Info("starting imagelist scheduler", "interval_seconds", globalUpdater.options.Interval)

	interval := globalUpdater.options.Interval
	if interval <= 0 {
		log.Info("invalid update interval configured, defaulting to 300 seconds")
		interval = 300
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	if err := globalUpdater.Update(ctx); err != nil {
		log.Error(err, "initial imagelist update failed")
	}

	for {
		select {
		case <-ctx.Done():
			log.Info("imagelist scheduler stopped")
			return
		case <-ticker.C:
			if err := globalUpdater.Update(ctx); err != nil {
				log.Error(err, "periodic imagelist update failed")
			}
		}
	}
}
