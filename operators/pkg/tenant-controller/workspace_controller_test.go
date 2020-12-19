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

	gocloak "github.com/Nerzal/gocloak/v7"
	"github.com/golang/mock/gomock"
	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/tenant-controller/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Workspace controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		wsNamespace  = ""
		wsPrettyName = "Workspace for testing"
		nsNamespace  = ""
		timeout      = time.Second * 10
		interval     = time.Millisecond * 250
	)
	var (
		// make workspace name time-sensitive to make test more independent since using external resources like a real keycloak instance
		wsName = fmt.Sprintf("test-%d", time.Now().Unix())
		nsName = fmt.Sprintf("workspace-%s", wsName)
	)

	mockCtrl := gomock.NewController(GinkgoT())
	BeforeEach(
		func() {
			mKcClient = nil
			mKcClient = mocks.NewMockGoCloak(mockCtrl)
			kcA.Client = mKcClient
			setupMocksForWorkspaceCreation(mKcClient, kcAccessToken, kcA.TargetRealm, kcTargetClientID, wsName)
		})

	It("Should create the related resources when creating a workspace", func() {
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

		By("By checking that the corresponding keycloak roles have been created")

		Eventually(func() bool {
			err := k8sClient.Get(ctx, wsLookupKey, ws)
			if err != nil {
				return false
			}
			if ws.Status.Subscriptions["keycloak"] != crownlabsv1alpha1.SubscrOk {
				return false
			}

			return true
		}, timeout, interval).Should(BeTrue())
		By("By checking that the keycloak deleteRole methods get called when a workspace is deleted")

	})

})

func setupMocksForWorkspaceCreation(mockKcCLient *mocks.MockGoCloak, kcAccessToken string, kcTargetRealm string, kcTargetClientID string, wsName string) {
	userKcRole := fmt.Sprintf("workspace-%s:user", wsName)
	adminKcRole := fmt.Sprintf("workspace-%s:admin", wsName)
	mockKcCLient.EXPECT().GetClientRole(
		gomock.AssignableToTypeOf(context.Background()),
		gomock.Eq(kcAccessToken),
		gomock.Eq(kcTargetRealm),
		gomock.Eq(kcTargetClientID),
		gomock.Eq(userKcRole),
	).Return(&gocloak.Role{Name: &userKcRole}, nil).MinTimes(1).MaxTimes(2)

	mockKcCLient.EXPECT().GetClientRole(
		gomock.AssignableToTypeOf(context.Background()),
		gomock.Eq(kcAccessToken),
		gomock.Eq(kcTargetRealm),
		gomock.Eq(kcTargetClientID),
		gomock.Eq(adminKcRole),
	).Return(&gocloak.Role{Name: &adminKcRole}, nil).MinTimes(1).MaxTimes(2)
}
