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

package activity

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	als "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v3"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// ALSServer implements the Envoy gRPC AccessLogService.
type ALSServer struct {
	als.UnimplementedAccessLogServiceServer
	tracker    *Tracker
	bindAddr   string
	grpcServer *grpc.Server
	log        logr.Logger
}

// NewALSServer creates a new Envoy ALS server instance.
func NewALSServer(tracker *Tracker, bindAddr string, log logr.Logger) *ALSServer {
	if bindAddr == "" {
		bindAddr = ":9000"
	}
	return &ALSServer{
		tracker:  tracker,
		bindAddr: bindAddr,
		log:      log.WithName("envoy-als-server"),
	}
}

// StreamAccessLogs handles incoming streaming access logs from Envoy proxies.
func (s *ALSServer) StreamAccessLogs(stream als.AccessLogService_StreamAccessLogsServer) error {
	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return stream.SendAndClose(&als.StreamAccessLogsResponse{})
		}
		if err != nil {
			s.log.V(1).Info("ALS stream terminated or errored", "error", err.Error())
			return err
		}

		httpLogs := msg.GetHttpLogs()
		if httpLogs == nil {
			continue
		}

		for _, entry := range httpLogs.GetLogEntry() {
			var path string
			if req := entry.GetRequest(); req != nil {
				path = req.GetPath()
			}

			// Discard immediately if the request path does not contain the instance prefix.
			if !strings.Contains(path, InstancePathPrefix) {
				continue
			}

			var cluster, route string
			if common := entry.GetCommonProperties(); common != nil {
				cluster = common.GetUpstreamCluster()
				route = common.GetRouteName()
			}

			s.log.Info("Incoming ALS instance log entry",
				"cluster", cluster,
				"route", route,
				"path", path,
			)

			var timestamp time.Time
			if common := entry.GetCommonProperties(); common != nil && common.GetStartTime() != nil {
				timestamp = common.GetStartTime().AsTime()
			} else {
				timestamp = time.Now()
			}

			// 1. Primary: Extract target namespace and service from UpstreamCluster or RouteName
			var targetNamespace, targetName string
			if common := entry.GetCommonProperties(); common != nil {
				targetNamespace, targetName, _ = ExtractTargetFromClusterOrRoute(common.GetUpstreamCluster())
				if targetNamespace == "" || targetName == "" {
					targetNamespace, targetName, _ = ExtractTargetFromClusterOrRoute(common.GetRouteName())
				}
			}

			if targetNamespace != "" && targetName != "" {
				if err := s.tracker.RecordActivityForService(ctx, targetNamespace, targetName, timestamp); err != nil {
					s.log.V(utils.LogDebugLevel).Info("Could not record activity for service entry", "namespace", targetNamespace, "name", targetName, "error", err.Error())
				}
			}
		}
	}
}

// Start begins listening on bindAddr and serves gRPC requests until the context is canceled.
// It implements controller-runtime's Runnable interface so it can be added directly to ctrl.Manager.
func (s *ALSServer) Start(ctx context.Context) error {
	var lc net.ListenConfig
	lis, err := lc.Listen(ctx, "tcp", s.bindAddr)
	if err != nil {
		return fmt.Errorf("failed listening on %s for ALS: %w", s.bindAddr, err)
	}

	s.grpcServer = grpc.NewServer()
	als.RegisterAccessLogServiceServer(s.grpcServer, s)

	s.log.Info("Envoy ALS gRPC server listening", "address", s.bindAddr)

	errChan := make(chan error, 1)
	go func() {
		if serveErr := s.grpcServer.Serve(lis); serveErr != nil && !errors.Is(serveErr, grpc.ErrServerStopped) {
			errChan <- serveErr
		}
		close(errChan)
	}()

	select {
	case <-ctx.Done():
		s.log.Info("Shutting down Envoy ALS gRPC server gracefully")
		s.grpcServer.GracefulStop()
		return nil
	case serveErr := <-errChan:
		return serveErr
	}
}

// NeedLeaderElection specifies if this runnable needs leader election.
// Returns false to ensure ALS can receive logs across all active replicas if needed.
func (s *ALSServer) NeedLeaderElection() bool {
	return false
}
