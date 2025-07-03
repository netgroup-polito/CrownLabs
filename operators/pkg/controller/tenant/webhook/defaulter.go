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

// Package webhook implements the webhook handlers for tenant resources.
package webhook

import (
	"context"
	"fmt"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// TenantDefaulter implements a defaulting webhook for Tenant resources.
type TenantDefaulter struct {
	admission.CustomDefaulter
	TenantWebhook

	Decoder         admission.Decoder
	OpSelectorLabel common.KVLabel
	BaseWorkspaces  []string
}

// Default adds operator selector labels to new tenants and prevents possible changes - this method is used by controller runtime.
func (tm *TenantDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	log := ctrl.LoggerFrom(ctx).WithName("labeler")
	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		log.Error(err, "failed to get request from context")
		return err
	}
	log = log.WithValues("username", req.UserInfo.Username, "tenant", req.Name)
	ctx = ctrl.LoggerInto(ctx, log)

	log.V(utils.LogDebugLevel).Info("processing mutation request", "groups", strings.Join(req.UserInfo.Groups, ","))

	tenant, ok := obj.(*v1alpha2.Tenant)
	if !ok {
		err := errors.NewBadRequest("expected a Tenant object")
		log.Error(err, "object is not a Tenant")
	}

	labels, err := tm.EnforceTenantLabels(ctx, &req, tenant.GetLabels())
	if err != nil {
		log.Error(err, "label enforcement failed")
		return err
	}

	tenant.SetLabels(labels)

	tm.EnforceTenantBaseWorkspaces(ctx, tenant)

	return nil
}

// DecodeTenant decodes the tenant from the incoming request.
func (tm *TenantDefaulter) DecodeTenant(obj runtime.RawExtension) (tenant *v1alpha2.Tenant, err error) {
	if tm.Decoder == nil {
		return nil, errors.NewInternalError(fmt.Errorf("decoder is not set"))
	}
	tenant = &v1alpha2.Tenant{}
	err = tm.Decoder.DecodeRaw(obj, tenant)
	return
}

// EnforceTenantLabels sets operator selector labels.
func (tm *TenantDefaulter) EnforceTenantLabels(ctx context.Context, req *admission.Request, oldLabels map[string]string) (labels map[string]string, err error) {
	log := ctrl.LoggerFrom(ctx)

	labels = oldLabels

	if labels == nil {
		labels = map[string]string{}
	}

	// enforce empty operator on svc tenant
	if req.Name == v1alpha2.SVCTenantName {
		if !tm.OpSelectorLabel.IsIncluded(labels) {
			labels[tm.OpSelectorLabel.GetKey()] = ""
			log.Info("attempted adding operator selector labels to svc tenant")
			return labels, nil
		}
		log.Info("service tenant processed")
		return labels, nil
	}

	// skip enforcement in case of override user and non-empty selector
	if labels[tm.OpSelectorLabel.GetKey()] != "" && tm.CheckWebhookOverride(req) {
		log.Info("webhook override: not changing labels")
		return labels, nil
	}

	// enforce labels on create
	if req.Operation == admissionv1.Create {
		log.Info("enforcing operator selection labels", "operation", "create")
		labels[tm.OpSelectorLabel.GetKey()] = tm.OpSelectorLabel.GetValue()
		return labels, nil
	}

	oldTenant, err := tm.DecodeTenant(req.OldObject)
	if err != nil {
		// if we get an error here it's not because we're on create
		log.Error(err, "previous tenant decode from request failed")
		return nil, err
	}

	oldTenantLabels := oldTenant.GetLabels()

	oldLabel, oldLabelExisted := oldTenantLabels[tm.OpSelectorLabel.GetKey()]
	newLabel := labels[tm.OpSelectorLabel.GetKey()]

	if newLabel != oldLabel {
		if oldLabelExisted {
			labels[tm.OpSelectorLabel.GetKey()] = oldLabel
		} else {
			delete(labels, tm.OpSelectorLabel.GetKey())
		}
		log.Info("operator selector label change prevented", "operation", "update", "requested", oldLabel, "applied", newLabel)
	} else {
		log.Info("correct operator selector label already present", "operation", "update")
	}

	return labels, nil
}

// EnforceTenantBaseWorkspaces ensure base workspaces are present in the given tenant.
func (tm *TenantDefaulter) EnforceTenantBaseWorkspaces(ctx context.Context, tenant *v1alpha2.Tenant) {
	log := ctrl.LoggerFrom(ctx)
	log.Info("enforcing base workspaces")

	for _, baseWs := range tm.BaseWorkspaces {
		found := false
		for _, tenantWs := range tenant.Spec.Workspaces {
			if tenantWs.Name == baseWs {
				found = true
				break
			}
		}
		if !found {
			tenant.Spec.Workspaces = append(tenant.Spec.Workspaces, v1alpha2.TenantWorkspaceEntry{
				Name: baseWs,
				Role: v1alpha2.User,
			})
			log.Info("base workspace added", "workspace", baseWs)
		}
	}
}
