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

package v1alpha1_test

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apicommon "github.com/netgroup-polito/CrownLabs/operators/api/common"
	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
)

var _ = Describe("The jobOptions of a Workspace", func() {
	ctx := context.Background()

	const limitsMessage = "queueLimits.limits must list at least one resource"

	// limits returns the limits of a job queue, with one resource.
	limits := func() map[string]resource.Quantity {
		return map[string]resource.Quantity{"cpu": resource.MustParse("8")}
	}

	// withLimits returns jobOptions with the given priority and a queue with limits.
	withLimits := func(priority *int32) *clv1alpha1.JobOptions {
		return &clv1alpha1.JobOptions{Priority: priority, QueueLimits: clv1alpha1.QueueLimits{Limits: limits()}}
	}

	// newWorkspace returns a Workspace with the given jobOptions.
	newWorkspace := func(name string, jobOptions *clv1alpha1.JobOptions) *clv1alpha1.Workspace {
		return &clv1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec: clv1alpha1.WorkspaceSpec{
				PrettyName: "Workspace with jobs",
				Quota: apicommon.WorkspaceResourceQuota{
					ResourceSpec: apicommon.ResourceSpec{CPU: 4, Memory: resource.MustParse("8Gi")},
					Instances:    2,
				},
				JobOptions: jobOptions,
			},
		}
	}

	// rawWorkspace returns a Workspace as a YAML file would describe it, with the given jobOptions.
	rawWorkspace := func(name string, jobOptions map[string]interface{}) *unstructured.Unstructured {
		return &unstructured.Unstructured{Object: map[string]interface{}{
			"apiVersion": clv1alpha1.GroupVersion.String(),
			"kind":       "Workspace",
			"metadata":   map[string]interface{}{"name": name},
			"spec": map[string]interface{}{
				"prettyName": "Workspace with jobs",
				"quota":      map[string]interface{}{"cpu": int64(4), "memory": "8Gi", "instances": int64(2)},
				"jobOptions": jobOptions,
			},
		}}
	}

	// create creates the Workspace, and deletes it at the end of the spec.
	create := func(ws client.Object) error {
		err := k8sClient.Create(ctx, ws)
		if err == nil {
			DeferCleanup(func() { Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, ws))).To(Succeed()) })
		}
		return err
	}

	// stored returns the spec.jobOptions of a Workspace as the API server
	// stores it, both as raw JSON fields and decoded into the Go type.
	stored := func(name string) (raw map[string]interface{}, decoded *clv1alpha1.JobOptions) {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(clv1alpha1.GroupVersion.WithKind("Workspace"))
		Expect(k8sClient.Get(ctx, client.ObjectKey{Name: name}, u)).To(Succeed())
		raw, found, err := unstructured.NestedMap(u.Object, "spec", "jobOptions")
		Expect(err).ToNot(HaveOccurred())
		Expect(found).To(BeTrue())

		ws := &clv1alpha1.Workspace{}
		Expect(k8sClient.Get(ctx, client.ObjectKey{Name: name}, ws)).To(Succeed())
		Expect(ws.Spec.JobOptions).ToNot(BeNil())
		return raw, ws.Spec.JobOptions
	}

	It("gives an absent priority the default of 100", func() {
		Expect(create(newWorkspace("ws-priority-absent", withLimits(nil)))).To(Succeed())
		raw, decoded := stored("ws-priority-absent")
		Expect(raw).To(HaveKeyWithValue("priority", BeNumerically("==", 100)))
		Expect(decoded.Priority).To(HaveValue(BeEquivalentTo(100)))
	})

	It("keeps a priority of 0, which is not the default", func() {
		Expect(create(newWorkspace("ws-priority-zero", withLimits(ptr.To[int32](0))))).To(Succeed())
		raw, decoded := stored("ws-priority-zero")
		Expect(raw).To(HaveKeyWithValue("priority", BeNumerically("==", 0)))
		Expect(decoded.Priority).To(HaveValue(BeEquivalentTo(0)))
	})

	It("keeps a priority of 10000, the maximum", func() {
		Expect(create(newWorkspace("ws-priority-max", withLimits(ptr.To[int32](10000))))).To(Succeed())
		raw, decoded := stored("ws-priority-max")
		Expect(raw).To(HaveKeyWithValue("priority", BeNumerically("==", 10000)))
		Expect(decoded.Priority).To(HaveValue(BeEquivalentTo(10000)))
	})

	DescribeTable("refuses a priority out of the range from 0 to 10000",
		func(name string, priority int32, message string) {
			err := create(newWorkspace(name, withLimits(ptr.To(priority))))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(message))
		},
		Entry("10001", "ws-priority-10001", int32(10001), "spec.jobOptions.priority in body should be less than or equal to 10000"),
		Entry("-1", "ws-priority-minus-1", int32(-1), "spec.jobOptions.priority in body should be greater than or equal to 0"),
	)

	It("refuses jobOptions without queueLimits", func() {
		err := create(rawWorkspace("ws-no-queuelimits", map[string]interface{}{"priority": int64(100)}))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.jobOptions.queueLimits: Required value"))
	})

	It("stores requests and limits", func() {
		jobOptions := withLimits(nil)
		jobOptions.QueueLimits.Limits["nvidia.com/gpu"] = resource.MustParse("2")
		jobOptions.QueueLimits.Requests = map[string]resource.Quantity{"cpu": resource.MustParse("2")}
		Expect(create(newWorkspace("ws-requests-limits", jobOptions))).To(Succeed())
		raw, decoded := stored("ws-requests-limits")
		Expect(raw).To(HaveKeyWithValue("queueLimits", HaveKeyWithValue("limits", HaveKeyWithValue("nvidia.com/gpu", "2"))))
		Expect(raw).To(HaveKeyWithValue("queueLimits", HaveKeyWithValue("requests", HaveKeyWithValue("cpu", "2"))))
		Expect(decoded.QueueLimits.Limits).To(HaveKeyWithValue("cpu", resource.MustParse("8")))
		Expect(decoded.QueueLimits.Requests).To(HaveKeyWithValue("cpu", resource.MustParse("2")))
	})

	It("refuses jobOptions without limits, as a Go client sends them when no limit is set", func() {
		jobOptions := &clv1alpha1.JobOptions{QueueLimits: clv1alpha1.QueueLimits{Requests: limits()}}
		sent, err := json.Marshal(jobOptions)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(sent)).ToNot(ContainSubstring(`"limits"`))

		err = create(newWorkspace("ws-no-limits", jobOptions))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(limitsMessage))
	})

	It("refuses limits that list no resource", func() {
		err := create(rawWorkspace("ws-limits-empty", map[string]interface{}{
			"queueLimits": map[string]interface{}{"limits": map[string]interface{}{}},
		}))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(limitsMessage))
	})

	It("accepts a window whose overrides have no limits", func() {
		Expect(create(rawWorkspace("ws-window-no-limits", map[string]interface{}{
			"queueLimits": map[string]interface{}{"limits": map[string]interface{}{"cpu": "8"}},
			"overrideWindows": []interface{}{map[string]interface{}{
				"days": []interface{}{"Sat"}, "start": "08:00", "end": "20:00",
				"overrides": map[string]interface{}{"requests": map[string]interface{}{"cpu": "4"}},
			}},
		}))).To(Succeed())
		raw, _ := stored("ws-window-no-limits")
		windows, found, err := unstructured.NestedSlice(raw, "overrideWindows")
		Expect(err).ToNot(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(windows).To(HaveLen(1))
		Expect(windows[0]).To(HaveKeyWithValue("overrides", And(HaveKey("requests"), Not(HaveKey("limits")))))
	})

	It("stores the limits of one job, maxRunTime and resources, under queueLimits.jobLimits", func() {
		Expect(create(rawWorkspace("ws-joblimits", map[string]interface{}{
			"queueLimits": map[string]interface{}{
				"limits": map[string]interface{}{"cpu": "8"},
				"jobLimits": map[string]interface{}{
					"maxRunTime": "6h",
					"resources":  map[string]interface{}{"cpu": "4", "nvidia.com/gpu": "1"},
				},
			},
		}))).To(Succeed())
		raw, decoded := stored("ws-joblimits")
		queueLimits, _, err := unstructured.NestedMap(raw, "queueLimits")
		Expect(err).ToNot(HaveOccurred())
		Expect(queueLimits).To(HaveKeyWithValue("jobLimits", HaveKeyWithValue("maxRunTime", "6h")))
		Expect(queueLimits).To(HaveKeyWithValue("jobLimits", HaveKey("resources")))
		Expect(queueLimits).ToNot(HaveKey("maxRunTime"))
		Expect(queueLimits).ToNot(HaveKey("resources"))
		Expect(decoded.QueueLimits.JobLimits).ToNot(BeNil())
		Expect(decoded.QueueLimits.JobLimits.MaxRunTime).To(Equal("6h"))
		Expect(decoded.QueueLimits.JobLimits.Resources).To(HaveKeyWithValue("nvidia.com/gpu", resource.MustParse("1")))
	})

	It("sends no jobLimits from a Go client that leaves them unset, so that a window keeps the default", func() {
		window := clv1alpha1.JobWindow{
			Days: []clv1alpha1.WeekDay{clv1alpha1.WeekDaySaturday}, Start: "08:00", End: "20:00",
			Overrides: clv1alpha1.QueueLimits{Requests: limits()},
		}
		sent, err := json.Marshal(window)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(sent)).ToNot(ContainSubstring(`"jobLimits"`))

		jobOptions := withLimits(nil)
		jobOptions.OverrideWindows = []clv1alpha1.JobWindow{window}
		Expect(create(newWorkspace("ws-joblimits-unset", jobOptions))).To(Succeed())
		raw, decoded := stored("ws-joblimits-unset")
		Expect(raw).To(HaveKeyWithValue("queueLimits", Not(HaveKey("jobLimits"))))
		Expect(raw).To(HaveKeyWithValue("overrideWindows", ContainElement(HaveKeyWithValue("overrides", Not(HaveKey("jobLimits"))))))
		Expect(decoded.QueueLimits.JobLimits).To(BeNil())
		Expect(decoded.OverrideWindows[0].Overrides.JobLimits).To(BeNil())
	})

	It("refuses a maxRunTime of jobLimits that does not match the pattern shared with the Template", func() {
		err := create(rawWorkspace("ws-maxruntime-30s", map[string]interface{}{
			"queueLimits": map[string]interface{}{
				"limits":    map[string]interface{}{"cpu": "8"},
				"jobLimits": map[string]interface{}{"maxRunTime": "30s"},
			},
		}))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.jobOptions.queueLimits.jobLimits.maxRunTime in body should match '^[1-9][0-9]{0,3}[mhd]$'"))
	})
})

var _ = Describe("The conditions of a Workspace", func() {
	ctx := context.Background()

	It("stores a QueueReady condition written through the status subresource, and keeps it on a spec update", func() {
		ws := &clv1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{Name: "ws-queue-ready"},
			Spec: clv1alpha1.WorkspaceSpec{
				PrettyName: "Workspace with a queue",
				Quota: apicommon.WorkspaceResourceQuota{
					ResourceSpec: apicommon.ResourceSpec{CPU: 4, Memory: resource.MustParse("8Gi")},
					Instances:    2,
				},
			},
		}
		Expect(k8sClient.Create(ctx, ws)).To(Succeed())
		DeferCleanup(func() { Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, ws))).To(Succeed()) })

		meta.SetStatusCondition(&ws.Status.Conditions, metav1.Condition{
			Type:    clv1alpha1.WorkspaceConditionQueueReady,
			Status:  metav1.ConditionFalse,
			Reason:  "CapacityExceeded",
			Message: "the requests of all the queues would exceed the job capacity",
		})
		Expect(k8sClient.Status().Update(ctx, ws)).To(Succeed())

		read := &clv1alpha1.Workspace{}
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(ws), read)).To(Succeed())
		condition := meta.FindStatusCondition(read.Status.Conditions, clv1alpha1.WorkspaceConditionQueueReady)
		Expect(condition).ToNot(BeNil())
		Expect(condition.Status).To(Equal(metav1.ConditionFalse))
		Expect(condition.Reason).To(Equal("CapacityExceeded"))

		// A spec update that sends no status at all: the API server keeps the stored one.
		read.Spec.PrettyName = "Workspace with a queue, renamed"
		read.Status = clv1alpha1.WorkspaceStatus{}
		Expect(k8sClient.Update(ctx, read)).To(Succeed())

		after := &clv1alpha1.Workspace{}
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(ws), after)).To(Succeed())
		Expect(after.Spec.PrettyName).To(Equal("Workspace with a queue, renamed"))
		Expect(meta.IsStatusConditionFalse(after.Status.Conditions, clv1alpha1.WorkspaceConditionQueueReady)).To(BeTrue())
	})
})
