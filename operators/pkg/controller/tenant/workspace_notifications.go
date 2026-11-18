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
	"maps"
	"slices"
	"time"

	"github.com/go-logr/logr"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
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
// removed from, and records the notified workspaces with one annotation each (see
// forge.NotifiedWorkspaceAnnotationPrefix): the annotation is created, with the current
// timestamp, once the tenant has been told it joined the workspace, and deleted once the
// tenant has been told it left it.
//
// The annotations are kept up to date in the same way even when no Mailer is configured (i.e.,
// notifications are disabled).
func (r *Reconciler) syncWorkspaceNotifications(
	ctx context.Context,
	log logr.Logger,
	tn *clv1alpha2.Tenant,
) error {
	notified := notifiedWorkspaces(tn.Annotations)

	var (
		set    = map[string]string{}
		remove []string
		errs   []error
	)

	// Removals: notified workspaces that are no longer in the spec. Failing workspaces
	// are not removals, as they are still in the spec.
	for _, name := range notified {
		if slices.ContainsFunc(tn.Spec.Workspaces, func(ws clv1alpha2.TenantWorkspaceEntry) bool {
			return ws.Name == name
		}) {
			continue
		}

		if err := r.notifyWorkspaceChange(ctx, log, tn, workspaceRemoved, clv1alpha2.TenantWorkspaceEntry{Name: name}); err != nil {
			errs = append(errs, err)
			continue // the annotation is kept, so that the removal is notified again
		}
		remove = append(remove, notifiedWorkspaceAnnotation(name))
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
		set[notifiedWorkspaceAnnotation(ws.Name)] = time.Now().UTC().Format(time.RFC3339)
	}

	if len(set) > 0 || len(remove) > 0 {
		if err := r.patchNotifiedWorkspaces(ctx, log, tn, set, remove); err != nil {
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

// patchNotifiedWorkspaces sets the given annotations and deletes the ones in remove, in a single patch.
func (r *Reconciler) patchNotifiedWorkspaces(
	ctx context.Context,
	log logr.Logger,
	tn *clv1alpha2.Tenant,
	set map[string]string,
	remove []string,
) error {
	if err := r.enforcePreservingStatus(ctx, log, tn, func(t *clv1alpha2.Tenant) *clv1alpha2.Tenant {
		if t.Annotations == nil {
			t.Annotations = map[string]string{}
		}
		maps.Copy(t.Annotations, set)
		for _, key := range remove {
			delete(t.Annotations, key)
		}
		return t
	}); err != nil {
		log.Error(err, "Error updating notified workspaces annotations")
		return fmt.Errorf("error updating notified workspaces annotations: %w", err)
	}
	return nil
}
