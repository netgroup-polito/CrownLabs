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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

var _ = Describe("Activity Tracker", func() {
	Context("ExtractTargetFromClusterOrRoute", func() {
		It("correctly extracts namespace and name from Envoy Gateway cluster name", func() {
			ns, name, ok := activity.ExtractTargetFromClusterOrRoute("httproute/tenant-mario/my-vm-default-env/rule/0/backend/0")
			Expect(ok).To(BeTrue())
			Expect(ns).To(Equal("tenant-mario"))
			Expect(name).To(Equal("my-vm-default-env"))
		})

		It("correctly extracts namespace and name from Envoy Gateway route name", func() {
			ns, name, ok := activity.ExtractTargetFromClusterOrRoute("httproute/tenant-lucia/lab-workstation-gui/rule/0")
			Expect(ok).To(BeTrue())
			Expect(ns).To(Equal("tenant-lucia"))
			Expect(name).To(Equal("lab-workstation-gui"))
		})

		It("fails when target string does not match expected format", func() {
			ns, name, ok := activity.ExtractTargetFromClusterOrRoute("passthrough_cluster")
			Expect(ok).To(BeFalse())
			Expect(ns).To(BeEmpty())
			Expect(name).To(BeEmpty())

			ns, name, ok = activity.ExtractTargetFromClusterOrRoute("")
			Expect(ok).To(BeFalse())
			Expect(ns).To(BeEmpty())
			Expect(name).To(BeEmpty())
		})
	})

	Context("Service-based Activity Tracking", func() {
		var (
			ctx         context.Context
			cl          client.Client
			scheme      *runtime.Scheme
			tracker     *activity.Tracker
			instance    *clv1alpha2.Instance
			service     *corev1.Service
			httpRoute   *gatewayv1.HTTPRoute
			now         time.Time
			namespace   = "tenant-tester"
			instName    = "test-instance"
			serviceName = "test-instance-default-env"
		)

		BeforeEach(func() {
			ctx = context.Background()
			scheme = runtime.NewScheme()
			Expect(corev1.AddToScheme(scheme)).To(Succeed())
			Expect(gatewayv1.Install(scheme)).To(Succeed())
			Expect(clv1alpha2.AddToScheme(scheme)).To(Succeed())

			instance = &clv1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      instName,
					Namespace: namespace,
					UID:       "instance-uid-12345",
				},
			}

			service = &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      serviceName,
					Namespace: namespace,
					Labels: map[string]string{
						forge.LabelInstanceKey: instName,
					},
				},
			}

			httpRoute = &gatewayv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "route-without-service",
					Namespace: namespace,
					Labels: map[string]string{
						forge.LabelInstanceKey: instName,
					},
				},
			}

			cl = fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(instance, service, httpRoute).
				Build()

			tracker = activity.NewTracker(cl, 15*time.Minute)
			now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
		})

		It("resolves instance from Service label and updates activity", func() {
			Expect(tracker.RecordActivityForService(ctx, namespace, serviceName, now)).To(Succeed())

			var retrieved clv1alpha2.Instance
			Expect(cl.Get(ctx, types.NamespacedName{Namespace: namespace, Name: instName}, &retrieved)).To(Succeed())
			Expect(retrieved.Annotations[forge.LastActivityAnnotation]).To(Equal(now.Format(time.RFC3339)))
		})

		It("resolves instance from HTTPRoute when Service is absent", func() {
			Expect(tracker.RecordActivityForService(ctx, namespace, "route-without-service", now)).To(Succeed())

			var retrieved clv1alpha2.Instance
			Expect(cl.Get(ctx, types.NamespacedName{Namespace: namespace, Name: instName}, &retrieved)).To(Succeed())
			Expect(retrieved.Annotations[forge.LastActivityAnnotation]).To(Equal(now.Format(time.RFC3339)))
		})

		It("returns error when service or route does not exist", func() {
			err := tracker.RecordActivityForService(ctx, namespace, "non-existent-svc", now)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Debounce logic and Kubernetes patching", func() {
		var (
			ctx         context.Context
			cl          client.Client
			scheme      *runtime.Scheme
			tracker     *activity.Tracker
			instanceUID = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
			instance    *clv1alpha2.Instance
			now         time.Time
		)

		BeforeEach(func() {
			ctx = context.Background()
			scheme = runtime.NewScheme()
			Expect(clv1alpha2.AddToScheme(scheme)).To(Succeed())

			instance = &clv1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-instance",
					Namespace: "tenant-tester",
					UID:       types.UID(instanceUID),
				},
			}

			cl = fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(instance).
				Build()

			tracker = activity.NewTracker(cl, 15*time.Minute)
			now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
		})

		It("should use default 15m interval when non-positive is passed", func() {
			t := activity.NewTracker(cl, 0)
			Expect(t.GetDebounceInterval()).To(Equal(15 * time.Minute))

			t2 := activity.NewTracker(cl, -5*time.Minute)
			Expect(t2.GetDebounceInterval()).To(Equal(15 * time.Minute))
		})

		It("should update when instance is seen for the first time", func() {
			key := instance.Namespace + "/" + instance.Name
			Expect(tracker.ShouldUpdate(key, now)).To(BeTrue())

			err := tracker.RecordActivityForInstance(ctx, instance, now)
			Expect(err).ToNot(HaveOccurred())

			// Verify annotation was written on the instance
			var retrieved clv1alpha2.Instance
			Expect(cl.Get(ctx, types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}, &retrieved)).To(Succeed())
			Expect(retrieved.Annotations[forge.LastActivityAnnotation]).To(Equal(now.Format(time.RFC3339)))
		})

		It("should debounce updates within 15 minutes", func() {
			key := instance.Namespace + "/" + instance.Name
			// First update at 12:00
			Expect(tracker.RecordActivityForInstance(ctx, instance, now)).To(Succeed())

			// Second update at 12:05 (only 5 minutes elapsed)
			timeFiveMinLater := now.Add(5 * time.Minute)
			Expect(tracker.ShouldUpdate(key, timeFiveMinLater)).To(BeFalse())

			// RecordActivity at 12:05 should not update the annotation
			Expect(tracker.RecordActivityForInstance(ctx, instance, timeFiveMinLater)).To(Succeed())

			var retrieved clv1alpha2.Instance
			Expect(cl.Get(ctx, types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}, &retrieved)).To(Succeed())
			// Timestamp should still be 12:00
			Expect(retrieved.Annotations[forge.LastActivityAnnotation]).To(Equal(now.Format(time.RFC3339)))
		})

		It("should perform update after 15 minutes have elapsed", func() {
			key := instance.Namespace + "/" + instance.Name
			// First update at 12:00
			Expect(tracker.RecordActivityForInstance(ctx, instance, now)).To(Succeed())

			// Next update at 12:15:01 (15m 1s elapsed)
			timeFifteenMinLater := now.Add(15*time.Minute + 1*time.Second)
			Expect(tracker.ShouldUpdate(key, timeFifteenMinLater)).To(BeTrue())

			Expect(tracker.RecordActivityForInstance(ctx, instance, timeFifteenMinLater)).To(Succeed())

			var retrieved clv1alpha2.Instance
			Expect(cl.Get(ctx, types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}, &retrieved)).To(Succeed())
			// Timestamp should now be updated to 12:15:01
			Expect(retrieved.Annotations[forge.LastActivityAnnotation]).To(Equal(timeFifteenMinLater.Format(time.RFC3339)))
		})
	})
})
