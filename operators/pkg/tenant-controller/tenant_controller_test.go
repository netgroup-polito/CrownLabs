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
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/tenant-controller/mocks"
)

var _ = Describe("Tenant controller", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	var (
		mockCtrl *gomock.Controller

		wsName        = "ws1"
		wsNamespace   = ""
		wsPrettyName  = "workspace 1"
		wsLabelKey    = "crownlabs.polito.it/workspace-ws1"
		wsUserRole    = "workspace-ws1:user"
		wsManagerRole = "workspace-ws1:manager"

		tnName          = "mariorossi"
		tnFirstName     = "mariò"
		tnLastName      = "ròssì verdò"
		tnWorkspaces    = []crownlabsv1alpha2.TenantWorkspaceEntry{{Name: "ws1", Role: crownlabsv1alpha2.User}}
		tnEmail         = "mario.rossi@email.com"
		userID          = "userID"
		tr              = true
		fa              = false
		userRoleName    = wsUserRole
		testUserRoleID  = "role1"
		testUserRole    = gocloak.Role{ID: &testUserRoleID, Name: &userRoleName}
		beforeUserRoles = []*gocloak.Role{}
		rolesToDelete   = []gocloak.Role{}
		rolesToSet      = []gocloak.Role{{ID: &testUserRoleID, Name: &userRoleName}}
	)

	const (
		tnNamespace = ""
		timeout     = time.Second * 10
		interval    = time.Millisecond * 250
		nsName      = "tenant-mariorossi"

		nfsServerName     = "rook-ceph-nfs-my-nfs-a"
		nfsPath           = "/path"
		rookCephNamespace = "rook-ceph"
	)

	BeforeEach(func() {

		mockCtrl = gomock.NewController(GinkgoT())
		mKcClient = mocks.NewMockGoCloak(mockCtrl)
		kcA.Client = mKcClient

		setupMocksForWorkspaceCreationExistingRoles(mKcClient, kcAccessToken, kcTargetRealm, kcTargetClientID, wsName, wsPrettyName)

		// the user did not exist
		mKcClient.EXPECT().GetUsers(
			gomock.Any(),
			gomock.Eq(kcAccessToken),
			gomock.Eq(kcTargetRealm),
			gomock.Eq(gocloak.GetUsersParams{Username: &tnName}),
		).Return([]*gocloak.User{}, nil).AnyTimes()

		mKcClient.EXPECT().CreateUser(
			gomock.Any(),
			gomock.Eq(kcAccessToken),
			gomock.Eq(kcTargetRealm),
			gomock.Eq(
				gocloak.User{
					Username:      &tnName,
					FirstName:     &tnFirstName,
					LastName:      &tnLastName,
					Email:         &tnEmail,
					Enabled:       &tr,
					EmailVerified: &fa,
				}),
		).Return(userID, nil).AnyTimes()

		mKcClient.EXPECT().ExecuteActionsEmail(
			gomock.Any(),
			gomock.Eq(kcAccessToken),
			gomock.Eq(kcTargetRealm),
			gomock.Eq(gocloak.ExecuteActionsEmail{
				UserID:   &userID,
				Lifespan: &emailActionLifespan,
				Actions:  &reqActions,
			})).Return(nil).AnyTimes()

		mKcClient.EXPECT().GetClientRole(
			gomock.Any(),
			gomock.Eq(kcAccessToken),
			gomock.Eq(kcTargetRealm),
			gomock.Eq(kcTargetClientID),
			gomock.Eq(userRoleName),
		).Return(&testUserRole, nil).AnyTimes()

		mKcClient.EXPECT().GetClientRolesByUserID(
			gomock.Any(),
			gomock.Eq(kcAccessToken),
			gomock.Eq(kcTargetRealm),
			gomock.Eq(kcTargetClientID),
			gomock.Eq(userID),
		).Return(beforeUserRoles, nil).AnyTimes()

		mKcClient.EXPECT().DeleteClientRoleFromUser(
			gomock.Any(),
			gomock.Eq(kcAccessToken),
			gomock.Eq(kcTargetRealm),
			gomock.Eq(kcTargetClientID),
			gomock.Eq(userID),
			gomock.AssignableToTypeOf(rolesToDelete),
		).Return(nil).AnyTimes()

		mKcClient.EXPECT().AddClientRoleToUser(
			gomock.Any(),
			gomock.Eq(kcAccessToken),
			gomock.Eq(kcTargetRealm),
			gomock.Eq(kcTargetClientID),
			gomock.Eq(userID),
			gomock.AssignableToTypeOf(rolesToSet),
		).Return(nil).AnyTimes()

		mKcClient.EXPECT().DeleteClientRole(
			gomock.Any(),
			gomock.Eq(kcAccessToken),
			gomock.Eq(kcTargetRealm),
			gomock.Eq(kcTargetClientID),
			gomock.Eq(wsUserRole),
		).Return(nil).AnyTimes()

		mKcClient.EXPECT().DeleteClientRole(
			gomock.Any(),
			gomock.Eq(kcAccessToken),
			gomock.Eq(kcTargetRealm),
			gomock.Eq(kcTargetClientID),
			gomock.Eq(wsManagerRole),
		).Return(nil).AnyTimes()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("Should create the related resources when creating a tenant", func() {
		SampleResourceQuota := crownlabsv1alpha1.WorkspaceResourceQuota{
			CPU:       *resource.NewQuantity(15, resource.DecimalSI),
			Memory:    *resource.NewQuantity(25*1024*1024*1024, resource.BinarySI),
			Instances: 5,
		}

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
				Quota:      SampleResourceQuota,
			},
		}
		Expect(k8sClient.Create(ctx, ws)).Should(Succeed())

		By("By creating a tenant")
		tn := &crownlabsv1alpha2.Tenant{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "crownlabs.polito.it/v1alpha1",
				Kind:       "Tenant",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      tnName,
				Namespace: tnNamespace,
				Labels:    map[string]string{targetLabelKey: targetLabelValue},
			},
			Spec: crownlabsv1alpha2.TenantSpec{
				FirstName:  tnFirstName,
				LastName:   tnLastName,
				Email:      tnEmail,
				LastLogin:  metav1.Now(),
				Workspaces: tnWorkspaces,
			},
		}
		Expect(k8sClient.Create(ctx, tn)).Should(Succeed())

		By("By checking that the tenant has been created")
		tnLookupKey := types.NamespacedName{Name: tnName, Namespace: tnNamespace}
		createdTn := &crownlabsv1alpha2.Tenant{}
		doesEventuallyExists(ctx, tnLookupKey, createdTn, BeTrue(), timeout, interval)

		By("By checking that the needed cluster resources for the tenant have been updated accordingly")
		checkTnClusterResourceCreation(ctx, tnName, nsName, timeout, interval, nfsServerName, rookCephNamespace, nfsPath)

		By("By checking that the tenant has been updated accordingly after creation")

		Eventually(func() bool {
			err := k8sClient.Get(ctx, tnLookupKey, tn)
			if err != nil {
				return false
			}
			// check if namespace status has been correctly updated
			if !tn.Status.PersonalNamespace.Created || tn.Status.PersonalNamespace.Name != nsName {
				return false
			}
			// check if external subscriptions has been correctly updated
			if tn.Status.Subscriptions["keycloak"] != crownlabsv1alpha2.SubscrOk {
				return false
			}
			// check if workspace inconsistence has been correctly updated
			if len(tn.Status.FailingWorkspaces) != 0 {
				return false
			}
			// check if labels have been correctly updated
			if tn.Labels[wsLabelKey] != string(crownlabsv1alpha2.User) {
				return false
			}
			if tn.Labels["crownlabs.polito.it/first-name"] != "mari" {
				return false
			}
			if tn.Labels["crownlabs.polito.it/last-name"] != "rss_verd" {
				return false
			}
			if !containsString(tn.Finalizers, "crownlabs.polito.it/tenant-operator") {
				return false
			}
			return true
		}, timeout, interval).Should(BeTrue())

		By("By deleting the workspace of the tenant")
		Expect(k8sClient.Delete(ctx, ws)).Should(Succeed())

		By("By checking that the tenant has been updated accordingly after workspace deletion")
		Eventually(func() bool {
			err := k8sClient.Get(ctx, tnLookupKey, tn)
			if err != nil {
				return false
			}
			// check if labels have been correctly updated
			if _, ok := tn.Labels[wsLabelKey]; ok {
				return false
			}
			return true
		}, timeout, interval).Should(BeTrue())
	})

})

func checkTnPVBound(ctx context.Context, tnName string, timeout, interval time.Duration, serverName, rookCephNamespace, path string) {
	var pvc v1.PersistentVolumeClaim
	doesEventuallyExists(ctx, types.NamespacedName{Name: myDrivePVCName(tnName), Namespace: testMyDrivePVCsNamespace}, &pvc, BeTrue(), timeout, interval)

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pv-test",
		},
		Spec: v1.PersistentVolumeSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteMany,
			},
			Capacity: v1.ResourceList{
				v1.ResourceStorage: *resource.NewQuantity(1*1024*1024*1024, resource.BinarySI),
			},
			ClaimRef: &v1.ObjectReference{
				APIVersion:      "v1",
				Kind:            "PersistentVolumeClaim",
				Name:            pvc.Name,
				Namespace:       pvc.Namespace,
				ResourceVersion: pvc.ResourceVersion,
				UID:             pvc.UID,
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				CSI: &v1.CSIPersistentVolumeSource{
					Driver:       "rook-ceph.nfs.csi.ceph.com",
					VolumeHandle: "path",
					VolumeAttributes: map[string]string{
						"server":    serverName,
						"share":     path,
						"clusterID": rookCephNamespace,
					},
				},
			},
		},
	}
	Expect(k8sClient.Create(ctx, pv)).Should(Succeed())
	pvc.Spec.VolumeName = "pv-test"
	Expect(k8sClient.Update(ctx, &pvc)).Should(Succeed())
	pvc.Status.Phase = v1.ClaimBound
	Expect(k8sClient.Status().Update(ctx, &pvc)).Should(Succeed())
}

func checkTnClusterResourceCreation(ctx context.Context, tnName, nsName string, timeout, interval time.Duration, serverName, rookCephNamespace, path string) {
	By("By checking that the corresponding namespace has been created")

	nsLookupKey := types.NamespacedName{Name: nsName, Namespace: ""}
	createdNs := &v1.Namespace{}

	doesEventuallyExists(ctx, nsLookupKey, createdNs, BeTrue(), timeout, interval)

	By("By checking that the corresponding namespace has a controller reference pointing to the tenant")

	Expect(createdNs.OwnerReferences).Should(ContainElement(MatchFields(IgnoreExtras, Fields{"Name": Equal(tnName)})))
	Expect(createdNs.Labels).Should(HaveKeyWithValue("crownlabs.polito.it/type", "tenant"))
	Expect(createdNs.Labels).Should(HaveKeyWithValue("crownlabs.polito.it/instance-resources-replication", "true"))

	By("By checking that the resource quota of the tenant has been created")
	rqLookupKey := types.NamespacedName{Name: "crownlabs-resource-quota", Namespace: nsName}
	createdRq := &v1.ResourceQuota{}
	doesEventuallyExists(ctx, rqLookupKey, createdRq, BeTrue(), timeout, interval)

	By("By checking that the resource quota has a label pointing to the tenant operator")
	Expect(createdRq.Labels).Should(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))

	By("By checking that the resource quota has a correct spec")
	limitCPU, _ := resource.ParseQuantity("15")
	limitMem, _ := resource.ParseQuantity("25Gi")
	reqCPU, _ := resource.ParseQuantity("15")
	reqMem, _ := resource.ParseQuantity("25Gi")
	instanceCount, _ := resource.ParseQuantity("5")
	Expect(createdRq.Spec.Hard["limits.cpu"]).Should(Equal(limitCPU))
	Expect(createdRq.Spec.Hard["limits.memory"]).Should(Equal(limitMem))
	Expect(createdRq.Spec.Hard["requests.cpu"]).Should(Equal(reqCPU))
	Expect(createdRq.Spec.Hard["requests.memory"]).Should(Equal(reqMem))
	Expect(createdRq.Spec.Hard["count/instances.crownlabs.polito.it"]).Should(Equal(instanceCount))

	By("By checking that the role binding of the tenant has been created")
	rbLookupKey := types.NamespacedName{Name: "crownlabs-manage-instances", Namespace: nsName}
	createdRb := &rbacv1.RoleBinding{}
	doesEventuallyExists(ctx, rbLookupKey, createdRb, BeTrue(), timeout, interval)

	By("By checking that the role binding of the tenant has a label pointing to the tenant operator")
	Expect(createdRb.Labels).Should(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))

	By("By checking that the role binding has a correct spec")
	Expect(createdRb.RoleRef.Name).Should(Equal("crownlabs-manage-instances"))
	Expect(createdRb.RoleRef.Kind).Should(Equal("ClusterRole"))
	Expect(createdRb.Subjects).Should(HaveLen(1))
	Expect(createdRb.Subjects[0]).Should(MatchFields(IgnoreExtras, Fields{"Name": Equal(tnName), "Kind": Equal("User")}))

	By("By checking that the cluster role of the tenant has been created")
	crName := fmt.Sprintf("crownlabs-manage-%s", nsName)
	crLookupKey := types.NamespacedName{Name: crName}
	createdCr := &rbacv1.ClusterRole{}
	doesEventuallyExists(ctx, crLookupKey, createdCr, BeTrue(), timeout, interval)

	By("By checking that the cluster role of the tenant has a controller reference pointing to the tenant")
	Expect(createdCr.OwnerReferences).Should(ContainElement(MatchFields(IgnoreExtras, Fields{"Name": Equal(tnName)})))
	Expect(createdCr.Labels).Should(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))

	By("By checking that the cluster role has a correct spec")
	Expect(createdCr.Rules).Should(HaveLen(1))
	Expect(createdCr.Rules[0].APIGroups).Should(ContainElement(Equal("crownlabs.polito.it")))
	Expect(createdCr.Rules[0].Resources).Should(ContainElement(Equal("tenants")))
	Expect(createdCr.Rules[0].ResourceNames).Should(ContainElement(Equal(tnName)))
	Expect(createdCr.Rules[0].Verbs).Should(Equal([]string{"get", "list", "watch", "patch", "update"}))

	By("By checking that the cluster role binding of the tenant has been created")
	crbName := fmt.Sprintf("crownlabs-manage-%s", nsName)
	crbLookupKey := types.NamespacedName{Name: crbName}
	createdCrb := &rbacv1.ClusterRoleBinding{}
	doesEventuallyExists(ctx, crbLookupKey, createdCrb, BeTrue(), timeout, interval)

	By("By checking that the cluster role binding of the tenant has a controller reference pointing to the tenant")
	Expect(createdCrb.OwnerReferences).Should(ContainElement(MatchFields(IgnoreExtras, Fields{"Name": Equal(tnName)})))
	Expect(createdCrb.Labels).Should(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))

	By("By checking that the cluster role binding has a correct spec")
	Expect(createdCrb.RoleRef.Name).Should(Equal(crName))
	Expect(createdCrb.RoleRef.Kind).Should(Equal("ClusterRole"))
	Expect(createdCrb.Subjects).Should(HaveLen(1))
	Expect(createdCrb.Subjects[0]).Should(MatchFields(IgnoreExtras, Fields{"Name": Equal(tnName), "Kind": Equal("User")}))

	By("By checking that the deny network policy of the tenant has been created")
	netPolDenyLookupKey := types.NamespacedName{Name: "crownlabs-deny-ingress-traffic", Namespace: nsName}
	createdNetPolDeny := &netv1.NetworkPolicy{}
	doesEventuallyExists(ctx, netPolDenyLookupKey, createdNetPolDeny, BeTrue(), timeout, interval)

	By("By checking that the deny network policy of the tenant has a label pointing to the tenant operator")
	Expect(createdNetPolDeny.Labels).Should(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))

	By("By checking that the deny network policy has a correct spec")
	Expect(createdNetPolDeny.Spec.PodSelector.MatchLabels).Should(HaveLen(0))
	Expect(createdNetPolDeny.Spec.Ingress).Should(HaveLen(1))
	Expect(createdNetPolDeny.Spec.Ingress[0].From).Should(HaveLen(1))
	Expect(createdNetPolDeny.Spec.Ingress[0].From[0].PodSelector.MatchLabels).Should(HaveLen(0))

	By("By checking that the allow network policy of the tenant has been created")
	netPolAllowLookupKey := types.NamespacedName{Name: "crownlabs-allow-trusted-ingress-traffic", Namespace: nsName}
	createdNetPolAllow := &netv1.NetworkPolicy{}
	doesEventuallyExists(ctx, netPolAllowLookupKey, createdNetPolAllow, BeTrue(), timeout, interval)

	By("By checking that the allow network policy of the tenant has a label pointing to the tenant operator")
	Expect(createdNetPolAllow.Labels).Should(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))

	By("By checking that the allow network policy has a correct spec")
	Expect(createdNetPolAllow.Spec.PodSelector.MatchLabels).Should(HaveLen(0))
	Expect(createdNetPolAllow.Spec.Ingress).Should(HaveLen(1))
	Expect(createdNetPolAllow.Spec.Ingress[0].From).Should(HaveLen(1))
	Expect(createdNetPolAllow.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels["crownlabs.polito.it/allow-instance-access"]).Should(Equal("true"))

	By("By checking that the mydrive-info secret of the tenant has been created")
	checkTnPVBound(ctx, tnName, timeout, interval, serverName, rookCephNamespace, path)

	By("By checking that the mydrive-info secret of the tenant has been created")
	NFSSecretLookupKey := types.NamespacedName{Name: NFSSecretName, Namespace: nsName}
	createdNFSSecret := &v1.Secret{}
	doesEventuallyExists(ctx, NFSSecretLookupKey, createdNFSSecret, BeTrue(), timeout, interval)

	By("By checking that the NFS pvc secret of the tenant has a owner reference and label pointing to the tenant operator")
	Expect(createdNFSSecret.OwnerReferences).Should(ContainElement(MatchFields(IgnoreExtras, Fields{"Name": Equal(tnName)})))
	Expect(createdNFSSecret.Labels).Should(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))

	By("By checking that the NFS pvc secret of the tenant has a correct spec")
	Expect(createdNFSSecret.Type).Should(Equal(v1.SecretTypeOpaque))
	Expect(createdNFSSecret.Data).Should(HaveKeyWithValue(NFSSecretPathKey, []byte(path)))
	Expect(createdNFSSecret.Data).Should(HaveKeyWithValue(NFSSecretServerNameKey, []byte(serverName+"."+rookCephNamespace)))
}
