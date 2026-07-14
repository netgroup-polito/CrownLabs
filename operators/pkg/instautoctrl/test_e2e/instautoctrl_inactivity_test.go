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
	"fmt"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instautoctrl"
)

var _ = Describe("Instautoctrl-inactivity", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		PersistentInstanceName                 = "test-inactivity-instance-persistent"
		PersistentInstanceName2                = "test-inactivity-instance-persistent2"
		NonPersistentInstanceName              = "test-inactivity-instance-non-persistent"
		WorkingNamespace                       = "test-inactivity-working-namespace"
		persistentTemplateName                 = "test-inactivity-test-template-persistent"
		persistentTemplateName2                = "test-inactivity-test-template-persistent-2"
		nonPersistentTemplateName              = "test-inactivity-template-non-persistent"
		TenantName                             = "test-inactivity-tenant"
		CustomDeleteAfter                      = instautoctrl.NeverTimeoutValue
		CustomstopAfterInactivity              = instautoctrl.NeverTimeoutValue
		CustomDeleteAfterNonPersistent         = instautoctrl.NeverTimeoutValue
		CustomstopAfterInactivityNonPersistent = "0m"
		CustomDeleteAfterPersistent2           = instautoctrl.NeverTimeoutValue
		CustomstopAfterInactivityPersistent2   = "2m"

		timeout  = time.Second * 60
		interval = time.Millisecond * 500
	)

	var (
		workingNs = corev1.Namespace{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: WorkingNamespace,
				Labels: map[string]string{
					"crownlabs.polito.it/operator-selector": "test-suite",
				},
			},
			Spec:   corev1.NamespaceSpec{},
			Status: corev1.NamespaceStatus{},
		}
		tenantNs = corev1.Namespace{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: TenantName,
				Labels: map[string]string{
					"crownlabs.polito.it/operator-selector": "test-suite",
					"crownlabs.polito.it/tenant":            TenantName,
				},
			},
			Spec:   corev1.NamespaceSpec{},
			Status: corev1.NamespaceStatus{},
		}
		templatePersistentEnvironment = clv1alpha2.TemplateSpec{
			WorkspaceRef: clv1alpha2.GenericRef{},
			PrettyName:   "My Template",
			Description:  "Description of my template",
			EnvironmentList: []clv1alpha2.Environment{
				{
					Name:       "env-1",
					GuiEnabled: true,
					Resources: clv1alpha2.EnvironmentResources{
						CPU:                   1,
						ReservedCPUPercentage: 1,
						Memory:                resource.MustParse("1024M"),
					},
					EnvironmentType: clv1alpha2.ClassVM,
					Persistent:      true,
					Image:           "crownlabs/vm",
				},
			},
			Cleanup: clv1alpha2.CleanupOptions{
				DeleteAfterCreation:   "never",
				StopAfterInactivity:   "never",
				DeleteAfterInactivity: "never",
			},
		}
		templateNonPersistentEnvironment = clv1alpha2.TemplateSpec{
			WorkspaceRef: clv1alpha2.GenericRef{},
			PrettyName:   "My Template",
			Description:  "Description of my template",
			EnvironmentList: []clv1alpha2.Environment{
				{
					Name:       "env-1",
					GuiEnabled: true,
					Resources: clv1alpha2.EnvironmentResources{
						CPU:                   1,
						ReservedCPUPercentage: 1,
						Memory:                resource.MustParse("1024M"),
					},
					EnvironmentType: clv1alpha2.ClassVM,
					Persistent:      false,
					Image:           "crownlabs/vm",
				},
			},
			Cleanup: clv1alpha2.CleanupOptions{
				DeleteAfterCreation:   CustomDeleteAfterNonPersistent,
				StopAfterInactivity:   CustomstopAfterInactivityNonPersistent,
				DeleteAfterInactivity: "never",
			},
		}
		templatePersistentEnvironmentWithCustomstopAfterInactivity = clv1alpha2.TemplateSpec{
			WorkspaceRef: clv1alpha2.GenericRef{},
			PrettyName:   "My Template",
			Description:  "Description of my template",
			EnvironmentList: []clv1alpha2.Environment{
				{
					Name:       "env-1",
					GuiEnabled: true,
					Resources: clv1alpha2.EnvironmentResources{
						CPU:                   1,
						ReservedCPUPercentage: 1,
						Memory:                resource.MustParse("1024M"),
					},
					EnvironmentType: clv1alpha2.ClassVM,
					Persistent:      false,
					Image:           "crownlabs/vm",
				},
			},
			Cleanup: clv1alpha2.CleanupOptions{
				DeleteAfterCreation:   CustomDeleteAfterPersistent2,
				StopAfterInactivity:   CustomstopAfterInactivityPersistent2,
				DeleteAfterInactivity: "never",
			},
		}
		persistentTemplate2 = clv1alpha2.Template{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      persistentTemplateName2,
				Namespace: WorkingNamespace,
			},
			Spec:   templatePersistentEnvironmentWithCustomstopAfterInactivity,
			Status: clv1alpha2.TemplateStatus{},
		}
		persistentTemplate = clv1alpha2.Template{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      persistentTemplateName,
				Namespace: WorkingNamespace,
			},
			Spec:   templatePersistentEnvironment,
			Status: clv1alpha2.TemplateStatus{},
		}
		nonPersistentTemplate = clv1alpha2.Template{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      nonPersistentTemplateName,
				Namespace: WorkingNamespace,
			},
			Spec:   templateNonPersistentEnvironment,
			Status: clv1alpha2.TemplateStatus{},
		}
		persistentInstance2 = clv1alpha2.Instance{
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
			Spec: clv1alpha2.InstanceSpec{
				Running: true,
				Template: clv1alpha2.GenericRef{
					Name:      persistentTemplateName2,
					Namespace: WorkingNamespace,
				},
				Tenant: clv1alpha2.GenericRef{
					Name:      TenantName,
					Namespace: tenantNs.Name,
				},
			},
			Status: clv1alpha2.InstanceStatus{},
		}

		persistentInstance = clv1alpha2.Instance{
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
			Spec: clv1alpha2.InstanceSpec{
				Running: true,
				Template: clv1alpha2.GenericRef{
					Name:      persistentTemplateName,
					Namespace: WorkingNamespace,
				},
				Tenant: clv1alpha2.GenericRef{
					Name:      TenantName,
					Namespace: tenantNs.Name,
				},
			},
			Status: clv1alpha2.InstanceStatus{},
		}
		nonPersistentInstance = clv1alpha2.Instance{
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
			Spec: clv1alpha2.InstanceSpec{
				Running: true,
				Template: clv1alpha2.GenericRef{
					Name:      nonPersistentTemplateName,
					Namespace: WorkingNamespace,
				},
				Tenant: clv1alpha2.GenericRef{
					Name:      TenantName,
					Namespace: tenantNs.Name,
				},
			},
			Status: clv1alpha2.InstanceStatus{},
		}

		tenant = clv1alpha2.Tenant{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      TenantName,
				Namespace: TenantName,
				Labels: map[string]string{
					"crownlabs.polito.it/operator-selector": "test-suite",
				},
			},
			Spec: clv1alpha2.TenantSpec{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@gmail.com",
				Workspaces: []clv1alpha2.TenantWorkspaceEntry{
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
		if (err1 != nil || err2 != nil) && (kerrors.IsAlreadyExists(err1) || kerrors.IsAlreadyExists(err2)) {
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
		createdPersitentTemplate := &clv1alpha2.Template{}
		createdNonPersitentTemplate := &clv1alpha2.Template{}
		createdPersistentTemplate2 := &clv1alpha2.Template{}

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

		It("Should succeed: the Persistent instance get the default stopAfterInactivity value and it is not stopped", func() {
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
			currentInstance := &clv1alpha2.Instance{}
			instanceLookupKey := types.NamespacedName{Name: PersistentInstanceName, Namespace: tenant.Namespace}
			doesEventuallyExists(ctx, instanceLookupKey, currentInstance, BeTrue(), timeout, interval, k8sClient)

			By("Getting current templates")
			currentTemplate := &clv1alpha2.Template{}

			templateLookupKey := types.NamespacedName{Name: currentInstance.Spec.Template.Name, Namespace: WorkingNamespace}
			doesEventuallyExists(ctx, templateLookupKey, currentTemplate, BeTrue(), timeout, interval, k8sClient)

			By("Checking the stopAfterInactivity field is the default one")
			currentstopAfterInactivity := currentTemplate.Spec.Cleanup.StopAfterInactivity
			defaultstopAfterInactivity := instautoctrl.NeverTimeoutValue
			Expect(currentstopAfterInactivity).To(Equal(defaultstopAfterInactivity))
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
			currentInstance := &clv1alpha2.Instance{}
			instanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenant.Namespace}
			doesEventuallyExists(ctx, instanceLookupKey, currentInstance, BeTrue(), timeout, interval, k8sClient)

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
			currentInstance := &clv1alpha2.Instance{}
			instanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenant.Namespace}
			doesEventuallyExists(ctx, instanceLookupKey, currentInstance, BeTrue(), timeout, interval, k8sClient)

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

	Context("Testing destruction after inactivity", func() {
		It("Should delete the persistent instance if destroy timer is exceeded", func() {
			By("Updating template with deleteAfterInactivity")
			currentTemplate := &clv1alpha2.Template{}
			templateLookupKey := types.NamespacedName{Name: persistentTemplateName2, Namespace: WorkingNamespace}
			Eventually(func() error {
				if err := k8sClient.Get(ctx, templateLookupKey, currentTemplate); err != nil {
					return err
				}
				currentTemplate.Spec.Cleanup.DeleteAfterInactivity = "100h"
				return k8sClient.Update(ctx, currentTemplate)
			}, timeout, interval).Should(Succeed())

			By("Waiting for the updated template to be observed by the manager cache")
			Eventually(func() string {
				observedTemplate := &clv1alpha2.Template{}
				if err := k8sClient.Get(ctx, templateLookupKey, observedTemplate); err != nil {
					return ""
				}
				return observedTemplate.Spec.Cleanup.DeleteAfterInactivity
			}, timeout, interval).Should(Equal("100h"))

			By("Setting instance as powered off with an expired destruction timestamp")
			currentInstance := &clv1alpha2.Instance{}
			instanceLookupKey := types.NamespacedName{Name: PersistentInstanceName2, Namespace: tenant.Namespace}
			Eventually(func() error {
				if err := k8sClient.Get(ctx, instanceLookupKey, currentInstance); err != nil {
					return err
				}
				currentInstance.Spec.Running = false
				if currentInstance.Annotations == nil {
					currentInstance.Annotations = make(map[string]string)
				}
				currentInstance.Annotations[forge.LastPoweredOffTimestampAnnotation] = time.Now().Add(-150 * time.Hour).Format(time.RFC3339)
				return k8sClient.Update(ctx, currentInstance)
			}, timeout, interval).Should(Succeed())

			By("Checking the instance is deleted")
			doesEventuallyExists(ctx, instanceLookupKey, currentInstance, BeFalse(), timeout, interval, k8sClient)
		})

		It("Should not delete the persistent instance if destroy timer is NOT exceeded", func() {
			By("Getting current instance")
			currentInstance := &clv1alpha2.Instance{}
			instanceLookupKey := types.NamespacedName{Name: PersistentInstanceName, Namespace: tenant.Namespace}
			doesEventuallyExists(ctx, instanceLookupKey, currentInstance, BeTrue(), timeout, interval, k8sClient)

			By("Setting instance as powered off and recent timestamp")
			Eventually(func() error {
				if err := k8sClient.Get(ctx, instanceLookupKey, currentInstance); err != nil {
					return err
				}
				currentInstance.Spec.Running = false
				if currentInstance.Annotations == nil {
					currentInstance.Annotations = make(map[string]string)
				}
				currentInstance.Annotations[forge.LastPoweredOffTimestampAnnotation] = time.Now().Add(-50 * time.Hour).Format(time.RFC3339)
				return k8sClient.Update(ctx, currentInstance)
			}, timeout, interval).Should(Succeed())

			By("Updating template with deleteAfterInactivity")
			currentTemplate := &clv1alpha2.Template{}
			templateLookupKey := types.NamespacedName{Name: persistentTemplateName, Namespace: WorkingNamespace}
			Eventually(func() error {
				if err := k8sClient.Get(ctx, templateLookupKey, currentTemplate); err != nil {
					return err
				}
				currentTemplate.Spec.Cleanup.DeleteAfterInactivity = "100h"
				return k8sClient.Update(ctx, currentTemplate)
			}, timeout, interval).Should(Succeed())

			By("Checking the instance is NOT deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, instanceLookupKey, currentInstance)
				if err != nil {
					return false
				}
				return !currentInstance.Spec.Running
			}, time.Second*5, interval).Should(BeTrue(), "The instance should not be deleted")
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
			currentInstance := &clv1alpha2.Instance{}
			instanceLookupKey := types.NamespacedName{Name: PersistentInstanceName, Namespace: tenant.Namespace}
			doesEventuallyExists(ctx, instanceLookupKey, currentInstance, BeTrue(), timeout, interval, k8sClient)

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
			currentInstance := &clv1alpha2.Instance{}
			instanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenant.Namespace}
			doesEventuallyExists(ctx, instanceLookupKey, currentInstance, BeTrue(), timeout, interval, k8sClient)

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
