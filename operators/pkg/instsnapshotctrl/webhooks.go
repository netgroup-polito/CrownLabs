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

package instsnapshotctrl

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

// InstanceSnapshotWebhook handles admission for InstanceSnapshot.
type InstanceSnapshotWebhook struct {
	Client                  client.Client
	SnapshotPublicNamespace string
}

// SetupWebhookWithManager registers the webhook with the manager.
func (w *InstanceSnapshotWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	w.Client = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(&clv1alpha2.InstanceSnapshot{}).
		WithDefaulter(w).
		WithValidator(w).
		Complete()
}

var _ webhook.CustomDefaulter = &InstanceSnapshotWebhook{}
var _ webhook.CustomValidator = &InstanceSnapshotWebhook{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type.
func (w *InstanceSnapshotWebhook) Default(_ context.Context, obj runtime.Object) error {
	snapshot, ok := obj.(*clv1alpha2.InstanceSnapshot)
	if !ok {
		return fmt.Errorf("expected an InstanceSnapshot object but got %T", obj)
	}

	// Resolve the destination namespace based on the requested scope.
	switch snapshot.Spec.Destination.Scope {
	case clv1alpha2.PrivateScope:
		snapshot.Spec.Destination.Namespace = snapshot.Spec.Source.InstanceRef.Namespace // Assuming same namespace as source for private
	case clv1alpha2.WorkspaceScope:
		if snapshot.Spec.Source.WorkspaceRef.Name != "" {
			snapshot.Spec.Destination.Namespace = forge.GetWorkspaceNamespaceName(&clv1alpha1.Workspace{
				ObjectMeta: metav1.ObjectMeta{Name: snapshot.Spec.Source.WorkspaceRef.Name},
			})
		}
	case clv1alpha2.PublicScope:
		snapshot.Spec.Destination.Namespace = w.SnapshotPublicNamespace
	}

	return nil
}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type.
func (w *InstanceSnapshotWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	snapshot, ok := obj.(*clv1alpha2.InstanceSnapshot)
	if !ok {
		return nil, fmt.Errorf("expected an InstanceSnapshot object but got %T", obj)
	}

	// 1. Check authorization (mock)
	// Normal user can only do private snapshot for their own instance.
	// Workspace manager can do workspace or private snapshots.
	// Cluster admin can do anything.
	// We'll mock the authorization check as "allowed" for now, based on the implementation plan.

	// 2. Validate source Instance exists and is powered off.
	var instance clv1alpha2.Instance
	if err := w.Client.Get(ctx, client.ObjectKey{
		Namespace: snapshot.Spec.Source.InstanceRef.Namespace,
		Name:      snapshot.Spec.Source.InstanceRef.Name,
	}, &instance); err != nil {
		return nil, fmt.Errorf("failed to get source instance: %w", err)
	}

	if instance.Spec.Running {
		return nil, fmt.Errorf("instance %s must be powered off before snapshotting", instance.Name)
	}

	// 3. Ensure destination namespace is set
	if snapshot.Spec.Destination.Namespace == "" {
		return nil, fmt.Errorf("destination namespace could not be resolved for scope %s", snapshot.Spec.Destination.Scope)
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator.
func (w *InstanceSnapshotWebhook) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oldSnapshot := oldObj.(*clv1alpha2.InstanceSnapshot)
	newSnapshot := newObj.(*clv1alpha2.InstanceSnapshot)

	// Enforce immutable spec (except displayName which might be allowed, but we freeze source and destination).
	if oldSnapshot.Spec.Destination.Namespace != newSnapshot.Spec.Destination.Namespace {
		return nil, fmt.Errorf("spec.destination.namespace is immutable")
	}

	if oldSnapshot.Spec.Destination.Scope != newSnapshot.Spec.Destination.Scope {
		return nil, fmt.Errorf("spec.destination.scope is immutable")
	}

	if oldSnapshot.Spec.Source.PVCName != newSnapshot.Spec.Source.PVCName {
		return nil, fmt.Errorf("spec.source.pvcName is immutable")
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator.
func (w *InstanceSnapshotWebhook) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}
