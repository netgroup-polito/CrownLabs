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

var _ = Describe("Instautoctrl-inactivity", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		PersistentInstanceName               = "test-inactivity-instance-persistent"
		PersistentInstanceName2              = "test-inactivity-instance-persistent2"
		NonPersistentInstanceName            = "test-inactivity-instance-non-persistent"
		WorkingNamespace                     = "test-inactivity-working-namespace"
		persistentTemplateName               = "test-inactivity-test-template-persistent"
		persistentTemplateName2              = "test-inactivity-test-template-persistent-2"
		nonPersistentTemplateName            = "test-inactivity-template-non-persistent"
		TenantName                           = "test-inactivity-tenant"
		CustomDeleteAfter                    = instautoctrl.NeverTimeoutValue
		CustomInactivityTimeout              = instautoctrl.NeverTimeoutValue
		CustomDeleteAfterNonPersistent       = instautoctrl.NeverTimeoutValue
		CustomInactivityTimeoutNonPersistent = "0m"
		CustomDeleteAfterPersistent2         = instautoctrl.NeverTimeoutValue
		CustomInactivityTimeoutPersistent2   = "2m"

		timeout  = time.Second * 60
		interval = time.Millisecond * 500
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
					"crownlabs.polito.it/instance-type":     "non-persistent",
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
		} else if err1 != nil || err2 != nil {
			Fail(fmt.Sprintf("Unable to create namespace -> %s %s", err1, err2))
		}
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

		By("Creating the instances")
		Expect(k8sClient.Create(ctx, newPersistentInstance)).Should(Succeed())
		Expect(k8sClient.Create(ctx, newNonPersistentInstance)).Should(Succeed())
		Expect(k8sClient.Create(ctx, newPersistentInstance2)).Should(Succeed())
	})

	Context("Testing default and custom inactivity value", func() {

		It("Should succeed: the Persistent instance get the default InactivityTimeout value and it is not stopped", func() {
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

			By("Getting current instance")
			currentInstance := &crownlabsv1alpha2.Instance{}
			instanceLookupKey := types.NamespacedName{Name: PersistentInstanceName, Namespace: tenant.Namespace}
			Expect(k8sClient.Get(ctx, instanceLookupKey, currentInstance)).Should(Succeed())

			By("Getting current templates")
			currentTemplate := &crownlabsv1alpha2.Template{}

			templateLookupKey := types.NamespacedName{Name: currentInstance.Spec.Template.Name, Namespace: WorkingNamespace}
			Expect(k8sClient.Get(ctx, templateLookupKey, currentTemplate)).Should(Succeed())

			By("Checking the InactivityTimeout field is the default one")
			currentInactivityTimeout := currentTemplate.Spec.InactivityTimeout
			defaultInactivityTimeout := instautoctrl.NeverTimeoutValue
			Expect(currentInactivityTimeout).To(Equal(defaultInactivityTimeout))
			Eventually(func() bool {
				err := k8sClient.Get(ctx, instanceLookupKey, currentInstance)
				if err != nil {
					return false
				}
				return currentInstance.Spec.Running
			}, timeout, interval).Should(BeTrue(), "The instance should be running")
		})

		It("The non-persistent VM is active and should not be deleted", func() {
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

			By("Getting current instance")
			currentInstance := &crownlabsv1alpha2.Instance{}
			instanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenant.Namespace}
			Expect(k8sClient.Get(ctx, instanceLookupKey, currentInstance)).Should(Succeed())

			By("Checking the instance is still running")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, instanceLookupKey, currentInstance)
				if err != nil {
					return false
				}
				return currentInstance.Spec.Running
			}, timeout, interval).Should(BeTrue(), "The instance should be running")
		})
		It("The non-persistent VM is inactive for a long time and it is deleted", func() {

			mockProm.EXPECT().
				IsPrometheusHealthy(gomock.Any(), gomock.Any()).
				Return(true, nil).
				AnyTimes()

			mockProm.EXPECT().
				GetLastActivityTime(gomock.Any(), gomock.Any()).
				Return(time.Now().Add(-1000*time.Hour), nil).
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

			By("Getting current instance")
			currentInstance := &crownlabsv1alpha2.Instance{}
			instanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenant.Namespace}
			Expect(k8sClient.Get(ctx, instanceLookupKey, currentInstance)).Should(Succeed())

			By("Checking the instance is deleted")
			doesEventuallyExists(ctx, instanceLookupKey, currentInstance, BeFalse(), timeout, interval, k8sClient)
		})

		It("The persistent VM is active and is not stopped", func() {
			mockProm.EXPECT().
				IsPrometheusHealthy(gomock.Any(), gomock.Any()).
				Return(true, nil).
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
				GetLastActivityTime(gomock.Any(), gomock.Any()).
				Return(time.Now(), nil).
				AnyTimes()

			mockProm.EXPECT().
				GetQueryWebSSHData().
				Return("").
				AnyTimes()

		})

	})

	Context("Testing errors", func() {
		It("Should fail: prometheus is not healthy", func() {
			mockProm.EXPECT().
				IsPrometheusHealthy(gomock.Any(), gomock.Any()).
				Return(false, nil).
				AnyTimes()

			mockProm.EXPECT().
				GetQuerySSHData().
				Return("").
				AnyTimes()

			mockProm.EXPECT().
				GetQueryNginxData().
				Return("").
				AnyTimes()

			mockProm.EXPECT().
				GetLastActivityTime(gomock.Any(), gomock.Any()).
				Return(time.Now().Add(-100*time.Hour), nil).
				AnyTimes()

			mockProm.EXPECT().
				GetQueryWebSSHData().
				Return("").
				AnyTimes()

			By("Getting current instance")
			currentInstance := &crownlabsv1alpha2.Instance{}
			instanceLookupKey := types.NamespacedName{Name: PersistentInstanceName, Namespace: tenant.Namespace}
			Expect(k8sClient.Get(ctx, instanceLookupKey, currentInstance)).Should(Succeed())

			By("Checking the instance is still running")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, instanceLookupKey, currentInstance)
				if err != nil {
					return false
				}
				return currentInstance.Spec.Running
			}, timeout, interval).Should(BeTrue(), "The instance should be running")
		})
		It("Should fail: activity time not correctly returned, the instance should be running", func() {
			mockProm.EXPECT().
				IsPrometheusHealthy(gomock.Any(), gomock.Any()).
				Return(true, nil).
				AnyTimes()

			mockProm.EXPECT().
				GetLastActivityTime(gomock.Any(), gomock.Any()).
				Return(time.Now(), fmt.Errorf("")).
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

			By("Getting current instance")
			currentInstance := &crownlabsv1alpha2.Instance{}
			instanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenant.Namespace}
			Expect(k8sClient.Get(ctx, instanceLookupKey, currentInstance)).Should(Succeed())

			By("Checking the instance is still running")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, instanceLookupKey, currentInstance)
				if err != nil {
					return false
				}
				return currentInstance.Spec.Running
			}, timeout, interval).Should(BeTrue(), "The instance should be running")
		})

	})

})
