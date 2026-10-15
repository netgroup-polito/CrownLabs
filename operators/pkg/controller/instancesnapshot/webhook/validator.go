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
	"reflect"
	"slices"

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
// The destination of the snapshot is the namespace the InstanceSnapshot itself is created in, which the
// API server already gates through RBAC. The public catalog is checked here as well: every tenant boots
// from it, so publishing is reserved to the members of PublisherGroup.
type InstanceSnapshotValidator struct {
	admission.CustomValidator
	Client                  client.Client
	PublicSnapshotNamespace string
	PublisherGroup          string
	BypassGroups            []string
}

// ValidateCreate validates the owner and the scope of a newly created InstanceSnapshot.
func (isv *InstanceSnapshotValidator) ValidateCreate(
	ctx context.Context,
	obj runtime.Object,
) (admission.Warnings, error) {
	snapshot, ok := obj.(*clv1alpha2.InstanceSnapshot)
	if !ok {
		return nil, fmt.Errorf("expected InstanceSnapshot resource but got %T", obj)
	}

	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get admission request from context: %w", err)
	}

	if utils.MatchOneInStringSlices(isv.BypassGroups, req.UserInfo.Groups) {
		return nil, nil
	}

	if err := validateOwner(snapshot, req.UserInfo.Username); err != nil {
		return nil, err
	}

	if err := isv.validateDestination(snapshot, req.UserInfo.Groups); err != nil {
		return nil, err
	}

	return nil, isv.validateSource(ctx, snapshot, req.UserInfo.Username)
}

// ValidateUpdate rejects every spec change after creation. Snapshot metadata and source references
// are consumed when the DataVolume is created and must remain consistent with the artifact.
// The tenant label is frozen as well, otherwise the owner checked at creation could be rewritten.
func (isv *InstanceSnapshotValidator) ValidateUpdate(
	_ context.Context,
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

	if !reflect.DeepEqual(oldSnapshot.Spec, newSnapshot.Spec) {
		return nil, fmt.Errorf("InstanceSnapshot spec is immutable")
	}

	if oldSnapshot.Labels[forge.LabelTenantKey] != newSnapshot.Labels[forge.LabelTenantKey] {
		return nil, fmt.Errorf("label %s is immutable", forge.LabelTenantKey)
	}

	return nil, nil
}

// validateOwner checks that the snapshot is labeled with the tenant creating it.
// Also, it makes sure a snapshot can only be created carrying the name of its
// creator, and the label can no longer change afterwards.
func validateOwner(snapshot *clv1alpha2.InstanceSnapshot, username string) error {
	owner, ok := snapshot.Labels[forge.LabelTenantKey]
	if !ok {
		return fmt.Errorf("label %s must be set to the tenant creating the snapshot", forge.LabelTenantKey)
	}

	if owner != username {
		return fmt.Errorf("label %s is %q, but the snapshot is being created by %q", forge.LabelTenantKey, owner, username)
	}

	return nil
}

// validateDestination checks that only publishers create snapshots in the public catalog.
func (isv *InstanceSnapshotValidator) validateDestination(snapshot *clv1alpha2.InstanceSnapshot, groups []string) error {
	if snapshot.Namespace != isv.PublicSnapshotNamespace {
		return nil
	}

	if isv.PublisherGroup == "" || !slices.Contains(groups, isv.PublisherGroup) {
		return fmt.Errorf("only snapshot publishers can create snapshots in the public catalog %q", isv.PublicSnapshotNamespace)
	}

	return nil
}

// validateSource checks that the requester is entitled to read the disk of the referenced instance.
func (isv *InstanceSnapshotValidator) validateSource(
	ctx context.Context,
	snapshot *clv1alpha2.InstanceSnapshot,
	username string,
) error {
	// The reference is read with no defaulting.
	srcNamespace := snapshot.Spec.Instance.Namespace
	if srcNamespace == "" {
		return fmt.Errorf("spec.instanceRef.namespace must be set explicitly")
	}

	tenant := &clv1alpha2.Tenant{}
	if err := isv.Client.Get(ctx, types.NamespacedName{Name: username}, tenant); err != nil {
		return fmt.Errorf("failed to get tenant %s: %w", username, err)
	}

	if !forge.TenantCanReadNamespace(tenant, srcNamespace, isv.PublicSnapshotNamespace) {
		return fmt.Errorf("cannot snapshot instance %q from namespace %q: the source must belong to your own "+
			"tenant, to a workspace you are subscribed to, or to the public snapshot catalog",
			snapshot.Spec.Instance.Name, srcNamespace)
	}

	return nil
}
