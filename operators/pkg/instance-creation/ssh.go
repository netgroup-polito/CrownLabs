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

package instance_creation

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// GetPublicKeys extracts and returns the set of public keys associated with a
// given tenant, along with the ones of the tenants having Manager role in the
// corresponding workspace.
func GetPublicKeys(ctx context.Context, c client.Reader, tenantRef, templateRef crownlabsv1alpha2.GenericRef, publicKeys *[]string) error {
	tenant := crownlabsv1alpha1.Tenant{}
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: tenantRef.Namespace,
		Name:      tenantRef.Name,
	}, &tenant); err != nil {
		return err
	}

	*publicKeys = append(*publicKeys, tenant.Spec.PublicKeys...)

	template := crownlabsv1alpha2.Template{}
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: templateRef.Namespace,
		Name:      templateRef.Name,
	}, &template); err != nil {
		return err
	}

	label := map[string]string{crownlabsv1alpha1.WorkspaceLabelPrefix + template.Spec.WorkspaceRef.Name: "manager"}

	var managers crownlabsv1alpha1.TenantList
	if err := c.List(context.Background(), &managers, client.MatchingLabels(label)); apierrors.IsNotFound(err) {
		// if there are no managers in this workspace there's nothing to do
		return nil
	} else if err != nil {
		return err
	}

	for i := range managers.Items {
		// avoid duplicates
		if managers.Items[i].Name != tenant.Name {
			*publicKeys = append(*publicKeys, managers.Items[i].Spec.PublicKeys...)
		}
	}

	return nil
}
