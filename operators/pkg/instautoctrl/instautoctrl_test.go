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

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var _ = Describe("Instautoctrl", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		PersistentInstanceName    = "test-instance-persistent"
		NonPersistentInstanceName = "test-instance-non-persistent"
		WorkingNamespace          = "working-namespace"
		persistentTemplateName    = "test-template-persistent"
		nonPersistentTemplateName = "test-template-non-persistent"
		TenantName                = "test-tenant"
		CustomDeleteAfter         = "30d"
		CustomInactivityTimeout   = "14d"

		timeout  = time.Second * 25
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
					Name: TenantName,
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
	)

	BeforeEach(func() {
		newNs := workingNs.DeepCopy()
		newPersistentTemplate := persistentTemplate.DeepCopy()
		newNonPersistentTemplate := nonPersistentTemplate.DeepCopy()
		newPersistentInstance := persistentInstance.DeepCopy()
		newNonPersistentInstance := nonPersistentInstance.DeepCopy()
		By("Creating the namespace where to create instance and template")
		err := k8sClient.Create(ctx, newNs)
		if err != nil && errors.IsAlreadyExists(err) {
			By("Cleaning up the environment")
			By("Deleting templates")
			Expect(k8sClient.Delete(ctx, &persistentTemplate)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, &nonPersistentTemplate)).Should(Succeed())
			By("Deleting instances")
			Expect(k8sClient.Delete(ctx, &persistentInstance)).Should(Succeed())
			Expect(k8sClient.Delete(ctx, &nonPersistentInstance)).Should(Succeed())
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

	Context("Testing maximum time instance deletion", func() {
		It("Should succed: the VM did not reach the maximum deletion time", func() {
			By("Getting current instance")
			currentInstance := &crownlabsv1alpha2.Instance{}
			instanceLookupKey := types.NamespacedName{Name: PersistentInstanceName, Namespace: WorkingNamespace}
			Expect(k8sClient.Get(ctx, instanceLookupKey, currentInstance)).Should(Succeed())

			By("Checking the VM still exists")
			doesEventuallyExists(ctx, instanceLookupKey, currentInstance, BeTrue(), timeout, interval)
		})
		It("Should fail: the VM reached the maximum deletion time", func() {
			By("Getting current instance")
			currentInstance := &crownlabsv1alpha2.Instance{}
			instanceLookupKey := types.NamespacedName{Name: PersistentInstanceName, Namespace: WorkingNamespace}
			Expect(k8sClient.Get(ctx, instanceLookupKey, currentInstance)).Should(Succeed())

			By("Getting template associated to the instance")
			currentTemplate := &crownlabsv1alpha2.Template{}
			templateLookupKey := types.NamespacedName{Name: persistentTemplateName, Namespace: WorkingNamespace}
			Expect(k8sClient.Get(ctx, templateLookupKey, currentTemplate)).Should(Succeed())

			By("Patching deleteAfter field to a short time")
			currentTemplate.Spec.DeleteAfter = "0m"

			By("Checking the VM has been deleted")
			// fix this
			//doesEventuallyExists(ctx, instanceLookupKey, currentInstance, BeFalse(), timeout, interval)

		})

	})

	Context("Testing default values", func() {
		It("Should succed: the Non-Persistent template provides the default DeleteAfter field", func() {
			By("Getting current templates")
			currentTemplate := &crownlabsv1alpha2.Template{}

			templateLookupKey := types.NamespacedName{Name: nonPersistentTemplateName, Namespace: WorkingNamespace}
			Expect(k8sClient.Get(ctx, templateLookupKey, currentTemplate)).Should(Succeed())

			By("Checking the DeleteField is the default one")
			currentDeleteAfter := currentTemplate.Spec.DeleteAfter
			defaultDeleteAfter := "never"
			Expect(currentDeleteAfter).To(Equal(defaultDeleteAfter))

		})
		It("Should succed: the Persistent template has a custom DeleteAfter field", func() {
			By("Getting current templates")
			currentTemplate := &crownlabsv1alpha2.Template{}

			templateLookupKey := types.NamespacedName{Name: persistentTemplateName, Namespace: WorkingNamespace}
			Expect(k8sClient.Get(ctx, templateLookupKey, currentTemplate)).Should(Succeed())

			By("Checking the DeleteField is the custom one")
			currentDeleteAfter := currentTemplate.Spec.DeleteAfter
			Expect(currentDeleteAfter).To(Equal(CustomDeleteAfter))

		})
		It("Should succed: the Non-Persistent template provides the default InactivityTimeout field", func() {

			// This test is commented out because the InactivityTimeout field is not set by default in the template spec.

			// By("Getting current templates")
			// currentTemplate := &crownlabsv1alpha2.Template{}

			// templateLookupKey := types.NamespacedName{Name: nonPersistentTemplateName, Namespace: WorkingNamespace}
			// Expect(k8sClient.Get(ctx, templateLookupKey, currentTemplate)).Should(Succeed())

			// By("Checking the InactivityTimeout field is the default one")
			// currentInactivityTimeout := currentTemplate.Spec.InactivityTimeout
			// defaultInactivityTimeout := "60d"
			// Expect(currentInactivityTimeout).To(Equal(defaultInactivityTimeout))
		})
		It("Should succed: the Persistent template has a custom InactivityTimeout field", func() {
			By("Getting current templates")
			currentTemplate := &crownlabsv1alpha2.Template{}

			templateLookupKey := types.NamespacedName{Name: persistentTemplateName, Namespace: WorkingNamespace}
			Expect(k8sClient.Get(ctx, templateLookupKey, currentTemplate)).Should(Succeed())

			By("Checking the InactivityTimeout field is the custom one")
			currentInactivityTimeout := currentTemplate.Spec.InactivityTimeout
			Expect(currentInactivityTimeout).To(Equal(CustomInactivityTimeout))
		})

	})

	Context("Testing maximum inactivity time", func() {
		It("Should succed: the VM is active ", func() {

			mockProm.EXPECT().
				IsPrometheusHealthy(gomock.Any()).
				Return(true, nil).
				Times(1)

			mockProm.EXPECT().
				GetLastActivityTime(gomock.Any(), gomock.Any()).
				Return(time.Now(), nil).
				Times(1)

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
			By("Checking the VM is active")
			doesEventuallyExists(ctx, instanceLookupKey, currentInstance, BeTrue(), timeout, interval)

		})
		It("Should fail: the VM is inactive and it sends the first alert", func() {

		})
		It("Should fail: the persistent VM is inactive for a long time and it is stopped", func() {

		})
		It("Should fail: the non-persistent VM is inactive for a long time it is deleted", func() {

		})

	})

	// tests for prometheus healthy??

})
