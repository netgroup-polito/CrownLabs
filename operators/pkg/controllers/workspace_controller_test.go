/*
Ideally, we should have one `<kind>_conroller_test.go` for each controller scaffolded and called in the `test_suite.go`.
So, let's write our example test for the CronJob controller (`cronjob_controller_test.go.`)
*/

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
// +kubebuilder:docs-gen:collapse=Apache License

/*
As usual, we start with the necessary imports. We also define some utility variables.
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

// +kubebuilder:docs-gen:collapse=Imports

/*
The first step to writing a simple integration test is to actually create an instance of CronJob you can run tests against.
Note that to create a CronJob, you’ll need to create a stub CronJob struct that contains your CronJob’s specifications.
Note that when we create a stub CronJob, the CronJob also needs stubs of its required downstream objects.
Without the stubbed Job template spec and the Pod template spec below, the Kubernetes API will not be able to
create the CronJob.
*/
var _ = Describe("Workspace controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		WSName       = "test-workspace"
		WSNamespace  = ""
		WSPrettyName = "Workspace for testing"
		NSName       = "workspace-test-workspace"
		NSNamespace  = ""

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
					Name:      WSName,
					Namespace: WSNamespace,
				},
				Spec: crownlabsv1alpha1.WorkspaceSpec{
					PrettyName: WSPrettyName,
				},
			}
			Expect(k8sClient.Create(ctx, ws)).Should(Succeed())

			By("By checking the workspace has been created")

			wsLookupKey := types.NamespacedName{Name: WSName, Namespace: WSNamespace}
			createdWS := &crownlabsv1alpha1.Workspace{}

			doesEventuallyExists(ctx, wsLookupKey, createdWS, BeTrue(), timeout, interval)

			By("By checking the workspace has the correct name")
			Expect(createdWS.Spec.PrettyName).Should(Equal(WSPrettyName))

			By("By checking the corresponding namespace has been created")

			nsLookupKey := types.NamespacedName{Name: NSName, Namespace: NSNamespace}
			createdNS := &v1.Namespace{}

			doesEventuallyExists(ctx, nsLookupKey, createdNS, BeTrue(), timeout, interval)

			By("By checking the corresponding namespace has a controller reference pointing to the workspace")

			Expect(createdNS.OwnerReferences).Should(ContainElement(MatchFields(IgnoreExtras, Fields{"Name": Equal(WSName)})))
			Expect(createdNS.Labels).Should(HaveKeyWithValue("crownlabs.polito.it/type", "workspace"))

		})
	})

})

func doesEventuallyExists(ctx context.Context, nsLookupKey types.NamespacedName, createdNS runtime.Object, expectedStatus gomegaTypes.GomegaMatcher, timeout time.Duration, interval time.Duration) {
	Eventually(func() bool {
		err := k8sClient.Get(ctx, nsLookupKey, createdNS)
		return err == nil
	}, timeout, interval).Should(expectedStatus)

}
