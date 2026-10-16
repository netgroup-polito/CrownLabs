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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/activity"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("Activity Tracker", func() {
	Context("ExtractInstanceUIDFromPath", func() {
		It("correctly extracts instance UID from standard GUI path", func() {
			uid, ok := activity.ExtractInstanceUIDFromPath("/instance/123e4567-e89b-12d3-a456-426614174000/env-default")
			Expect(ok).To(BeTrue())
			Expect(uid).To(Equal("123e4567-e89b-12d3-a456-426614174000"))
		})

		It("correctly extracts instance UID with app prefix", func() {
			uid, ok := activity.ExtractInstanceUIDFromPath("/app/instance/my-instance-uid/gui/novnc")
			Expect(ok).To(BeTrue())
			Expect(uid).To(Equal("my-instance-uid"))
		})

		It("correctly extracts instance UID when query params are present", func() {
			uid, ok := activity.ExtractInstanceUIDFromPath("/instance/inst-abc-123/env?token=xyz&foo=bar")
			Expect(ok).To(BeTrue())
			Expect(uid).To(Equal("inst-abc-123"))
		})

		It("correctly extracts instance UID without trailing segments", func() {
			uid, ok := activity.ExtractInstanceUIDFromPath("/instance/target-uid-1")
			Expect(ok).To(BeTrue())
			Expect(uid).To(Equal("target-uid-1"))
		})

		It("fails when path does not contain instance prefix", func() {
			uid, ok := activity.ExtractInstanceUIDFromPath("/graph")
			Expect(ok).To(BeFalse())
			Expect(uid).To(BeEmpty())
		})

		It("fails on empty path or trailing slash without UID", func() {
			uid, ok := activity.ExtractInstanceUIDFromPath("")
			Expect(ok).To(BeFalse())
			Expect(uid).To(BeEmpty())

			uid, ok = activity.ExtractInstanceUIDFromPath("/instance/")
			Expect(ok).To(BeFalse())
			Expect(uid).To(BeEmpty())
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
			Expect(tracker.ShouldUpdate(instanceUID, now)).To(BeTrue())

			err := tracker.RecordActivity(ctx, instanceUID, now)
			Expect(err).ToNot(HaveOccurred())

			// Verify annotation was written on the instance
			var retrieved clv1alpha2.Instance
			Expect(cl.Get(ctx, types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}, &retrieved)).To(Succeed())
			Expect(retrieved.Annotations[forge.LastActivityAnnotation]).To(Equal(now.Format(time.RFC3339)))
		})

		It("should debounce updates within 15 minutes", func() {
			// First update at 12:00
			Expect(tracker.RecordActivity(ctx, instanceUID, now)).To(Succeed())

			// Second update at 12:05 (only 5 minutes elapsed)
			timeFiveMinLater := now.Add(5 * time.Minute)
			Expect(tracker.ShouldUpdate(instanceUID, timeFiveMinLater)).To(BeFalse())

			// RecordActivity at 12:05 should not update the annotation
			Expect(tracker.RecordActivity(ctx, instanceUID, timeFiveMinLater)).To(Succeed())

			var retrieved clv1alpha2.Instance
			Expect(cl.Get(ctx, types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}, &retrieved)).To(Succeed())
			// Timestamp should still be 12:00
			Expect(retrieved.Annotations[forge.LastActivityAnnotation]).To(Equal(now.Format(time.RFC3339)))
		})

		It("should perform update after 15 minutes have elapsed", func() {
			// First update at 12:00
			Expect(tracker.RecordActivity(ctx, instanceUID, now)).To(Succeed())

			// Next update at 12:15:01 (15m 1s elapsed)
			timeFifteenMinLater := now.Add(15*time.Minute + 1*time.Second)
			Expect(tracker.ShouldUpdate(instanceUID, timeFifteenMinLater)).To(BeTrue())

			Expect(tracker.RecordActivity(ctx, instanceUID, timeFifteenMinLater)).To(Succeed())

			var retrieved clv1alpha2.Instance
			Expect(cl.Get(ctx, types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}, &retrieved)).To(Succeed())
			// Timestamp should now be updated to 12:15:01
			Expect(retrieved.Annotations[forge.LastActivityAnnotation]).To(Equal(timeFifteenMinLater.Format(time.RFC3339)))
		})

		It("returns error when instance UID does not exist", func() {
			err := tracker.RecordActivity(ctx, "non-existent-uid", now)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})

		It("returns error when instance UID is empty", func() {
			err := tracker.RecordActivity(ctx, "", now)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("empty instance UID"))
		})
	})
})
