// Copyright 2020-2026 Politecnico di Torino
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
	pkgcontext "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instautoctrl"
)

var _ = Describe("Instautoctrl inactivity unit test", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		PersistentInstanceName               = "test-instance-persistent"
		PersistentInstanceName2              = "test-instance-persistent2"
		NonPersistentInstanceName            = "test-instance-non-persistent"
		WorkingNamespace                     = "working-namespace"
		persistentTemplateName               = "test-template-persistent"
		persistentTemplateName2              = "test-template-persistent-2"
		nonPersistentTemplateName            = "test-template-non-persistent"
		TenantName                           = "test-tenant"
		CustomDeleteAfter                    = instautoctrl.NeverTimeoutValue
		CustomInactivityTimeout              = instautoctrl.NeverTimeoutValue
		CustomDeleteAfterNonPersistent       = "1m"
		CustomInactivityTimeoutNonPersistent = "1m"
		CustomDeleteAfterPersistent2         = "0m"
		CustomInactivityTimeoutPersistent2   = "0m"

		timeout  = time.Second * 150
		interval = time.Millisecond * 500
	)
	whiteListMap := map[string]string{
		"crownlabs.polito.it/operator-selector": "test-suite",
	}

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
			DeleteAfter:       CustomDeleteAfterNonPersistent,
			InactivityTimeout: CustomInactivityTimeoutNonPersistent,
		}
		templatePersistentEnvironmentWithCustomInactivityTimeout = crownlabsv1alpha2.TemplateSpec{
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
			DeleteAfter:       CustomDeleteAfterPersistent2,
			InactivityTimeout: CustomInactivityTimeoutPersistent2,
		}
		persistentTemplate2 = crownlabsv1alpha2.Template{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      persistentTemplateName2,
				Namespace: WorkingNamespace,
			},
			Spec:   templatePersistentEnvironmentWithCustomInactivityTimeout,
			Status: crownlabsv1alpha2.TemplateStatus{},
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
		persistentInstance2 = crownlabsv1alpha2.Instance{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      PersistentInstanceName2,
				Namespace: tenantNs.Name,
				Labels: map[string]string{
					"crownlabs.polito.it/operator-selector": "test-suite",
					"crownlabs.polito.it/tenant":            TenantName,
					"crownlabs.polito.it/workspace":         WorkingNamespace,
					"crownlabs.polito.it/template":          persistentTemplateName2,
					"crownlabs.polito.it/instance-type":     "persistent",
				},
			},
			Spec: crownlabsv1alpha2.InstanceSpec{
				Running: true,
				Template: crownlabsv1alpha2.GenericRef{
					Name:      persistentTemplateName2,
					Namespace: WorkingNamespace,
				},
				Tenant: crownlabsv1alpha2.GenericRef{
					Name:      TenantName,
					Namespace: tenantNs.Name,
				},
			},
			Status: crownlabsv1alpha2.InstanceStatus{},
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
					"crownlabs.polito.it/template":          nonPersistentTemplateName,
					"crownlabs.polito.it/instance-type":     "persistent",
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
		Expect(k8sClient.Delete(ctx, &persistentTemplate2)).Should(Succeed())

		By("Deleting instances")
		Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, &persistentInstance))).To(Succeed())
		Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, &nonPersistentInstance))).To(Succeed())
		Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, &persistentInstance2))).To(Succeed())

		By("Deleting tenant")
		Expect(k8sClient.Delete(ctx, &tenant)).Should(Succeed())
	}

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
		tenNs := tenantNs.DeepCopy()
		newPersistentTemplate := persistentTemplate.DeepCopy()
		newNonPersistentTemplate := nonPersistentTemplate.DeepCopy()
		newPersistentTemplate2 := persistentTemplate2.DeepCopy()
		newPersistentInstance := persistentInstance.DeepCopy()
		newNonPersistentInstance := nonPersistentInstance.DeepCopy()
		newPersistentInstance2 := persistentInstance2.DeepCopy()
		newTenant := tenant.DeepCopy()

		By("Creating the namespace where to create instance and template")
		err1 := k8sClient.Create(ctx, tenNs)
		err2 := k8sClient.Create(ctx, newNs)
		if (err1 != nil || err2 != nil) && (errors.IsAlreadyExists(err1) || errors.IsAlreadyExists(err2)) {
			cleanupEnvironment()
		} else if err1 != nil || err2 != nil {
			Fail(fmt.Sprintf("Unable to create namespace -> %s %s", err1, err2))
		}

		By("By checking that the namespace has been created")
		workNs := &v1.Namespace{}
		tenantNs := &v1.Namespace{}

		nsLookupKey := types.NamespacedName{Name: WorkingNamespace}
		doesEventuallyExists(ctx, nsLookupKey, workNs, BeTrue(), timeout, interval, k8sClient)
		tenantNsLookupKey := types.NamespacedName{Name: TenantName}
		doesEventuallyExists(ctx, tenantNsLookupKey, tenantNs, BeTrue(), timeout, interval, k8sClient)

		By("Creating the templates")

		Expect(k8sClient.Create(ctx, newPersistentTemplate)).Should(Succeed())
		Expect(k8sClient.Create(ctx, newNonPersistentTemplate)).Should(Succeed())
		Expect(k8sClient.Create(ctx, newPersistentTemplate2)).Should(Succeed())

		By("By checking that the template has been created")
		persistentTemplateLookupKey := types.NamespacedName{Name: persistentTemplateName, Namespace: WorkingNamespace}
		nonPersistentTemplateLookupKey := types.NamespacedName{Name: nonPersistentTemplateName, Namespace: WorkingNamespace}
		persistentTemplate2LookupKey := types.NamespacedName{Name: persistentTemplateName2, Namespace: WorkingNamespace}
		createdPersitentTemplate := &crownlabsv1alpha2.Template{}
		createdNonPersitentTemplate := &crownlabsv1alpha2.Template{}
		createdPersistentTemplate2 := &crownlabsv1alpha2.Template{}

		doesEventuallyExists(ctx, persistentTemplateLookupKey, createdPersitentTemplate, BeTrue(), timeout, interval, k8sClient)
		doesEventuallyExists(ctx, nonPersistentTemplateLookupKey, createdNonPersitentTemplate, BeTrue(), timeout, interval, k8sClient)
		doesEventuallyExists(ctx, persistentTemplate2LookupKey, createdPersistentTemplate2, BeTrue(), timeout, interval, k8sClient)

		By("Creating the tenant")
		Expect(k8sClient.Create(ctx, newTenant)).Should(Succeed())

		By("Checking that the tenant has been created")
		tenantLookupKey := types.NamespacedName{Name: TenantName, Namespace: WorkingNamespace}
		createdTenant := &crownlabsv1alpha2.Tenant{}
		doesEventuallyExists(ctx, tenantLookupKey, createdTenant, BeTrue(), timeout, interval, k8sClient)

		By("Creating the instances")
		Expect(k8sClient.Create(ctx, newPersistentInstance)).Should(Succeed())
		Expect(k8sClient.Create(ctx, newNonPersistentInstance)).Should(Succeed())
		Expect(k8sClient.Create(ctx, newPersistentInstance2)).Should(Succeed())

		By("Checking that the instances has been created")
		persistanteInstanceLookupKey := types.NamespacedName{Name: PersistentInstanceName, Namespace: tenantNs.Name}
		nonPersistentInstanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenNs.Name}
		persistentInstance2LookupKey := types.NamespacedName{Name: PersistentInstanceName, Namespace: tenantNs.Name}
		createdPersistentInstance := &crownlabsv1alpha2.Instance{}
		createdNonPersistentInstance := &crownlabsv1alpha2.Instance{}
		createdPersistentInstance2 := &crownlabsv1alpha2.Instance{}

		doesEventuallyExists(ctx, persistanteInstanceLookupKey, createdPersistentInstance, BeTrue(), timeout, interval, k8sClient)
		doesEventuallyExists(ctx, nonPersistentInstanceLookupKey, createdNonPersistentInstance, BeTrue(), timeout, interval, k8sClient)
		doesEventuallyExists(ctx, persistentInstance2LookupKey, createdPersistentInstance2, BeTrue(), timeout, interval, k8sClient)

	})

	Describe("testing TerminateInstance function", func() {
		It("should delete the instance successfully when terminated", func() {
			r := &instautoctrl.InstanceInactiveTerminationReconciler{
				Client:             k8sClient,
				NamespaceWhitelist: metav1.LabelSelector{MatchLabels: whiteListMap, MatchExpressions: []metav1.LabelSelectorRequirement{}},
				Prometheus:         mockProm,
			}

			By("Checking that the instance is running")
			InstanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenantNs.Name}
			currentInstance := &crownlabsv1alpha2.Instance{}
			doesEventuallyExists(ctx, InstanceLookupKey, currentInstance, BeTrue(), timeout, interval, k8sClient)
			TemplateLookupKey := types.NamespacedName{Name: nonPersistentTemplateName, Namespace: WorkingNamespace}
			currentTemplate := &crownlabsv1alpha2.Template{}
			doesEventuallyExists(ctx, TemplateLookupKey, currentTemplate, BeTrue(), timeout, interval, k8sClient)
			tenantLookupKey := types.NamespacedName{Name: TenantName, Namespace: tenantNs.Name}
			currentTenant := &crownlabsv1alpha2.Tenant{}
			doesEventuallyExists(ctx, tenantLookupKey, currentTenant, BeTrue(), timeout, interval, k8sClient)
			ctx := context.Background()
			ctx, _ = pkgcontext.InstanceInto(ctx, currentInstance)
			ctx, _ = pkgcontext.TemplateInto(ctx, currentTemplate)
			ctx, _ = pkgcontext.TenantInto(ctx, currentTenant)
			By("Calling TerminateInstance function")
			err := r.TerminateInstance(ctx)
			Expect(err).ToNot(HaveOccurred())

			By("Checking that the instance has been deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, InstanceLookupKey, currentInstance)
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue(), "Instance should be deleted")
		})
	})

	Describe("Testing UpdateInstanceLastLogin function", func() {
		It("should update the last login time of the instance", func() {

			r := &instautoctrl.InstanceInactiveTerminationReconciler{
				Client:             k8sClient,
				NamespaceWhitelist: metav1.LabelSelector{MatchLabels: whiteListMap, MatchExpressions: []metav1.LabelSelectorRequirement{}},
				Prometheus:         mockProm,
			}

			By("Checking that the instance is running")
			InstanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenantNs.Name}
			currentInstance := &crownlabsv1alpha2.Instance{}
			doesEventuallyExists(ctx, InstanceLookupKey, currentInstance, BeTrue(), timeout, interval, k8sClient)
			TemplateLookupKey := types.NamespacedName{Name: nonPersistentTemplateName, Namespace: WorkingNamespace}
			currentTemplate := &crownlabsv1alpha2.Template{}
			doesEventuallyExists(ctx, TemplateLookupKey, currentTemplate, BeTrue(), timeout, interval, k8sClient)
			tenantLookupKey := types.NamespacedName{Name: TenantName, Namespace: tenantNs.Name}
			currentTenant := &crownlabsv1alpha2.Tenant{}
			doesEventuallyExists(ctx, tenantLookupKey, currentTenant, BeTrue(), timeout, interval, k8sClient)
			ctx := context.Background()
			ctx, _ = pkgcontext.InstanceInto(ctx, currentInstance)
			ctx, _ = pkgcontext.TemplateInto(ctx, currentTemplate)
			ctx, _ = pkgcontext.TenantInto(ctx, currentTenant)

			oldLastLogin := currentInstance.GetAnnotations()[forge.LastActivityAnnotation]
			err := r.SetupInstanceAnnotations(ctx)
			Expect(err).ToNot(HaveOccurred(), "SetupInstanceAnnotations should not return an error")
			inactivityTimeoutDuration := time.Hour * 24 * 14
			err = r.UpdateInstanceLastLogin(ctx, inactivityTimeoutDuration)
			Expect(err).ToNot(HaveOccurred(), "UpdateInstanceLastLogin should not return an error")

			By("Checking that the instance has been updated")
			Eventually(func() bool {
				updatedInstance := &crownlabsv1alpha2.Instance{}
				err := k8sClient.Get(ctx, InstanceLookupKey, updatedInstance)
				if err != nil {
					return false
				}
				annotations := updatedInstance.GetAnnotations()
				newLoginTime, ok := annotations[forge.LastActivityAnnotation]
				if !ok {
					return false
				}
				return newLoginTime != oldLastLogin
			}, timeout, interval).Should(BeTrue(), "Instance should be updated with a new last login time")
		})
	})

	Describe("Testing GetRemainingInactivityTime function", func() {
		var (
			r                         *instautoctrl.InstanceInactiveTerminationReconciler
			ctx                       context.Context
			currentInstance           *crownlabsv1alpha2.Instance
			currentTemplate           *crownlabsv1alpha2.Template
			currentTenant             *crownlabsv1alpha2.Tenant
			inactivityTimeoutDuration time.Duration
		)

		JustBeforeEach(func() {
			ctx = context.Background()

			r = &instautoctrl.InstanceInactiveTerminationReconciler{
				Client:             k8sClient,
				NamespaceWhitelist: metav1.LabelSelector{MatchLabels: whiteListMap},
				Prometheus:         mockProm,
			}

			By("Checking that the instance is running")
			InstanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenantNs.Name}
			currentInstance = &crownlabsv1alpha2.Instance{}
			doesEventuallyExists(ctx, InstanceLookupKey, currentInstance, BeTrue(), timeout, interval, k8sClient)

			TemplateLookupKey := types.NamespacedName{Name: nonPersistentTemplateName, Namespace: WorkingNamespace}
			currentTemplate = &crownlabsv1alpha2.Template{}
			doesEventuallyExists(ctx, TemplateLookupKey, currentTemplate, BeTrue(), timeout, interval, k8sClient)

			tenantLookupKey := types.NamespacedName{Name: TenantName, Namespace: tenantNs.Name}
			currentTenant = &crownlabsv1alpha2.Tenant{}
			doesEventuallyExists(ctx, tenantLookupKey, currentTenant, BeTrue(), timeout, interval, k8sClient)

			ctx, _ = pkgcontext.InstanceInto(ctx, currentInstance)
			ctx, _ = pkgcontext.TemplateInto(ctx, currentTemplate)
			ctx, _ = pkgcontext.TenantInto(ctx, currentTenant)

			inactivityTimeoutDuration = time.Hour * 24 * 14
			err := r.SetupInstanceAnnotations(ctx)
			Expect(err).ToNot(HaveOccurred(), "SetupInstanceAnnotations should not return an error")
		})

		It("should return remaining time if instance is still active", func() {
			lastLogin := time.Now().Add(-10 * time.Minute).Format(time.RFC3339)
			currentInstance.Annotations[forge.LastActivityAnnotation] = lastLogin

			remaining, err := r.GetRemainingInactivityTime(ctx, inactivityTimeoutDuration)
			Expect(err).ToNot(HaveOccurred())
			Expect(remaining.Seconds()).To(BeNumerically(">", 0))
		})

		It("should return <=0 if inactivity timeout has been exceeded", func() {
			lastLogin := time.Now().Add(-1000 * time.Hour).Format(time.RFC3339)
			currentInstance.Annotations[forge.LastActivityAnnotation] = lastLogin

			remaining, err := r.GetRemainingInactivityTime(ctx, inactivityTimeoutDuration)
			Expect(err).ToNot(HaveOccurred())
			Expect(remaining).To(BeNumerically("<=", 0))
		})

		It("should return error if annotation is missing", func() {
			delete(currentInstance.Annotations, forge.LastActivityAnnotation)

			_, err := r.GetRemainingInactivityTime(ctx, inactivityTimeoutDuration)
			Expect(err).To(HaveOccurred())
		})

		It("should return error if annotation is not parseable", func() {
			currentInstance.Annotations[forge.LastActivityAnnotation] = "not-a-valid-time"

			_, err := r.GetRemainingInactivityTime(ctx, inactivityTimeoutDuration)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Testing IncrementAnnotation function", func() {
		var (
			reconciler *instautoctrl.InstanceInactiveTerminationReconciler
			ctx        context.Context
		)

		JustBeforeEach(func() {
			reconciler = &instautoctrl.InstanceInactiveTerminationReconciler{}
			ctx = context.Background()
		})

		It("should increment a valid numeric annotation string", func() {
			input := "3"
			output, err := reconciler.IncrementAnnotation(ctx, input)

			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal("4"))
		})

		It("should return error for invalid numeric input", func() {
			input := "abc"
			output, err := reconciler.IncrementAnnotation(ctx, input)

			Expect(err).To(HaveOccurred())
			Expect(output).To(Equal("0"))
		})
	})

	Describe("Testing SetupInstanceAnnotations function", func() {
		var (
			r               *instautoctrl.InstanceInactiveTerminationReconciler
			ctx             context.Context
			currentInstance *crownlabsv1alpha2.Instance
			currentTemplate *crownlabsv1alpha2.Template
			currentTenant   *crownlabsv1alpha2.Tenant
		)

		JustBeforeEach(func() {

			ctx = context.Background()

			r = &instautoctrl.InstanceInactiveTerminationReconciler{
				Client:             k8sClient,
				NamespaceWhitelist: metav1.LabelSelector{MatchLabels: whiteListMap},
				Prometheus:         mockProm,
			}

			By("Checking that the instance is running")
			InstanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenantNs.Name}
			currentInstance = &crownlabsv1alpha2.Instance{}
			doesEventuallyExists(ctx, InstanceLookupKey, currentInstance, BeTrue(), timeout, interval, k8sClient)

			TemplateLookupKey := types.NamespacedName{Name: nonPersistentTemplateName, Namespace: WorkingNamespace}
			currentTemplate = &crownlabsv1alpha2.Template{}
			doesEventuallyExists(ctx, TemplateLookupKey, currentTemplate, BeTrue(), timeout, interval, k8sClient)

			tenantLookupKey := types.NamespacedName{Name: TenantName, Namespace: tenantNs.Name}
			currentTenant = &crownlabsv1alpha2.Tenant{}
			doesEventuallyExists(ctx, tenantLookupKey, currentTenant, BeTrue(), timeout, interval, k8sClient)

			ctx, _ = pkgcontext.InstanceInto(ctx, currentInstance)
			ctx, _ = pkgcontext.TemplateInto(ctx, currentTemplate)
			ctx, _ = pkgcontext.TenantInto(ctx, currentTenant)
		})

		It("should add all missing annotations and patch", func() {
			err := r.SetupInstanceAnnotations(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(currentInstance.Annotations).To(HaveKeyWithValue(forge.AlertAnnotationNum, "0"))
			Expect(currentInstance.Annotations).To(HaveKey(forge.LastActivityAnnotation))
			Expect(currentInstance.Annotations).To(HaveKey(forge.LastNotificationTimestampAnnotation))
		})
	})

	Describe("Testing shouldSendWarningNotification function", func() {
		var (
			reconciler *instautoctrl.InstanceInactiveTerminationReconciler
			instance   *crownlabsv1alpha2.Instance
			ctx        context.Context
			template   *crownlabsv1alpha2.Template
		)

		BeforeEach(func() {
			reconciler = &instautoctrl.InstanceInactiveTerminationReconciler{
				EnableInactivityNotifications: true,
				NotificationInterval:          time.Hour,
				InstanceMaxNumberOfAlerts:     3,
			}
			instance = &crownlabsv1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-instance",
					Annotations: map[string]string{},
				},
				Spec: crownlabsv1alpha2.InstanceSpec{
					Running: true,
				},
			}
			template = &crownlabsv1alpha2.Template{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						forge.CustomNumberOfAlertsAnnotation: "4",
					},
				},
			}
			ctx = context.Background()
			ctx, _ = pkgcontext.TemplateInto(ctx, template)
		})

		It("returns false if inactivity notifications are disabled", func() {
			reconciler.EnableInactivityNotifications = false
			send, err := reconciler.ShouldSendWarningNotification(ctx, instance)
			Expect(send).To(BeFalse())
			Expect(err).To(BeNil())
		})

		It("returns error if AlertAnnotationNum is invalid", func() {
			instance.Annotations[forge.AlertAnnotationNum] = "invalid"
			send, err := reconciler.ShouldSendWarningNotification(ctx, instance)
			Expect(send).To(BeFalse())
			Expect(err).To(HaveOccurred())
		})

		It("returns true if last notification timestamp is missing", func() {
			instance.Annotations[forge.AlertAnnotationNum] = "0"
			send, err := reconciler.ShouldSendWarningNotification(ctx, instance)
			Expect(send).To(BeTrue())
			Expect(err).To(BeNil())
		})

		It("returns false if last notification is within interval", func() {
			instance.Annotations[forge.AlertAnnotationNum] = "1"
			instance.Annotations[forge.LastNotificationTimestampAnnotation] = time.Now().Add(-30 * time.Minute).Format(time.RFC3339)

			send, err := reconciler.ShouldSendWarningNotification(ctx, instance)
			Expect(send).To(BeFalse())
			Expect(err).To(BeNil())
		})

		It("returns false if instance is not running", func() {
			instance.Spec.Running = false
			instance.Annotations[forge.AlertAnnotationNum] = "0"
			send, err := reconciler.ShouldSendWarningNotification(ctx, instance)
			Expect(send).To(BeFalse())
			Expect(err).To(BeNil())
		})

		It("respects custom max alerts annotation", func() {
			instance.Annotations[forge.AlertAnnotationNum] = "2"
			send, err := reconciler.ShouldSendWarningNotification(ctx, instance)
			Expect(send).To(BeTrue())
			Expect(err).To(BeNil())
		})

		It("returns false if numAlerts exceeds default max", func() {
			instance.Annotations[forge.AlertAnnotationNum] = "5"
			instance.Annotations[forge.LastNotificationTimestampAnnotation] = time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
			send, err := reconciler.ShouldSendWarningNotification(ctx, instance)
			Expect(send).To(BeFalse())
			Expect(err).To(BeNil())
		})
	})

	Describe("Testing ResetAlertAnnotation function", func() {
		var (
			r               *instautoctrl.InstanceInactiveTerminationReconciler
			ctx             context.Context
			currentInstance *crownlabsv1alpha2.Instance
			currentTemplate *crownlabsv1alpha2.Template
			currentTenant   *crownlabsv1alpha2.Tenant
		)
		JustBeforeEach(func() {
			ctx = context.Background()

			r = &instautoctrl.InstanceInactiveTerminationReconciler{
				Client:             k8sClient,
				NamespaceWhitelist: metav1.LabelSelector{MatchLabels: whiteListMap},
				Prometheus:         mockProm,
			}

			By("Checking that the instance is running")
			InstanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenantNs.Name}
			currentInstance = &crownlabsv1alpha2.Instance{}
			doesEventuallyExists(ctx, InstanceLookupKey, currentInstance, BeTrue(), timeout, interval, k8sClient)

			TemplateLookupKey := types.NamespacedName{Name: nonPersistentTemplateName, Namespace: WorkingNamespace}
			currentTemplate = &crownlabsv1alpha2.Template{}
			doesEventuallyExists(ctx, TemplateLookupKey, currentTemplate, BeTrue(), timeout, interval, k8sClient)

			tenantLookupKey := types.NamespacedName{Name: TenantName, Namespace: tenantNs.Name}
			currentTenant = &crownlabsv1alpha2.Tenant{}
			doesEventuallyExists(ctx, tenantLookupKey, currentTenant, BeTrue(), timeout, interval, k8sClient)

			ctx, _ = pkgcontext.InstanceInto(ctx, currentInstance)
			ctx, _ = pkgcontext.TemplateInto(ctx, currentTemplate)
			ctx, _ = pkgcontext.TenantInto(ctx, currentTenant)
			err := r.SetupInstanceAnnotations(ctx)
			Expect(err).ToNot(HaveOccurred(), "SetupInstanceAnnotations should not return an error")
		})

		It("should reset the AlertAnnotationNum to 0", func() {
			err := r.ResetAnnotations(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(currentInstance.Annotations[forge.AlertAnnotationNum]).To(Equal("0"))
		})

	})

})
