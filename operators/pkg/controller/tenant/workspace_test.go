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

package tenant_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var _ = Describe("Workspace management", func() {
	ws1 := &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ws1",
		},
		Spec: v1alpha1.WorkspaceSpec{
			AutoEnroll: v1alpha1.AutoenrollWithApproval,
		},
	}
	ws2 := &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ws2",
		},
	}
	bws1 := &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "base-ws1",
		},
	}

	BeforeEach(func() {
		addObjToObjectsList(ws1)
		addObjToObjectsList(ws2)
		addObjToObjectsList(bws1)
	})

	AfterEach(func() {
		removeObjFromObjectsList(ws1)
		removeObjFromObjectsList(ws2)
		removeObjFromObjectsList(bws1)
	})

	Context("When a workspace is present and valid", func() {
		Context("and the role is manager", func() {
			BeforeEach(func() {
				tnResource.Spec.Workspaces = []v1alpha2.TenantWorkspaceEntry{{
					Name: "ws1",
					Role: v1alpha2.Manager,
				}}
			})

			It("Should add the workspace label to the tenant", func() {
				tn := &v1alpha2.Tenant{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)

				Expect(tn.Labels).To(HaveKeyWithValue("crownlabs.polito.it/workspace-ws1", "manager"))
			})
		})

		Context("and the role is user", func() {
			BeforeEach(func() {
				tnResource.Spec.Workspaces = []v1alpha2.TenantWorkspaceEntry{{
					Name: "ws1",
					Role: v1alpha2.User,
				}}
			})

			It("Should add the workspace label to the tenant", func() {
				tn := &v1alpha2.Tenant{}
				DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)

				Expect(tn.Labels).To(HaveKeyWithValue("crownlabs.polito.it/workspace-ws1", "user"))
			})
		})

		Context("and the role is candidate", func() {
			Context("and the workspace is auto-enrollable", func() {
				BeforeEach(func() {
					tnResource.Spec.Workspaces = []v1alpha2.TenantWorkspaceEntry{{
						Name: "ws1",
						Role: v1alpha2.Candidate,
					}}
				})

				It("Should add the workspace label to the tenant", func() {
					tn := &v1alpha2.Tenant{}
					DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)

					Expect(tn.Labels).To(HaveKeyWithValue("crownlabs.polito.it/workspace-ws1", "candidate"))
				})
			})

			Context("and the workspace is not auto-enrollable", func() {
				BeforeEach(func() {
					tnResource.Spec.Workspaces = []v1alpha2.TenantWorkspaceEntry{{
						Name: "ws2",
						Role: v1alpha2.Candidate,
					}}
				})

				It("Should not add the workspace label to the tenant", func() {
					tn := &v1alpha2.Tenant{}
					DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)

					Expect(tn.Labels).NotTo(HaveKey("crownlabs.polito.it/workspace-ws2"))
				})

				It("Should add the workspace to the failing workspaces list", func() {
					tn := &v1alpha2.Tenant{}
					DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)

					Expect(tn.Status.FailingWorkspaces).To(ContainElement("ws2"))
				})
			})
		})
	})

	Context("When a workspace is present but not valid", func() {
		BeforeEach(func() {
			tnResource.Spec.Workspaces = []v1alpha2.TenantWorkspaceEntry{{
				Name: "ws3", // ws3 does not exist
				Role: v1alpha2.Manager,
			}}
		})

		It("Should not add the workspace label to the tenant", func() {
			tn := &v1alpha2.Tenant{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)

			Expect(tn.Labels).NotTo(HaveKey("crownlabs.polito.it/workspace-ws3"))
		})

		It("Should add the workspace to the failing workspaces list", func() {
			tn := &v1alpha2.Tenant{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)

			Expect(tn.Status.FailingWorkspaces).To(ContainElement("ws3"))
		})
	})

	Context("When a workspace is not present but the related label exists", func() {
		BeforeEach(func() {
			tnResource.Labels["crownlabs.polito.it/workspace-ws1"] = "manager"
			tnResource.Spec.Workspaces = []v1alpha2.TenantWorkspaceEntry{}
		})

		It("Should remove the workspace label from the tenant", func() {
			tn := &v1alpha2.Tenant{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)

			Expect(tn.Labels).NotTo(HaveKey("crownlabs.polito.it/workspace-ws1"))
		})
	})

	Context("When no workspaces are present", func() {
		It("Should add the no-workspaces label to the tenant", func() {
			tn := &v1alpha2.Tenant{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)

			Expect(tn.Labels).To(HaveKeyWithValue("crownlabs.polito.it/no-workspaces", "true"))
		})
	})

	Context("When only default workspaces are present", func() {
		BeforeEach(func() {
			tnResource.Spec.Workspaces = []v1alpha2.TenantWorkspaceEntry{{
				Name: "base-ws1",
				Role: v1alpha2.Manager,
			}}
		})

		It("Should add the related label to the tenant", func() {
			tn := &v1alpha2.Tenant{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)

			Expect(tn.Labels).To(HaveKeyWithValue("crownlabs.polito.it/workspace-base-ws1", "manager"))
		})

		It("Should add the no-workspaces label to the tenant", func() {
			tn := &v1alpha2.Tenant{}
			DoesEventuallyExists(ctx, cl, client.ObjectKey{Name: tnName}, tn, BeTrue(), timeout, interval)

			Expect(tn.Labels).To(HaveKeyWithValue("crownlabs.polito.it/no-workspaces", "true"))
		})
	})
})
