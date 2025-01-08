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

package bastion_controller

import (
	"bytes"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	crownlabsalpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var _ = Describe("Bastion controller - creating two tenants", func() {

	const (
		NameTenant1 = "s11111"
		NameTenant2 = "s22222"

		testFile = "./authorized_keys_test"
		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	var (
		PubKeysToBeChecked = map[string][]string{}
		PublicKeysTenant1  []string
		PublicKeysTenant2  []string
		tenant1LookupKey   = types.NamespacedName{Name: NameTenant1}
		tenant2LookupKey   = types.NamespacedName{Name: NameTenant2}
	)

	// this function checks if the keys are properly placed in the file.
	checkFile := func() (bool, error) {

		data, err := os.ReadFile(testFile)
		if err != nil {
			return false, err
		}

		for id, t := range PubKeysToBeChecked {

			for i := range t {
				entry, err := Create(t[i], id)
				if err != nil {
					continue
				}

				if !bytes.Contains(data, []byte(entry.Compose())) {
					return false, nil
				}
			}

		}
		return true, nil

	}

	BeforeEach(func() {
		PubKeysToBeChecked = make(map[string][]string)

		PublicKeysTenant1 = []string{
			"ssh-ed25519 publicKeyString_1 comment_1",
			"ssh-ed25519 publicKeyString_2 comment_2 with spaces",
			"ssh-rsa publicKeyString_3",
			"invalid_entry",
		}
		PublicKeysTenant2 = []string{
			"ssh-rsa abcdefghi comment",
		}

		tenant1 := &crownlabsalpha2.Tenant{}
		tenant2 := &crownlabsalpha2.Tenant{}

		// create or update tenant in order to reset the specs

		if err1 := k8sClient.Get(ctx, tenant1LookupKey, tenant1); err1 != nil {
			tenant1 = &crownlabsalpha2.Tenant{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "crownlabs.polito.it/v1alpha1",
					Kind:       "Tenant",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: NameTenant1,
				},
				Spec: crownlabsalpha2.TenantSpec{
					FirstName:  "Mario",
					LastName:   "Rossi",
					Email:      "mario.rossi@fakemail.com",
					Workspaces: []crownlabsalpha2.TenantWorkspaceEntry{},
					PublicKeys: PublicKeysTenant1,
				},
			}
			Expect(k8sClient.Create(ctx, tenant1)).Should(Succeed())
		} else {
			tenant1.Spec.PublicKeys = PublicKeysTenant1
			Expect(k8sClient.Update(ctx, tenant1)).Should(Succeed())
		}

		updatedTenant1 := &crownlabsalpha2.Tenant{}

		Eventually(func() []string {
			err := k8sClient.Get(ctx, tenant1LookupKey, updatedTenant1)
			if err != nil {
				return nil
			}
			return updatedTenant1.Spec.PublicKeys
		}, timeout, interval).Should(Equal(PublicKeysTenant1))

		if err2 := k8sClient.Get(ctx, tenant2LookupKey, tenant2); err2 != nil {
			tenant2 = &crownlabsalpha2.Tenant{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "crownlabs.polito.it/v1alpha1",
					Kind:       "Tenant",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: NameTenant2,
				},
				Spec: crownlabsalpha2.TenantSpec{
					FirstName:  "Fabio",
					LastName:   "Bianchi",
					Email:      "fabio.bianchi@fakemail.com",
					Workspaces: []crownlabsalpha2.TenantWorkspaceEntry{},
					PublicKeys: PublicKeysTenant2,
				},
			}
			Expect(k8sClient.Create(ctx, tenant2)).Should(Succeed())
		} else {
			tenant2.Spec.PublicKeys = PublicKeysTenant2
			Expect(k8sClient.Update(ctx, tenant2)).Should(Succeed())
		}

		updatedTenant2 := &crownlabsalpha2.Tenant{}

		Eventually(func() []string {
			err := k8sClient.Get(ctx, tenant2LookupKey, updatedTenant2)
			if err != nil {
				return nil
			}
			return updatedTenant2.Spec.PublicKeys
		}, timeout, interval).Should(Equal(PublicKeysTenant2))

	})

	Context("When creating two tenants", func() {
		It("Should create the file authorized_keys and write the tenant's public keys into it.", func() {

			By("Checking if the file exists.")
			Eventually(func() bool {
				if _, err := os.Stat(testFile); err == nil {
					return true
				}
				return false
			}).Should(BeTrue())

			By("Checking the file for both tenants' pub keys")
			PubKeysToBeChecked[NameTenant2] = PublicKeysTenant1
			PubKeysToBeChecked[NameTenant2] = PublicKeysTenant2
			Eventually(checkFile, timeout, interval).Should(BeTrue())

		})
	})

	Context("When updating the keys of the one tenant", func() {
		BeforeEach(func() {

			createdTenant := &crownlabsalpha2.Tenant{}

			Eventually(func() []string {
				err := k8sClient.Get(ctx, tenant1LookupKey, createdTenant)
				if err != nil {
					return nil
				}
				return createdTenant.Spec.PublicKeys
			}, timeout, interval).Should(Equal(PublicKeysTenant1))

			Expect(retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: NameTenant1}, createdTenant)
				if err != nil {
					return err
				}
				PublicKeysTenant1[0] = "ecdsa-sha2-nistp256 yet_another_public_key comment_3"

				createdTenant.Spec.PublicKeys = PublicKeysTenant1
				return k8sClient.Update(ctx, createdTenant)
			})).Should(Succeed())

			updatedTenant := &crownlabsalpha2.Tenant{}

			Eventually(func() []string {
				err := k8sClient.Get(ctx, tenant1LookupKey, updatedTenant)
				if err != nil {
					return nil
				}
				return updatedTenant.Spec.PublicKeys
			}, timeout, interval).Should(Equal(createdTenant.Spec.PublicKeys))
		})

		It("Should update the keys in the file", func() {

			By("Checking the file after updating")
			PubKeysToBeChecked[NameTenant1] = PublicKeysTenant1
			Eventually(checkFile, timeout, interval).Should(BeTrue())

		})
	})

	Context("When deleting a Tenant", func() {
		BeforeEach(func() {
			Expect(k8sClient.Delete(ctx, &crownlabsalpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name: NameTenant1,
				},
			})).Should(Succeed())
		})

		It("Should contain only the keys of the remaining tenant", func() {

			PubKeysToBeChecked[NameTenant1] = PublicKeysTenant1
			PubKeysToBeChecked[NameTenant2] = PublicKeysTenant2
			By("Checking the file for first tenant's keys after deleting it")
			Eventually(checkFile, timeout, interval).Should(BeFalse())

			By("Checking the file for second tenant's keys after the deletion of the first")
			delete(PubKeysToBeChecked, NameTenant1)
			Eventually(checkFile, timeout, interval).Should(BeTrue())

		})

	})

})
