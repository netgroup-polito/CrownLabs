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

var _ = Describe("Instautoctrl", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		PersistentInstanceName    = "test-instance-persistent"
		NonPersistentInstanceName = "test-instance-non-persistent"
		WorkingNamespace          = "working-namespace"
		persistentTemplateName    = "test-template-persistent"
		nonPersistentTemplateName = "test-template-non-persistent"
		TenantName                = "test-tenant"
		CustomDeleteAfter         = instautoctrl.NEVER_TIMEOUT_VALUE
		CustomInactivityTimeout   = instautoctrl.NEVER_TIMEOUT_VALUE
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
		newNs := workingNs.DeepCopy()
		newPersistentTemplate := persistentTemplate.DeepCopy()
		newNonPersistentTemplate := nonPersistentTemplate.DeepCopy()
		newPersistentInstance := persistentInstance.DeepCopy()
		newNonPersistentInstance := nonPersistentInstance.DeepCopy()
		newTenantNs := tenantNs.DeepCopy()
		newTenant := tenant.DeepCopy()
		By("Creating the namespace where to create instance and template")
		err := k8sClientExpiration.Create(ctx, newNs)
		err = k8sClientExpiration.Create(ctx, newTenantNs)
		if err != nil && errors.IsAlreadyExists(err) {
			By("Cleaning up the environment")
			By("Deleting templates")
			Expect(k8sClientExpiration.Delete(ctx, &persistentTemplate)).Should(Succeed())
			Expect(k8sClientExpiration.Delete(ctx, &nonPersistentTemplate)).Should(Succeed())
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
		tenantNs := &v1.Namespace{}

		nsLookupKey := types.NamespacedName{Name: WorkingNamespace}
		doesEventuallyExists(ctx, nsLookupKey, createdNs, BeTrue(), timeout, interval, k8sClientExpiration)
		tenantNsLookupKey := types.NamespacedName{Name: TenantName}
		doesEventuallyExists(ctx, tenantNsLookupKey, tenantNs, BeTrue(), timeout, interval, k8sClientExpiration)

		By("Creating the templates")
		Expect(k8sClientExpiration.Create(ctx, newPersistentTemplate)).Should(Succeed())
		Expect(k8sClientExpiration.Create(ctx, newNonPersistentTemplate)).Should(Succeed())

		By("By checking that the template has been created")
		persistentTemplateLookupKey := types.NamespacedName{Name: persistentTemplateName, Namespace: WorkingNamespace}
		nonPersistentTemplateLookupKey := types.NamespacedName{Name: nonPersistentTemplateName, Namespace: WorkingNamespace}
		createdPersitentTemplate := &crownlabsv1alpha2.Template{}
		createdNonPersitentTemplate := &crownlabsv1alpha2.Template{}

		doesEventuallyExists(ctx, persistentTemplateLookupKey, createdPersitentTemplate, BeTrue(), timeout, interval, k8sClientExpiration)
		doesEventuallyExists(ctx, nonPersistentTemplateLookupKey, createdNonPersitentTemplate, BeTrue(), timeout, interval, k8sClientExpiration)

		By("Creating the tenant")
		Expect(k8sClientExpiration.Create(ctx, newTenant)).Should(Succeed())

		By("Checking that the tenant has been created")
		tenantLookupKey := types.NamespacedName{Name: TenantName, Namespace: tenantNs.Name}
		createdTenant := &crownlabsv1alpha2.Tenant{}
		doesEventuallyExists(ctx, tenantLookupKey, createdTenant, BeTrue(), timeout, interval, k8sClientExpiration)

		By("Creating the instances")
		Expect(k8sClientExpiration.Create(ctx, newPersistentInstance)).Should(Succeed())
		Expect(k8sClientExpiration.Create(ctx, newNonPersistentInstance)).Should(Succeed())

		By("Checking that the instances has been created")
		persistanteInstanceLookupKey := types.NamespacedName{Name: PersistentInstanceName, Namespace: tenant.Name}
		nonPersistentInstanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenant.Name}
		createdPersistentInstance := &crownlabsv1alpha2.Instance{}
		createdNonPersistentInstance := &crownlabsv1alpha2.Instance{}

		doesEventuallyExists(ctx, persistanteInstanceLookupKey, createdPersistentInstance, BeTrue(), timeout, interval, k8sClientExpiration)
		doesEventuallyExists(ctx, nonPersistentInstanceLookupKey, createdNonPersistentInstance, BeTrue(), timeout, interval, k8sClientExpiration)

	})

	It("testing CheckInstanceExpiration function", func() {
		r := &instautoctrl.InstanceExpirationReconciler{
			Client: k8sClientExpiration,
		}
		By("Checking that the instance is running")
		InstanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenantNs.Name}
		currentInstance := &crownlabsv1alpha2.Instance{}
		doesEventuallyExists(ctx, InstanceLookupKey, currentInstance, BeTrue(), timeout, interval, k8sClientExpiration)
		TemplateLookupKey := types.NamespacedName{Name: nonPersistentTemplateName, Namespace: WorkingNamespace}
		currentTemplate := &crownlabsv1alpha2.Template{}
		doesEventuallyExists(ctx, TemplateLookupKey, currentTemplate, BeTrue(), timeout, interval, k8sClientExpiration)
		tenantLookupKey := types.NamespacedName{Name: TenantName, Namespace: tenantNs.Name}
		currentTenant := &crownlabsv1alpha2.Tenant{}
		doesEventuallyExists(ctx, tenantLookupKey, currentTenant, BeTrue(), timeout, interval, k8sClientExpiration)
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
			Client: k8sClientExpiration,
		}
		By("Checking that the instance is running")
		InstanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenantNs.Name}
		currentInstance := &crownlabsv1alpha2.Instance{}
		doesEventuallyExists(ctx, InstanceLookupKey, currentInstance, BeTrue(), timeout, interval, k8sClientExpiration)
		TemplateLookupKey := types.NamespacedName{Name: nonPersistentTemplateName, Namespace: WorkingNamespace}
		currentTemplate := &crownlabsv1alpha2.Template{}
		doesEventuallyExists(ctx, TemplateLookupKey, currentTemplate, BeTrue(), timeout, interval, k8sClientExpiration)
		tenantLookupKey := types.NamespacedName{Name: TenantName, Namespace: tenantNs.Name}
		currentTenant := &crownlabsv1alpha2.Tenant{}
		doesEventuallyExists(ctx, tenantLookupKey, currentTenant, BeTrue(), timeout, interval, k8sClientExpiration)
		ctx := context.Background()
		ctx, _ = pkgcontext.InstanceInto(ctx, currentInstance)
		ctx, _ = pkgcontext.TemplateInto(ctx, currentTemplate)
		ctx, _ = pkgcontext.TenantInto(ctx, currentTenant)
		By("Calling deleteInstance function")
		err := r.DeleteInstance(ctx)
		Expect(err).ToNot(HaveOccurred())
		By("Checking that the instance has been deleted")
		doesEventuallyExists(ctx, InstanceLookupKey, currentInstance, BeFalse(), timeout, interval, k8sClientExpiration)
	})
	It("Testing ShouldSendWarningNotification function", func() {
		r := &instautoctrl.InstanceExpirationReconciler{
			Client: k8sClientExpiration,
		}
		By("Checking that the instance is running")
		InstanceLookupKey := types.NamespacedName{Name: NonPersistentInstanceName, Namespace: tenantNs.Name}
		currentInstance := &crownlabsv1alpha2.Instance{}
		doesEventuallyExists(ctx, InstanceLookupKey, currentInstance, BeTrue(), timeout, interval, k8sClientExpiration)
		TemplateLookupKey := types.NamespacedName{Name: nonPersistentTemplateName, Namespace: WorkingNamespace}
		currentTemplate := &crownlabsv1alpha2.Template{}
		doesEventuallyExists(ctx, TemplateLookupKey, currentTemplate, BeTrue(), timeout, interval, k8sClientExpiration)
		tenantLookupKey := types.NamespacedName{Name: TenantName, Namespace: tenantNs.Name}
		currentTenant := &crownlabsv1alpha2.Tenant{}
		doesEventuallyExists(ctx, tenantLookupKey, currentTenant, BeTrue(), timeout, interval, k8sClientExpiration)
		ctx := context.Background()
		ctx, _ = pkgcontext.InstanceInto(ctx, currentInstance)
		ctx, _ = pkgcontext.TemplateInto(ctx, currentTemplate)
		ctx, _ = pkgcontext.TenantInto(ctx, currentTenant)

		By("Calling ShouldSendNotification function")
		shouldSend, err := r.ShouldSendWarningNotification(ctx)
		Expect(err).ToNot(HaveOccurred())
		Expect(shouldSend).To(BeTrue(), "Should send notification because the instance is not deleted yet")

		By("Verifying the ExpiringWarningNotificationAnnotation annotation has been added to the instance")
		Expect(currentInstance.Annotations[forge.ExpiringWarningNotificationAnnotation]).To(Equal("true"), "ExpiringWarningNotificationAnnotation annotation should be set to true")
	})

})
