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

package activity_test

import (
	"context"
	"net"
	"time"

	data "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v3"
	als "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v3"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/activity"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("Envoy ALS Server", func() {
	var (
		ctx         context.Context
		cancel      context.CancelFunc
		cl          client.Client
		scheme      *runtime.Scheme
		tracker     *activity.Tracker
		instanceUID = "als-test-instance-uid-12345"
		instance    *clv1alpha2.Instance
		service     *corev1.Service
		alsServer   *activity.ALSServer
		serverAddr  string
	)

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(gatewayv1.Install(scheme)).To(Succeed())
		Expect(clv1alpha2.AddToScheme(scheme)).To(Succeed())

		instance = &clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "als-test-instance",
				Namespace: "tenant-als",
				UID:       types.UID(instanceUID),
			},
		}

		service = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "als-test-instance-default-env",
				Namespace: "tenant-als",
				Labels: map[string]string{
					forge.LabelInstanceKey: "als-test-instance",
				},
			},
		}

		cl = fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(instance, service).
			Build()

		tracker = activity.NewTracker(cl, 15*time.Minute)

		// Find a free port for testing
		var lc net.ListenConfig
		lis, err := lc.Listen(ctx, "tcp", "127.0.0.1:0")
		Expect(err).ToNot(HaveOccurred())
		serverAddr = lis.Addr().String()
		Expect(lis.Close()).To(Succeed())

		alsServer = activity.NewALSServer(tracker, serverAddr, logr.Discard())

		// Start server in background
		go func() {
			_ = alsServer.Start(ctx)
		}()

		// Wait briefly for server to bind
		Eventually(func() error {
			conn, err := net.DialTimeout("tcp", serverAddr, 100*time.Millisecond)
			if err != nil {
				return err
			}
			return conn.Close()
		}, 3*time.Second, 50*time.Millisecond).Should(Succeed())
	})

	AfterEach(func() {
		cancel()
	})

	It("processes access log stream and updates instance activity via Service in UpstreamCluster", func() {
		conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		Expect(err).ToNot(HaveOccurred())
		defer conn.Close()

		client := als.NewAccessLogServiceClient(conn)
		stream, err := client.StreamAccessLogs(ctx)
		Expect(err).ToNot(HaveOccurred())

		logTime := time.Date(2026, 10, 16, 15, 30, 0, 0, time.UTC)
		reqMsg := &als.StreamAccessLogsMessage{
			LogEntries: &als.StreamAccessLogsMessage_HttpLogs{
				HttpLogs: &als.StreamAccessLogsMessage_HTTPAccessLogEntries{
					LogEntry: []*data.HTTPAccessLogEntry{
						{
							CommonProperties: &data.AccessLogCommon{
								StartTime:       timestamppb.New(logTime),
								UpstreamCluster: "httproute/tenant-als/als-test-instance-default-env/rule/0/backend/0",
							},
							Request: &data.HTTPRequestProperties{
								Path: "/instance/als-test-instance-uid-12345/default-env",
							},
						},
						// Non-matching path / unknown cluster (should be safely ignored)
						{
							CommonProperties: &data.AccessLogCommon{
								StartTime:       timestamppb.New(logTime),
								UpstreamCluster: "other-cluster",
							},
							Request: &data.HTTPRequestProperties{
								Path: "/graph",
							},
						},
						// Non-instance website route (should be safely ignored)
						{
							CommonProperties: &data.AccessLogCommon{
								StartTime:       timestamppb.New(logTime),
								UpstreamCluster: "httproute/website/website/rule/0",
								RouteName:       "httproute/website/website/rule/0/match/0/crownlabs_polito_it",
							},
							Request: &data.HTTPRequestProperties{
								Path: "/cloud/ocs/v1.php/cloud/users/test",
							},
						},
					},
				},
			},
		}

		Expect(stream.Send(reqMsg)).To(Succeed())
		resp, err := stream.CloseAndRecv()
		Expect(err).ToNot(HaveOccurred())
		Expect(resp).ToNot(BeNil())

		// Verify annotation on instance was updated
		Eventually(func() string {
			var retrieved clv1alpha2.Instance
			if err := cl.Get(ctx, types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}, &retrieved); err != nil {
				return ""
			}
			return retrieved.Annotations[forge.LastActivityAnnotation]
		}, 3*time.Second, 100*time.Millisecond).Should(Equal(logTime.Format(time.RFC3339)))
	})
})
