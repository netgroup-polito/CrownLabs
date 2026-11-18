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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
		Expect(create(newWorkspace("ws-priority-absent", &clv1alpha1.JobOptions{}))).To(Succeed())
		raw, decoded := stored("ws-priority-absent")
		Expect(raw).To(HaveKeyWithValue("priority", BeNumerically("==", 100)))
		Expect(decoded.Priority).To(HaveValue(BeEquivalentTo(100)))
	})

	It("keeps a priority of 0, which is not the default", func() {
		Expect(create(newWorkspace("ws-priority-zero", &clv1alpha1.JobOptions{Priority: ptr.To[int32](0)}))).To(Succeed())
		raw, decoded := stored("ws-priority-zero")
		Expect(raw).To(HaveKeyWithValue("priority", BeNumerically("==", 0)))
		Expect(decoded.Priority).To(HaveValue(BeEquivalentTo(0)))
	})

	It("keeps a priority of 10000, the maximum", func() {
		Expect(create(newWorkspace("ws-priority-max", &clv1alpha1.JobOptions{Priority: ptr.To[int32](10000)}))).To(Succeed())
		raw, decoded := stored("ws-priority-max")
		Expect(raw).To(HaveKeyWithValue("priority", BeNumerically("==", 10000)))
		Expect(decoded.Priority).To(HaveValue(BeEquivalentTo(10000)))
	})

	DescribeTable("refuses a priority out of the range from 0 to 10000",
		func(name string, priority int32, message string) {
			err := create(newWorkspace(name, &clv1alpha1.JobOptions{Priority: ptr.To(priority)}))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(message))
		},
		Entry("10001", "ws-priority-10001", int32(10001), "spec.jobOptions.priority in body should be less than or equal to 10000"),
		Entry("-1", "ws-priority-minus-1", int32(-1), "spec.jobOptions.priority in body should be greater than or equal to 0"),
	)

	It("refuses jobOptions without queueLimits", func() {
		ws := &unstructured.Unstructured{Object: map[string]interface{}{
			"apiVersion": clv1alpha1.GroupVersion.String(),
			"kind":       "Workspace",
			"metadata":   map[string]interface{}{"name": "ws-no-queuelimits"},
			"spec": map[string]interface{}{
				"prettyName": "Workspace with jobs",
				"quota":      map[string]interface{}{"cpu": int64(4), "memory": "8Gi", "instances": int64(2)},
				"jobOptions": map[string]interface{}{"priority": int64(100)},
			},
		}}
		err := create(ws)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.jobOptions.queueLimits: Required value"))
	})

	It("always sends queueLimits from Go, as an empty object if no limit is set", func() {
		Expect(create(newWorkspace("ws-queuelimits-empty", &clv1alpha1.JobOptions{}))).To(Succeed())
		raw, decoded := stored("ws-queuelimits-empty")
		Expect(raw).To(HaveKeyWithValue("queueLimits", BeEmpty()))
		Expect(decoded.QueueLimits.MaxAtOnce).To(BeNil())
	})

	It("stores the limits of one job, maxRunTime and resources, in queueLimits", func() {
		limits := apicommon.JobLimits{
			MaxRunTime: "6h",
			Resources:  map[string]resource.Quantity{"cpu": resource.MustParse("4"), "nvidia.com/gpu": resource.MustParse("1")},
		}
		Expect(create(newWorkspace("ws-joblimits", &clv1alpha1.JobOptions{QueueLimits: clv1alpha1.QueueLimits{JobLimits: limits}}))).To(Succeed())
		raw, decoded := stored("ws-joblimits")
		Expect(raw).To(HaveKeyWithValue("queueLimits", HaveKeyWithValue("maxRunTime", "6h")))
		Expect(raw).To(HaveKeyWithValue("queueLimits", HaveKey("resources")))
		Expect(decoded.QueueLimits.MaxRunTime).To(Equal("6h"))
		Expect(decoded.QueueLimits.Resources).To(HaveKeyWithValue("nvidia.com/gpu", resource.MustParse("1")))
	})

	It("refuses a maxRunTime of queueLimits that does not match the pattern shared with the Template", func() {
		limits := apicommon.JobLimits{MaxRunTime: "30s"}
		err := create(newWorkspace("ws-maxruntime-30s", &clv1alpha1.JobOptions{QueueLimits: clv1alpha1.QueueLimits{JobLimits: limits}}))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.jobOptions.queueLimits.maxRunTime in body should match '^[1-9][0-9]{0,3}[mhd]$'"))
	})
})
