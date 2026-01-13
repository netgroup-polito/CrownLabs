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

package main

import (
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	instancewebhook "github.com/netgroup-polito/CrownLabs/operators/pkg/controller/instance/webhook"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	// InstanceValidatorWebhookPath is the path on which the validator webhook will be bound.
	InstanceValidatorWebhookPath = "/validator-v1alpha2-instance"
)

// setupInstance configures the Instance controller.
func setupInstance(mgr manager.Manager) error {
	// Setup the webhook if enabled
	if enableWebhooks {
		if err := setupInstanceWebhook(mgr); err != nil {
			return err
		}
	}

	return nil
}

// setupInstanceWebhook configures the Webhook that validates the resources available for the Tenant in the Workspace.
func setupInstanceWebhook(
	mgr ctrl.Manager,
) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&crownlabsv1alpha2.Instance{}).
		WithValidator(&instancewebhook.InstanceValidator{
			Client: mgr.GetClient(),
		}).
		WithValidatorCustomPath(InstanceValidatorWebhookPath).
		Complete()
}
