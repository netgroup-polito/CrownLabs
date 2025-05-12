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
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/workspace"
)

func init() {}

func setupWorkspace(
	mgr manager.Manager,
	log logr.Logger,
	targetLabel common.KVLabel,
) error {
	// Create the Workspace Reconciler
	wr := &workspace.Reconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		TargetLabel:   targetLabel,
		KeycloakActor: common.GetKeycloakActor(),
		Reschedule:    reschedule,
	}

	// Register the WorkspaceReconciler with the manager
	return wr.SetupWithManager(mgr, log)
}
