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

// Package tenantwh groups the functionalities related to the Tenant webhook.
package tenantwh

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// LastLoginToleration defines the maximum skew with respect to the current time that is accepted by the webhook for the LastLogin field.
const LastLoginToleration = time.Hour * 24

// TenantValidator validates Tenants.
type TenantValidator struct{ TenantWebhook }

// MakeTenantValidator creates a new webhook handler suitable for controller runtime based on TenantValidator.
func MakeTenantValidator(c client.Client, webhookBypassGroups []string, scheme *runtime.Scheme) *webhook.Admission {
	return &webhook.Admission{Handler: &TenantValidator{TenantWebhook{
		Client:       c,
		BypassGroups: webhookBypassGroups,
		decoder:      admission.NewDecoder(scheme),
	}}}
}

// Handle admits a tenant if user is editing its own tenant or a user is adding/removing workspaces
// they own to/from another user - this method is used by controller runtime.
func (tv *TenantValidator) Handle(ctx context.Context, req admission.Request) admission.Response { //nolint:gocritic // the signature of this method is imposed by controller runtime.
	log := ctrl.LoggerFrom(ctx).WithName("validator").WithValues("username", req.UserInfo.Username, "tenant", req.Name)

	log.V(utils.LogDebugLevel).Info("processing admission request", "groups", strings.Join(req.UserInfo.Groups, ","))

	if tv.CheckWebhookOverride(&req) {
		log.Info("admitted: successful override")
		return admission.Allowed("")
	}

	tenant, err := tv.DecodeTenant(req.Object)
	if err != nil {
		log.Error(err, "new tenant decode from request failed")
		return admission.Errored(http.StatusBadRequest, err)
	}
	oldTenant, err := tv.DecodeTenant(req.OldObject)
	if err != nil && req.Operation != admissionv1.Create {
		log.Error(err, "previous tenant decode from request failed")
		return admission.Errored(http.StatusBadRequest, err)
	}

	if req.UserInfo.Username == req.Name {
		ctx = ctrl.LoggerInto(ctx, log.WithValues("operation", "self-edit"))
		return tv.HandleSelfEdit(ctx, tenant, oldTenant)
	}

	manager, err := tv.GetClusterTenant(ctx, req.UserInfo.Username)
	if err != nil {
		log.Error(err, "failed fetching a (manager) tenant associated to the current actor")
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("could not fetch a tenant for the current user: %w", err))
	}

	ctx = ctrl.LoggerInto(ctx, log.WithValues("operation", "workspaces-edit"))
	return tv.HandleWorkspaceEdit(ctx, tenant, oldTenant, manager, req.Operation)
}

// HandleSelfEdit checks every field but public keys for changes:
// - LastLogin must be within a certain tolerance;
// - Workspaces can be changed only if autoenroll is enabled and within the allowed roles;
// - Other fields must be unchanged.
func (tv *TenantValidator) HandleSelfEdit(ctx context.Context, newTenant, oldTenant *clv1alpha2.Tenant) admission.Response {
	log := ctrl.LoggerFrom(ctx)
	newTenant.Spec.PublicKeys = nil
	oldTenant.Spec.PublicKeys = nil

	lastLoginDelta := time.Until(newTenant.Spec.LastLogin.Time).Abs()
	if newTenant.Spec.LastLogin != oldTenant.Spec.LastLogin && lastLoginDelta > LastLoginToleration {
		return admission.Denied(fmt.Sprintf("unacceptable last login time, must be within +/-%v since local server time: %v", LastLoginToleration, time.Now()))
	}
	newTenant.Spec.LastLogin = metav1.Time{}
	oldTenant.Spec.LastLogin = metav1.Time{}

	// manage workspaces
	newWorkspaces := newTenant.Spec.Workspaces
	oldWorkspaces := oldTenant.Spec.Workspaces
	newTenant.Spec.Workspaces = nil
	oldTenant.Spec.Workspaces = nil

	if !reflect.DeepEqual(newTenant.Spec, oldTenant.Spec) {
		log.Info("denied: unexpected tenant spec change")
		return admission.Denied("only changes to public keys or workspaces that have autoenroll enabled are allowed in the owned tenant")
	}

	newTenant.Spec.Workspaces = newWorkspaces
	oldTenant.Spec.Workspaces = oldWorkspaces

	res, err := tv.checkValidWorkspaces(ctx, newTenant, oldTenant)
	if err != nil {
		log.Error(err, "failed to check workspace changes")
		return admission.Errored(http.StatusInternalServerError, err)
	}
	if !res {
		log.Info("denied: workspaces validation failed")
		return admission.Denied("you have changed workspaces you are not allowed to change")
	}

	log.Info("allowed")
	return admission.Allowed("")
}

// checkValidWorkspaces checks that the user is not changing workspaces they are not allowed to change.
func (tv *TenantValidator) checkValidWorkspaces(ctx context.Context, newTenant, oldTenant *clv1alpha2.Tenant) (bool, error) {
	workspaceDiff := CalculateWorkspacesDiff(newTenant, oldTenant)
	newWorkspacesMap := mapFromWorkspacesList(newTenant)

	for ws, changed := range workspaceDiff {
		if !changed {
			// it's always ok to keep the same role
			continue
		}
		wsObj := clv1alpha1.Workspace{}
		err := tv.Client.Get(ctx, client.ObjectKey{Name: ws}, &wsObj)
		if err != nil {
			return false, fmt.Errorf("failed to fetch workspace %s: %w", ws, err)
		}
		if !utils.AutoEnrollEnabled(wsObj.Spec.AutoEnroll) {
			// Tenant cannot change workspaces with autoenroll disabled
			return false, nil
		}
		if _, ok := newWorkspacesMap[ws]; !ok {
			// it's always possible to remove a Workspace from the Tenant if the target Workspace has autoenroll enabled
			continue
		}
		if wsObj.Spec.AutoEnroll == clv1alpha1.AutoenrollImmediate && newWorkspacesMap[ws] != clv1alpha2.User {
			// if AutoEnroll is Immediate, then the user has to enroll with User role
			return false, nil
		}
		if wsObj.Spec.AutoEnroll == clv1alpha1.AutoenrollWithApproval && newWorkspacesMap[ws] != clv1alpha2.Candidate {
			// if AutoEnroll is WithApproval, then the user has to enroll with Candidate role (to be approved by a Manager)
			return false, nil
		}
	}

	return true, nil
}

// HandleWorkspaceEdit checks that changes made to the workspaces have been made by a valid manager, then checks other fields not to have been modified through DeepEqual.
func (tv *TenantValidator) HandleWorkspaceEdit(ctx context.Context, newTenant, oldTenant, manager *clv1alpha2.Tenant, operation admissionv1.Operation) admission.Response {
	log := ctrl.LoggerFrom(ctx)

	workspacesDiff := CalculateWorkspacesDiff(newTenant, oldTenant)
	managerWorkspaces := mapFromWorkspacesList(manager)

	for ws, changed := range workspacesDiff {
		if changed && managerWorkspaces[ws] != clv1alpha2.Manager {
			log.Info("denied: unexpected tenant spec change", "not-a-manager-for", ws)
			return admission.Denied("you are not a manager for workspace <" + ws + ">")
		}
	}

	newTenant.Spec.Workspaces = nil
	oldTenant.Spec.Workspaces = nil
	if operation != admissionv1.Create && !reflect.DeepEqual(newTenant.Spec, oldTenant.Spec) {
		log.Info("denied: unexpected tenant spec change")
		return admission.Denied("only changes to workspaces are allowed to workspace managers")
	}

	log.Info("allowed")
	return admission.Allowed("")
}

func calculateWorkspacesOneWayDiff(a, b *clv1alpha2.Tenant, changes map[string]bool) map[string]bool {
	aAsMap := mapFromWorkspacesList(a)
	for _, v := range b.Spec.Workspaces {
		if aAsMap[v.Name] != v.Role {
			changes[v.Name] = true
		}
	}
	return changes
}

// CalculateWorkspacesDiff returns the list of workspaces that are different between two tenants.
func CalculateWorkspacesDiff(a, b *clv1alpha2.Tenant) map[string]bool {
	changes := calculateWorkspacesOneWayDiff(a, b, map[string]bool{})

	return calculateWorkspacesOneWayDiff(b, a, changes)
}

func mapFromWorkspacesList(tenant *clv1alpha2.Tenant) map[string]clv1alpha2.WorkspaceUserRole {
	wss := make(map[string]clv1alpha2.WorkspaceUserRole, len(tenant.Spec.Workspaces))

	for _, v := range tenant.Spec.Workspaces {
		wss[v.Name] = v.Role
	}

	return wss
}
