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

package instmetrics

import (
	"context"
	"encoding/json"
	"strings"
	sync "sync"

	"github.com/docker/docker/client"
	"golang.org/x/sync/errgroup"
	criapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"k8s.io/utils/trace"

	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
)

// ContainerStatsResponse is the response from the docker API for container stats.
type ContainerStatsResponse struct {
	MemoryStats struct {
		Usage uint64 `json:"usage"`
	} `json:"memory_stats"`

	CPUStats struct {
		Timestamp int64  `json:"system_cpu_usage"`
		Cores     uint64 `json:"online_cpus"`

		Usage struct {
			Total uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
	} `json:"cpu_stats"`
}

// DockerMetricsScraper extracts container stats using the docker API client.
type DockerMetricsScraper struct {
	// Client is the API client that performs all operations against a docker server.
	DockerClient       *client.Client
	ContStatsListMutex *sync.RWMutex
}

func (s DockerMetricsScraper) getStats(ctx context.Context, containers []*criapi.Container) ([]ContainerStats, error) {
	log := clctx.LoggerFromContext(ctx).WithName("docker-metrics-scraper")
	errs, ctx := errgroup.WithContext(ctx)
	tracer := trace.New("docker-metrics-scraper")

	var containerStatsList []ContainerStats

	for _, container := range containers {
		container := container // https://golang.org/doc/faq#closures_and_goroutines
		containerID := container.GetId()

		errs.Go(func() error {
			log = log.WithValues("containerID", containerID)

			stats, err := getContainerStatsOneShot(ctx, containerID, s.DockerClient)
			log.V(3).Info("getContainerStatsOneShot obtained")

			if err != nil {
				if strings.Contains(err.Error(), "No such container") {
					s.ContStatsListMutex.Lock()
					containerStatsList = append(containerStatsList, ContainerStats{container: container, runningContainer: false})
					s.ContStatsListMutex.Unlock()
					return nil
				}
				return err
			}

			s.ContStatsListMutex.Lock()
			containerStatsList = append(containerStatsList, ContainerStats{
				CPUTimestamp:         stats.CPUStats.Timestamp,
				UsageCoreNanoSeconds: stats.CPUStats.Usage.Total * stats.CPUStats.Cores, // Total cores usage.
				MemoryUsageInBytes:   stats.MemoryStats.Usage,
				DiskUsageInBytes:     0, // Not supported.
				container:            container,
				runningContainer:     true})
			s.ContStatsListMutex.Unlock()

			return nil
		})
	}

	if err := errs.Wait(); err != nil {
		return nil, err
	}
	tracer.Log()

	return containerStatsList, nil
}

func getContainerStatsOneShot(ctx context.Context, cID string, cli *client.Client) (*ContainerStatsResponse, error) {
	stats, err := cli.ContainerStatsOneShot(ctx, cID)
	if err != nil {
		return nil, err
	}

	out := stats.Body
	defer out.Close()

	containerStatsResponse := ContainerStatsResponse{}
	if err := json.NewDecoder(out).Decode(&containerStatsResponse); err != nil {
		return nil, err
	}

	return &containerStatsResponse, nil
}
