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

package forge

import (
	"fmt"
	"maps"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// GetWorkspaceNamespaceName returns the name of the namespace for a workspace.
func GetWorkspaceNamespaceName(ws *v1alpha1.Workspace) string {
	return fmt.Sprintf("workspace-%s", ws.Name)
}

// ConfigureWorkspaceNamespace configures a namespace for a workspace.
func ConfigureWorkspaceNamespace(ns *corev1.Namespace, labels map[string]string) {
	// Set the labels for the namespace
	if ns.Labels == nil {
		ns.Labels = make(map[string]string)
	}

	// Copy the provided labels
	maps.Copy(ns.Labels, labels)

	// Set the workspace type
	ns.Labels[labelTypeKey] = labelTypeWorkspaceValue
}

// GetTenantNamespaceName returns the name of the namespace for a tenant.
func GetTenantNamespaceName(tn *v1alpha2.Tenant) string {
	return fmt.Sprintf("tenant-%s", strings.ReplaceAll(tn.Name, ".", "-"))
}

// ConfigureTenantNamespace configures a namespace for a tenant.
func ConfigureTenantNamespace(ns *corev1.Namespace, tn *v1alpha2.Tenant, labels map[string]string) {
	// Set the labels for the namespace
	if ns.Labels == nil {
		ns.Labels = make(map[string]string)
	}

	// Copy the provided labels
	maps.Copy(ns.Labels, labels)

	// Set tenant-specific labels
	ns.Labels["crownlabs.polito.it/type"] = "tenant"
	ns.Labels["crownlabs.polito.it/name"] = tn.Name
	ns.Labels["crownlabs.polito.it/instance-resources-replication"] = "true"
}
