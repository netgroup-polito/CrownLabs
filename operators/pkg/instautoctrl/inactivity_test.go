package instautoctrl_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var _ = Describe("Inactivity", func() {
	BeforeEach(func() {
		const (
			WorkingNamespace = "test-namespace"
			TenantName       = "test-tenant"
		)

		// Persistent Template and Instance, used to test the behavior of the controller functions with persistent instances.
		var (
			persistentTemplateEnvironment = crownlabsv1alpha2.TemplateSpec{
				WorkspaceRef: crownlabsv1alpha2.GenericRef{},
				PrettyName:   "Persistent Template",
				Description:  "Description the template",
				EnvironmentList: []crownlabsv1alpha2.Environment{
					{
						Name:       "Env-Persistent",
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
				DeleteAfter:       "30d",
				InactivityTimeout: "72h",
			}

			persistentTemplate = crownlabsv1alpha2.Template{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "persistent-template",
					Namespace: WorkingNamespace,
				},
				Spec:   persistentTemplateEnvironment,
				Status: crownlabsv1alpha2.TemplateStatus{},
			}

			persistentInstance = crownlabsv1alpha2.Instance{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "persistent-instance",
					Namespace: WorkingNamespace,
				},
				Spec: crownlabsv1alpha2.InstanceSpec{
					Running: false,
					Template: crownlabsv1alpha2.GenericRef{
						Name:      "persistent-template",
						Namespace: WorkingNamespace,
					},
					Tenant: crownlabsv1alpha2.GenericRef{
						Name: TenantName,
					},
				},
				Status: crownlabsv1alpha2.InstanceStatus{},
			}
		)

		// Not Persistent Template and Instance, used to test the behavior of the controller functions with non-persistent instances.
		var (
			NotPersistentTemplateEnvironment = crownlabsv1alpha2.TemplateSpec{
				WorkspaceRef: crownlabsv1alpha2.GenericRef{},
				PrettyName:   "Not Persistent Template",
				Description:  "Description the template",
				EnvironmentList: []crownlabsv1alpha2.Environment{
					{
						Name:       "Env-Not-Persistent",
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
				DeleteAfter:       "30d",
				InactivityTimeout: "72h",
			}

			NotPersistentTemplate = crownlabsv1alpha2.Template{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "not-persistent-template",
					Namespace: WorkingNamespace,
				},
				Spec:   NotPersistentTemplateEnvironment,
				Status: crownlabsv1alpha2.TemplateStatus{},
			}

			NotPersistentInstance = crownlabsv1alpha2.Instance{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "not-persistent-instance",
					Namespace: WorkingNamespace,
				},
				Spec: crownlabsv1alpha2.InstanceSpec{
					Running: true,
					Template: crownlabsv1alpha2.GenericRef{
						Name:      "not-persistent-template",
						Namespace: WorkingNamespace,
					},
					Tenant: crownlabsv1alpha2.GenericRef{
						Name: TenantName,
					},
				},
				Status: crownlabsv1alpha2.InstanceStatus{},
			}
		)

		// Never Template and Instance, used to test the behavior of the controller functions with instances that should never be deleted due to inactivity.
		var (
			NeverTemplateEnvironment = crownlabsv1alpha2.TemplateSpec{
				WorkspaceRef: crownlabsv1alpha2.GenericRef{},
				PrettyName:   "Never Template",
				Description:  "Description the template",
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
				DeleteAfter:       "30d",
				InactivityTimeout: "never",
			}

			NeverPersistentTemplate = crownlabsv1alpha2.Template{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "never-template",
					Namespace: WorkingNamespace,
				},
				Spec:   NeverTemplateEnvironment,
				Status: crownlabsv1alpha2.TemplateStatus{},
			}

			NeverPersistentInstance = crownlabsv1alpha2.Instance{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "never-instance",
					Namespace: WorkingNamespace,
				},
				Spec: crownlabsv1alpha2.InstanceSpec{
					Running: false,
					Template: crownlabsv1alpha2.GenericRef{
						Name:      "never-template",
						Namespace: WorkingNamespace,
					},
					Tenant: crownlabsv1alpha2.GenericRef{
						Name: TenantName,
					},
				},
				Status: crownlabsv1alpha2.InstanceStatus{},
			}
		)
	})

})
