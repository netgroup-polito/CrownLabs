// Copyright 2020-2021 Politecnico di Torino
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

package v1alpha1

import (
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// ConvertTo converts a v1alpha1.Tenant to a v1alpha2.Tenant.
func (src *Tenant) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha2.Tenant)

	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.CreateSandbox = src.Spec.CreateSandbox
	dst.Spec.Email = src.Spec.Email
	dst.Spec.FirstName = src.Spec.FirstName
	dst.Spec.LastName = src.Spec.LastName
	dst.Spec.PublicKeys = src.Spec.PublicKeys
	dst.Spec.Workspaces = make([]v1alpha2.TenantWorkspaceEntry, len(src.Spec.Workspaces))
	dst.Spec.Quota = src.Spec.Quota
	for i, w := range src.Spec.Workspaces {
		dst.Spec.Workspaces[i].Name = w.WorkspaceRef.Name
		dst.Spec.Workspaces[i].Role = w.Role
	}

	dst.Status = src.Status

	return nil
}

// ConvertFrom creates a v1alpha1.Tenant from a v1alpha2.Tenant.
//nolint:stylecheck,revive // allow different receiver name as it is clearer (as in kubebuilder example).
func (dst *Tenant) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha2.Tenant)

	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.CreateSandbox = src.Spec.CreateSandbox
	dst.Spec.Email = src.Spec.Email
	dst.Spec.FirstName = src.Spec.FirstName
	dst.Spec.LastName = src.Spec.LastName
	dst.Spec.PublicKeys = src.Spec.PublicKeys
	dst.Spec.Workspaces = make([]TenantWorkspaceEntry, len(src.Spec.Workspaces))
	dst.Spec.Quota = src.Spec.Quota
	for i, w := range src.Spec.Workspaces {
		dst.Spec.Workspaces[i].WorkspaceRef.Name = w.Name
		dst.Spec.Workspaces[i].Role = w.Role
	}

	dst.Status = src.Status

	return nil
}
