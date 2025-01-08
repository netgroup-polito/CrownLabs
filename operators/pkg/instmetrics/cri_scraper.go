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

package instmetrics

import (
	"context"

	criapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/utils/trace"

	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
)

// CRIMetricsScraper extracts container stats using the CRI-API.
type CRIMetricsScraper struct {
	RuntimeClient *RemoteRuntimeServiceClient
}

func (s CRIMetricsScraper) getStats(ctx context.Context, containers []*criapi.Container) ([]ContainerStats, error) {
	log := clctx.LoggerFromContext(ctx).WithName("cri-metrics-scraper")
	tracer := trace.New("cri-metrics-scraper")

	var containerStatsList []ContainerStats

	for _, container := range containers {
		containerID := container.GetId()

		log = log.WithValues("containerId", containerID)

		filter := &criapi.ContainerStatsFilter{
			Id: containerID,
		}
		ctx = clctx.LoggerIntoContext(ctx, log.WithName("list-containers-stats"))
		containerStatsResponse, err := s.RuntimeClient.ListContainerStats(ctx, filter)
		if err != nil {
			return nil, err
		}

		log.V(3).Info("ListContainerStats obtained")

		// If no stats are found for containerID, Pod may no longer be running
		if len(containerStatsResponse.GetStats()) == 0 {
			containerStatsList = append(containerStatsList, ContainerStats{container: container, runningContainer: false})
			continue
		}

		containerStats := containerStatsResponse.GetStats()[0]
		containerStatsList = append(containerStatsList, ContainerStats{
			CPUTimestamp:         containerStats.GetCpu().GetTimestamp(),
			UsageCoreNanoSeconds: containerStats.GetCpu().GetUsageCoreNanoSeconds().GetValue(),
			MemoryUsageInBytes:   containerStats.GetMemory().GetWorkingSetBytes().GetValue(),
			DiskUsageInBytes:     containerStats.GetWritableLayer().GetUsedBytes().GetValue(),
			container:            container,
			runningContainer:     true})
	}

	tracer.Log()

	return containerStatsList, nil
}
