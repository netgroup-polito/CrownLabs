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
	"errors"
	"fmt"
	"slices"

	"github.com/go-logr/logr"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/mail"
)

const (
	// WorkspaceAddedMailTemplatePath is the path, relative to the mail template directory,
	// of the template used to notify a tenant that it has been added to a workspace.
	WorkspaceAddedMailTemplatePath = "tenant_workspace_added_notification.yaml"
	// WorkspaceRemovedMailTemplatePath is the path, relative to the mail template directory,
	// of the template used to notify a tenant that it has been removed from a workspace.
	WorkspaceRemovedMailTemplatePath = "tenant_workspace_removed_notification.yaml"
)

// Mailer sends templated CrownLabs emails. It is implemented by *mail.Client.
type Mailer interface {
	SendCrownLabsMail(ctx context.Context, emailContentTemplatePath string, ph *mail.Placeholders) error
}

// workspaceChange is a change in the workspaces of a tenant, notified via email.
type workspaceChange string

const (
	workspaceAdded   workspaceChange = "added"
	workspaceRemoved workspaceChange = "removed"
)

// syncWorkspaceNotifications emails the tenant about each workspace it has been added to or
// removed from, and records the notified workspaces in the NotifiedWorkspacesAnnotation:
// a workspace is added once the tenant has been told it joined it, and dropped once the
// tenant has been told it left it.
//
// The annotation is kept up to date even when no Mailer is configured (i.e., notifications
// are disabled): this way, enabling notifications never sends emails for changes that happened
// in the meantime. A failed email leaves that workspace unchanged in the annotation, so it is
// sent again at the next reconcile.
func (r *Reconciler) syncWorkspaceNotifications(
	ctx context.Context,
	log logr.Logger,
	tn *clv1alpha2.Tenant,
) error {
	raw, tracked := tn.Annotations[forge.NotifiedWorkspacesAnnotation]
	if !tracked {
		// The tenant was created before workspace notifications were introduced (new
		// tenants get an empty annotation from the mutating webhook): record its current
		// workspaces without sending any email, as it is already aware of them.
		var current []string
		for _, ws := range tn.Spec.Workspaces {
			if r.isNotifiableRole(ws) {
				current = append(current, ws.Name)
			}
		}
		log.Info("Recording current workspaces as already notified", "workspaces", current, "for tenant", tn.Name)
		return r.patchNotifiedWorkspaces(ctx, log, tn, current)
	}

	notified := parseWorkspaceList(raw)

	var (
		next []string
		errs []error
	)

	// Removals: notified workspaces that are no longer in the spec. Failing workspaces
	// are not removals, as they are still in the spec.
	for _, name := range notified {
		if slices.ContainsFunc(tn.Spec.Workspaces, func(ws clv1alpha2.TenantWorkspaceEntry) bool {
			return ws.Name == name
		}) {
			next = append(next, name)
			continue
		}

		if err := r.notifyWorkspaceChange(ctx, log, tn, workspaceRemoved, clv1alpha2.TenantWorkspaceEntry{Name: name}); err != nil {
			errs = append(errs, err)
			next = append(next, name) // kept, so that the removal is notified again
		}
	}

	// Additions: workspaces in the spec that deserve a notification and have not been notified yet.
	for _, ws := range tn.Spec.Workspaces {
		if slices.Contains(notified, ws.Name) || !r.isNotifiableRole(ws) ||
			slices.Contains(tn.Status.FailingWorkspaces, ws.Name) {
			continue
		}

		if err := r.notifyWorkspaceChange(ctx, log, tn, workspaceAdded, ws); err != nil {
			errs = append(errs, err)
			continue // not recorded, so that the addition is notified again
		}
		next = append(next, ws.Name)
	}

	if formatWorkspaceList(next) != raw {
		if err := r.patchNotifiedWorkspaces(ctx, log, tn, next); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// isNotifiableRole returns whether being part of the given workspace deserves a notification:
// candidates are not enrolled yet, and base workspaces are enforced on every tenant.
func (r *Reconciler) isNotifiableRole(ws clv1alpha2.TenantWorkspaceEntry) bool {
	return ws.Role != clv1alpha2.Candidate && !slices.Contains(r.BaseWorkspaces, ws.Name)
}

// notifyWorkspaceChange emails the tenant about a workspace it has been added to or removed from,
// unless notifications are disabled. An error means that the email has not been sent, and is to
// be sent again at the next reconcile.
func (r *Reconciler) notifyWorkspaceChange(
	ctx context.Context,
	log logr.Logger,
	tn *clv1alpha2.Tenant,
	change workspaceChange,
	ws clv1alpha2.TenantWorkspaceEntry,
) error {
	log = log.WithValues("workspace", ws.Name, "change", change)

	if r.Mailer == nil {
		log.Info("Workspace notifications disabled, recording the change without sending")
		return nil
	}

	if err := r.sendWorkspaceNotification(ctx, tn, change, ws); err != nil {
		log.Error(err, "Failed sending workspace notification")
		tnOpinternalErrors.WithLabelValues("tenant", "workspace-notification").Inc()
		return fmt.Errorf("failed sending %s notification for workspace %s: %w", change, ws.Name, err)
	}

	log.Info("Workspace notification sent")
	return nil
}

func (r *Reconciler) sendWorkspaceNotification(
	ctx context.Context,
	tn *clv1alpha2.Tenant,
	change workspaceChange,
	ws clv1alpha2.TenantWorkspaceEntry,
) error {
	prettyName := ws.Name
	var workspace clv1alpha1.Workspace
	switch err := r.Get(ctx, types.NamespacedName{Name: ws.Name}, &workspace); {
	case err == nil:
		if workspace.Spec.PrettyName != "" {
			prettyName = workspace.Spec.PrettyName
		}
	case kerrors.IsNotFound(err) && change == workspaceRemoved:
		// Deleting a workspace also removes all its tenants, so it might be gone already:
		// fall back to the workspace name.
	default:
		return fmt.Errorf("failed retrieving workspace: %w", err)
	}

	templatePath := WorkspaceAddedMailTemplatePath
	if change == workspaceRemoved {
		templatePath = WorkspaceRemovedMailTemplatePath
	}

	return r.Mailer.SendCrownLabsMail(ctx, templatePath, &mail.Placeholders{
		TenantName:          tn.Name,
		TenantEmail:         tn.Spec.Email,
		WorkspacePrettyName: prettyName,
		WorkspaceRole:       string(ws.Role),
	})
}

func (r *Reconciler) patchNotifiedWorkspaces(
	ctx context.Context,
	log logr.Logger,
	tn *clv1alpha2.Tenant,
	workspaces []string,
) error {
	value := formatWorkspaceList(workspaces)
	if err := r.enforcePreservingStatus(ctx, log, tn, func(t *clv1alpha2.Tenant) *clv1alpha2.Tenant {
		if t.Annotations == nil {
			t.Annotations = map[string]string{}
		}
		t.Annotations[forge.NotifiedWorkspacesAnnotation] = value
		return t
	}); err != nil {
		log.Error(err, "Error updating notified workspaces annotation")
		return fmt.Errorf("error updating notified workspaces annotation: %w", err)
	}
	return nil
}
