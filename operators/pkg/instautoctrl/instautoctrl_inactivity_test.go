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
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instautoctrl"
)

var _ = Describe("Instautoctrl", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		PersistentInstanceName               = "test-instance-persistent"
		NonPersistentInstanceName            = "test-instance-non-persistent"
		WorkingNamespace                     = "working-namespace"
		persistentTemplateName               = "test-template-persistent"
		nonPersistentTemplateName            = "test-template-non-persistent"
		TenantName                           = "test-tenant"
		CustomDeleteAfter                    = instautoctrl.NEVER_TIMEOUT_VALUE
		CustomInactivityTimeout              = instautoctrl.NEVER_TIMEOUT_VALUE
		CustomDeleteAfterNonPersistent       = "1m"
		CustomInactivityTimeoutNonPersistent = "10m"

		timeout  = time.Second * 80
		interval = time.Millisecond * 500
	)

	var (
		workingNs = v1.Namespace{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: WorkingNamespace,
				Labels: map[string]string{
					"test-suite": "true",
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
					Name:       "Env-1",
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
					Name:       "Env-1",
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
		newNs := workingNs.DeepCopy()
		newPersistentTemplate := persistentTemplate.DeepCopy()
		newNonPersistentTemplate := nonPersistentTemplate.DeepCopy()
		newPersistentInstance := persistentInstance.DeepCopy()
		newNonPersistentInstance := nonPersistentInstance.DeepCopy()
		newTenant := tenant.DeepCopy()
		By("Creating the namespace where to create instance and template")
		err := k8sClient.Create(ctx, newNs)
		if err != nil && errors.IsAlreadyExists(err) {
			By("Cleaning up the environment")
			By("Deleting templates")
			Expect(k8sClient.Delete(ctx, &persistentTemplate)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, &nonPersistentTemplate)).Should(Succeed())
			By("Deleting instances")
			Expect(client.IgnoreNotFound(k8sClientExpiration.Delete(ctx, &persistentInstance))).To(Succeed())
			Expect(client.IgnoreNotFound(k8sClientExpiration.Delete(ctx, &nonPersistentInstance))).To(Succeed())
			By("Deleting tenant")
			Expect(k8sClientExpiration.Delete(ctx, &tenant)).Should(Succeed())
		} else if err != nil {
			Fail(fmt.Sprintf("Unable to create namespace -> %s", err))
		}

		By("By checking that the namespace has been created")
		createdNs := &v1.Namespace{}

		nsLookupKey := types.NamespacedName{Name: WorkingNamespace}
		doesEventuallyExists(ctx, nsLookupKey, createdNs, BeTrue(), timeout, interval)

		By("Creating the templates")
		Expect(k8sClient.Create(ctx, newPersistentTemplate)).Should(Succeed())
		Expect(k8sClient.Create(ctx, newNonPersistentTemplate)).Should(Succeed())

		By("By checking that the template has been created")
		persistentTemplateLookupKey := types.NamespacedName{Name: persistentTemplateName, Namespace: WorkingNamespace}
		nonPersistentTemplateLookupKey := types.NamespacedName{Name: nonPersistentTemplateName, Namespace: WorkingNamespace}
		createdPersitentTemplate := &crownlabsv1alpha2.Template{}
		createdNonPersitentTemplate := &crownlabsv1alpha2.Template{}

		doesEventuallyExists(ctx, persistentTemplateLookupKey, createdPersitentTemplate, BeTrue(), timeout, interval)
		doesEventuallyExists(ctx, nonPersistentTemplateLookupKey, createdNonPersitentTemplate, BeTrue(), timeout, interval)

		By("Creating the tenant")
		Expect(k8sClient.Create(ctx, newTenant)).Should(Succeed())

		By("Checking that the tenant has been created")
		tenantLookupKey := types.NamespacedName{Name: TenantName, Namespace: WorkingNamespace}
		createdTenant := &crownlabsv1alpha2.Tenant{}
		doesEventuallyExists(ctx, tenantLookupKey, createdTenant, BeTrue(), timeout, interval)

		By("Creating the instances")
		Expect(k8sClient.Create(ctx, newPersistentInstance)).Should(Succeed())
		Expect(k8sClient.Create(ctx, newNonPersistentInstance)).Should(Succeed())

		By("Checking that the instances has been created")
		persistanteInstanceLookupKey := types.NamespacedName{Name: PersistentInstanceName, Namespace: WorkingNamespace}
		nonPersistentInstanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: WorkingNamespace}
		createdPersistentInstance := &crownlabsv1alpha2.Instance{}
		createdNonPersistentInstance := &crownlabsv1alpha2.Instance{}

		doesEventuallyExists(ctx, persistanteInstanceLookupKey, createdPersistentInstance, BeTrue(), timeout, interval)
		doesEventuallyExists(ctx, nonPersistentInstanceLookupKey, createdNonPersistentInstance, BeTrue(), timeout, interval)

	})

	Context("Testing value", func() {

		It("Should succeed: the Persistent instance is not stopped because it is set with the default InactivityTimeout value", func() {
			mockProm.EXPECT().
				IsPrometheusHealthy(gomock.Any()).
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

			By("Getting current instance")
			currentInstance := &crownlabsv1alpha2.Instance{}
			instanceLookupKey := types.NamespacedName{Name: PersistentInstanceName, Namespace: WorkingNamespace}
			Expect(k8sClient.Get(ctx, instanceLookupKey, currentInstance)).Should(Succeed())

			By("Getting current templates")
			currentTemplate := &crownlabsv1alpha2.Template{}

			templateLookupKey := types.NamespacedName{Name: currentInstance.Spec.Template.Name, Namespace: WorkingNamespace}
			Expect(k8sClient.Get(ctx, templateLookupKey, currentTemplate)).Should(Succeed())

			By("Checking the InactivityTimeout field is the default one")
			currentInactivityTimeout := currentTemplate.Spec.InactivityTimeout
			defaultInactivityTimeout := instautoctrl.NEVER_TIMEOUT_VALUE
			Expect(currentInactivityTimeout).To(Equal(defaultInactivityTimeout))
		})

		It("Should succeed: the persistent instance get the annotations", func() {
			mockProm.EXPECT().
				IsPrometheusHealthy(gomock.Any()).
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

			By("Getting current instance")
			currentInstance := &crownlabsv1alpha2.Instance{}
			instanceLookupKey := types.NamespacedName{Name: PersistentInstanceName, Namespace: WorkingNamespace}
			Expect(k8sClient.Get(ctx, instanceLookupKey, currentInstance)).Should(Succeed())

			By("Checking the instance got the annotations defined  for the inactivity controller")
			//Eventually(currentInstance.Annotations, timeout, interval).Should(HaveKey(forge.AlertAnnotationNum))
			Eventually(currentInstance.Annotations, timeout/4, interval).Should(HaveKey(forge.LastActivityAnnotation))

		})
	})

	Context("Testing maximum inactivity time", func() {
		It("Should succeed: the non-persistent instance is inactive because it has reached the maximum inactivity time", func() {
			mockProm.EXPECT().
				IsPrometheusHealthy(gomock.Any()).
				Return(true, nil).
				AnyTimes()

			mockProm.EXPECT().
				GetLastActivityTime(gomock.Any(), gomock.Any()).
				Return(time.Now().Add(-48*time.Hour), nil).
				AnyTimes()

			mockProm.EXPECT().
				GetQueryNginxData().
				Return("").
				AnyTimes()

			mockProm.EXPECT().
				GetQuerySSHData().
				Return("").
				AnyTimes()

			By("Getting current instance")
			currentInstance := &crownlabsv1alpha2.Instance{}
			instanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: WorkingNamespace}
			Expect(k8sClient.Get(ctx, instanceLookupKey, currentInstance)).Should(Succeed())

			By("Checking the instance is inactive")
			Expect(currentInstance.Spec.Running).To(BeFalse())
		})

		// It("Should fail: the persistent VM is inactive for a long time and it is stopped", func() {
		// 	By("Getting current instance")
		// 	currentInstance := &crownlabsv1alpha2.Instance{}
		// 	instanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: WorkingNamespace}
		// 	Expect(k8sClient.Get(ctx, instanceLookupKey, currentInstance)).Should(Succeed())

		// })

		// It("Should fail: the non-persistent VM is inactive for a long time it is deleted", func() {

		// })

	})

})
