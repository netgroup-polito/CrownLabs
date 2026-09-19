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

package webhook_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	apicommon "github.com/netgroup-polito/CrownLabs/operators/api/common"
	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/instance/webhook"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

func TestInstanceValidator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "InstanceValidator Suite")
}

var _ = Describe("InstanceValidator", func() {
	var (
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("should allow creation when under quota", func() {
		ws := &clv1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{Name: testWorkspace},
			Spec: clv1alpha1.WorkspaceSpec{
				Quota: apicommon.WorkspaceResourceQuota{
					Instances: 2,
					ResourceSpec: apicommon.ResourceSpec{
						CPU:    4,
						Memory: resource.MustParse("8Gi"),
					},
				},
			},
		}
		tmpl := &clv1alpha2.Template{
			ObjectMeta: metav1.ObjectMeta{Name: testTemplate, Namespace: testWorkspaceNamespace},
			Spec: clv1alpha2.TemplateSpec{
				EnvironmentList: []clv1alpha2.Environment{{
					Name: testEnvironment,
					Resources: clv1alpha2.EnvironmentResources{
						ResourceSpec: apicommon.ResourceSpec{
							CPU:    2,
							Memory: resource.MustParse("2Gi"),
						},
					},
				}},
				WorkspaceRef: clv1alpha2.GenericRef{Name: testWorkspace},
			},
		}
		inst := &clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testNewInstance,
				Namespace: testTenantNamespace,
				Labels:    map[string]string{forge.LabelWorkspaceKey: testWorkspace},
			},
			Spec: clv1alpha2.InstanceSpec{
				Template: clv1alpha2.GenericRef{Name: testTemplate, Namespace: testWorkspaceNamespace},
				Running:  true,
			},
		}
		// Add another instance to exceed quota
		otherInst := &clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testExistingInstance,
				Namespace: testTenantNamespace,
				Labels:    map[string]string{forge.LabelWorkspaceKey: testWorkspace},
			},
			Spec: clv1alpha2.InstanceSpec{
				Template: clv1alpha2.GenericRef{Name: testTemplate, Namespace: testWorkspaceNamespace},
				Running:  true,
			},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ws, tmpl, otherInst).Build()
		validator := &webhook.InstanceValidator{Client: fakeClient}
		warnings, err := validator.ValidateCreate(ctx, inst)
		Expect(err).To(BeNil())
		Expect(warnings).To(BeEmpty())
	})

	It("should deny creation when quota exceeded", func() {
		ws := &clv1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{Name: testWorkspace},
			Spec: clv1alpha1.WorkspaceSpec{
				Quota: apicommon.WorkspaceResourceQuota{
					Instances: 1,
					ResourceSpec: apicommon.ResourceSpec{
						CPU:    2,
						Memory: resource.MustParse("2Gi"),
					},
				},
			},
		}
		tmpl := &clv1alpha2.Template{
			ObjectMeta: metav1.ObjectMeta{Name: testTemplate, Namespace: testWorkspaceNamespace},
			Spec: clv1alpha2.TemplateSpec{
				EnvironmentList: []clv1alpha2.Environment{{
					Name: testEnvironment,
					Resources: clv1alpha2.EnvironmentResources{
						ResourceSpec: apicommon.ResourceSpec{
							CPU:    2,
							Memory: resource.MustParse("2Gi"),
						},
					},
				}},
				WorkspaceRef: clv1alpha2.GenericRef{Name: testWorkspace},
			},
		}
		inst := &clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testNewInstance,
				Namespace: testTenantNamespace,
				Labels:    map[string]string{forge.LabelWorkspaceKey: testWorkspace},
			},
			Spec: clv1alpha2.InstanceSpec{
				Template: clv1alpha2.GenericRef{Name: testTemplate, Namespace: testWorkspaceNamespace},
				Running:  true,
			},
		}
		// Add another instance to exceed quota
		otherInst := &clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testExistingInstance,
				Namespace: testTenantNamespace,
				Labels:    map[string]string{forge.LabelWorkspaceKey: testWorkspace},
			},
			Spec: clv1alpha2.InstanceSpec{
				Template: clv1alpha2.GenericRef{Name: testTemplate, Namespace: testWorkspaceNamespace},
				Running:  true,
			},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ws, tmpl, otherInst).Build()
		validator := &webhook.InstanceValidator{Client: fakeClient}
		warnings, err := validator.ValidateCreate(ctx, inst)
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(ContainSubstring("quota exceeded"))
		Expect(warnings).To(BeEmpty())
	})

	It("should deny creation when disk quota is exceeded", func() {
		ws := &clv1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{Name: testWorkspace},
			Spec: clv1alpha1.WorkspaceSpec{
				Quota: apicommon.WorkspaceResourceQuota{
					Instances: 2,
					ResourceSpec: apicommon.ResourceSpec{
						CPU:    4,
						Memory: resource.MustParse("8Gi"),
						Disk:   resource.MustParse("10Gi"),
					},
				},
			},
		}
		tmpl := &clv1alpha2.Template{
			ObjectMeta: metav1.ObjectMeta{Name: testTemplate, Namespace: testWorkspaceNamespace},
			Spec: clv1alpha2.TemplateSpec{
				EnvironmentList: []clv1alpha2.Environment{{
					Name: testEnvironment,
					Resources: clv1alpha2.EnvironmentResources{
						ResourceSpec: apicommon.ResourceSpec{
							CPU:    4,
							Memory: resource.MustParse("8Gi"),
							Disk:   resource.MustParse("12Gi"),
						},
					},
				}},
				WorkspaceRef: clv1alpha2.GenericRef{Name: testWorkspace},
			},
		}
		inst := &clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testNewInstance,
				Namespace: testTenantNamespace,
				Labels:    map[string]string{forge.LabelWorkspaceKey: testWorkspace},
			},
			Spec: clv1alpha2.InstanceSpec{
				Template: clv1alpha2.GenericRef{Name: testTemplate, Namespace: testWorkspaceNamespace},
				Running:  true,
			},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ws, tmpl).Build()
		validator := &webhook.InstanceValidator{Client: fakeClient}
		warnings, err := validator.ValidateCreate(ctx, inst)
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(ContainSubstring("quota exceeded: Disk"))
		Expect(warnings).To(BeEmpty())
	})

	It("should deny creation when extended resource quota is exceeded", func() {
		ws := &clv1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{Name: testWorkspace},
			Spec: clv1alpha1.WorkspaceSpec{
				Quota: apicommon.WorkspaceResourceQuota{
					Instances: 2,
					ResourceSpec: apicommon.ResourceSpec{
						CPU:    4,
						Memory: resource.MustParse("8Gi"),
						OtherResources: map[string]resource.Quantity{
							"nvidia.com/gpu": resource.MustParse("1"),
						},
					},
				},
			},
		}
		tmpl := &clv1alpha2.Template{
			ObjectMeta: metav1.ObjectMeta{Name: testTemplate, Namespace: testWorkspaceNamespace},
			Spec: clv1alpha2.TemplateSpec{
				EnvironmentList: []clv1alpha2.Environment{{
					Name: testEnvironment,
					Resources: clv1alpha2.EnvironmentResources{
						ResourceSpec: apicommon.ResourceSpec{
							CPU:    4,
							Memory: resource.MustParse("8Gi"),
							OtherResources: map[string]resource.Quantity{
								"nvidia.com/gpu": resource.MustParse("2"),
							},
						},
					},
				}},
				WorkspaceRef: clv1alpha2.GenericRef{Name: testWorkspace},
			},
		}
		inst := &clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testNewInstance,
				Namespace: testTenantNamespace,
				Labels:    map[string]string{forge.LabelWorkspaceKey: testWorkspace},
			},
			Spec: clv1alpha2.InstanceSpec{
				Template: clv1alpha2.GenericRef{Name: testTemplate, Namespace: testWorkspaceNamespace},
				Running:  true,
			},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ws, tmpl).Build()
		validator := &webhook.InstanceValidator{Client: fakeClient}
		warnings, err := validator.ValidateCreate(ctx, inst)
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(ContainSubstring("quota exceeded: nvidia.com/gpu"))
		Expect(warnings).To(BeEmpty())
	})

	It("should warn if template is missing for an instance", func() {
		ws := &clv1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{Name: testWorkspace},
			Spec: clv1alpha1.WorkspaceSpec{
				Quota: apicommon.WorkspaceResourceQuota{
					Instances: 2,
					ResourceSpec: apicommon.ResourceSpec{
						CPU:    4,
						Memory: resource.MustParse("8Gi"),
					},
				},
			},
		}
		inst := &clv1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testNewInstance,
				Namespace: testTenantNamespace,
				Labels:    map[string]string{forge.LabelWorkspaceKey: testWorkspace},
			},
			Spec: clv1alpha2.InstanceSpec{
				Template: clv1alpha2.GenericRef{Name: testMissingTemplate, Namespace: testWorkspaceNamespace},
				Running:  true,
			},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ws).Build()
		validator := &webhook.InstanceValidator{Client: fakeClient}
		warnings, err := validator.ValidateCreate(ctx, inst)
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(ContainSubstring("failed to get instance template"))
		Expect(warnings).To(BeEmpty())
	})

	Context("LocalVM PVC Access Validation", func() {
		var (
			tenant          *clv1alpha2.Tenant
			localVMTemplate *clv1alpha2.Template
			instance        *clv1alpha2.Instance
		)

		BeforeEach(func() {
			req := admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					UserInfo: authenticationv1.UserInfo{
						Username: testTenant,
					},
				},
			}
			ctx = admission.NewContextWithRequest(ctx, req)

			tenant = &clv1alpha2.Tenant{
				ObjectMeta: metav1.ObjectMeta{Name: testTenant},
				Spec: clv1alpha2.TenantSpec{
					PersonalWorkspace: &apicommon.WorkspaceResourceQuota{
						Instances: 10,
						ResourceSpec: apicommon.ResourceSpec{
							CPU:    10,
							Memory: resource.MustParse("20Gi"),
						},
					},
					Workspaces: []clv1alpha2.TenantWorkspaceEntry{
						{Name: "my-shared-ws", Role: clv1alpha2.User},
					},
				},
			}

			localVMTemplate = &clv1alpha2.Template{
				ObjectMeta: metav1.ObjectMeta{Name: "personal-template", Namespace: testTenantNamespace},
				Spec: clv1alpha2.TemplateSpec{
					EnvironmentList: []clv1alpha2.Environment{{
						Name:            "localvm-env",
						EnvironmentType: clv1alpha2.ClassLocalVM,
						Resources: clv1alpha2.EnvironmentResources{
							ResourceSpec: apicommon.ResourceSpec{
								CPU:    2,
								Memory: resource.MustParse("2Gi"),
							},
						},
					}},
					WorkspaceRef: clv1alpha2.GenericRef{Name: "personal"},
				},
			}

			instance = &clv1alpha2.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testNewInstance,
					Namespace: testTenantNamespace,
					Labels:    map[string]string{forge.LabelWorkspaceKey: "personal"},
				},
				Spec: clv1alpha2.InstanceSpec{
					Template: clv1alpha2.GenericRef{Name: "personal-template", Namespace: testTenantNamespace},
					Tenant:   clv1alpha2.GenericRef{Name: testTenant},
					Running:  true,
				},
			}
		})

		It("should allow LocalVM creation using PVC from the same tenant's personal namespace", func() {
			localVMTemplate.Spec.EnvironmentList[0].Image = testTenantNamespace + "/my-pvc"

			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tenant, localVMTemplate).Build()
			validator := &webhook.InstanceValidator{Client: fakeClient, PublicNamespace: "cldprog-5-block-vms-tests"}

			warnings, err := validator.ValidateCreate(ctx, instance)
			Expect(err).To(BeNil())
			Expect(warnings).To(BeEmpty())
		})

		It("should deny LocalVM creation using PVC from another tenant's personal namespace", func() {
			otherTenantNamespace := "tenant-other"
			localVMTemplate.Spec.EnvironmentList[0].Image = otherTenantNamespace + "/my-pvc"

			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tenant, localVMTemplate).Build()
			validator := &webhook.InstanceValidator{Client: fakeClient, PublicNamespace: "cldprog-5-block-vms-tests"}

			warnings, err := validator.ValidateCreate(ctx, instance)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("uses a PVC from an unauthorized namespace"))
			Expect(warnings).To(BeEmpty())
		})

		It("should allow LocalVM creation using PVC from an allowed workspace namespace", func() {
			workspaceNamespace := "workspace-my-shared-ws"
			localVMTemplate.Spec.EnvironmentList[0].Image = workspaceNamespace + "/my-pvc"

			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tenant, localVMTemplate).Build()
			validator := &webhook.InstanceValidator{Client: fakeClient, PublicNamespace: "cldprog-5-block-vms-tests"}

			warnings, err := validator.ValidateCreate(ctx, instance)
			Expect(err).To(BeNil())
			Expect(warnings).To(BeEmpty())
		})

		It("should allow LocalVM creation using PVC from the public namespace", func() {
			publicNamespace := "cldprog-5-block-vms-tests"
			localVMTemplate.Spec.EnvironmentList[0].Image = publicNamespace + "/my-pvc"

			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tenant, localVMTemplate).Build()
			validator := &webhook.InstanceValidator{Client: fakeClient, PublicNamespace: "cldprog-5-block-vms-tests"}

			warnings, err := validator.ValidateCreate(ctx, instance)
			Expect(err).To(BeNil())
			Expect(warnings).To(BeEmpty())
		})

		It("should bypass PVC check for templates in shared workspaces", func() {
			// Template points to another tenant's PVC, but it's part of a shared workspace
			otherTenantNamespace := "tenant-other"
			localVMTemplate.Spec.EnvironmentList[0].Image = otherTenantNamespace + "/my-pvc"
			localVMTemplate.Spec.WorkspaceRef.Name = testWorkspace

			ws := &clv1alpha1.Workspace{
				ObjectMeta: metav1.ObjectMeta{Name: testWorkspace},
				Spec: clv1alpha1.WorkspaceSpec{
					Quota: apicommon.WorkspaceResourceQuota{
						Instances: 10,
						ResourceSpec: apicommon.ResourceSpec{
							CPU:    10,
							Memory: resource.MustParse("20Gi"),
						},
					},
				},
			}

			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ws, localVMTemplate).Build()
			validator := &webhook.InstanceValidator{Client: fakeClient, PublicNamespace: "cldprog-5-block-vms-tests"}

			warnings, err := validator.ValidateCreate(ctx, instance)
			Expect(err).To(BeNil())
			Expect(warnings).To(BeEmpty())
		})
	})
})
