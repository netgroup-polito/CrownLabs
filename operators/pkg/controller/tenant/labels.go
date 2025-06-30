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

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"k8s.io/klog/v2"
)

func (r *Reconciler) updateTenantBaseLabels(
	ctx context.Context,
	log *klog.Logger,
	tn *v1alpha2.Tenant,
) error {
	if tn.Labels == nil {
		tn.Labels = make(map[string]string, 1)
	}

	tn.Labels[r.TargetLabel.GetKey()] = r.TargetLabel.GetValue()

	tn.Labels["crownlabs.polito.it/first-name"] = cleanName(tn.Spec.FirstName)
	tn.Labels["crownlabs.polito.it/last-name"] = cleanName(tn.Spec.LastName)

	if err := r.updatePreservingStatus(ctx, tn); err != nil {
		klog.Errorf("Error updating labels for tenant %s: %v", tn.Name, err)
		return err
	}
	log.Info("Updated basic tenant labels")

	return nil
}
