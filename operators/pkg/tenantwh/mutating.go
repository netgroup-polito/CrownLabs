// Copyright 2020-2021 Politecnico di Torino
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

package tenantwh

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	admissionv1 "k8s.io/api/admission/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// TenantLabeler labels Tenants.
type TenantLabeler struct {
	opSelectorKey, opSelectorValue string
	TenantWebhook
}

// MakeTenantLabeler creates a new webhook handler suitable for controller runtime based on TenantLabeler.
func MakeTenantLabeler(c client.Client, webhookBypassGroups []string, opSelectorKey, opSelectorValue string) *webhook.Admission {
	return &webhook.Admission{Handler: &TenantLabeler{
		opSelectorKey, opSelectorValue,
		TenantWebhook{Client: c, BypassGroups: webhookBypassGroups},
	}}
}

// Handle on TenantLabeler adds operator selector labels to new tenants and prevents possible changes - this method is used by controller runtime.
func (tl *TenantLabeler) Handle(ctx context.Context, req admission.Request) admission.Response { //nolint:gocritic,hugeParam // the signature of this method is imposed by controller runtime.
	log := ctrl.LoggerFrom(ctx).WithName("labeler").WithValues("username", req.UserInfo.Username, "tenant", req.Name)
	ctx = ctrl.LoggerInto(ctx, log)

	log.V(utils.LogDebugLevel).Info("processing mutation request", "groups", strings.Join(req.UserInfo.Groups, ","))

	tenant, err := tl.DecodeTenant(req.Object)
	if err != nil {
		log.Error(err, "tenant decode from request failed")
		return admission.Errored(http.StatusBadRequest, err)
	}

	labels, warnings, err := tl.EnforceTenantLabels(ctx, &req, tenant.GetLabels())
	if err != nil {
		log.Error(err, "label enforcement failed")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	tenant.SetLabels(labels)

	return tl.CreatePatchResponse(ctx, &req, tenant).WithWarnings(warnings...)
}

// EnforceTenantLabels sets operator selector labels.
func (tl *TenantLabeler) EnforceTenantLabels(ctx context.Context, req *admission.Request, oldLabels map[string]string) (labels map[string]string, warnings []string, err error) {
	log := ctrl.LoggerFrom(ctx)

	labels = oldLabels

	if labels == nil {
		labels = map[string]string{}
	}

	// enforce empty operator on svc tenant
	if req.Name == clv1alpha2.SVCTenantName {
		if labels[tl.opSelectorKey] != "" {
			labels[tl.opSelectorKey] = ""
			log.Info("attempted adding operator selector labels to svc tenant")
			return labels, []string{"operator selector label must not be present on service tenant and has been removed"}, nil
		}
		log.Info("service tenant processed")
		return labels, nil, nil
	}

	// skip enforcement in case of override user and non-empty selector
	if labels[tl.opSelectorKey] != "" && tl.CheckWebhookOverride(req) {
		log.Info("webhook override: not changing labels")
		return labels, nil, nil
	}

	// enforce labels on create
	if req.Operation == admissionv1.Create {
		log.Info("enforcing operator selection labels", "operation", "create")
		labels[tl.opSelectorKey] = tl.opSelectorValue
		return labels, nil, nil
	}

	oldTenant, err := tl.DecodeTenant(req.OldObject)
	if err != nil {
		// if we get an error here it's not because we're on create
		log.Error(err, "previous tenant decode from request failed")
		return nil, nil, err
	}

	oldTenantLabels := oldTenant.GetLabels()

	oldLabel, oldLabelExisted := oldTenantLabels[tl.opSelectorKey]
	newLabel := labels[tl.opSelectorKey]

	if newLabel != oldLabel {
		if oldLabelExisted {
			labels[tl.opSelectorKey] = oldLabel
		} else {
			delete(labels, tl.opSelectorKey)
		}
		warnings = []string{"operator selector label change is prohibited and has been reverted"}
		log.Info("operator selector label change prevented", "operation", "update", "requested", oldLabel, "applied", newLabel)
	} else {
		log.Info("correct operator selector label already present", "operation", "update")
	}

	return labels, warnings, nil
}

// CreatePatchResponse creates and admission response with the given tenant.
func (tl *TenantLabeler) CreatePatchResponse(ctx context.Context, req *admission.Request, tenant *clv1alpha2.Tenant) admission.Response {
	marshaledTenant, err := json.Marshal(tenant)
	if err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "patch response creation failed")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledTenant)
}
