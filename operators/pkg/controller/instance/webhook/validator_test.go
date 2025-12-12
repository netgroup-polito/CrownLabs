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

package webhook_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
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
		ws := &v1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{Name: testWorkspace},
			Spec: v1alpha1.WorkspaceSpec{
				Quota: v1alpha1.WorkspaceResourceQuota{
					Instances: 2,
					CPU:       resource.MustParse("4"),
					Memory:    resource.MustParse("8Gi"),
				},
			},
		}
		tmpl := &v1alpha2.Template{
			ObjectMeta: metav1.ObjectMeta{Name: testTemplate, Namespace: testWorkspaceNamespace},
			Spec: v1alpha2.TemplateSpec{
				EnvironmentList: []v1alpha2.Environment{{
					Name: testEnvironment,
					Resources: v1alpha2.EnvironmentResources{
						CPU:    2,
						Memory: resource.MustParse("2Gi"),
					},
				}},
			},
		}
		inst := &v1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{Name: testNewInstance, Namespace: testTenantNamespace},
			Spec: v1alpha2.InstanceSpec{
				Template: v1alpha2.GenericRef{Name: testTemplate, Namespace: testWorkspaceNamespace},
			},
		}
		// Add another instance to exceed quota
		otherInst := &v1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{Name: testExistingInstance, Namespace: testTenantNamespace, Labels: map[string]string{forge.LabelWorkspaceKey: testWorkspace}},
			Spec: v1alpha2.InstanceSpec{
				Template: v1alpha2.GenericRef{Name: testTemplate, Namespace: testWorkspaceNamespace},
			},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ws, tmpl, otherInst).Build()
		validator := &webhook.InstanceValidator{Client: fakeClient}
		warnings, err := validator.ValidateCreate(ctx, inst)
		Expect(err).To(BeNil())
		Expect(warnings).To(BeEmpty())
	})

	It("should deny creation when quota exceeded", func() {
		ws := &v1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{Name: testWorkspace},
			Spec: v1alpha1.WorkspaceSpec{
				Quota: v1alpha1.WorkspaceResourceQuota{
					Instances: 1,
					CPU:       resource.MustParse("2"),
					Memory:    resource.MustParse("2Gi"),
				},
			},
		}
		tmpl := &v1alpha2.Template{
			ObjectMeta: metav1.ObjectMeta{Name: testTemplate, Namespace: testWorkspaceNamespace},
			Spec: v1alpha2.TemplateSpec{
				EnvironmentList: []v1alpha2.Environment{{
					Name: testEnvironment,
					Resources: v1alpha2.EnvironmentResources{
						CPU:    2,
						Memory: resource.MustParse("2Gi"),
					},
				}},
			},
		}
		inst := &v1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{Name: testNewInstance, Namespace: testTenantNamespace},
			Spec: v1alpha2.InstanceSpec{
				Template: v1alpha2.GenericRef{Name: testTemplate, Namespace: testWorkspaceNamespace},
			},
		}
		// Add another instance to exceed quota
		otherInst := &v1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{Name: testExistingInstance, Namespace: testTenantNamespace, Labels: map[string]string{forge.LabelWorkspaceKey: testWorkspace}},
			Spec: v1alpha2.InstanceSpec{
				Template: v1alpha2.GenericRef{Name: testTemplate, Namespace: testWorkspaceNamespace},
			},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ws, tmpl, otherInst).Build()
		validator := &webhook.InstanceValidator{Client: fakeClient}
		warnings, err := validator.ValidateCreate(ctx, inst)
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(ContainSubstring("quota exceeded"))
		Expect(warnings).To(BeEmpty())
	})

	It("should warn if template is missing for an instance", func() {
		ws := &v1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{Name: testWorkspace},
			Spec: v1alpha1.WorkspaceSpec{
				Quota: v1alpha1.WorkspaceResourceQuota{
					Instances: 2,
					CPU:       resource.MustParse("4"),
					Memory:    resource.MustParse("8Gi"),
				},
			},
		}
		inst := &v1alpha2.Instance{
			ObjectMeta: metav1.ObjectMeta{Name: testNewInstance, Namespace: testTenantNamespace},
			Spec: v1alpha2.InstanceSpec{
				Template: v1alpha2.GenericRef{Name: testMissingTemplate, Namespace: testWorkspaceNamespace},
			},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ws).Build()
		validator := &webhook.InstanceValidator{Client: fakeClient}
		warnings, err := validator.ValidateCreate(ctx, inst)
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(ContainSubstring("template missing"))
		Expect(warnings).To(BeEmpty())
	})
})
