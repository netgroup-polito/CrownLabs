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

package tenant

import (
	"context"

	"github.com/go-logr/logr"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

func (r *Reconciler) enforceTenantBaseLabels(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
) error {
	if tn.Labels == nil {
		tn.Labels = make(map[string]string, 1)
	}

	// Use forge function for target label
	tn.Labels = forge.UpdateTenantResourceCommonLabels(tn.Labels, r.TargetLabel)

	// Add specific tenant identity labels
	tn.Labels["crownlabs.polito.it/first-name"] = forge.CleanTenantName(tn.Spec.FirstName)
	tn.Labels["crownlabs.polito.it/last-name"] = forge.CleanTenantName(tn.Spec.LastName)

	if err := r.enforcePreservingStatus(ctx, log, tn); err != nil {
		log.Error(err, "Error updating labels for tenant", "tenant", tn.Name)
		return err
	}
	log.Info("Updated basic tenant labels")

	return nil
}
