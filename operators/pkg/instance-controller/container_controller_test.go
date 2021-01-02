package instance_controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	errors "k8s.io/apimachinery/pkg/api/errors"
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
		webdavSecret = v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "webdav-secret",
				Namespace: InstanceNamespace,
			},
			Data: map[string][]byte{
				"username": []byte("username"),
				"password": []byte("password"),
			},
			StringData: nil,
			Type:       "",
		}
		environment = crownlabsv1alpha2.Environment{
			Name:            InstanceName,
			Image:           "crownlabs/pycharm",
			EnvironmentType: crownlabsv1alpha2.ClassContainer,
			Resources: crownlabsv1alpha2.EnvironmentResources{
				CPU:                   1,
				ReservedCPUPercentage: 1,
				Memory:                resource.MustParse("1024M"),
			},
		}
		template = crownlabsv1alpha2.Template{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      TemplateName,
				Namespace: TemplateNamespace,
			},
			Spec: crownlabsv1alpha2.TemplateSpec{
				WorkspaceRef: crownlabsv1alpha2.GenericRef{},
				PrettyName:   "App Container Template",
				Description:  "This is the container template",
				EnvironmentList: []crownlabsv1alpha2.Environment{
					environment,
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
			},
			Status: crownlabsv1alpha2.InstanceStatus{},
		}
		ns   = v1.Namespace{}
		tmpl = crownlabsv1alpha2.Template{}
		inst = crownlabsv1alpha2.Instance{}
		depl = appsv1.Deployment{}
	)

	Context("When creating containerized apps", func() {
		It("Should create the Instance and Template namespaces", func() {
			ctx = context.Background()
			By("Creating the Template namespace")
			Expect(k8sClient.Create(ctx, &templateNs)).Should(Succeed())
			doesEventuallyExist(ctx, types.NamespacedName{
				Name: TemplateNamespace,
			}, &ns, BeTrue(), timeout, interval)

			By("Creating the Instance namespace")
			Expect(k8sClient.Create(ctx, &instanceNs)).Should(Succeed())
			Expect(k8sClient.Create(ctx, &webdavSecret)).Should(Succeed())
			doesEventuallyExist(ctx, types.NamespacedName{
				Name: InstanceNamespace,
			}, &ns, BeTrue(), timeout, interval)
		})

		It("Should create the container Template and the related container Instance", func() {
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

		It("Should check the deployment exists", func() {
			By("Checking for the existence of the deployment")
			doesEventuallyExist(ctx, types.NamespacedName{
				Name:      InstanceName,
				Namespace: InstanceNamespace,
			}, &depl, BeTrue(), timeout, interval)

			flag := true
			expectedOwnerReference := metav1.OwnerReference{
				Kind:               "Instance",
				APIVersion:         "crownlabs.polito.it/v1alpha2",
				UID:                instance.ObjectMeta.UID,
				Name:               InstanceName,
				Controller:         &flag,
				BlockOwnerDeletion: &flag,
			}
			By("Ensuring the deployment has got an OwnerReference")
			Expect(depl.ObjectMeta.OwnerReferences).To(ContainElement(expectedOwnerReference))
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

	Context("When deleting containerized apps", func() {
		It("Should delete the container Instance", func() {
			By("Deleting the container Instance")
			Expect(k8sClient.Delete(ctx, &instance)).Should(Succeed())

			By("Ensuring the container Instance is deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      InstanceName,
					Namespace: InstanceNamespace,
				}, &inst)
				if err != nil && errors.IsNotFound(err) {
					return true
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})
	})
})
