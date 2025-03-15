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

package instctrl_test

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/instctrl"
	tntctrl "github.com/netgroup-polito/CrownLabs/operators/pkg/tenant-controller"
	. "github.com/netgroup-polito/CrownLabs/operators/pkg/utils/tests"
)

var _ = Describe("Generation of the cloud-init configuration", func() {
	var (
		ctx           context.Context
		clientBuilder fake.ClientBuilder
		reconciler    instctrl.InstanceReconciler

		instance    clv1alpha2.Instance
		template    clv1alpha2.Template
		tenant      clv1alpha2.Tenant
		environment clv1alpha2.Environment

		pvcSecretName types.NamespacedName
		objectName    types.NamespacedName
		secret        corev1.Secret

		ownerRef metav1.OwnerReference

		err error
	)

	const (
		instanceName      = "kubernetes-0000"
		instanceNamespace = "tenant-tester"
		templateName      = "kubernetes"
		templateNamespace = "workspace-netgroup"
		workspaceName     = "netgroup"
		environmentName   = "control-plane"
		tenantName        = "tester"

		NFSServiceName = "rook-nfs-server-name"
		NFSServicePath = "/path"
	)

	NewTenant := func(suffix string, workspace string, role clv1alpha2.WorkspaceUserRole, keys []string) clv1alpha2.Tenant {
		return clv1alpha2.Tenant{
			ObjectMeta: metav1.ObjectMeta{
				Name:   fmt.Sprintf("%s-%s", tenantName, suffix),
				Labels: map[string]string{clv1alpha2.WorkspaceLabelPrefix + workspace: string(role)},
			},
			Spec: clv1alpha2.TenantSpec{PublicKeys: keys},
		}
	}

	BeforeEach(func() {
		ctx = ctrl.LoggerInto(context.Background(), logr.Discard())
		clientBuilder = *fake.NewClientBuilder().WithScheme(scheme.Scheme)

		instance = clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{Name: instanceName, Namespace: instanceNamespace},
			Spec: clv1alpha2.InstanceSpec{
				Running:  true,
				Template: clv1alpha2.GenericRef{Name: templateName, Namespace: templateNamespace},
				Tenant:   clv1alpha2.GenericRef{Name: tenantName},
			},
		}
		template = clv1alpha2.Template{
			ObjectMeta: metav1.ObjectMeta{Name: templateName, Namespace: templateNamespace},
			Spec: clv1alpha2.TemplateSpec{
				WorkspaceRef: clv1alpha2.GenericRef{Name: workspaceName},
			},
		}
		environment = clv1alpha2.Environment{Name: environmentName, MountMyDriveVolume: true}
		tenant = NewTenant("user", workspaceName, clv1alpha2.User, []string{"tenant-key-1", "tenant-key-2"})

		pvcSecretName = types.NamespacedName{Namespace: instanceNamespace, Name: tntctrl.NFSSecretName}
		objectName = forge.NamespacedName(&instance)
		secret = corev1.Secret{}

		ownerRef = metav1.OwnerReference{
			APIVersion:         clv1alpha2.GroupVersion.String(),
			Kind:               "Instance",
			Name:               instance.GetName(),
			UID:                instance.GetUID(),
			BlockOwnerDeletion: ptr.To(true),
			Controller:         ptr.To(true),
		}
	})

	ForgePvcSecret := func(serviceNameKey, servicePathKey string) *corev1.Secret {
		return &corev1.Secret{
			ObjectMeta: forge.NamespacedNameToObjectMeta(pvcSecretName),
			Data: map[string][]byte{
				serviceNameKey: []byte(NFSServiceName),
				servicePathKey: []byte(NFSServicePath),
			},
		}
	}

	JustBeforeEach(func() {
		client := FakeClientWrapped{Client: clientBuilder.Build()}
		reconciler = instctrl.InstanceReconciler{
			Client: client, Scheme: scheme.Scheme,
		}

		ctx, _ = clctx.InstanceInto(ctx, &instance)
		ctx, _ = clctx.TemplateInto(ctx, &template)
		ctx, _ = clctx.TenantInto(ctx, &tenant)
		ctx, _ = clctx.EnvironmentInto(ctx, &environment)
	})

	Describe("The EnforceCloudInitSecret function", func() {
		var expected []byte

		Extractor := func(content map[string][]byte) string {
			return string(content[instctrl.UserDataKey])
		}

		BeforeEach(func() {
			clientBuilder = *clientBuilder.WithObjects(ForgePvcSecret(tntctrl.NFSSecretServerNameKey, tntctrl.NFSSecretPathKey))

			expected, err = forge.CloudInitUserData(tenant.Spec.PublicKeys, []forge.NFSVolumeMountInfo{
				forge.MyDriveNFSVolumeMountInfo(NFSServiceName, NFSServicePath),
			})
			Expect(err).ToNot(HaveOccurred())
		})

		JustBeforeEach(func() {
			err = reconciler.EnforceCloudInitSecret(ctx)
		})

		When("the secret does not yet exist", func() {
			It("Should succeed", func() { Expect(err).ToNot(HaveOccurred()) })

			It("Should be present and have the common attributes", func() {
				Expect(reconciler.Get(ctx, objectName, &secret)).To(Succeed())
				Expect(secret.GetLabels()).To(Equal(forge.InstanceObjectLabels(nil, &instance)))
				Expect(secret.GetOwnerReferences()).To(ContainElement(ownerRef))
			})

			It("Should be present and have the expected content", func() {
				Expect(reconciler.Get(ctx, objectName, &secret)).To(Succeed())
				Expect(secret.Data).To(WithTransform(Extractor, Equal(string(expected))))
				Expect(secret.Type).To(Equal(corev1.SecretTypeOpaque))
			})
		})

		When("the secret already exists", func() {
			BeforeEach(func() {
				scrt := corev1.Secret{ObjectMeta: forge.NamespacedNameToObjectMeta(objectName)}
				clientBuilder.WithObjects(&scrt)
			})

			It("Should succeed", func() { Expect(err).ToNot(HaveOccurred()) })

			It("Should still be present and have the common attributes", func() {
				Expect(reconciler.Get(ctx, objectName, &secret)).To(Succeed())
				Expect(secret.GetLabels()).To(Equal(forge.InstanceObjectLabels(nil, &instance)))
				Expect(secret.GetOwnerReferences()).To(ContainElement(ownerRef))
			})

			It("Should still be present and have unmodified content", func() {
				Expect(reconciler.Get(ctx, objectName, &secret)).To(Succeed())
				Expect(secret.Data).To(WithTransform(Extractor, Equal(string(expected))))
				Expect(secret.Type).To(Equal(corev1.SecretTypeOpaque))
			})
		})

	})

	Describe("The NFSSpecs function", func() {
		var serviceName, servicePath string

		JustBeforeEach(func() {
			serviceName, servicePath, err = reconciler.GetNFSSpecs(ctx)
		})

		Context("The user-pvc secret does not exist", func() {
			It("Should return a not found error", func() { Expect(err).To(FailBecauseNotFound()) })
		})

		Context("The user-pvc secret exists", func() {
			When("the secret contains the expected data", func() {
				BeforeEach(func() {
					clientBuilder = *clientBuilder.WithObjects(ForgePvcSecret(tntctrl.NFSSecretServerNameKey, tntctrl.NFSSecretPathKey))
				})

				It("Should not return an error", func() { Expect(err).ToNot(HaveOccurred()) })
				It("The retrieved dns name should be correct", func() { Expect(serviceName).To(BeIdenticalTo(NFSServiceName)) })
				It("The retrieved path should be correct", func() { Expect(servicePath).To(BeIdenticalTo(NFSServicePath)) })
			})

			When("the secret does not contain the dns name", func() {
				BeforeEach(func() {
					clientBuilder = *clientBuilder.WithObjects(ForgePvcSecret("invalid-name-key", tntctrl.NFSSecretPathKey))
				})

				It("Should return an error", func() { Expect(err).To(HaveOccurred()) })
			})

			When("the secret does not contain the path", func() {
				BeforeEach(func() {
					clientBuilder = *clientBuilder.WithObjects(ForgePvcSecret(tntctrl.NFSSecretServerNameKey, "invalid-path-key"))
				})

				It("Should return an error", func() { Expect(err).To(HaveOccurred()) })
			})
		})
	})

	Describe("The GetPublicKeys function", func() {
		var (
			keys  []string
			other clv1alpha2.Tenant
		)

		JustBeforeEach(func() {
			keys, err = reconciler.GetPublicKeys(ctx)
		})

		When("no managers are associated with the instance workspace", func() {
			It("Should succeed", func() { Expect(err).ToNot(HaveOccurred()) })
			It("Should return the tenant public keys", func() { Expect(keys).To(ConsistOf(tenant.Spec.PublicKeys)) })
		})

		When("there is a manager associated with the instance workspace", func() {
			BeforeEach(func() {
				other = NewTenant("mgr", workspaceName, clv1alpha2.Manager, []string{"manager-key-1", "manager-key-2"})
				clientBuilder.WithObjects(&other)
			})

			It("Should succeed", func() { Expect(err).ToNot(HaveOccurred()) })
			It("Should return both the tenant and the manager public keys", func() {
				Expect(keys).To(ContainElements(tenant.Spec.PublicKeys))
				Expect(keys).To(ContainElements(other.Spec.PublicKeys))
			})
		})

		When("there is a manager associated with another workspace", func() {
			BeforeEach(func() {
				other = NewTenant("mgr", "another", clv1alpha2.Manager, []string{"manager-key-1", "manager-key-2"})
				clientBuilder.WithObjects(&other)
			})

			It("Should succeed", func() { Expect(err).ToNot(HaveOccurred()) })
			It("Should return the tenant public keys only", func() { Expect(keys).To(ConsistOf(tenant.Spec.PublicKeys)) })
		})

		When("there is another user associated with the instance workspace", func() {
			BeforeEach(func() {
				other = NewTenant("other", workspaceName, clv1alpha2.User, []string{"other-key-1", "other-key-2"})
				clientBuilder.WithObjects(&other)
			})

			It("Should succeed", func() { Expect(err).ToNot(HaveOccurred()) })
			It("Should return the tenant public keys only", func() { Expect(keys).To(ConsistOf(tenant.Spec.PublicKeys)) })
		})
	})
})
