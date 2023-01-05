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
	"context"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Resources utilization derived from ContainerMetrics.
type Resources struct {
	Connections []ConnInfo `json:"connections"`
	ConnCount   uint32     `json:"connectionsCount"`
	CPUPerc     uint16     `json:"cpu"`
	MemPerc     uint16     `json:"mem"`
	Net         int64      `json:"net"`
	Timestamp   time.Time  `json:"timestamp,omitempty"`
	ErrorMsg    string     `json:"error,omitempty"`
}

// InstanceMetricsHandler is the main handler for instance metrics.
type InstanceMetricsHandler struct {
	instanceMetricsClient *InstanceMetricsClient
	// Application container resources.limits.cpu.
	cpuLimit string
	// Application container resources.limits.memory.
	memoryLimit string
	// PodName where application container is running.
	podName string
	// Last Resources extracted from CustomMetricsServer
	cachedResources      *Resources
	cachedResourcesMutex sync.RWMutex
	// <UID, ConnInfo> map
	connectionsTracking *sync.Map
	connectionsCount    uint32
}

// ServeHTTP serves WS server.
func (h *InstanceMetricsHandler) serveWs(w http.ResponseWriter, r *http.Request) {
	var err error
	var updatePeriod int64 = 2

	if h.instanceMetricsClient == nil {
		http.Error(w, "Instmetrics server unavailable", http.StatusServiceUnavailable)
		return
	}

	if up, ok := r.URL.Query()["updatePeriod"]; ok {
		updatePeriod, err = strconv.ParseInt(up[0], 10, 16)
		if err != nil {
			http.Error(w, "Invalid updatePeriod: it must be Integer (update seconds)", http.StatusBadRequest)
			return
		}
	}
	updatePeriodD := time.Duration(updatePeriod) * time.Second

	var connUID *string
	if uid, ok := r.URL.Query()["connUid"]; ok {
		// noVNC page request
		uids := uid[0]
		connUID = &uids
		atomic.AddUint32(&h.connectionsCount, 1)
	} else {
		// metricsDashboard request
		connUID = nil
	}

	ip := r.Header.Get(headerXForwardedFor)
	log.Printf("Incoming websocket connection on path /usages with update period %d from IP=%s", updatePeriod, ip)

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}

	go func() {
		ticker := time.NewTicker(updatePeriodD)
		for range ticker.C {
			if err := h.sendInstMetrics(ws, updatePeriodD, connUID); err != nil {
				log.Printf("[sendCustomMetricsCycle] Error sending data on websocket. Closing WebSocket connection. Error: %v", err)
				_ = ws.Close()
				ticker.Stop()
				return
			}
		}
	}()
}

func (h *InstanceMetricsHandler) sendInstMetrics(ws *websocket.Conn, updatePeriod time.Duration, connUID *string) error {
	ctx, cancel := context.WithTimeout(context.Background(), updatePeriod)
	defer cancel()

	if !h.cachedResourcesIsUpdated(updatePeriod) {
		if err := h.updateCachedResources(ctx, updatePeriod); err != nil {
			return err
		}
	}
	if err := h.sendCachedResources(ws, updatePeriod, connUID); err != nil {
		return err
	}

	return nil
}

// cachedResourcesIsUpdated returns true if chachedResources is no older than updatePeriod.
func (h *InstanceMetricsHandler) cachedResourcesIsUpdated(updatePeriod time.Duration) bool {
	h.cachedResourcesMutex.RLock()
	defer h.cachedResourcesMutex.RUnlock()
	if cr := h.cachedResources; cr == nil {
		return false
	}
	return time.Since(h.cachedResources.Timestamp) < updatePeriod
}

// updateCachedResources retrieves updated ContainerMetrics from instMetrics server.
func (h *InstanceMetricsHandler) updateCachedResources(ctx context.Context, updatePeriod time.Duration) error {
	h.cachedResourcesMutex.RLock()
	cr := h.cachedResources
	h.cachedResourcesMutex.RUnlock()
	if cr != nil && time.Since(cr.Timestamp) < updatePeriod {
		return nil
	}

	if h.podName == "" {
		return errors.New("[InstanceMetricsHandler] podName is required")
	}

	response, err := h.instanceMetricsClient.ContainerMetrics(ctx, &h.podName)
	if err != nil {
		return err
	}
	log.Println("InstMetrics Response: ", response)

	resources := percentagesFromContainerMetrics(response.CpuPerc, response.MemBytes, h.cpuLimit, h.memoryLimit)
	resources.Timestamp = time.Now()

	h.cachedResources = resources
	return nil
}

func (h *InstanceMetricsHandler) sendCachedResources(ws *websocket.Conn, updatePeriod time.Duration, connUID *string) error {
	h.cachedResourcesMutex.RLock()
	cr := Resources{
		CPUPerc:   h.cachedResources.CPUPerc,
		MemPerc:   h.cachedResources.MemPerc,
		Timestamp: time.Now(),
	}
	h.cachedResourcesMutex.RUnlock()

	if connUID != nil {
		// send noVNC page latency.
		if ci, ok := h.connectionsTracking.Load(*connUID); ok {
			cr.Net = ci.(ConnInfo).Latency
		}
	} else {
		// send all existing connections info and connections count.
		h.connectionsTracking.Range(func(k, v interface{}) bool {
			cr.Connections = append(cr.Connections, v.(ConnInfo))
			return true
		})
		cr.ConnCount = h.connectionsCount
	}

	message, err := json.Marshal(cr)
	if err != nil {
		log.Println("Error marhaling cachedResources")
		return err
	}

	if err = ws.WriteMessage(websocket.TextMessage, message); err != nil {
		log.Println("InstMetrics WebSocket error:", err)
		return err
	}
	return nil
}

func percentagesFromContainerMetrics(cpuBasePerc float32, memoryBytes uint64, cpuLimit, memoryLimit string) *Resources {
	memoryLimitQuantity := resource.MustParse(memoryLimit)
	memoryLimitBytes := memoryLimitQuantity.Value()

	floatCPULimit, _ := strconv.ParseFloat(cpuLimit, 32)
	cpuLimitCores := float32(floatCPULimit / 1000)

	memoryPerc := memoryBytes * 100 / uint64(memoryLimitBytes)
	cpuPerc := math.Round(float64(cpuBasePerc / cpuLimitCores))

	return &Resources{CPUPerc: uint16(cpuPerc), MemPerc: uint16(memoryPerc)}
}
