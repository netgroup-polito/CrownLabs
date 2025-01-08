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
	"net"
	"net/url"
	"time"

	// criapi "k8s.io/cri-api/pkg/apis/runtime/v1".
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	criapi "k8s.io/cri-api/pkg/apis/runtime/v1"

	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
)

// unixProtocol is the network protocol of unix socket.
const unixProtocol = "unix"

// RemoteRuntimeServiceClient is a gRPC implementation of criapi.RuntimeServiceClient.
type RemoteRuntimeServiceClient struct {
	runtimeClient criapi.RuntimeServiceClient
}

// GetRuntimeService creates and returns a new RuntimeServiceClient.
func GetRuntimeService(ctx context.Context, connectionTimeout time.Duration, runtimeEndpoint string) (*RemoteRuntimeServiceClient, error) {
	log := clctx.LoggerFromContext(ctx).WithName("get-runtime-service").WithValues("timeout", connectionTimeout)
	ctx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	log.Info("Initializing connection to remote runtime service ")

	ctx = clctx.LoggerIntoContext(ctx, log)
	res, err := newRemoteRuntimeServiceClient(ctx, runtimeEndpoint)
	if err != nil {
		log.Error(err, "Error creating new remoteRuntimeServiceClient")
		return nil, err
	}

	return res, nil
}

// newRemoteRuntimeServiceClient creates a new criapi.RuntimeService.
func newRemoteRuntimeServiceClient(ctx context.Context, endpoint string) (*RemoteRuntimeServiceClient, error) {
	log := clctx.LoggerFromContext(ctx)
	log.Info("Connecting to runtime service", "endpoint", endpoint)

	URL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	addr := URL.Path

	// Open gRPC connection.
	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dial))
	if err != nil {
		log.Error(err, "Connect remote runtime failed", "address", addr)
		return nil, err
	}
	log.Info("Successfully connected to runtime service")

	serviceClient := &RemoteRuntimeServiceClient{}
	serviceClient.runtimeClient = criapi.NewRuntimeServiceClient(conn)
	return serviceClient, nil
}

// ListPodSandbox returns a list of PodSandboxes.
func (r *RemoteRuntimeServiceClient) ListPodSandbox(ctx context.Context, filter *criapi.PodSandboxFilter) ([]*criapi.PodSandbox, error) {
	log := clctx.LoggerFromContext(ctx).WithValues("PodSandboxFilter", filter)

	resp, err := r.runtimeClient.ListPodSandbox(ctx, &criapi.ListPodSandboxRequest{
		Filter: filter,
	})
	if err != nil {
		log.Error(err, "Remote runtime call failed")
		return nil, err
	}
	log.V(5).Info("Remote runtime call succeeded", "ListPodSandboxResponse", resp.GetItems())
	return resp.GetItems(), nil
}

// ListContainers lists containers by filters.
func (r *RemoteRuntimeServiceClient) ListContainers(ctx context.Context, filter *criapi.ContainerFilter) ([]*criapi.Container, error) {
	log := clctx.LoggerFromContext(ctx).WithValues("ContainerFilter", filter)

	resp, err := r.runtimeClient.ListContainers(ctx, &criapi.ListContainersRequest{
		Filter: filter,
	})
	if err != nil {
		log.Error(err, "Remote runtime call failed")
		return nil, err
	}
	log.V(5).Info("Remote runtime call succeeded", "ListContainersResponse", resp.GetContainers())
	return resp.Containers, nil
}

// ListContainerStats returns the list of ContainerStats given the filter.
func (r *RemoteRuntimeServiceClient) ListContainerStats(ctx context.Context, filter *criapi.ContainerStatsFilter) (*criapi.ListContainerStatsResponse, error) {
	log := clctx.LoggerFromContext(ctx).WithValues("ContainerStatsFilter", filter)

	resp, err := r.runtimeClient.ListContainerStats(ctx, &criapi.ListContainerStatsRequest{
		Filter: filter,
	})
	if err != nil {
		log.Error(err, "Remote runtime call failed")
		return nil, err
	}
	log.V(5).Info("Remote runtime call succeeded", "ListContainerStatsResponse", resp.GetStats())
	return resp, nil
}

func dial(ctx context.Context, addr string) (net.Conn, error) {
	return (&net.Dialer{}).DialContext(ctx, unixProtocol, addr)
}
