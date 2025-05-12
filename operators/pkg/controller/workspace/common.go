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

package workspace

import (
	"context"

	"github.com/go-logr/logr"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

func (r *Reconciler) enforcePreservingStatus(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha1.Workspace,
	mutating func(*v1alpha1.Workspace) *v1alpha1.Workspace,
) error {
	// the update function will overwrite the status, so we need to save it
	// and restore it after the update
	prevStatus := tn.Status
	if err := utils.PatchObject(ctx, r.Client, tn, mutating); err != nil {
		log.Error(err, "Error when updating workspace")
		return err
	}
	tn.Status = prevStatus

	return nil
}
