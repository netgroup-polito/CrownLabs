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

package instautoctrl_test

import (
	"fmt"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instautoctrl"
)

var _ = Describe("Instautoctrl-expiration", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		PersistentInstanceName     = "test-expiration-instance-persistent"
		NonPersistentInstanceName  = "test-expiration-instance-non-persistent"
		NonPersistentInstanceName2 = "test-expiration-instance-non-persistent-2"
		WorkingNamespace           = "test-expiration-working-namespace"
		persistentTemplateName     = "test-expiration-template-persistent"
		nonPersistentTemplateName  = "test-expiration-template-non-persistent"
		TenantName                 = "test-expiration-tenant"
		CustomDeleteAfter          = instautoctrl.NeverTimeoutValue
		CustomInactivityTimeout    = instautoctrl.NeverTimeoutValue
		CustomDeleteAfter2         = "10s"
		CustomInactivityTimeout2   = "1m"
	)

	var (
		workingNs = v1.Namespace{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: WorkingNamespace,
				Labels: map[string]string{
					"crownlabs.polito.it/operator-selector": "test-suite",
				},
			},
			Spec:   v1.NamespaceSpec{},
			Status: v1.NamespaceStatus{},
		}
		templatePersistentEnvironment = crownlabsv1alpha2.TemplateSpec{
			WorkspaceRef: crownlabsv1alpha2.GenericRef{},
			PrettyName:   "My Template",
			Description:  "Description of my template",
			EnvironmentList: []crownlabsv1alpha2.Environment{
				{
					Name:       "env-1",
					GuiEnabled: true,
					Resources: crownlabsv1alpha2.EnvironmentResources{
						CPU:                   1,
						ReservedCPUPercentage: 1,
						Memory:                resource.MustParse("1024M"),
					},
					EnvironmentType: crownlabsv1alpha2.ClassVM,
					Persistent:      true,
					Image:           "crownlabs/vm",
				},
			},
			DeleteAfter:       CustomDeleteAfter,
			InactivityTimeout: CustomInactivityTimeout,
		}
		templateNonPersistentEnvironment = crownlabsv1alpha2.TemplateSpec{
			WorkspaceRef: crownlabsv1alpha2.GenericRef{},
			PrettyName:   "My Template",
			Description:  "Description of my template",
			EnvironmentList: []crownlabsv1alpha2.Environment{
				{
					Name:       "env-1",
					GuiEnabled: true,
					Resources: crownlabsv1alpha2.EnvironmentResources{
						CPU:                   1,
						ReservedCPUPercentage: 1,
						Memory:                resource.MustParse("1024M"),
					},
					EnvironmentType: crownlabsv1alpha2.ClassVM,
					Persistent:      false,
					Image:           "crownlabs/vm",
				},
			},
			DeleteAfter:       CustomDeleteAfter2,
			InactivityTimeout: CustomInactivityTimeout2,
		}
		persistentTemplate = crownlabsv1alpha2.Template{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      persistentTemplateName,
				Namespace: WorkingNamespace,
			},
			Spec:   templatePersistentEnvironment,
			Status: crownlabsv1alpha2.TemplateStatus{},
		}
		nonPersistentTemplate = crownlabsv1alpha2.Template{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      nonPersistentTemplateName,
				Namespace: WorkingNamespace,
			},
			Spec:   templateNonPersistentEnvironment,
			Status: crownlabsv1alpha2.TemplateStatus{},
		}

		persistentInstance = crownlabsv1alpha2.Instance{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      PersistentInstanceName,
				Namespace: WorkingNamespace,
			},
			Spec: crownlabsv1alpha2.InstanceSpec{
				Running: false,
				Template: crownlabsv1alpha2.GenericRef{
					Name:      persistentTemplateName,
					Namespace: WorkingNamespace,
				},
				Tenant: crownlabsv1alpha2.GenericRef{
					Name:      TenantName,
					Namespace: WorkingNamespace,
				},
			},
			Status: crownlabsv1alpha2.InstanceStatus{},
		}
		nonPersistentInstance = crownlabsv1alpha2.Instance{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      NonPersistentInstanceName,
				Namespace: WorkingNamespace,
			},
			Spec: crownlabsv1alpha2.InstanceSpec{
				Running: false,
				Template: crownlabsv1alpha2.GenericRef{
					Name:      nonPersistentTemplateName,
					Namespace: WorkingNamespace,
				},
				Tenant: crownlabsv1alpha2.GenericRef{
					Name: TenantName,
				},
			},
			Status: crownlabsv1alpha2.InstanceStatus{},
		}

		nonPersistentInstance2 = crownlabsv1alpha2.Instance{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      NonPersistentInstanceName2,
				Namespace: WorkingNamespace,
			},
			Spec: crownlabsv1alpha2.InstanceSpec{
				Running: false,
				Template: crownlabsv1alpha2.GenericRef{
					Name:      nonPersistentTemplateName,
					Namespace: WorkingNamespace,
				},
				Tenant: crownlabsv1alpha2.GenericRef{
					Name: TenantName,
				},
			},
			Status: crownlabsv1alpha2.InstanceStatus{},
		}

		tenant = crownlabsv1alpha2.Tenant{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: TenantName,
				Labels: map[string]string{
					"crownlabs.polito.it/operator-selector": "test-suite",
				},
			},
			Spec: crownlabsv1alpha2.TenantSpec{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@gmail.com",
				Workspaces: []crownlabsv1alpha2.TenantWorkspaceEntry{
					{Name: workingNs.Name,
						Role: "user"},
				},
			}}
	)

	BeforeEach(func() {
		mockProm.EXPECT().
			IsPrometheusHealthy(gomock.Any(), gomock.Any()).
			Return(true, nil).
			AnyTimes()

		mockProm.EXPECT().
			GetLastActivityTime(gomock.Any(), gomock.Any()).
			Return(time.Now(), nil).
			AnyTimes()

		mockProm.EXPECT().
			GetQueryNginxData().
			Return("").
			AnyTimes()

		mockProm.EXPECT().
			GetQuerySSHData().
			Return("").
			AnyTimes()

		mockProm.EXPECT().
			GetQueryWebSSHData().
			Return("").
			AnyTimes()
		newNs := workingNs.DeepCopy()
		newPersistentTemplate := persistentTemplate.DeepCopy()
		newNonPersistentTemplate := nonPersistentTemplate.DeepCopy()

		newTenant := tenant.DeepCopy()
		By("Creating the namespace where to create instance and template")
		err := k8sClientExpiration.Create(ctx, newNs)
		if err != nil && errors.IsAlreadyExists(err) {
			By("Cleaning up the environment")
			By("Deleting templates")
			Expect(k8sClientExpiration.Delete(ctx, &persistentTemplate)).Should(Succeed())
			Expect(k8sClientExpiration.Delete(ctx, &nonPersistentTemplate)).Should(Succeed())
			By("Deleting instances")
			Expect(client.IgnoreNotFound(k8sClientExpiration.Delete(ctx, &persistentInstance))).To(Succeed())
			Expect(client.IgnoreNotFound(k8sClientExpiration.Delete(ctx, &nonPersistentInstance))).To(Succeed())
			Expect(client.IgnoreNotFound(k8sClientExpiration.Delete(ctx, &nonPersistentInstance2))).To(Succeed())
			By("Deleting tenant")
			Expect(k8sClientExpiration.Delete(ctx, &tenant)).Should(Succeed())

		} else if err != nil {
			Fail(fmt.Sprintf("Unable to create namespace -> %s", err))
		}

		By("Creating the templates")
		Expect(k8sClientExpiration.Create(ctx, newPersistentTemplate)).Should(Succeed())
		Expect(k8sClientExpiration.Create(ctx, newNonPersistentTemplate)).Should(Succeed())

		By("Creating the tenant")
		Expect(k8sClientExpiration.Create(ctx, newTenant)).Should(Succeed())

	})

	Context("Testing never deletion", func() {
		It("Should succeed: the persistent VM is not deleted because it has a never deletion time", func() {
			By("Checking that the persistent template has a never deletion time")
			currentTemplate := &crownlabsv1alpha2.Template{}
			templateLookupKey := types.NamespacedName{Name: persistentTemplateName, Namespace: WorkingNamespace}
			Expect(k8sClientExpiration.Get(ctx, templateLookupKey, currentTemplate)).Should(Succeed())
			Expect(currentTemplate.Spec.DeleteAfter).To(Equal(instautoctrl.NeverTimeoutValue))
		})
		It("Should succeed: the persistent VM has a valid deletion time and should be deleted", func() {
			currentTemplate := &crownlabsv1alpha2.Template{}
			templateLookupKey := types.NamespacedName{Name: nonPersistentTemplateName, Namespace: WorkingNamespace}
			Expect(k8sClientExpiration.Get(ctx, templateLookupKey, currentTemplate)).Should(Succeed())
			Expect(currentTemplate.Spec.DeleteAfter).ToNot(Equal(CustomDeleteAfter))

		})

	})

})
