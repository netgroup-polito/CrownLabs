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
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var _ = Describe("Labels", func() {
	It("Should add the operator target label", func() {
		tn := &v1alpha2.Tenant{}

		DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)
		Expect(tn.Labels).To(HaveKeyWithValue("crownlabs.polito.it/operator-selector", "test"))
	})

	It("Should add the first and last name labels", func() {
		tn := &v1alpha2.Tenant{}

		DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)
		Expect(tn.Labels).To(HaveKeyWithValue("crownlabs.polito.it/first-name", "Test"))
		Expect(tn.Labels).To(HaveKeyWithValue("crownlabs.polito.it/last-name", "Tenant"))
	})

	Context("When there is an error updating the labels", func() {
		BeforeEach(func() {
			builder.WithInterceptorFuncs(interceptor.Funcs{
				Patch: func(_ context.Context, _ client.WithWatch, obj client.Object, _ client.Patch, _ ...client.PatchOption) error {
					if tn, ok := obj.(*v1alpha2.Tenant); ok && tn.Name == tnName && len(tn.Labels) > 1 {
						return fmt.Errorf("error updating labels")
					}
					return nil
				},
			})

			tnReconcileErrExpected = HaveOccurred()
		})

		It("Should return an error", func() {
			// checked in BeforeEach
		})

		It("Should set status to not ready", func() {
			tn := &v1alpha2.Tenant{}

			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)
			Expect(tn.Status.Ready).To(BeFalse())
		})
	})
})
