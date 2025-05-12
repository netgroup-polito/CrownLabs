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
	"maps"

	"github.com/go-logr/logr"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

func (r *Reconciler) enforceTenantBaseLabels(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
) error {
	labels := make(map[string]string)
	maps.Copy(labels, tn.Labels)

	labels = forge.TenantLabels(labels, tn, r.TargetLabel)

	if err := r.enforcePreservingStatus(ctx, log, tn, func(t *v1alpha2.Tenant) *v1alpha2.Tenant {
		t.Labels = labels
		return t
	}); err != nil {
		log.Error(err, "Error updating labels for tenant")
		return err
	}
	log.Info("Updated basic tenant labels")

	return nil
}
