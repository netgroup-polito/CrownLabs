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

// Package webhook implements the webhook handlers for instance resources.
package webhook

import (
	"context"
	"fmt"
	"strings"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// InstanceValidator implements a validating webhook for Instance resources.
type InstanceValidator struct {
	admission.CustomValidator
	Client client.Client
}

// ValidateCreate validates a new instance creation request.
func (iv *InstanceValidator) ValidateCreate(
	ctx context.Context,
	obj runtime.Object,
) (admission.Warnings, error) {
	// TODO: handle personal workspace quotas

	var warnings admission.Warnings

	instance, ok := obj.(*v1alpha2.Instance)
	if !ok {
		return warnings, fmt.Errorf("expected Instance resource but got %T", obj)
	}

	tenantNamespace := instance.Namespace

	// Get the workspace
	wsNamespace := instance.Spec.Template.Namespace
	wsName := strings.TrimPrefix(wsNamespace, "workspace-")

	ws := &v1alpha1.Workspace{}
	if err := iv.Client.Get(ctx, types.NamespacedName{Name: wsName}, ws); err != nil {
		return warnings, fmt.Errorf("failed to get workspace: %v", err)
	}

	// Get all the templates in the workspace namespace
	// They are needed to calculate the resource usage
	wsTemplateList := &v1alpha2.TemplateList{}
	if err := iv.Client.List(
		ctx,
		wsTemplateList,
		client.InNamespace(wsNamespace),
	); err != nil {
		return warnings, fmt.Errorf("failed to list templates in workspace namespace: %v", err)
	}

	wsTemplates := make(map[string]v1alpha2.Template)
	for _, tmpl := range wsTemplateList.Items {
		wsTemplates[tmpl.Name] = tmpl
	}

	// Find the other instances in the same workspace owned by the same user
	workspaceInstances := &v1alpha2.InstanceList{}
	if err := iv.Client.List(
		ctx,
		workspaceInstances,
		client.InNamespace(tenantNamespace),
		client.MatchingLabels{forge.LabelWorkspaceKey: wsName},
	); err != nil {
		return warnings, fmt.Errorf("failed to list instances in workspace: %v", err)
	}

	// Get the instance's template
	instanceTemplate, exists := wsTemplates[instance.Spec.Template.Name]
	if !exists {
		return warnings, fmt.Errorf("template %s not found in workspace namespace", instance.Spec.Template.Name)
	}

	// Calculate total resource usage
	var totalInstances int64 = 1 // Count the instance being created.
	var totalCPU int64 = 0
	var totalMemory resource.Quantity = resource.MustParse("0")

	// TODO: handle suspended instances

	// Add the resources of the instance being created
	for _, env := range instanceTemplate.Spec.EnvironmentList {
		totalCPU += int64(env.Resources.CPU)
		totalMemory.Add(env.Resources.Memory)
	}

	// Add the resources of the other instances
	for _, inst := range workspaceInstances.Items {
		// Skip the instance being created if found in the list
		if inst.Name == instance.Name {
			continue
		}

		totalInstances++

		instanceTemplate, exists := wsTemplates[inst.Spec.Template.Name]
		if !exists {
			warnings = append(warnings, fmt.Sprintf("template %s not found in workspace namespace for instance %s; skipping resource calculation for this instance", inst.Spec.Template.Name, inst.Name))
			continue
		}

		for _, env := range instanceTemplate.Spec.EnvironmentList {
			totalCPU += int64(env.Resources.CPU)
			totalMemory.Add(env.Resources.Memory)
		}
	}

	// Check against the workspace quota
	quota := ws.Spec.Quota

	if quota.Instances > 0 && totalInstances > quota.Instances {
		return warnings, fmt.Errorf("quota exceeded: Instances (%d > %d)", totalInstances, quota.Instances)
	}

	if !quota.CPU.IsZero() && totalCPU > quota.CPU.Value() {
		return warnings, fmt.Errorf("quota exceeded: CPU (%d > %d)", totalCPU, quota.CPU.Value())
	}

	if !quota.Memory.IsZero() && totalMemory.Cmp(quota.Memory) > 0 {
		return warnings, fmt.Errorf("quota exceeded: Memory (%s > %s)", totalMemory.String(), quota.Memory.String())
	}

	return warnings, nil
}
