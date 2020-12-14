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
package tenant_controller

import (
	"context"
	"fmt"
	"time"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Tenant controller", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		tnNamespace = ""
		tnFirstName = "mario"
		tnLastName  = "rossi"
		tnEmail     = "mario.rossi@email.com"
		nsNamespace = ""
		timeout     = time.Second * 10
		interval    = time.Millisecond * 250
	)
	var (
		// make tenant name time-sensitive to make test more independent since using external resources like a real keycloak instance
		tnName = fmt.Sprintf("test-%d", time.Now().Unix())
		nsName = fmt.Sprintf("tenant-%s", tnName)
	)

	It("Should create the related resources when creating a tenant", func() {
		By("By creating a tenant")
		ctx := context.Background()
		tn := &crownlabsv1alpha1.Tenant{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "crownlabs.polito.it/v1alpha1",
				Kind:       "Tenant",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      tnName,
				Namespace: tnNamespace,
			},
			Spec: crownlabsv1alpha1.TenantSpec{
				FirstName: tnFirstName,
				LastName:  tnLastName,
				Email:     tnEmail,
			},
		}
		Expect(k8sClient.Create(ctx, tn)).Should(Succeed())

		By("By checking that the tenant has been created")
		tnLookupKey := types.NamespacedName{Name: tnName, Namespace: tnNamespace}
		createdTn := &crownlabsv1alpha1.Tenant{}

		doesEventuallyExists(ctx, tnLookupKey, createdTn, BeTrue(), timeout, interval)

		// By("By checking that the tenant has the correct name")
		// Expect(createdWs.Spec.PrettyName).Should(Equal(wsPrettyName))

		By("By checking that the corresponding namespace has been created")

		nsLookupKey := types.NamespacedName{Name: nsName, Namespace: nsNamespace}
		createdNs := &v1.Namespace{}

		doesEventuallyExists(ctx, nsLookupKey, createdNs, BeTrue(), timeout, interval)

		By("By checking that the corresponding namespace has a controller reference pointing to the tenant")

		Expect(createdNs.OwnerReferences).Should(ContainElement(MatchFields(IgnoreExtras, Fields{"Name": Equal(tnName)})))
		Expect(createdNs.Labels).Should(HaveKeyWithValue("crownlabs.polito.it/type", "tenant"))

		By("By checking that the status of the tenant has been updated accordingly")

		Eventually(func() bool {
			err := k8sClient.Get(ctx, tnLookupKey, tn)
			if err != nil {
				return false
			}
			if !tn.Status.PersonalNamespace.Created || tn.Status.PersonalNamespace.Name != nsName {
				return false
			}
			return true
		}, timeout, interval).Should(BeTrue())

	})
})
