package instautoctrl

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Prometheus struct {
	address                  string
	queryNginxAvailable      string
	queryBastionSSHAvailable string
	queryNginxData           string
	queryBastionSSHData      string
	client                   v1.API
}

func NewPrometheusObj(
	address string,
	queryNginxAvailable, queryBastionSSHAvailable, queryNginxData, queryBastionSSHData string,
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
		queryNginxData:           queryNginxData,
		queryBastionSSHData:      queryBastionSSHData,
		client:                   v1api,
	}, nil
}

// IsPrometheusHealthy checks if Prometheus and required metrics are available.
func (p *Prometheus) IsPrometheusHealthy(ctx context.Context) (bool, error) {
	log := ctrl.LoggerFrom(ctx).WithName("prometheus-health")

	// Verify connection to Prometheus health endpoint
	healthEndpoint := fmt.Sprintf("%s/-/healthy", p.address)

	statusCode, _, err := utils.HTTPGet(ctx, healthEndpoint, 5*time.Second)
	if err != nil {
		log.Error(err, "Failed to connect to Prometheus health endpoint")
		return false, fmt.Errorf("prometheus health check failed: %w", err)
	}

	if statusCode != http.StatusOK {
		log.Info("Prometheus health check returned non-OK status", "statusCode", statusCode)
		return false, nil
	}

	// Check if ingress metrics and bastion metrics are available on worker nodes
	query1 := p.queryNginxAvailable
	query2 := p.queryBastionSSHAvailable

	result1, _, err1 := p.client.Query(ctx, query1, time.Now())
	result2, _, err2 := p.client.Query(ctx, query2, time.Now())

	if err1 != nil && err2 != nil {
		log.Error(err1, "Failed to query Prometheus for ingress metrics")
		log.Error(err2, "Failed to query Prometheus for bastion SSH metrics")
		return false, fmt.Errorf("both Prometheus queries failed: %v, %v", err1, err2)
	}

	active1 := false
	active2 := false

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

	if !active1 && !active2 {
		log.Info("Neither ingress metrics nor bastion SSH metrics are available on worker nodes")
		return false, nil
	}

	// At least one node has ingress metrics available
	return true, nil
}

// GetLastActivityTime retrieves the last time an instance was accessed.
func (p *Prometheus) GetLastActivityTime(query string, interval time.Duration) (time.Time, error) {
	end := time.Now()
	start := end.Add(-interval)

	r := v1.Range{
		Start: start,
		End:   end,
		Step:  time.Minute,
	}

	result, warnings, err := p.client.QueryRange(context.Background(), query, r)
	if err != nil {
		return time.Time{}, fmt.Errorf("query failed: %w", err)
	}
	if len(warnings) > 0 {
		fmt.Println("Warnings:", warnings)
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

	if lastChange.IsZero() {
		return time.Time{}, fmt.Errorf("no changes detected")
	}
	return lastChange, nil
}

type PrometheusClientInterface interface {
	IsPrometheusHealthy(ctx context.Context) (bool, error)
	GetLastActivityTime(query string, interval time.Duration) (time.Time, error)
	GetQueryNginxData() string
	GetQuerySSHData() string
}

// client, err := api.NewClient(
// 		api.Config{
// 			Address: address,
// 		},
// 	)
// 	if err != nil {
// 		return nil, err
// 	}

// 	v1api := v1.NewAPI(client)
// 	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
// 	defer cancel()

// 	return v1api, nil

func (p *Prometheus) GetQueryNginxData() string {
	return p.queryNginxData
}

func (p *Prometheus) GetQuerySSHData() string {
	return p.queryBastionSSHData
}
