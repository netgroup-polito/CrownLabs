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

// Package main contains the entrypoint for the Crownlabs unified operator.
package main

import (
	"time"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/tenant"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/utils"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func setup_tenant(
	mgr manager.Manager,
	targetLabel utils.Label,
	TenantNSKeepAlive time.Duration,
) error {
	// TODO manage webhook
	// TODO setup tenant reconciler
	if err := (&tenant.TenantReconciler{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		TargetLabel:       targetLabel,
		TenantNSKeepAlive: TenantNSKeepAlive,
	}).SetupWithManager(mgr); err != nil {
		return err
	}

	return nil
}
