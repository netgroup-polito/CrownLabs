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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var _ = Describe("The instance and the prettyName of a CrownJob", func() {
	ctx := context.Background()

	It("stores the name of the instance and the prettyName", func() {
		cj := &clv1alpha2.CrownJob{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "instance-1-", Namespace: testNamespace},
			Spec:       clv1alpha2.CrownJobSpec{Instance: "instance-1", PrettyName: "Training, 3 epochs"},
		}
		Expect(k8sClient.Create(ctx, cj)).To(Succeed())
		DeferCleanup(func() { Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, cj))).To(Succeed()) })
		Expect(cj.Name).To(HavePrefix("instance-1-"))

		read := &clv1alpha2.CrownJob{}
		Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(cj), read)).To(Succeed())
		Expect(read.Spec.Instance).To(Equal("instance-1"))
		Expect(read.Spec.PrettyName).To(Equal("Training, 3 epochs"))
	})

	It("refuses an empty instance", func() {
		cj := &clv1alpha2.CrownJob{
			ObjectMeta: metav1.ObjectMeta{Name: "cj-instance-empty", Namespace: testNamespace},
			Spec:       clv1alpha2.CrownJobSpec{Instance: ""},
		}
		err := k8sClient.Create(ctx, cj)
		if err == nil {
			DeferCleanup(func() { Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, cj))).To(Succeed()) })
		}
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.instance in body should be at least 1 chars long"))
	})
})

var _ = Describe("The args of a CrownJob", func() {
	ctx := context.Background()

	// create creates a CrownJob with the given args through the typed client.
	create := func(name string, args []string) *clv1alpha2.CrownJob {
		cj := &clv1alpha2.CrownJob{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testNamespace},
			Spec:       clv1alpha2.CrownJobSpec{Instance: "instance-1", Args: args},
		}
		Expect(k8sClient.Create(ctx, cj)).To(Succeed())
		DeferCleanup(func() { Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, cj))).To(Succeed()) })
		return cj
	}

	// storedSpec returns the spec of a CrownJob as the API server returns it, as raw JSON fields.
	storedSpec := func(name string) map[string]interface{} {
		raw := &unstructured.Unstructured{}
		raw.SetGroupVersionKind(clv1alpha2.GroupVersion.WithKind("CrownJob"))
		Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: name}, raw)).To(Succeed())
		spec, found, err := unstructured.NestedMap(raw.Object, "spec")
		Expect(err).ToNot(HaveOccurred())
		Expect(found).To(BeTrue())
		return spec
	}

	// decodedArgs returns the args of a CrownJob read through the typed client.
	decodedArgs := func(name string) []string {
		cj := &clv1alpha2.CrownJob{}
		Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: name}, cj)).To(Succeed())
		return cj.Spec.Args
	}

	// addLabel updates the whole object (as a controller does with Update), changing only a label.
	addLabel := func(name string) {
		cj := &clv1alpha2.CrownJob{}
		Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: name}, cj)).To(Succeed())
		cj.SetLabels(map[string]string{"crownlabs.polito.it/test": "updated"})
		Expect(k8sClient.Update(ctx, cj)).To(Succeed())
	}

	It("sends nil args as null and empty ones as []", func() {
		absent, err := json.Marshal(clv1alpha2.CrownJobSpec{Instance: "instance-1"})
		Expect(err).ToNot(HaveOccurred())
		Expect(string(absent)).To(ContainSubstring(`"args":null`))

		empty, err := json.Marshal(clv1alpha2.CrownJobSpec{Instance: "instance-1", Args: []string{}})
		Expect(err).ToNot(HaveOccurred())
		Expect(string(empty)).To(ContainSubstring(`"args":[]`))
	})

	When("they are absent", func() {
		const name = "cj-args-absent"
		BeforeEach(func() { create(name, nil) })

		It("are stored as absent and decode to nil", func() {
			Expect(storedSpec(name)).ToNot(HaveKey("args"))
			Expect(decodedArgs(name)).To(BeNil())
		})

		It("stay absent after an update of the metadata", func() {
			addLabel(name)
			Expect(storedSpec(name)).ToNot(HaveKey("args"))
			Expect(decodedArgs(name)).To(BeNil())
		})
	})

	When("they are empty", func() {
		const name = "cj-args-empty"
		BeforeEach(func() { create(name, []string{}) })

		It("are stored as [] and decode to an empty slice, not nil", func() {
			Expect(storedSpec(name)).To(HaveKeyWithValue("args", BeEquivalentTo([]interface{}{})))
			args := decodedArgs(name)
			Expect(args).ToNot(BeNil())
			Expect(args).To(BeEmpty())
		})

		It("stay [] after an update of the metadata, which the immutable spec accepts", func() {
			addLabel(name)
			Expect(storedSpec(name)).To(HaveKeyWithValue("args", BeEquivalentTo([]interface{}{})))
			Expect(decodedArgs(name)).ToNot(BeNil())
		})
	})

	When("they are a list", func() {
		const name = "cj-args-list"
		BeforeEach(func() { create(name, []string{"python", "train.py"}) })

		It("are stored and decoded as that list", func() {
			Expect(storedSpec(name)).To(HaveKeyWithValue("args", BeEquivalentTo([]interface{}{"python", "train.py"})))
			Expect(decodedArgs(name)).To(Equal([]string{"python", "train.py"}))
		})
	})

	It("cannot be changed once the CrownJob exists", func() {
		const name = "cj-args-immutable"
		create(name, nil)
		cj := &clv1alpha2.CrownJob{}
		Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: name}, cj)).To(Succeed())
		cj.Spec.Args = []string{}
		err := k8sClient.Update(ctx, cj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("CrownJob spec is immutable"))
	})
})
