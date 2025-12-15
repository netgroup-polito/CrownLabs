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
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	pkgcontext "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instautoctrl"
)

var _ = Describe("Instautoctrl-expiration-unit", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		PersistentInstanceName    = "test-instance-persistent-unit-test-expiration"
		NonPersistentInstanceName = "test-instance-non-persistent-unit-test-expiration"
		WorkingNamespace          = "working-namespace-instautoctrl-test-expiration"
		persistentTemplateName    = "test-template-persistent-unit-test-expiration"
		nonPersistentTemplateName = "test-template-non-persistent-unit-test-expiration"
		TenantName                = "test-tenant-unit-test-expiration"
		CustomDeleteAfter         = instautoctrl.NeverTimeoutValue
		CustomInactivityTimeout   = instautoctrl.NeverTimeoutValue
		CustomDeleteAfter2        = "1m"
		CustomInactivityTimeout2  = "2m"
		tolerance                 = time.Minute

		timeout  = time.Second * 130
		interval = time.Millisecond * 1000
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
		tenantNs = v1.Namespace{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: TenantName,
				Labels: map[string]string{
					"crownlabs.polito.it/operator-selector": "test-suite",
					"crownlabs.polito.it/tenant":            TenantName,
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
				Namespace: tenantNs.Name,
				Labels: map[string]string{
					"crownlabs.polito.it/operator-selector": "test-suite",
					"crownlabs.polito.it/tenant":            TenantName,
					"crownlabs.polito.it/workspace":         WorkingNamespace,
					"crownlabs.polito.it/template":          persistentTemplateName,
					"crownlabs.polito.it/instance-type":     "non-persistent",
				},
			},
			Spec: crownlabsv1alpha2.InstanceSpec{
				Running: true,
				Template: crownlabsv1alpha2.GenericRef{
					Name:      persistentTemplateName,
					Namespace: WorkingNamespace,
				},
				Tenant: crownlabsv1alpha2.GenericRef{
					Name:      TenantName,
					Namespace: tenantNs.Name,
				},
			},
			Status: crownlabsv1alpha2.InstanceStatus{},
		}
		nonPersistentInstance = crownlabsv1alpha2.Instance{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      NonPersistentInstanceName,
				Namespace: tenantNs.Name,
				Labels: map[string]string{
					"crownlabs.polito.it/operator-selector": "test-suite",
					"crownlabs.polito.it/tenant":            TenantName,
					"crownlabs.polito.it/workspace":         WorkingNamespace,
					"crownlabs.polito.it/template":          nonPersistentTemplateName,
					"crownlabs.polito.it/instance-type":     "non-persistent",
				},
			},
			Spec: crownlabsv1alpha2.InstanceSpec{
				Running: true,
				Template: crownlabsv1alpha2.GenericRef{
					Name:      nonPersistentTemplateName,
					Namespace: WorkingNamespace,
				},
				Tenant: crownlabsv1alpha2.GenericRef{
					Name:      TenantName,
					Namespace: tenantNs.Name,
				},
			},
			Status: crownlabsv1alpha2.InstanceStatus{},
		}

		tenant = crownlabsv1alpha2.Tenant{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      TenantName,
				Namespace: TenantName,
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
	var cleanupEnvironment = func() {
		By("Cleaning up the environment")
		By("Deleting templates")
		Expect(k8sClient.Delete(ctx, &persistentTemplate)).Should(Succeed())
		Expect(k8sClient.Delete(ctx, &nonPersistentTemplate)).Should(Succeed())

		By("Deleting instances")
		Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, &persistentInstance))).To(Succeed())
		Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, &nonPersistentInstance))).To(Succeed())

		By("Deleting tenant")
		Expect(k8sClient.Delete(ctx, &tenant)).Should(Succeed())
	}

	BeforeEach(func() {
		newPersistentTemplate := persistentTemplate.DeepCopy()
		newNonPersistentTemplate := nonPersistentTemplate.DeepCopy()
		newPersistentInstance := persistentInstance.DeepCopy()
		newNonPersistentInstance := nonPersistentInstance.DeepCopy()
		newTenantNs := tenantNs.DeepCopy()
		newWorkingNs := workingNs.DeepCopy()
		newTenant := tenant.DeepCopy()
		By("Creating the namespace where to create instance and template")
		err1 := k8sClient.Create(ctx, newTenantNs)
		err2 := k8sClient.Create(ctx, newWorkingNs)
		if (err1 != nil || err2 != nil) && (errors.IsAlreadyExists(err1) || errors.IsAlreadyExists(err2)) {
			cleanupEnvironment()
		} else if err1 != nil || err2 != nil {
			Fail(fmt.Sprintf("Unable to create namespace -> %s %s", err1, err2))
		}

		By("By checking that the namespace has been created")
		createdNs := &v1.Namespace{}
		tenantNs := &v1.Namespace{}

		nsLookupKey := types.NamespacedName{Name: WorkingNamespace}
		Expect(k8sClient.Get(ctx, nsLookupKey, createdNs)).To(Succeed())
		tenantNsLookupKey := types.NamespacedName{Name: TenantName}
		Expect(k8sClient.Get(ctx, tenantNsLookupKey, tenantNs)).To(Succeed())

		By("Creating the tenant")
		Expect(k8sClient.Create(ctx, newTenant)).Should(Succeed())

		By("Creating the templates")
		Expect(k8sClient.Create(ctx, newPersistentTemplate)).Should(Succeed())
		Expect(k8sClient.Create(ctx, newNonPersistentTemplate)).Should(Succeed())

		By("By checking that the template has been created")
		persistentTemplateLookupKey := types.NamespacedName{Name: persistentTemplateName, Namespace: WorkingNamespace}
		nonPersistentTemplateLookupKey := types.NamespacedName{Name: nonPersistentTemplateName, Namespace: WorkingNamespace}
		createdPersitentTemplate := &crownlabsv1alpha2.Template{}
		createdNonPersitentTemplate := &crownlabsv1alpha2.Template{}

		Expect(k8sClient.Get(ctx, persistentTemplateLookupKey, createdPersitentTemplate)).To(Succeed())
		Expect(k8sClient.Get(ctx, nonPersistentTemplateLookupKey, createdNonPersitentTemplate)).To(Succeed())

		By("Creating the instances")
		Expect(k8sClient.Create(ctx, newPersistentInstance)).Should(Succeed())
		Expect(k8sClient.Create(ctx, newNonPersistentInstance)).Should(Succeed())

		By("Checking that the instances has been created")
		persistanteInstanceLookupKey := types.NamespacedName{Name: PersistentInstanceName, Namespace: tenant.Name}
		nonPersistentInstanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenant.Name}
		createdPersistentInstance := &crownlabsv1alpha2.Instance{}
		createdNonPersistentInstance := &crownlabsv1alpha2.Instance{}

		Expect(k8sClient.Get(ctx, persistanteInstanceLookupKey, createdPersistentInstance)).To(Succeed())
		Expect(k8sClient.Get(ctx, nonPersistentInstanceLookupKey, createdNonPersistentInstance)).To(Succeed())

	})

	It("testing CheckInstanceExpiration function", func() {
		r := &instautoctrl.InstanceExpirationReconciler{
			Client: k8sClient,
		}
		By("Checking that the instance is running")
		InstanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenantNs.Name}
		currentInstance := &crownlabsv1alpha2.Instance{}
		Expect(k8sClient.Get(ctx, InstanceLookupKey, currentInstance)).To(Succeed())
		TemplateLookupKey := types.NamespacedName{Name: nonPersistentTemplateName, Namespace: WorkingNamespace}
		currentTemplate := &crownlabsv1alpha2.Template{}
		Expect(k8sClient.Get(ctx, TemplateLookupKey, currentTemplate)).To(Succeed())
		tenantLookupKey := types.NamespacedName{Name: TenantName, Namespace: tenantNs.Name}
		currentTenant := &crownlabsv1alpha2.Tenant{}
		Expect(k8sClient.Get(ctx, tenantLookupKey, currentTenant)).To(Succeed())
		ctx := context.Background()
		ctx, _ = pkgcontext.InstanceInto(ctx, currentInstance)
		ctx, _ = pkgcontext.TemplateInto(ctx, currentTemplate)
		ctx, _ = pkgcontext.TenantInto(ctx, currentTenant)
		remainingTime, err := r.CheckInstanceExpiration(ctx, currentTemplate.Spec.InactivityTimeout)
		Expect(err).ToNot(HaveOccurred())
		deleteAfterDuration, err := time.ParseDuration(CustomDeleteAfter2)
		Expect(err).ToNot(HaveOccurred())
		Expect(remainingTime).To(BeNumerically("<=", deleteAfterDuration+tolerance))

	})
	It("testing deleteInstance function", func() {
		r := &instautoctrl.InstanceExpirationReconciler{
			Client: k8sClient,
		}
		By("Checking that the instance is running")
		InstanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenantNs.Name}
		currentInstance := &crownlabsv1alpha2.Instance{}
		Expect(k8sClient.Get(ctx, InstanceLookupKey, currentInstance)).To(Succeed())
		TemplateLookupKey := types.NamespacedName{Name: nonPersistentTemplateName, Namespace: WorkingNamespace}
		currentTemplate := &crownlabsv1alpha2.Template{}
		Expect(k8sClient.Get(ctx, TemplateLookupKey, currentTemplate)).To(Succeed())
		tenantLookupKey := types.NamespacedName{Name: TenantName, Namespace: tenantNs.Name}
		currentTenant := &crownlabsv1alpha2.Tenant{}
		Expect(k8sClient.Get(ctx, tenantLookupKey, currentTenant)).To(Succeed())
		ctx := context.Background()
		ctx, _ = pkgcontext.InstanceInto(ctx, currentInstance)
		ctx, _ = pkgcontext.TemplateInto(ctx, currentTemplate)
		ctx, _ = pkgcontext.TenantInto(ctx, currentTenant)
		By("Calling deleteInstance function")
		err := r.DeleteInstance(ctx)
		Expect(err).ToNot(HaveOccurred())
		By("Checking that the instance has been deleted")
		doesEventuallyExists(ctx, InstanceLookupKey, currentInstance, BeFalse(), timeout, interval, k8sClient)
	})
	It("Testing ShouldSendWarningNotification function", func() {
		r := &instautoctrl.InstanceExpirationReconciler{
			Client: k8sClient,
		}
		By("Checking that the instance is running")
		InstanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenantNs.Name}
		currentInstance := &crownlabsv1alpha2.Instance{}
		Expect(k8sClient.Get(ctx, InstanceLookupKey, currentInstance)).To(Succeed())
		TemplateLookupKey := types.NamespacedName{Name: nonPersistentTemplateName, Namespace: WorkingNamespace}
		currentTemplate := &crownlabsv1alpha2.Template{}
		Expect(k8sClient.Get(ctx, TemplateLookupKey, currentTemplate)).To(Succeed())
		tenantLookupKey := types.NamespacedName{Name: TenantName, Namespace: tenantNs.Name}
		currentTenant := &crownlabsv1alpha2.Tenant{}
		Expect(k8sClient.Get(ctx, tenantLookupKey, currentTenant)).To(Succeed())
		ctx := context.Background()
		ctx, _ = pkgcontext.InstanceInto(ctx, currentInstance)
		ctx, _ = pkgcontext.TemplateInto(ctx, currentTemplate)
		ctx, _ = pkgcontext.TenantInto(ctx, currentTenant)

		By("Calling ShouldSendWarningNotification function")
		shouldSend, err := r.ShouldSendWarningNotification(ctx)
		Expect(err).ToNot(HaveOccurred())
		Expect(shouldSend).To(BeTrue(), "Should send notification because the instance is not deleted yet")

		By("Verifying the ExpiringWarningNotificationTimestampAnnotation annotation has been added to the instance")
		Expect(currentInstance.Annotations[forge.ExpiringWarningNotificationTimestampAnnotation]).ToNot(BeEmpty())
	})

})
