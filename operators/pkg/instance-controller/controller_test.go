package instance_controller

import (
	"context"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gomegaTypes "github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:docs-gen:collapse=Imports

var _ = Describe("LabOperator controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		InstanceName      = "test-instance"
		InstanceNamespace = "instance-namespace"
		TemplateName      = "test-template"
		TemplateNamespace = "template-namespace"
		TenantName        = "test-tenant"
		TenantNamespace   = "test-namespace"

		timeout  = time.Second * 10
		interval = time.Millisecond * 500
	)

	Context("LabOperator controller", func() {
		It("Should create the related resources when creating an instance", func() {
			By("By creating a Instance")
			ctx := context.Background()
			templateNs := v1.Namespace{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: TemplateNamespace,
				},
				Spec:   v1.NamespaceSpec{},
				Status: v1.NamespaceStatus{},
			}
			instanceNs := v1.Namespace{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: InstanceNamespace,
				},
				Spec:   v1.NamespaceSpec{},
				Status: v1.NamespaceStatus{},
			}

			template := crownlabsv1alpha2.Template{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      TemplateName,
					Namespace: TemplateNamespace,
				},
				Spec: crownlabsv1alpha2.TemplateSpec{
					WorkspaceRef: crownlabsv1alpha2.GenericRef{},
					PrettyName:   "Wonderful Template",
					Description:  "A description",
					EnvironmentList: []crownlabsv1alpha2.Environment{
						{
							Name:       "Test",
							GuiEnabled: true,
							Resources: crownlabsv1alpha2.EnvironmentResources{
								CPU:                   1,
								ReservedCPUPercentage: 1,
								Memory:                resource.MustParse("1024M"),
							},
							EnvironmentType: crownlabsv1alpha2.ClassVM,
							Persistent:      false,
							Image:           "trololo/vm",
						},
					},
					DeleteAfter: "",
				},
				Status: crownlabsv1alpha2.TemplateStatus{},
			}
			instance := crownlabsv1alpha2.Instance{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      InstanceName,
					Namespace: InstanceNamespace,
				},
				Spec: crownlabsv1alpha2.InstanceSpec{
					Template: crownlabsv1alpha2.GenericRef{
						Name:      TemplateName,
						Namespace: TemplateNamespace,
					},
					Tenant: crownlabsv1alpha2.GenericRef{
						Name:      TenantName,
						Namespace: TenantNamespace,
					},
				},
				Status: crownlabsv1alpha2.InstanceStatus{},
			}
			Expect(k8sClient.Create(ctx, &templateNs)).Should(Succeed())
			Expect(k8sClient.Create(ctx, &instanceNs)).Should(Succeed())
			ns1 := v1.Namespace{}
			doesEventuallyExists(ctx, types.NamespacedName{
				Name: TemplateNamespace,
			}, &ns1, BeTrue(), timeout, interval)
			doesEventuallyExists(ctx, types.NamespacedName{
				Name: InstanceNamespace,
			}, &ns1, BeTrue(), timeout, interval)

			By("By creating a Instance")

			Expect(k8sClient.Create(ctx, &template)).Should(Succeed())
			Expect(k8sClient.Create(ctx, &instance)).Should(Succeed())

		})
	})

})

func doesEventuallyExists(ctx context.Context, nsLookupKey types.NamespacedName, createdObject runtime.Object, expectedStatus gomegaTypes.GomegaMatcher, timeout time.Duration, interval time.Duration) {
	Eventually(func() bool {
		err := k8sClient.Get(ctx, nsLookupKey, createdObject)
		return err == nil
	}, timeout, interval).Should(expectedStatus)

}
