// Copyright 2020-2024 Politecnico di Torino
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
	"log"
	"time"

	pb "github.com/novnc/websockify-other/websockify/instmetrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InstanceMetricsClient is a gRPC implementation of pb.InstanceMetricsClient.
type InstanceMetricsClient struct {
	client pb.InstanceMetricsClient
}

// GetInstanceMetricsClient creates and returns a new InstanceMetricsClient.
func GetInstanceMetricsClient(ctx context.Context, connectionTimeout time.Duration, instanceMetricsEndpoint string) (*InstanceMetricsClient, error) {
	ctx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	log.Println("Initializing connection to remote runtime service")

	res, err := newInstanceMetricsClient(ctx, instanceMetricsEndpoint)
	if err != nil {
		log.Println(err, "Error creating new InstanceMetricsClient")
		return nil, err
	}
	return res, nil
}

// newRemoteRuntimeServiceClient creates a new InstanceMetricsClient.
func newInstanceMetricsClient(ctx context.Context, endpoint string) (*InstanceMetricsClient, error) {
	log.Println("Connecting to instance metrics server", "endpoint", endpoint)

	connection, err := grpc.DialContext(ctx, endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Printf("An error occurred while creating instance metrics server client: %v", err)
		return nil, err
	}
	log.Println("Successfully connected to runtime service")

	instMetricsClient := &InstanceMetricsClient{client: pb.NewInstanceMetricsClient(connection)}
	return instMetricsClient, nil
}

// ContainerMetrics returns the containerMetrics given a podName.
func (c *InstanceMetricsClient) ContainerMetrics(ctx context.Context, podName *string) (*pb.ContainerMetricsResponse, error) {
	resp, err := c.client.ContainerMetrics(ctx, &pb.ContainerMetricsRequest{
		PodName: *podName,
	})
	if err != nil {
		log.Printf("ContainerMetrics with podName filter '%s' failed. Error: %v", *podName, err)
		return nil, err
	}
	return resp, nil
}
