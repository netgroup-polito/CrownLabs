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

package instautoctrl

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// Prometheus is a client for interacting with a Prometheus server.
type Prometheus struct {
	address                  string
	queryNginxAvailable      string
	queryBastionSSHAvailable string
	queryWebSSHAvailable     string
	queryNginxData           string
	queryBastionSSHData      string
	queryWebSSHData          string
	queryStep                time.Duration
	client                   v1.API
}

// NewPrometheusObj creates a new Prometheus client with the given configuration.
func NewPrometheusObj(
	address string,
	queryNginxAvailable, queryBastionSSHAvailable, queryWebSSHAvailable, queryNginxData, queryBastionSSHData, queryWebSSHData string,
	queryStep time.Duration,
) (PrometheusClientInterface, error) {
	client, err := api.NewClient(api.Config{
		Address: address,
	})
	if err != nil {
		return nil, err
	}

	v1api := v1.NewAPI(client)

	return &Prometheus{
		address:                  address,
		queryNginxAvailable:      queryNginxAvailable,
		queryBastionSSHAvailable: queryBastionSSHAvailable,
		queryWebSSHAvailable:     queryWebSSHAvailable,
		queryNginxData:           queryNginxData,
		queryBastionSSHData:      queryBastionSSHData,
		queryWebSSHData:          queryWebSSHData,
		queryStep:                queryStep,
		client:                   v1api,
	}, nil
}

// IsPrometheusHealthy checks if Prometheus and required metrics are available.
func (p *Prometheus) IsPrometheusHealthy(ctx context.Context, timeout time.Duration) (bool, error) {
	log := ctrl.LoggerFrom(ctx).WithName("prometheus-health")

	// Verify connection to Prometheus health endpoint
	healthEndpoint := fmt.Sprintf("%s/-/healthy", p.address)

	statusCode, _, err := utils.HTTPGet(ctx, healthEndpoint, timeout)
	if err != nil {
		log.Error(err, "Failed to connect to Prometheus health endpoint")
		return false, fmt.Errorf("prometheus health check failed: %w", err)
	}

	if statusCode != http.StatusOK {
		log.Error(fmt.Errorf("prometheus health check returned non-OK status"), "statusCode", statusCode)
		return false, nil
	}

	// Check if ingress metrics and bastion metrics are available on worker nodes
	query1 := p.queryNginxAvailable
	query2 := p.queryBastionSSHAvailable
	query3 := p.queryWebSSHAvailable

	result1, _, err1 := p.client.Query(ctx, query1, time.Now())
	result2, _, err2 := p.client.Query(ctx, query2, time.Now())
	result3, _, err3 := p.client.Query(ctx, query3, time.Now())

	if err1 != nil && err2 != nil && err3 != nil {
		log.Error(err1, "Failed to query Prometheus for ingress metrics")
		log.Error(err2, "Failed to query Prometheus for bastion SSH metrics")
		log.Error(err3, "Failed to query Prometheus for web SSH metrics")
		return false, fmt.Errorf("all Prometheus queries failed: %w, %w, %w", err1, err2, err3)
	}

	active1 := false
	active2 := false
	active3 := false

	if err1 == nil {
		vec1, ok1 := result1.(model.Vector)
		if ok1 && len(vec1) > 0 && int(vec1[0].Value) > 0 {
			active1 = true
		}
	}
	if err2 == nil {
		vec2, ok2 := result2.(model.Vector)
		if ok2 && len(vec2) > 0 && int(vec2[0].Value) > 0 {
			active2 = true
		}
	}
	if err3 == nil {
		vec3, ok3 := result3.(model.Vector)
		if ok3 && len(vec3) > 0 && int(vec3[0].Value) > 0 {
			active3 = true
		}
	}

	if !active1 && !active2 && !active3 {
		log.Error(fmt.Errorf("no metrics are available on worker nodes"), "No metrics are available on worker nodes")
		return false, nil
	}

	// At least one node has metrics available
	return true, nil
}

// GetLastActivityTime retrieves the last time an instance was accessed.
func (p *Prometheus) GetLastActivityTime(query string, interval time.Duration) (time.Time, error) {
	end := time.Now()
	start := end.Add(-interval)

	r := v1.Range{
		Start: start,
		End:   end,
		Step:  p.queryStep,
	}

	result, _, err := p.client.QueryRange(context.Background(), query, r)
	if err != nil {
		return time.Time{}, fmt.Errorf("query failed: %w", err)
	}

	matrix, ok := result.(model.Matrix)
	if !ok {
		return time.Time{}, fmt.Errorf("unexpected result format")
	}

	var lastChange time.Time

	for _, stream := range matrix {
		var prevValue model.SampleValue
		first := true
		for _, sample := range stream.Values {
			if first {
				prevValue = sample.Value
				first = false
				continue
			}
			if sample.Value != prevValue {
				lastChange = sample.Timestamp.Time()
				prevValue = sample.Value
			}
		}
	}

	return lastChange, nil
}

// PrometheusClientInterface defines the methods for interacting with a Prometheus client.
type PrometheusClientInterface interface {
	IsPrometheusHealthy(ctx context.Context, timeout time.Duration) (bool, error)
	GetLastActivityTime(query string, interval time.Duration) (time.Time, error)
	GetQueryNginxData() string
	GetQuerySSHData() string
	GetQueryWebSSHData() string
}

// GetQueryNginxData returns the query string for Nginx data.
func (p *Prometheus) GetQueryNginxData() string {
	return p.queryNginxData
}

// GetQuerySSHData returns the query string for SSH data.
func (p *Prometheus) GetQuerySSHData() string {
	return p.queryBastionSSHData
}

// GetQueryWebSSHData returns the query string for WebSSH data.
func (p *Prometheus) GetQueryWebSSHData() string {
	return p.queryWebSSHData
}
