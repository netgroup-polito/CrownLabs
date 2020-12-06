/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controllers

import (
	"context"
	"time"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	gomegaTypes "github.com/onsi/gomega/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Workspace controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		wsName       = "test-ws"
		wsNamespace  = ""
		wsPrettyName = "Workspace for testing"
		nsName       = "workspace-test-ws"
		nsNamespace  = ""

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("Workspace controller", func() {
		It("Should create the related namespace when creating a workspace", func() {
			By("By creating a workspace")
			ctx := context.Background()
			ws := &crownlabsv1alpha1.Workspace{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "crownlabs.polito.it/v1alpha1",
					Kind:       "Workspace",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      wsName,
					Namespace: wsNamespace,
				},
				Spec: crownlabsv1alpha1.WorkspaceSpec{
					PrettyName: wsPrettyName,
				},
			}
			Expect(k8sClient.Create(ctx, ws)).Should(Succeed())

			By("By checking that the workspace has been created")

			wsLookupKey := types.NamespacedName{Name: wsName, Namespace: wsNamespace}
			createdWs := &crownlabsv1alpha1.Workspace{}

			doesEventuallyExists(ctx, wsLookupKey, createdWs, BeTrue(), timeout, interval)

			By("By checking that the workspace has the correct name")
			Expect(createdWs.Spec.PrettyName).Should(Equal(wsPrettyName))

			By("By checking that the corresponding namespace has been created")

			nsLookupKey := types.NamespacedName{Name: nsName, Namespace: nsNamespace}
			createdNs := &v1.Namespace{}

			doesEventuallyExists(ctx, nsLookupKey, createdNs, BeTrue(), timeout, interval)

			By("By checking that the corresponding namespace has a controller reference pointing to the workspace")

			Expect(createdNs.OwnerReferences).Should(ContainElement(MatchFields(IgnoreExtras, Fields{"Name": Equal(wsName)})))
			Expect(createdNs.Labels).Should(HaveKeyWithValue("crownlabs.polito.it/type", "workspace"))

			By("By checking that the status of the workspace has been updated accordingly")

			Eventually(func() bool {
				err := k8sClient.Get(ctx, wsLookupKey, ws)
				if err != nil {
					return false
				}
				if !ws.Status.Namespace.Created || ws.Status.Namespace.Name != nsName {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

		})
	})
})

func doesEventuallyExists(ctx context.Context, objLookupKey types.NamespacedName, targetObj runtime.Object, expectedStatus gomegaTypes.GomegaMatcher, timeout time.Duration, interval time.Duration) {
	Eventually(func() bool {
		err := k8sClient.Get(ctx, objLookupKey, targetObj)
		return err == nil
	}, timeout, interval).Should(expectedStatus)
}
