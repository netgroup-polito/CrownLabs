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

package tenant_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var _ = Describe("Namespace management", func() {
	Context("When tenant needs namespace resources", func() {
		It("Should create the personal namespace", func() {
			namespace := &v1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-" + tnName},
				namespace, BeTrue(), 10*time.Second, 250*time.Millisecond)

			// Verifica labels
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/type", "tenant"))
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/name", tnName))
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))
			Expect(namespace.Labels).To(HaveKeyWithValue("crownlabs.polito.it/operator-selector", "test"))
		})

		It("Should create the resource quota", func() {
			rq := &v1.ResourceQuota{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-resource-quota",
				Namespace: "tenant-" + tnName,
			}, rq, BeTrue(), 10*time.Second, 250*time.Millisecond)

			// Verifica configurazione
			Expect(rq.Labels).To(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))
			Expect(rq.Labels).To(HaveKeyWithValue("crownlabs.polito.it/operator-selector", "test"))
			Expect(rq.Spec.Hard).ToNot(BeEmpty())
		})

		It("Should create the role binding", func() {
			rb := &rbacv1.RoleBinding{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-manage-instances",
				Namespace: "tenant-" + tnName,
			}, rb, BeTrue(), 10*time.Second, 250*time.Millisecond)

			// Verifica configurazione
			Expect(rb.Labels).To(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))
			Expect(rb.Labels).To(HaveKeyWithValue("crownlabs.polito.it/operator-selector", "test"))
			Expect(rb.RoleRef.Kind).To(Equal("ClusterRole"))
			Expect(rb.RoleRef.Name).To(Equal("crownlabs-manage-instances"))
			Expect(rb.Subjects).To(HaveLen(1))
			Expect(rb.Subjects[0].Kind).To(Equal("User"))
			Expect(rb.Subjects[0].Name).To(Equal("" + tnName))
			Expect(rb.Subjects[0].APIGroup).To(Equal("rbac.authorization.k8s.io"))
		})

		It("Should create the deny network policy", func() {
			netPol := &netv1.NetworkPolicy{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-deny-ingress-traffic",
				Namespace: "tenant-" + tnName,
			}, netPol, BeTrue(), 10*time.Second, 250*time.Millisecond)

			// Verifica configurazione
			Expect(netPol.Labels).To(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))
			Expect(netPol.Labels).To(HaveKeyWithValue("crownlabs.polito.it/operator-selector", "test"))
			Expect(netPol.Spec.PodSelector.MatchLabels).To(HaveLen(0))
			Expect(netPol.Spec.Ingress).To(HaveLen(1))
			Expect(netPol.Spec.Ingress[0].From).To(HaveLen(1))
			Expect(netPol.Spec.Ingress[0].From[0].PodSelector.MatchLabels).To(HaveLen(0))
		})

		It("Should create the allow network policy", func() {
			netPol := &netv1.NetworkPolicy{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-allow-trusted-ingress-traffic",
				Namespace: "tenant-" + tnName,
			}, netPol, BeTrue(), 10*time.Second, 250*time.Millisecond)

			// Verifica configurazione
			Expect(netPol.Labels).To(HaveKeyWithValue("crownlabs.polito.it/managed-by", "tenant"))
			Expect(netPol.Labels).To(HaveKeyWithValue("crownlabs.polito.it/operator-selector", "test"))
			Expect(netPol.Spec.PodSelector.MatchLabels).To(HaveLen(0))
			Expect(netPol.Spec.Ingress).To(HaveLen(1))
			Expect(netPol.Spec.Ingress[0].From).To(HaveLen(1))
			Expect(netPol.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels).To(HaveKeyWithValue("crownlabs.polito.it/allow-instance-access", "true"))
		})
	})

	Context("When tenant namespace should be deleted", func() {
		JustBeforeEach(func() {
			// Verifica che le risorse siano state create
			namespace := &v1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-" + tnName},
				namespace, BeTrue(), 10*time.Second, 250*time.Millisecond)

			// Ottieni la versione aggiornata del tenant dal cluster
			updatedTenant := &v1alpha2.Tenant{}
			err := cl.Get(ctx, types.NamespacedName{Name: tnResource.Name}, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			// Configura per eliminazione: last login molto vecchio
			updatedTenant.Spec.LastLogin = metav1.NewTime(time.Now().Add(-48 * time.Hour))
			err = cl.Update(ctx, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			// Aggiorna la reference locale
			tnResource = updatedTenant

			// Trigger reconciliation per eliminazione
			_, err = tenantReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: tnResource.Name,
				},
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should delete the namespace resources", func() {
			namespace := &v1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-" + tnName},
				namespace, BeFalse(), 10*time.Second, 250*time.Millisecond)
		})

		It("Should delete the resource quota", func() {
			rq := &v1.ResourceQuota{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-resource-quota",
				Namespace: "tenant-" + tnName,
			}, rq, BeFalse(), 10*time.Second, 250*time.Millisecond)
		})

		It("Should delete the role binding", func() {
			rb := &rbacv1.RoleBinding{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-manage-instances",
				Namespace: "tenant-" + tnName,
			}, rb, BeFalse(), 10*time.Second, 250*time.Millisecond)
		})

		It("Should delete the deny network policy", func() {
			netPol := &netv1.NetworkPolicy{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-deny-ingress-traffic",
				Namespace: "tenant-" + tnName,
			}, netPol, BeFalse(), 10*time.Second, 250*time.Millisecond)
		})

		It("Should delete the allow network policy", func() {
			netPol := &netv1.NetworkPolicy{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-allow-trusted-ingress-traffic",
				Namespace: "tenant-" + tnName,
			}, netPol, BeFalse(), 10*time.Second, 250*time.Millisecond)
		})
	})

	Context("When tenant has instances running", func() {
		JustBeforeEach(func() {
			// Verifica che il namespace sia stato creato
			namespace := &v1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-" + tnName},
				namespace, BeTrue(), 10*time.Second, 250*time.Millisecond)

			// Crea un'istanza nel namespace del tenant
			instance := &v1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-instance",
					Namespace: "tenant-" + tnName,
					Labels: map[string]string{
						"crownlabs.polito.it/operator-selector": "test",
					},
				},
				Spec: v1alpha2.InstanceSpec{
					Running: true,
				},
			}

			// Aggiungi l'istanza al fake client esistente
			err := cl.Create(ctx, instance)
			Expect(err).ToNot(HaveOccurred())

			// Ottieni la versione aggiornata del tenant
			updatedTenant := &v1alpha2.Tenant{}
			err = cl.Get(ctx, types.NamespacedName{Name: tnResource.Name}, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			// Configura last login molto vecchio (dovrebbe eliminare il namespace ma non lo farà a causa dell'istanza)
			updatedTenant.Spec.LastLogin = metav1.NewTime(time.Now().Add(-48 * time.Hour))
			err = cl.Update(ctx, updatedTenant)
			Expect(err).ToNot(HaveOccurred())

			// Aggiorna la reference locale
			tnResource = updatedTenant

			// Trigger reconciliation
			_, err = tenantReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: tnResource.Name},
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should keep the namespace when instances are running", func() {
			namespace := &v1.Namespace{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: "tenant-" + tnName},
				namespace, BeTrue(), 10*time.Second, 250*time.Millisecond)
		})

		It("Should keep the resource quota when instances are running", func() {
			rq := &v1.ResourceQuota{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-resource-quota",
				Namespace: "tenant-" + tnName,
			}, rq, BeTrue(), 10*time.Second, 250*time.Millisecond)
		})

		It("Should keep the role binding when instances are running", func() {
			rb := &rbacv1.RoleBinding{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-manage-instances",
				Namespace: "tenant-" + tnName,
			}, rb, BeTrue(), 10*time.Second, 250*time.Millisecond)
		})

		It("Should keep the deny network policy when instances are running", func() {
			netPol := &netv1.NetworkPolicy{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-deny-ingress-traffic",
				Namespace: "tenant-" + tnName,
			}, netPol, BeTrue(), 10*time.Second, 250*time.Millisecond)
		})

		It("Should keep the allow network policy when instances are running", func() {
			netPol := &netv1.NetworkPolicy{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{
				Name:      "crownlabs-allow-trusted-ingress-traffic",
				Namespace: "tenant-" + tnName,
			}, netPol, BeTrue(), 10*time.Second, 250*time.Millisecond)
		})
	})
})
