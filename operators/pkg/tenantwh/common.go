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

package tenantwh

import (
	"context"
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// TenantWebhook holds data needed by webhooks.
type TenantWebhook struct {
	Client       client.Client
	decoder      *admission.Decoder
	BypassGroups []string // current ns SAs group: system:serviceaccounts:NAMESPACE
}

// CheckWebhookOverride verifies the subject who triggered the request can override the webhooks behavior.
func (twh *TenantWebhook) CheckWebhookOverride(req *admission.Request) bool {
	return utils.MatchOneInStringSlices(twh.BypassGroups, req.UserInfo.Groups)
}

// DecodeTenant decodes the tenant from the incoming request.
func (twh *TenantWebhook) DecodeTenant(obj runtime.RawExtension) (tenant *clv1alpha2.Tenant, err error) {
	if twh.decoder == nil {
		return nil, errors.New("missing decoder")
	}
	tenant = &clv1alpha2.Tenant{}
	err = twh.decoder.DecodeRaw(obj, tenant)
	return
}

// GetClusterTenant retrieves the tenant from the cluster given the name.
func (twh *TenantWebhook) GetClusterTenant(ctx context.Context, name string) (tenant *clv1alpha2.Tenant, err error) {
	tenant = &clv1alpha2.Tenant{}
	err = twh.Client.Get(ctx, types.NamespacedName{Name: name}, tenant)
	return
}
