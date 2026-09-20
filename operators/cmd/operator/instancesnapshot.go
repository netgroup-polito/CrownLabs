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
	"flag"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	instsnapwebhook "github.com/netgroup-polito/CrownLabs/operators/pkg/controller/instancesnapshot/webhook"
)

var (
	snapshotPublicNamespace     string
	snapshotWebhookBypassGroups string
)

const (
	// InstanceSnapshotValidatorWebhookPath is the path on which the validator webhook will be bound.
	InstanceSnapshotValidatorWebhookPath = "/validator-v1alpha2-instancesnapshot"
)

func init() {
	flag.StringVar(&snapshotPublicNamespace, "snapshot-public-namespace", "",
		"The namespace hosting the public snapshot catalog, whose volumes every tenant is entitled to read")
	flag.StringVar(&snapshotWebhookBypassGroups, "snapshot-webhook-bypass-groups", "system:masters",
		"The list of groups which can skip the snapshot scope checks, comma separated values")
}

// setupInstanceSnapshotWebhook configures the Webhook that validates the scope of InstanceSnapshot resources.
func setupInstanceSnapshotWebhook(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&clv1alpha2.InstanceSnapshot{}).
		WithValidator(&instsnapwebhook.InstanceSnapshotValidator{
			Client:                  mgr.GetClient(),
			PublicSnapshotNamespace: snapshotPublicNamespace,
			BypassGroups:            strings.Split(snapshotWebhookBypassGroups, ","),
		}).
		WithValidatorCustomPath(InstanceSnapshotValidatorWebhookPath).
		Complete()
}
