package instance_controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var _ = Describe("Instance Operator controller for containers", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		TemplateName      = "template-name-cont"
		TemplateNamespace = "template-namespace-cont"
		InstanceName      = "instance-name-cont"
		InstanceNamespace = "instance-namespace-cont"

		timeout  = time.Second * 20
		interval = time.Millisecond * 500
	)

	var (
		ctx        context.Context
		templateNs = v1.Namespace{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: TemplateNamespace,
				Labels: map[string]string{
					"test-suite": "true",
				},
			},
			Spec:   v1.NamespaceSpec{},
			Status: v1.NamespaceStatus{},
		}
		instanceNs = v1.Namespace{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: InstanceNamespace,
				Labels: map[string]string{
					"production": "true",
					"test-suite": "true",
				},
			},
			Spec:   v1.NamespaceSpec{},
			Status: v1.NamespaceStatus{},
		}
		template = crownlabsv1alpha2.Template{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      TemplateName,
				Namespace: TemplateNamespace,
			},
			Spec: crownlabsv1alpha2.TemplateSpec{
				WorkspaceRef: crownlabsv1alpha2.GenericRef{},
				PrettyName:   "Container template",
				Description:  "This is the container template",
				EnvironmentList: []crownlabsv1alpha2.Environment{
					{
						Name:            TemplateName,
						Image:           "crownlabs/pycharm",
						EnvironmentType: crownlabsv1alpha2.ClassContainer,
						GuiEnabled:      true,
						Persistent:      false,
						Resources: crownlabsv1alpha2.EnvironmentResources{
							CPU:                   1,
							ReservedCPUPercentage: 1,
							Memory:                resource.MustParse("1024M"),
						},
					},
				},
				DeleteAfter: "30d",
			},
			Status: crownlabsv1alpha2.TemplateStatus{},
		}
		instance = crownlabsv1alpha2.Instance{
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
					Name:      "tenant-name",
					Namespace: "tenant-namespace",
				},
			},
			Status: crownlabsv1alpha2.InstanceStatus{},
		}
		environmentWithPVC = crownlabsv1alpha2.Environment{
			Name:            InstanceName,
			Image:           "crownlabs/pycharm",
			EnvironmentType: crownlabsv1alpha2.ClassContainer,
			Resources: crownlabsv1alpha2.EnvironmentResources{
				CPU:                   1,
				ReservedCPUPercentage: 1,
				Memory:                resource.MustParse("1024M"),
				Disk:                  resource.MustParse("5G"),
			},
		}
		templateWithPVC = crownlabsv1alpha2.Template{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      TemplateName + "with-pvc",
				Namespace: TemplateNamespace,
			},
			Spec: crownlabsv1alpha2.TemplateSpec{
				WorkspaceRef: crownlabsv1alpha2.GenericRef{},
				PrettyName:   "App Container Template",
				Description:  "This is the container template",
				EnvironmentList: []crownlabsv1alpha2.Environment{
					environmentWithPVC,
				},
				DeleteAfter: "30d",
			},
			Status: crownlabsv1alpha2.TemplateStatus{},
		}
		instanceWithPVC = crownlabsv1alpha2.Instance{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      InstanceName + "with-pvc",
				Namespace: InstanceNamespace,
			},
			Spec: crownlabsv1alpha2.InstanceSpec{
				Template: crownlabsv1alpha2.GenericRef{
					Name:      TemplateName + "with-pvc",
					Namespace: TemplateNamespace,
				},
			},
			Status: crownlabsv1alpha2.InstanceStatus{},
		}
		tmpl = crownlabsv1alpha2.Template{}
		inst = crownlabsv1alpha2.Instance{}
		svc  = v1.Service{}
		ingr = networkingv1.Ingress{}
		depl = appsv1.Deployment{}
		pvc  = v1.PersistentVolumeClaim{}
		flag = true
	)

	Context("When creating the container-based Instance", func() {
		It("Should create the Instance and Template namespaces", func() {
			ctx = context.Background()
			ns := v1.Namespace{}
			By("Creating the Template namespace")
			Expect(k8sClient.Create(ctx, &templateNs)).Should(Succeed())
			doesEventuallyExist(ctx, types.NamespacedName{
				Name: TemplateNamespace,
			}, &ns, BeTrue(), timeout, interval)

			By("Creating the Instance namespace")
			Expect(k8sClient.Create(ctx, &instanceNs)).Should(Succeed())
			doesEventuallyExist(ctx, types.NamespacedName{
				Name: InstanceNamespace,
			}, &ns, BeTrue(), timeout, interval)
		})

		It("Should create the Template and the related Instance", func() {
			By("Creating the Template")
			Expect(k8sClient.Create(ctx, &template)).Should(Succeed())
			doesEventuallyExist(ctx, types.NamespacedName{
				Name:      TemplateName,
				Namespace: TemplateNamespace,
			}, &tmpl, BeTrue(), timeout, interval)

			By("Creating the Instance associated to the Template")
			Expect(k8sClient.Create(ctx, &instance)).Should(Succeed())
			doesEventuallyExist(ctx, types.NamespacedName{
				Name:      InstanceName,
				Namespace: InstanceNamespace,
			}, &inst, BeTrue(), timeout, interval)
		})

		It("Should create the exposition environment", func() {
			expectedOwnerReference := metav1.OwnerReference{
				Kind:               "Instance",
				APIVersion:         "crownlabs.polito.it/v1alpha2",
				UID:                inst.ObjectMeta.UID,
				Name:               InstanceName,
				Controller:         &flag,
				BlockOwnerDeletion: &flag,
			}

			By("Checking that the service exposing the pod (remote desktop + FileBrowser containers) exists")
			doesEventuallyExist(ctx, types.NamespacedName{
				Name:      InstanceName,
				Namespace: InstanceNamespace,
			}, &svc, BeTrue(), timeout, interval)
			By("Checking that the service has got an OwnerReference")
			Expect(svc.ObjectMeta.OwnerReferences).To(ContainElement(expectedOwnerReference))

			By("Checking that the ingress exists")
			doesEventuallyExist(ctx, types.NamespacedName{
				Name:      InstanceName,
				Namespace: InstanceNamespace,
			}, &ingr, BeTrue(), timeout, interval)
			By("Checking that the ingress has got an OwnerReference")
			Expect(ingr.ObjectMeta.OwnerReferences).To(ContainElement(expectedOwnerReference))

			By("Checking that the dedicated FileBrowser ingress exists")
			doesEventuallyExist(ctx, types.NamespacedName{
				Name:      InstanceName + "-filebrowser",
				Namespace: InstanceNamespace,
			}, &ingr, BeTrue(), timeout, interval)
			By("Checking that the dedicated FileBrowser ingress has got an OwnerReference")
			Expect(ingr.ObjectMeta.OwnerReferences).To(ContainElement(expectedOwnerReference))

			By("Checking that the OAUTH service exists")
			doesEventuallyExist(ctx, types.NamespacedName{
				Name:      InstanceName + "-oauth2",
				Namespace: InstanceNamespace,
			}, &svc, BeTrue(), timeout, interval)
			By("Checking that the auth service has got an OwnerReference")
			Expect(svc.ObjectMeta.OwnerReferences).To(ContainElement(expectedOwnerReference))

			By("Checking that the OAUTH ingress exists")
			doesEventuallyExist(ctx, types.NamespacedName{
				Name:      InstanceName + "-oauth2",
				Namespace: InstanceNamespace,
			}, &ingr, BeTrue(), timeout, interval)
			By("Checking that the auth ingress has got an OwnerReference")
			Expect(ingr.ObjectMeta.OwnerReferences).To(ContainElement(expectedOwnerReference))

			By("Checking that the OAUTH deployment exists")
			doesEventuallyExist(ctx, types.NamespacedName{
				Name:      InstanceName + "-oauth2",
				Namespace: InstanceNamespace,
			}, &depl, BeTrue(), timeout, interval)
			By("Checking that the auth deployment has got an OwnerReference")
			Expect(depl.ObjectMeta.OwnerReferences).To(ContainElement(expectedOwnerReference))
		})

		It("Should create the deployment", func() {
			doesEventuallyExist(ctx, types.NamespacedName{
				Name:      InstanceName,
				Namespace: InstanceNamespace,
			}, &inst, BeTrue(), timeout, interval)

			expectedOwnerReference := metav1.OwnerReference{
				Kind:               "Instance",
				APIVersion:         "crownlabs.polito.it/v1alpha2",
				UID:                inst.ObjectMeta.UID,
				Name:               InstanceName,
				Controller:         &flag,
				BlockOwnerDeletion: &flag,
			}

			By("Checking that the deployment exists")
			doesEventuallyExist(ctx, types.NamespacedName{
				Name:      InstanceName,
				Namespace: InstanceNamespace,
			}, &depl, BeTrue(), timeout, interval)

			By("Checking that the deployment has got an OwnerReference")
			Expect(depl.ObjectMeta.OwnerReferences).To(ContainElement(expectedOwnerReference))
		})

		It("Should have the Instance ready", func() {
			By("Setting the number of ready replicas")
			depl.Status.Replicas = 1
			depl.Status.ReadyReplicas = 1
			Expect(k8sClient.Status().Update(ctx, &depl)).Should(Succeed())

			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      InstanceName,
					Namespace: InstanceNamespace,
				}, &instance)
				return err == nil && instance.Status.Phase == "VmiReady"
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When creating containerized apps with attached PVC", func() {
		It("Should create the container Template and the related container Instance", func() {
			By("Creating the Template")
			Expect(k8sClient.Create(ctx, &templateWithPVC)).Should(Succeed())
			doesEventuallyExist(ctx, types.NamespacedName{
				Name:      TemplateName + "with-pvc",
				Namespace: TemplateNamespace,
			}, &tmpl, BeTrue(), timeout, interval)

			By("Creating the Instance associated to the Template")
			Expect(k8sClient.Create(ctx, &instanceWithPVC)).Should(Succeed())
			doesEventuallyExist(ctx, types.NamespacedName{
				Name:      InstanceName + "with-pvc",
				Namespace: InstanceNamespace,
			}, &inst, BeTrue(), timeout, interval)
		})

		It("Should correctly generate a PVC", func() {
			By("Checking for the existence of the PVC")
			doesEventuallyExist(ctx, types.NamespacedName{
				Name:      InstanceName + "with-pvc",
				Namespace: InstanceNamespace,
			}, &pvc, BeTrue(), timeout, interval)

			By("Checking that the PVC Has An OwnerReference")
			flag := true
			expectedOwnerReference := metav1.OwnerReference{
				Kind:               "Instance",
				APIVersion:         "crownlabs.polito.it/v1alpha2",
				UID:                instanceWithPVC.ObjectMeta.UID,
				Name:               InstanceName + "with-pvc",
				Controller:         &flag,
				BlockOwnerDeletion: &flag,
			}
			Expect(pvc.ObjectMeta.OwnerReferences).To(ContainElement(expectedOwnerReference))
			By("Checking that the PVC size is correct")
			Expect(*(pvc.Spec.Resources.Requests.Storage())).To(BeEquivalentTo(templateWithPVC.Spec.EnvironmentList[0].Resources.Disk))
		})

	})

	Context("When deleting the deployment", func() {
		It("Should delete the deployment", func() {
			doesEventuallyExist(ctx, types.NamespacedName{
				Name:      InstanceName,
				Namespace: InstanceNamespace,
			}, &depl, BeTrue(), timeout, interval)

			Expect(k8sClient.Delete(ctx, &depl)).Should(Succeed())
		})

		It("Should re-create the deployment", func() {
			doesEventuallyExist(ctx, types.NamespacedName{
				Name:      InstanceName,
				Namespace: InstanceNamespace,
			}, &depl, BeTrue(), timeout, interval)
		})
	})
})
