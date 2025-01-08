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
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	criapi "k8s.io/cri-api/pkg/apis/runtime/v1"
	"k8s.io/utils/trace"

	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// StatsScraper interface.
type StatsScraper interface {
	getStats(ctx context.Context, containers []*criapi.Container) ([]ContainerStats, error)
}

// ContainerStats used to compute metrics.
type ContainerStats struct {
	// Timestamp in nanoseconds at which the information were collected
	CPUTimestamp int64
	// Cumulative CPU usage (sum across all cores) since object creation.
	UsageCoreNanoSeconds uint64
	// Total memory in use.
	MemoryUsageInBytes uint64
	// UsedBytes represents the bytes used for images on the filesystem.
	// This may differ from the total bytes used on the filesystem and may not
	// equal CapacityBytes - AvailableBytes.
	DiskUsageInBytes uint64
	container        *criapi.Container
	runningContainer bool
}

// CustomMetrics stores desired metrics of a container.
type CustomMetrics struct {
	CPUPerc   float32 `json:"cpu"`
	MemBytes  uint64  `json:"mem"`
	DiskBytes uint64  `json:"disk"`
}

func (c CustomMetrics) String() string {
	return fmt.Sprintf("[CPU] %v \t[MEM] %v \t[DISK] %v", c.CPUPerc, c.MemBytes, c.DiskBytes)
}

// MetricsScraper interface.
type MetricsScraper struct {
	Log           logr.Logger
	UpdatePeriod  time.Duration
	RuntimeClient *RemoteRuntimeServiceClient
	// Containers ignored while retrieving metrics.
	ignoredContainerNames []string
	// Stats used to evaluate resource usages changes
	// <key=podName, val=ContainerStats>
	oldStats map[string]*ContainerStats
	// Stats used to evaluate resource usages changes
	// <key=podName, val=CustomMetrics>
	cachedMetrics      map[string]*CustomMetrics
	cachedMetricsMutex sync.RWMutex
	scraper            StatsScraper
}

// Start scraping metrics.
func (ms *MetricsScraper) Start(ctx context.Context) {
	ms.ignoredContainerNames = []string{
		forge.XVncName,
		forge.WebsockifyName,
		forge.ContentDownloaderName,
	}
	ms.oldStats = map[string]*ContainerStats{}

	// Periodically scrape updated metrics for watched containers.
	ticker := time.NewTicker(ms.UpdatePeriod)
	ms.Log.Info("MetricsScraper started", "updatePeriod", ms.UpdatePeriod)
	for range ticker.C {
		ctx, cancel := context.WithTimeout(ctx, ms.UpdatePeriod)

		if err := ms.fillCachedMetrics(ctx); err != nil {
			ms.Log.Error(err, "Error retrieving containerStats", "timeout", ms.UpdatePeriod)
		}

		// Operation running the context completed
		cancel()
	}
}

func (ms *MetricsScraper) fillCachedMetrics(ctx context.Context) error {
	log := ms.Log.WithName("fill-cached-metrics")

	// Update containerIDs list.
	containerIDs, err := ms.findApplicationContainerIDs(ctx)
	if err != nil {
		log.Error(err, "Error updating containerIDs list")
		return err
	}

	tracer := trace.New("Started scraping metrics")
	defer tracer.Log()

	ctx = clctx.LoggerIntoContext(ctx, log)
	containerStatsList, err := ms.scraper.getStats(ctx, containerIDs)
	if err != nil {
		return err
	}

	for i, containerStats := range containerStatsList {
		podName := containerStats.container.GetLabels()["io.kubernetes.pod.name"]

		// If no stats are found for containerID, Pod may no longer be running,
		// related cache information must be removed.
		if !containerStats.runningContainer {
			log.V(2).Info("Removing containerId related cached metrics", "ID", containerStats.container.Id)
			ms.removeCachedMetric(podName)
			continue
		}

		metrics := &CustomMetrics{}
		metrics.MemBytes = containerStats.MemoryUsageInBytes
		metrics.DiskBytes = containerStats.DiskUsageInBytes

		if oldCPU, ok := ms.oldStats[podName]; ok {
			duration := containerStats.CPUTimestamp - oldCPU.CPUTimestamp
			newNs := containerStats.UsageCoreNanoSeconds
			oldNs := oldCPU.UsageCoreNanoSeconds
			cpuPerc := float32(newNs-oldNs) * 100 / float32(duration)
			metrics.CPUPerc = cpuPerc
		}

		ms.oldStats[podName] = &containerStatsList[i]

		ms.cachedMetricsMutex.Lock()
		ms.cachedMetrics[podName] = metrics
		ms.cachedMetricsMutex.Unlock()
	}

	return nil
}

// findApplicationContainerIDs retrieves a list of running application containers.
func (ms *MetricsScraper) findApplicationContainerIDs(ctx context.Context) ([]*criapi.Container, error) {
	log := ms.Log.WithName("find-application-container-ids")

	// get instance pod names.
	ctxp := clctx.LoggerIntoContext(ctx, log.WithName("list-pods"))

	podFilter := &criapi.PodSandboxFilter{
		State: &criapi.PodSandboxStateValue{
			State: criapi.PodSandboxState_SANDBOX_READY,
		},
	}
	pods, err := ms.RuntimeClient.ListPodSandbox(ctxp, podFilter)
	if err != nil {
		return nil, err
	}

	instancePodIDs := []string{}
	for _, pod := range pods {
		if _, ok := pod.Labels["crownlabs.polito.it/instance"]; !ok {
			continue
		}
		if _, ok := pod.Labels["kubevirt.io"]; ok {
			continue
		}
		instancePodIDs = append(instancePodIDs, pod.Id)
	}

	// get application container ids.
	ctxc := clctx.LoggerIntoContext(ctx, log.WithName("list-containers"))

	containerFilter := &criapi.ContainerFilter{
		State: &criapi.ContainerStateValue{
			State: criapi.ContainerState_CONTAINER_RUNNING,
		},
	}
	containers, err := ms.RuntimeClient.ListContainers(ctxc, containerFilter)
	if err != nil {
		return nil, err
	}

	watchedContainers := []*criapi.Container{}
	for _, container := range containers {
		if utils.Contains(instancePodIDs, container.PodSandboxId) && !utils.Contains(ms.ignoredContainerNames, container.Metadata.Name) {
			watchedContainers = append(watchedContainers, container)
		}
	}

	log.V(3).Info("application containers found ", "containers", containers)
	return watchedContainers, nil
}

func (ms *MetricsScraper) removeCachedMetric(podName string) {
	delete(ms.oldStats, podName)
	ms.cachedMetricsMutex.Lock()
	delete(ms.cachedMetrics, podName)
	ms.cachedMetricsMutex.Unlock()
}
