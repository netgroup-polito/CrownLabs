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

package tenant

import (
	"context"
	"slices"
	"strings"

	"github.com/go-logr/logr"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

func (r *Reconciler) enforcePreservingStatus(
	ctx context.Context,
	log logr.Logger,
	tn *clv1alpha2.Tenant,
	mutating func(*clv1alpha2.Tenant) *clv1alpha2.Tenant,
) error {
	// the update function will overwrite the status, so we need to save it
	// and restore it after the update
	prevStatus := tn.Status
	if err := utils.PatchObject(ctx, r.Client, tn, mutating); err != nil {
		log.Error(err, "Error when updating tenant")
		return err
	}
	tn.Status = prevStatus

	return nil
}

// notifiedWorkspaceAnnotation returns the key of the annotation recording that the tenant
// has been notified about being added to the given workspace.
func notifiedWorkspaceAnnotation(workspace string) string {
	return forge.NotifiedWorkspaceAnnotationPrefix + workspace
}

// notifiedWorkspaces returns the sorted names of the workspaces the tenant has been notified about,
// as recorded in the given annotations.
func notifiedWorkspaces(annotations map[string]string) []string {
	var names []string
	for key := range annotations {
		if name, ok := strings.CutPrefix(key, forge.NotifiedWorkspaceAnnotationPrefix); ok && name != "" {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}
