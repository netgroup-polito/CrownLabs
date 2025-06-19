// Copyright 2020-2025 Politecnico di Torino
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

package tenant_controller

import (
	"context"
	"fmt"
	"time"

	gocloak "github.com/Nerzal/gocloak/v7"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/tenant-controller/mocks"
)

var _ = Describe("Workspace controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)
	var (
		mockCtrl *gomock.Controller

		wsNamespace  = ""
		wsPrettyName = "Workspace for testing"
		// make workspace name time-sensitive to make test more independent since using external resources like a real keycloak instance
		wsName = fmt.Sprintf("test-%d", time.Now().Unix())
		nsName = fmt.Sprintf("workspace-%s", wsName)
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mKcClient = mocks.NewMockGoCloak(mockCtrl)
		kcA.Client = mKcClient

		setupMocksForWorkspaceCreationExistingRoles(mKcClient, kcAccessToken, kcA.TargetRealm, kcTargetClientID, wsName, wsPrettyName)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("Should create the related resources when creating a workspace", func() {
		By("By creating a workspace")
		ws := &crownlabsv1alpha1.Workspace{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "crownlabs.polito.it/v1alpha1",
				Kind:       "Workspace",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      wsName,
				Namespace: wsNamespace,
				Labels:    map[string]string{targetLabelKey: targetLabelValue},
			},
			Spec: crownlabsv1alpha1.WorkspaceSpec{
				PrettyName: wsPrettyName,
				Quota: crownlabsv1alpha1.WorkspaceResourceQuota{
					Instances: 1,
				},
			},
		}
		Expect(k8sClient.Create(ctx, ws)).Should(Succeed())

		By("By checking that the workspace has been created")
		wsLookupKey := types.NamespacedName{Name: wsName, Namespace: wsNamespace}
		createdWs := &crownlabsv1alpha1.Workspace{}

		doesEventuallyExists(ctx, wsLookupKey, createdWs, BeTrue(), timeout, interval)

		By("By checking that the workspace has the correct name")
		Expect(createdWs.Spec.PrettyName).Should(Equal(wsPrettyName))

		By("By checking that the needed cluster resources for the workspace have been updated accordingly")
		checkWsClusterResourceCreation(ctx, wsName, nsName, timeout, interval)

		By("By checking that the status of the workspace has been updated accordingly")

		Eventually(func() bool {
			err := k8sClient.Get(ctx, wsLookupKey, ws)
			if err != nil {
				return false
			}
			if !ws.Status.Namespace.Created || ws.Status.Namespace.Name != nsName {
				return false
			}
			if ws.Status.Subscriptions["keycloak"] != crownlabsv1alpha2.SubscrOk {
				return false
			}
			if !containsString(ws.Finalizers, "crownlabs.polito.it/tenant-operator") {
				return false
			}
			return true
		}, timeout, interval).Should(BeTrue())
	})
})

func setupMocksForWorkspaceCreationExistingRoles(mockKcClient *mocks.MockGoCloak, kcAccessToken, kcTargetRealm, kcTargetClientID, wsName, wsPrettyName string) {
	userKcRole := fmt.Sprintf("workspace-%s:user", wsName)
	managerKcRole := fmt.Sprintf("workspace-%s:manager", wsName)
	mockKcClient.EXPECT().GetClientRole(
		gomock.Any(),
		gomock.Eq(kcAccessToken),
		gomock.Eq(kcTargetRealm),
		gomock.Eq(kcTargetClientID),
		gomock.Eq(userKcRole),
	).Return(&gocloak.Role{Name: &userKcRole, Description: &wsPrettyName}, nil).AnyTimes()

	mockKcClient.EXPECT().GetClientRole(
		gomock.Any(),
		gomock.Eq(kcAccessToken),
		gomock.Eq(kcTargetRealm),
		gomock.Eq(kcTargetClientID),
		gomock.Eq(managerKcRole),
	).Return(&gocloak.Role{Name: &managerKcRole, Description: &wsPrettyName}, nil).AnyTimes()

	mockKcClient.EXPECT().UpdateRole(
		gomock.Any(),
		gomock.Eq(kcAccessToken),
		gomock.Eq(kcTargetRealm),
		gomock.Eq(kcTargetClientID),
		gomock.AssignableToTypeOf(gocloak.Role{Name: &userKcRole, Description: &wsPrettyName}),
	).Return(nil).AnyTimes()

	mockKcClient.EXPECT().UpdateRole(
		gomock.Any(),
		gomock.Eq(kcAccessToken),
		gomock.Eq(kcTargetRealm),
		gomock.Eq(kcTargetClientID),
		gomock.AssignableToTypeOf(gocloak.Role{Name: &managerKcRole, Description: &wsPrettyName}),
	).Return(nil).AnyTimes()
}

func checkWsClusterResourceCreation(ctx context.Context, wsName, nsName string, timeout, interval time.Duration) {
	By("By checking that the corresponding namespace has been created")

	nsLookupKey := types.NamespacedName{Name: nsName, Namespace: ""}
	createdNs := &v1.Namespace{}

	doesEventuallyExists(ctx, nsLookupKey, createdNs, BeTrue(), timeout, interval)

	By("By checking that the corresponding namespace has a controller reference pointing to the workspace")

	Expect(createdNs.OwnerReferences).Should(ContainElement(MatchFields(IgnoreExtras, Fields{"Name": Equal(wsName)})))
	Expect(createdNs.Labels).Should(HaveKeyWithValue("crownlabs.polito.it/type", "workspace"))

	By("By checking that the cluster role binding of the workspace has been created")
	crbName := fmt.Sprintf("crownlabs-manage-instances-%s", wsName)
	crbLookupKey := types.NamespacedName{Name: crbName}
	createdCrb := &rbacv1.ClusterRoleBinding{}
	doesEventuallyExists(ctx, crbLookupKey, createdCrb, BeTrue(), timeout, interval)

	By("By checking that the cluster role binding of the workspace has a controller reference pointing to the workspace")
	Expect(createdCrb.OwnerReferences).Should(ContainElement(MatchFields(IgnoreExtras, Fields{"Name": Equal(wsName)})))
	Expect(createdCrb.Labels).Should(HaveKeyWithValue("crownlabs.polito.it/managed-by", "workspace"))

	By("By checking that the cluster role binding has a correct spec")
	crGroupName := fmt.Sprintf("kubernetes:workspace-%s:manager", wsName)
	Expect(createdCrb.RoleRef.Name).Should(Equal("crownlabs-manage-instances"))
	Expect(createdCrb.RoleRef.Kind).Should(Equal("ClusterRole"))
	Expect(createdCrb.Subjects).Should(HaveLen(1))
	Expect(createdCrb.Subjects[0]).Should(MatchFields(IgnoreExtras, Fields{"Name": Equal(crGroupName), "Kind": Equal("Group")}))

	By("By checking that the role binding of the workspace has been created")
	rbLookupKey := types.NamespacedName{Name: "crownlabs-view-templates", Namespace: nsName}
	createdRb := &rbacv1.RoleBinding{}
	doesEventuallyExists(ctx, rbLookupKey, createdRb, BeTrue(), timeout, interval)

	By("By checking that the role binding of the workspace has a label pointing to the workspace")
	Expect(createdRb.Labels).Should(HaveKeyWithValue("crownlabs.polito.it/managed-by", "workspace"))

	By("By checking that the role binding has a correct spec")
	rGroupName := fmt.Sprintf("kubernetes:workspace-%s:user", wsName)
	Expect(createdRb.RoleRef.Name).Should(Equal("crownlabs-view-templates"))
	Expect(createdRb.RoleRef.Kind).Should(Equal("ClusterRole"))
	Expect(createdRb.Subjects).Should(HaveLen(1))
	Expect(createdRb.Subjects[0]).Should(MatchFields(IgnoreExtras, Fields{"Name": Equal(rGroupName), "Kind": Equal("Group")}))

	By("By checking that the manager role binding of the workspace has been created")
	mngRbLookupKey := types.NamespacedName{Name: "crownlabs-manage-templates", Namespace: nsName}
	createdMngRb := &rbacv1.RoleBinding{}
	doesEventuallyExists(ctx, mngRbLookupKey, createdMngRb, BeTrue(), timeout, interval)

	By("By checking that the manager role binding of the workspace has a label pointing to the workspace")
	Expect(createdMngRb.Labels).Should(HaveKeyWithValue("crownlabs.polito.it/managed-by", "workspace"))

	By("By checking that the role binding has a correct spec")
	mngRGroupName := fmt.Sprintf("kubernetes:workspace-%s:manager", wsName)
	Expect(createdMngRb.RoleRef.Name).Should(Equal("crownlabs-manage-templates"))
	Expect(createdMngRb.RoleRef.Kind).Should(Equal("ClusterRole"))
	Expect(createdMngRb.Subjects).Should(HaveLen(1))
	Expect(createdMngRb.Subjects[0]).Should(MatchFields(IgnoreExtras, Fields{"Name": Equal(mngRGroupName), "Kind": Equal("Group")}))
}
