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

package v1alpha2_test

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var _ = Describe("The job section of a Template", func() {
	ctx := context.Background()

	// template returns a Template as a YAML file would describe it, with the given job section.
	template := func(name string, job map[string]interface{}) *unstructured.Unstructured {
		return &unstructured.Unstructured{Object: map[string]interface{}{
			"apiVersion": clv1alpha2.GroupVersion.String(),
			"kind":       "Template",
			"metadata":   map[string]interface{}{"name": name, "namespace": testNamespace},
			"spec": map[string]interface{}{
				"prettyName":  "Template with jobs",
				"description": "A Template that can run jobs",
				"environmentList": []interface{}{map[string]interface{}{
					"name":            "job-env",
					"environmentType": "Standalone",
					"image":           "harbor.crownlabs.polito.it/crownlabs-standalone/jupyterlab:1.0.0",
					"persistent":      true,
					"resources":       map[string]interface{}{"cpu": int64(2), "memory": "2Gi", "reservedCPUPercentage": int64(25)},
				}},
				"job": job,
			},
		}}
	}

	// create creates the Template, and deletes it at the end of the spec.
	create := func(tpl *unstructured.Unstructured) error {
		err := k8sClient.Create(ctx, tpl)
		if err == nil {
			DeferCleanup(func() { Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, tpl))).To(Succeed()) })
		}
		return err
	}

	// stored returns the job section of a Template as the API server stores
	// it, both as raw JSON fields and decoded into the Go type.
	stored := func(name string) (raw map[string]interface{}, decoded *clv1alpha2.TemplateJob) {
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(clv1alpha2.GroupVersion.WithKind("Template"))
		Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: name}, u)).To(Succeed())
		raw, found, err := unstructured.NestedMap(u.Object, "spec", "job")
		Expect(err).ToNot(HaveOccurred())
		Expect(found).To(BeTrue())

		tpl := &clv1alpha2.Template{}
		Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: name}, tpl)).To(Succeed())
		Expect(tpl.Spec.Job).ToNot(BeNil())
		return raw, tpl.Spec.Job
	}

	// addLabel updates the whole object through the typed client (as a
	// controller does with Update), changing only a label.
	addLabel := func(name string) {
		tpl := &clv1alpha2.Template{}
		Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: name}, tpl)).To(Succeed())
		tpl.SetLabels(map[string]string{"crownlabs.polito.it/test": "updated"})
		Expect(k8sClient.Update(ctx, tpl)).To(Succeed())
	}

	It("gives editable the default false", func() {
		Expect(create(template("tpl-editable-absent", map[string]interface{}{"environmentName": "job-env"}))).To(Succeed())
		raw, decoded := stored("tpl-editable-absent")
		Expect(raw).To(HaveKeyWithValue("editable", false))
		Expect(raw).ToNot(HaveKey("args"))
		Expect(decoded.Editable).To(BeFalse())
		Expect(decoded.Args).To(BeNil())
	})

	DescribeTable("accepts every combination of editable and args",
		func(name string, job map[string]interface{}) {
			job["environmentName"] = "job-env"
			Expect(create(template(name, job))).To(Succeed())
			raw, _ := stored(name)
			Expect(raw).To(HaveKeyWithValue("editable", job["editable"]))
		},
		Entry("not editable, no args: the args of the container", "tpl-fixed-noargs",
			map[string]interface{}{"editable": false}),
		Entry("not editable, empty args: no args", "tpl-fixed-emptyargs",
			map[string]interface{}{"editable": false, "args": []interface{}{}}),
		Entry("not editable, args: these args, fixed", "tpl-fixed-args",
			map[string]interface{}{"editable": false, "args": []interface{}{"make", "run"}}),
		Entry("editable, no args", "tpl-editable-noargs",
			map[string]interface{}{"editable": true}),
		Entry("editable, args: the default of the dashboard", "tpl-editable-args",
			map[string]interface{}{"editable": true, "args": []interface{}{"make", "run"}}),
	)

	It("sends nil args as null and empty ones as []", func() {
		absent, err := json.Marshal(clv1alpha2.TemplateJob{EnvironmentName: "job-env"})
		Expect(err).ToNot(HaveOccurred())
		Expect(string(absent)).To(ContainSubstring(`"args":null`))

		empty, err := json.Marshal(clv1alpha2.TemplateJob{EnvironmentName: "job-env", Args: []string{}})
		Expect(err).ToNot(HaveOccurred())
		Expect(string(empty)).To(ContainSubstring(`"args":[]`))
	})

	DescribeTable("keeps absent args and empty args different, also after an update through the typed client",
		func(name string, args interface{}, beDecoded OmegaMatcher) {
			job := map[string]interface{}{"environmentName": "job-env", "editable": false}
			if args != nil {
				job["args"] = args
			}
			Expect(create(template(name, job))).To(Succeed())
			for _, step := range []string{"created", "updated"} {
				if step == "updated" {
					addLabel(name)
				}
				raw, decoded := stored(name)
				if args == nil {
					Expect(raw).ToNot(HaveKey("args"), step)
				} else {
					Expect(raw).To(HaveKeyWithValue("args", BeEquivalentTo(args)), step)
				}
				Expect(decoded.Args).To(beDecoded, step)
			}
		},
		Entry("absent: stored as absent, decoded as nil", "tpl-args-absent", nil, BeNil()),
		Entry("empty: stored as [], decoded as an empty slice, not nil", "tpl-args-empty", []interface{}{}, And(Not(BeNil()), BeEmpty())),
		Entry("a list: stored and decoded as that list", "tpl-args-list", []interface{}{"python", "train.py"}, Equal([]string{"python", "train.py"})),
	)

	It("stores resources of any kind, and maxRunTime, the limits shared with the Workspace", func() {
		Expect(create(template("tpl-joblimits", map[string]interface{}{
			"environmentName": "job-env",
			"maxRunTime":      "4h",
			"resources":       map[string]interface{}{"cpu": int64(4), "memory": "8Gi", "nvidia.com/gpu": int64(1)},
		}))).To(Succeed())
		_, decoded := stored("tpl-joblimits")
		Expect(decoded.MaxRunTime).To(Equal("4h"))
		Expect(decoded.Resources).To(HaveKeyWithValue("cpu", resource.MustParse("4")))
		Expect(decoded.Resources).To(HaveKeyWithValue("memory", resource.MustParse("8Gi")))
		Expect(decoded.Resources).To(HaveKeyWithValue("nvidia.com/gpu", resource.MustParse("1")))
	})

	DescribeTable("refuses limits that the Workspace would refuse too",
		func(name string, job map[string]interface{}, message string) {
			job["environmentName"] = "job-env"
			err := create(template(name, job))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(message))
		},
		Entry("a maxRunTime in seconds", "tpl-maxruntime-30s", map[string]interface{}{"maxRunTime": "30s"},
			"spec.job.maxRunTime in body should match '^[1-9][0-9]{0,3}[mhd]$'"),
		Entry("a resource that is not a quantity", "tpl-resource-two", map[string]interface{}{"resources": map[string]interface{}{"nvidia.com/gpu": "two"}},
			"spec.job.resources.nvidia.com/gpu"),
	)
})
