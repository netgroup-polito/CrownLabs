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

var _ = Describe("The command of a CrownJob", func() {
	ctx := context.Background()

	// create creates a CrownJob with the given command through the typed client.
	create := func(name string, command []string) *clv1alpha2.CrownJob {
		cj := &clv1alpha2.CrownJob{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testNamespace},
			Spec:       clv1alpha2.CrownJobSpec{Instance: "instance-1", Command: command},
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

	// decodedCommand returns the command of a CrownJob read through the typed client.
	decodedCommand := func(name string) []string {
		cj := &clv1alpha2.CrownJob{}
		Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: name}, cj)).To(Succeed())
		return cj.Spec.Command
	}

	// addLabel updates the whole object (as a controller does with Update), changing only a label.
	addLabel := func(name string) {
		cj := &clv1alpha2.CrownJob{}
		Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: name}, cj)).To(Succeed())
		cj.SetLabels(map[string]string{"crownlabs.polito.it/test": "updated"})
		Expect(k8sClient.Update(ctx, cj)).To(Succeed())
	}

	It("sends a nil command as null and an empty one as []", func() {
		absent, err := json.Marshal(clv1alpha2.CrownJobSpec{Instance: "instance-1"})
		Expect(err).ToNot(HaveOccurred())
		Expect(string(absent)).To(ContainSubstring(`"command":null`))

		empty, err := json.Marshal(clv1alpha2.CrownJobSpec{Instance: "instance-1", Command: []string{}})
		Expect(err).ToNot(HaveOccurred())
		Expect(string(empty)).To(ContainSubstring(`"command":[]`))
	})

	When("it is absent", func() {
		const name = "cj-command-absent"
		BeforeEach(func() { create(name, nil) })

		It("is stored as absent and decodes to nil", func() {
			Expect(storedSpec(name)).ToNot(HaveKey("command"))
			Expect(decodedCommand(name)).To(BeNil())
		})

		It("stays absent after an update of the metadata", func() {
			addLabel(name)
			Expect(storedSpec(name)).ToNot(HaveKey("command"))
			Expect(decodedCommand(name)).To(BeNil())
		})
	})

	When("it is empty", func() {
		const name = "cj-command-empty"
		BeforeEach(func() { create(name, []string{}) })

		It("is stored as [] and decodes to an empty slice, not nil", func() {
			Expect(storedSpec(name)).To(HaveKeyWithValue("command", BeEquivalentTo([]interface{}{})))
			command := decodedCommand(name)
			Expect(command).ToNot(BeNil())
			Expect(command).To(BeEmpty())
		})

		It("stays [] after an update of the metadata, which the immutable spec accepts", func() {
			addLabel(name)
			Expect(storedSpec(name)).To(HaveKeyWithValue("command", BeEquivalentTo([]interface{}{})))
			Expect(decodedCommand(name)).ToNot(BeNil())
		})
	})

	When("it is a list", func() {
		const name = "cj-command-list"
		BeforeEach(func() { create(name, []string{"python", "train.py"}) })

		It("is stored and decoded as that list", func() {
			Expect(storedSpec(name)).To(HaveKeyWithValue("command", BeEquivalentTo([]interface{}{"python", "train.py"})))
			Expect(decodedCommand(name)).To(Equal([]string{"python", "train.py"}))
		})
	})

	It("cannot be changed once the CrownJob exists", func() {
		const name = "cj-command-immutable"
		create(name, nil)
		cj := &clv1alpha2.CrownJob{}
		Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: name}, cj)).To(Succeed())
		cj.Spec.Command = []string{}
		err := k8sClient.Update(ctx, cj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("CrownJob spec is immutable"))
	})
})
