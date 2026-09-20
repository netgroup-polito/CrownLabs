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

// Package webhook implements the webhook handlers for instance snapshot resources.
package webhook

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// InstanceSnapshotValidator implements a validating webhook for InstanceSnapshot resources.
//
// It is the only barrier constraining which disk a snapshot may read: the clone is carried out by
// the instance-operator service account, which holds cluster-wide permissions on
// cdi.kubevirt.io/datavolumes/source, so by the time the controller runs the identity of the
// requester is gone and neither RBAC nor CDI can tell on whose behalf they are acting.
//
// The destination of the snapshot is deliberately left out: it is the namespace the InstanceSnapshot
// itself is created in, which the API server already gates through RBAC. Keeping it there is also
// what lets administrators publish into the public catalog through a plain RoleBinding.
type InstanceSnapshotValidator struct {
	admission.CustomValidator
	Client                  client.Client
	PublicSnapshotNamespace string
	BypassGroups            []string
}

// ValidateCreate validates the scope of a newly created InstanceSnapshot.
func (isv *InstanceSnapshotValidator) ValidateCreate(
	ctx context.Context,
	obj runtime.Object,
) (admission.Warnings, error) {
	snapshot, ok := obj.(*clv1alpha2.InstanceSnapshot)
	if !ok {
		return nil, fmt.Errorf("expected InstanceSnapshot resource but got %T", obj)
	}

	return nil, isv.validateSource(ctx, snapshot)
}

// ValidateUpdate re-validates the scope whenever the source reference changes, catching an already
// admitted snapshot being repointed at somebody else's instance before the controller reads it.
func (isv *InstanceSnapshotValidator) ValidateUpdate(
	ctx context.Context,
	oldObj, newObj runtime.Object,
) (admission.Warnings, error) {
	oldSnapshot, ok := oldObj.(*clv1alpha2.InstanceSnapshot)
	if !ok {
		return nil, fmt.Errorf("expected InstanceSnapshot resource but got %T", oldObj)
	}

	newSnapshot, ok := newObj.(*clv1alpha2.InstanceSnapshot)
	if !ok {
		return nil, fmt.Errorf("expected InstanceSnapshot resource but got %T", newObj)
	}

	if oldSnapshot.Spec.Instance == newSnapshot.Spec.Instance {
		return nil, nil
	}

	return nil, isv.validateSource(ctx, newSnapshot)
}

// validateSource checks that the actor is entitled to read the disk of the referenced instance.
func (isv *InstanceSnapshotValidator) validateSource(
	ctx context.Context,
	snapshot *clv1alpha2.InstanceSnapshot,
) error {
	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get admission request from context: %w", err)
	}

	if utils.MatchOneInStringSlices(isv.BypassGroups, req.UserInfo.Groups) {
		return nil
	}

	// The reference is read with no defaulting.
	srcNamespace := snapshot.Spec.Instance.Namespace
	if srcNamespace == "" {
		return fmt.Errorf("spec.instanceRef.namespace must be set explicitly")
	}

	tenant := &clv1alpha2.Tenant{}
	if err := isv.Client.Get(ctx, types.NamespacedName{Name: req.UserInfo.Username}, tenant); err != nil {
		return fmt.Errorf("failed to get tenant %s: %w", req.UserInfo.Username, err)
	}

	if !forge.TenantCanReadNamespace(tenant, srcNamespace, isv.PublicSnapshotNamespace) {
		return fmt.Errorf("cannot snapshot instance %q from namespace %q: the source must belong to your own "+
			"tenant, to a workspace you are subscribed to, or to the public snapshot catalog",
			snapshot.Spec.Instance.Name, srcNamespace)
	}

	return nil
}
