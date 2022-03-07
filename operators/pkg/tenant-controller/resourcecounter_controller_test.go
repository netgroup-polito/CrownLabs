// Copyright 2020-2022 Politecnico di Torino
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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var _ = Describe("Resource counter controller", func() {

	ctx := context.Background()
	var (
		clientBuilder fake.ClientBuilder
		tenant        crownlabsv1alpha2.Tenant
		instance1     crownlabsv1alpha2.Instance
		instance2     crownlabsv1alpha2.Instance
		template1     crownlabsv1alpha2.Template
		template2     crownlabsv1alpha2.Template
		environment1  crownlabsv1alpha2.Environment
		environment2  crownlabsv1alpha2.Environment
		tnName        = "test"
		tnEmail       = "test@email.com"
		nsName        = "tenant-test"
		instance1Name = "inst1"
		instance2Name = "inst2"
		template1Name = "temp1"
		template2Name = "temp2"
	)

	RunReconciler := func() error {
		resourceCounterReconciler := ResourceCounterReconciler{
			Client: clientBuilder.Build(),
		}
		namespacedName := types.NamespacedName{
			Name:      tnName,
			Namespace: nsName,
		}
		_, err := resourceCounterReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: namespacedName,
		})
		if err != nil {
			return err
		}
		return resourceCounterReconciler.Client.Get(ctx, namespacedName, &tenant)
	}

	BeforeEach(func() {
		environment1 = crownlabsv1alpha2.Environment{
			EnvironmentType: crownlabsv1alpha2.ClassContainer,
			Resources: crownlabsv1alpha2.EnvironmentResources{
				CPU:                   1,
				ReservedCPUPercentage: 20,
			},
		}

		environment2 = crownlabsv1alpha2.Environment{
			EnvironmentType: crownlabsv1alpha2.ClassContainer,
			Resources: crownlabsv1alpha2.EnvironmentResources{
				CPU:                   2,
				ReservedCPUPercentage: 20,
			},
		}

	})

	JustBeforeEach(func() {
		ns := corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: nsName}}

		//		Expect(k8sClient.Create(ctx, &ns)).To(Succeed())

		tenant = crownlabsv1alpha2.Tenant{
			ObjectMeta: v1.ObjectMeta{
				Name:      tnName,
				Namespace: nsName,
			},
			Spec: crownlabsv1alpha2.TenantSpec{
				Email: tnEmail,
				ResourceToken: &crownlabsv1alpha2.TenantToken{
					Used:     0,
					Reserved: 100000,
					TokenCounterLastUpdated: v1.Time{
						Time: time.Now().Add(time.Minute * -10), // ten minutes before now.
					},
				},
			},
		}

		//		Expect(k8sClient.Create(ctx, &tenant)).Should(Succeed())

		template1 = crownlabsv1alpha2.Template{
			ObjectMeta: v1.ObjectMeta{Name: template1Name, Namespace: nsName},
			Spec: crownlabsv1alpha2.TemplateSpec{
				EnvironmentList: []crownlabsv1alpha2.Environment{environment1},
			},
		}

		//		Expect(k8sClient.Create(ctx, &template1)).Should(Succeed())

		template2 = crownlabsv1alpha2.Template{
			ObjectMeta: v1.ObjectMeta{Name: template2Name, Namespace: nsName},
			Spec: crownlabsv1alpha2.TemplateSpec{
				EnvironmentList: []crownlabsv1alpha2.Environment{environment2},
			},
		}

		//		Expect(k8sClient.Create(ctx, &template2)).Should(Succeed())

		instance1 = crownlabsv1alpha2.Instance{
			ObjectMeta: v1.ObjectMeta{
				Name:      instance1Name,
				Namespace: nsName,
				CreationTimestamp: v1.Time{
					Time: time.Now().Add(time.Minute * -20), // twenty minutes before now.
				},
			},
			Spec: crownlabsv1alpha2.InstanceSpec{
				Template: crownlabsv1alpha2.GenericRef{Name: template1.Name, Namespace: nsName},
				Tenant:   crownlabsv1alpha2.GenericRef{Name: tnName, Namespace: nsName},
			},
		}

		//		Expect(k8sClient.Create(ctx, &instance1)).Should(Succeed())

		instance2 = crownlabsv1alpha2.Instance{
			ObjectMeta: v1.ObjectMeta{
				Name:      instance2Name,
				Namespace: nsName,
				CreationTimestamp: v1.Time{
					Time: time.Now().Add(time.Minute * -5), // five minutes before now.
				},
			},
			Spec: crownlabsv1alpha2.InstanceSpec{
				Template: crownlabsv1alpha2.GenericRef{Name: template2.Name, Namespace: nsName},
				Tenant:   crownlabsv1alpha2.GenericRef{Name: tnName, Namespace: nsName},
			},
		}

		//		Expect(k8sClient.Create(ctx, &instance2)).Should(Succeed())
		clientBuilder = *fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(
			&ns,
			&tenant,
			&template1,
			&template2,
			&instance1,
			&instance2,
		)
	})

	It("should consider the more recent timestamp between instance creation and token last update", func() {
		expectedUsedToken := uint32((10 + 2*5) * 60)
		Expect(RunReconciler()).To(Succeed())
		Expect(tenant.Spec.ResourceToken.Used).To(Equal(expectedUsedToken))
	})
})
