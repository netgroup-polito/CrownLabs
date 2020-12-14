package bastion_controller

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	crownlabsalpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
)

var _ = Describe("Bastion controller", func() {

	const (
		NameTenant1          = "s11111"
		NamespaceTenant1     = ""
		FirstNameTenant1     = "Mario"
		LastNameTenant1      = "Rossi"
		EmailTenant1         = "mario.rossi@fakemail.com"
		CreateSandboxTenant1 = true

		NameTenant2          = "s22222"
		NamespaceTenant2     = ""
		FirstNameTenant2     = "Fabio"
		LastNameTenant2      = "Bianchi"
		EmailTenant2         = "fabio.bianchi@fakemail.com"
		CreateSandboxTenant2 = true

		testFile = "./authorized_keys_test"
		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	PublicKeysTenant1 := []string{
		"ssh-ed25519 publicKeyString_1 comment_1",
		"ssh-rsa publicKeyString_2 comment_2",
	}
	PublicKeysTenant2 := []string{
		"ssh-rsa abcdefghi fabio_comment",
	}

	var PubKeysToBeChecked []string
	WorkspacesTenants := []crownlabsalpha1.UserWorkspaceData{}

	Context("When updating a Tenant resource", func() {
		It("Should create the authorized_keys file if not already existing and insert tenant's pub keys", func() {
			By("By creating a new Tenant")
			ctx := context.Background()
			tenant1 := &crownlabsalpha1.Tenant{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "crownlabs.polito.it/v1alpha1",
					Kind:       "Tenant",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      NameTenant1,
					Namespace: NamespaceTenant1,
				},
				Spec: crownlabsalpha1.TenantSpec{
					FirstName:     FirstNameTenant1,
					LastName:      LastNameTenant1,
					Email:         EmailTenant1,
					Workspaces:    WorkspacesTenants,
					PublicKeys:    PublicKeysTenant1,
					CreateSandbox: CreateSandboxTenant1,
				},
			}
			Expect(k8sClient.Create(ctx, tenant1)).Should(Succeed())

			tenantLookupKey := types.NamespacedName{Name: NameTenant1, Namespace: ""}
			createdTenant := &crownlabsalpha1.Tenant{}

			Eventually(func() []string {
				err := k8sClient.Get(ctx, tenantLookupKey, createdTenant)
				if err != nil {
					return nil
				}
				return createdTenant.Spec.PublicKeys
			}, timeout, interval).Should(Equal(PublicKeysTenant1))

			checkFile := func() (bool, error) {

				data, err := ioutil.ReadFile(testFile)
				if err != nil {
					return false, err
				}

				for i := range PubKeysToBeChecked {
					if !bytes.Contains(data, []byte(PubKeysToBeChecked[i])) {
						return false, nil
					}
				}
				return true, nil

			}

			By("Checking the file after creation")
			PubKeysToBeChecked = PublicKeysTenant1
			Eventually(checkFile, timeout, interval).Should(BeTrue())

			By("Updating the keys on an already existing tenant resource")

			PublicKeysTenant1[0] = "ecdsa-sha2-nistp256 yet_another_public_key comment_3"

			createdTenant.Spec.PublicKeys = PublicKeysTenant1
			Eventually(func() bool {
				err := k8sClient.Update(ctx, createdTenant)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			updatedTenant := &crownlabsalpha1.Tenant{}

			Eventually(func() []string {
				err := k8sClient.Get(ctx, tenantLookupKey, updatedTenant)
				if err != nil {
					return nil
				}
				return updatedTenant.Spec.PublicKeys
			}, timeout, interval).Should(Equal(createdTenant.Spec.PublicKeys))

			By("Checking the file after updating")
			Eventually(checkFile, timeout, interval).Should(BeTrue())

			By("Adding another tenant")
			tenant2 := &crownlabsalpha1.Tenant{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "crownlabs.polito.it/v1alpha1",
					Kind:       "Tenant",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      NameTenant2,
					Namespace: NamespaceTenant2,
				},
				Spec: crownlabsalpha1.TenantSpec{
					FirstName:     FirstNameTenant2,
					LastName:      LastNameTenant2,
					Email:         EmailTenant2,
					Workspaces:    WorkspacesTenants,
					PublicKeys:    PublicKeysTenant2,
					CreateSandbox: CreateSandboxTenant2,
				},
			}
			Expect(k8sClient.Create(ctx, tenant2)).Should(Succeed())

			newTenantLookupKey := types.NamespacedName{Name: NameTenant2, Namespace: ""}
			newTenantGet := &crownlabsalpha1.Tenant{}

			Eventually(func() []string {
				err := k8sClient.Get(ctx, newTenantLookupKey, newTenantGet)
				if err != nil {
					return nil
				}
				return newTenantGet.Spec.PublicKeys
			}, timeout, interval).Should(Equal(PublicKeysTenant2))

			By("Checking the file both tenants' pub keys")
			PubKeysToBeChecked = append(PubKeysToBeChecked, PublicKeysTenant2...)
			Eventually(checkFile, timeout, interval).Should(BeTrue())

			By("Deleting the first tenant")
			Eventually(func() bool {
				err := k8sClient.Delete(ctx, &crownlabsalpha1.Tenant{
					ObjectMeta: metav1.ObjectMeta{
						Name:      NameTenant1,
						Namespace: NamespaceTenant1,
					},
				})
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Checking the file for first tenant's keys after deleting it")
			PubKeysToBeChecked = PublicKeysTenant1
			Eventually(checkFile, timeout, interval).Should(BeFalse())

			By("Checking the file for second tenant's keys after the deletion of the first")
			PubKeysToBeChecked = PublicKeysTenant2
			Eventually(checkFile, timeout, interval).Should(BeTrue())

			// remove the file
			Expect(os.Remove(testFile)).Should(Succeed())
		})
	})

})
