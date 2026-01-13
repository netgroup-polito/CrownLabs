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

// Package webhook implements the webhook handlers for instance resources.
package webhook

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/netgroup-polito/CrownLabs/operators/api/common"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

const (
	personalWorkspaceName = "personal"
)

// InstanceValidator implements a validating webhook for Instance resources.
type InstanceValidator struct {
	admission.CustomValidator
	Client client.Client
}

func validateQuota(ctx context.Context, instance *v1alpha2.Instance, cl client.Client) (admission.Warnings, error) {
	var warnings admission.Warnings

	tenantNamespace := instance.Namespace

	// Get the instance's template
	instanceTemplate := &v1alpha2.Template{}
	if err := cl.Get(ctx, types.NamespacedName{Name: instance.Spec.Template.Name, Namespace: instance.Spec.Template.Namespace}, instanceTemplate); err != nil {
		return warnings, fmt.Errorf("failed to get instance template: %w", err)
	}

	// Get the workspace details (quota, templates namespace)
	wsName := instanceTemplate.Spec.WorkspaceRef.Name
	wsQuota := common.WorkspaceResourceQuota{}
	templatesNamespace := ""

	if wsName == personalWorkspaceName {
		req, err := admission.RequestFromContext(ctx)
		if err != nil {
			return warnings, fmt.Errorf("failed to get admission request from context: %w", err)
		}

		tenant := &v1alpha2.Tenant{}
		if err := cl.Get(ctx, types.NamespacedName{Name: req.UserInfo.Username}, tenant); err != nil {
			return warnings, fmt.Errorf("failed to get tenant %s: %w", req.UserInfo.Username, err)
		}

		wsQuota.CPU = tenant.Spec.PersonalWorkspace.CPU
		wsQuota.Memory = tenant.Spec.PersonalWorkspace.Memory
		wsQuota.Instances = tenant.Spec.PersonalWorkspace.Instances

		templatesNamespace = tenantNamespace
	} else {
		ws := &v1alpha1.Workspace{}
		if err := cl.Get(ctx, types.NamespacedName{Name: wsName}, ws); err != nil {
			return warnings, fmt.Errorf("failed to get workspace: %w", err)
		}

		wsQuota = ws.Spec.Quota
		templatesNamespace = forge.GetWorkspaceNamespaceName(ws)
	}

	// Get all the templates in the workspace namespace, they are needed to calculate the resource usage.
	// Instead of querying the cluster for each instance's template, we get them all at once and store them in a map.
	wsTemplateList := &v1alpha2.TemplateList{}
	if err := cl.List(
		ctx,
		wsTemplateList,
		client.InNamespace(templatesNamespace),
	); err != nil {
		return warnings, fmt.Errorf("failed to list templates in workspace namespace: %w", err)
	}

	wsTemplates := make(map[string]v1alpha2.Template)
	for i := range wsTemplateList.Items {
		wsTemplates[wsTemplateList.Items[i].Name] = wsTemplateList.Items[i]
	}

	// Find the other instances in the same workspace owned by the same user
	workspaceInstances := &v1alpha2.InstanceList{}
	if err := cl.List(
		ctx,
		workspaceInstances,
		client.InNamespace(tenantNamespace),
		client.MatchingLabels{forge.LabelWorkspaceKey: wsName},
	); err != nil {
		return warnings, fmt.Errorf("failed to list instances in workspace: %w", err)
	}

	// Calculate total resource usage
	var totalInstances int64 = 1 // Count the instance being created.
	var totalCPU int64
	totalMemory := resource.MustParse("0")

	// Add the resources of the instance being created
	for i := range instanceTemplate.Spec.EnvironmentList {
		totalCPU += int64(instanceTemplate.Spec.EnvironmentList[i].Resources.CPU)
		totalMemory.Add(instanceTemplate.Spec.EnvironmentList[i].Resources.Memory)
	}

	// Add the resources of the other instances
	for i := range workspaceInstances.Items {
		// Skip the instance being created if found in the list
		if workspaceInstances.Items[i].Name == instance.Name {
			continue
		}

		// Skip suspended instances
		if !workspaceInstances.Items[i].Spec.Running {
			continue
		}

		totalInstances++

		instanceTemplate, exists := wsTemplates[workspaceInstances.Items[i].Spec.Template.Name]
		if !exists {
			warnings = append(warnings, fmt.Sprintf("template %s not found in workspace namespace for instance %s; skipping resource calculation for this instance", workspaceInstances.Items[i].Spec.Template.Name, workspaceInstances.Items[i].Name))
			continue
		}

		for j := range instanceTemplate.Spec.EnvironmentList {
			totalCPU += int64(instanceTemplate.Spec.EnvironmentList[j].Resources.CPU)
			totalMemory.Add(instanceTemplate.Spec.EnvironmentList[j].Resources.Memory)
		}
	}

	// Check against the workspace quota
	if wsQuota.Instances > 0 && totalInstances > wsQuota.Instances {
		return warnings, fmt.Errorf("quota exceeded: Instances (%d > %d)", totalInstances, wsQuota.Instances)
	}

	if !wsQuota.CPU.IsZero() && totalCPU > wsQuota.CPU.Value() {
		return warnings, fmt.Errorf("quota exceeded: CPU (%d > %d)", totalCPU, wsQuota.CPU.Value())
	}

	if !wsQuota.Memory.IsZero() && totalMemory.Cmp(wsQuota.Memory) > 0 {
		return warnings, fmt.Errorf("quota exceeded: Memory (%s > %s)", totalMemory.String(), wsQuota.Memory.String())
	}

	return warnings, nil
}

// ValidateCreate validates a new instance creation request.
func (iv *InstanceValidator) ValidateCreate(
	ctx context.Context,
	obj runtime.Object,
) (admission.Warnings, error) {
	var warnings admission.Warnings

	// Get the instance being created
	instance, ok := obj.(*v1alpha2.Instance)
	if !ok {
		return warnings, fmt.Errorf("expected Instance resource but got %T", obj)
	}

	return validateQuota(ctx, instance, iv.Client)
}

// ValidateUpdate checks if a paused instance can be started again.
func (iv *InstanceValidator) ValidateUpdate(
	ctx context.Context,
	oldObj, newObj runtime.Object,
) (admission.Warnings, error) {
	var warnings admission.Warnings

	// Get the instance objects
	oldInstance, ok := oldObj.(*v1alpha2.Instance)
	if !ok {
		return warnings, fmt.Errorf("expected Instance resource but got %T", oldObj)
	}

	newInstance, ok := newObj.(*v1alpha2.Instance)
	if !ok {
		return warnings, fmt.Errorf("expected Instance resource but got %T", newObj)
	}

	// If the instance is not being started, no further checks are needed
	if oldInstance.Spec.Running || !newInstance.Spec.Running {
		return warnings, nil
	}

	return validateQuota(ctx, newInstance, iv.Client)
}
