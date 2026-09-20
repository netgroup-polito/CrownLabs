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

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	apicommon "github.com/netgroup-polito/CrownLabs/operators/api/common"
	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

const (
	personalWorkspaceName = "personal"
)

// InstanceValidator implements a validating webhook for Instance resources.
type InstanceValidator struct {
	admission.CustomValidator
	Client                  client.Client
	APIReader               client.Reader
	PublicSnapshotNamespace string
	BypassGroups            []string
}

// accumulateEnvResources aggregates the resource footprints from a list of environments into the running totals.
func accumulateEnvResources(envList []clv1alpha2.Environment, total *apicommon.ResourceSpec) {
	for i := range envList {
		total.Accumulate(&envList[i].Resources.ResourceSpec)
	}
}

func validateQuota(ctx context.Context, instance *clv1alpha2.Instance, cl client.Client) (admission.Warnings, error) {
	var warnings admission.Warnings

	tenantNamespace := instance.Namespace

	// Get the instance's template
	instanceTemplate := &clv1alpha2.Template{}
	if err := cl.Get(ctx, forge.NamespacedNameFromGenericRef(instance.Spec.Template), instanceTemplate); err != nil {
		return warnings, fmt.Errorf("failed to get instance template: %w", err)
	}

	// Get the workspace details (quota, templates namespace)
	wsName := instanceTemplate.Spec.WorkspaceRef.Name
	wsQuota := apicommon.WorkspaceResourceQuota{}
	templatesNamespace := ""

	if wsName == personalWorkspaceName {
		req, err := admission.RequestFromContext(ctx)
		if err != nil {
			return warnings, fmt.Errorf("failed to get admission request from context: %w", err)
		}

		tenant := &clv1alpha2.Tenant{}
		if err := cl.Get(ctx, types.NamespacedName{Name: req.UserInfo.Username}, tenant); err != nil {
			return warnings, fmt.Errorf("failed to get tenant %s: %w", req.UserInfo.Username, err)
		}

		wsQuota = *tenant.Spec.PersonalWorkspace.DeepCopy()
		templatesNamespace = tenantNamespace
	} else {
		ws := &clv1alpha1.Workspace{}
		if err := cl.Get(ctx, types.NamespacedName{Name: wsName}, ws); err != nil {
			return warnings, fmt.Errorf("failed to get workspace: %w", err)
		}

		wsQuota = ws.Spec.Quota
		templatesNamespace = forge.GetWorkspaceNamespaceName(ws)
	}

	// Get all the templates in the workspace namespace, they are needed to calculate the resource usage.
	// Instead of querying the cluster for each instance's template, we get them all at once and store them in a map.
	wsTemplateList := &clv1alpha2.TemplateList{}
	if err := cl.List(
		ctx,
		wsTemplateList,
		client.InNamespace(templatesNamespace),
	); err != nil {
		return warnings, fmt.Errorf("failed to list templates in workspace namespace: %w", err)
	}

	wsTemplates := make(map[string]clv1alpha2.Template)
	for i := range wsTemplateList.Items {
		wsTemplates[wsTemplateList.Items[i].Name] = wsTemplateList.Items[i]
	}

	// Find the other instances in the same workspace owned by the same user
	workspaceInstances := &clv1alpha2.InstanceList{}
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
	totalResources := apicommon.ResourceSpec{
		CPU:            0,
		Memory:         resource.MustParse("0"),
		Disk:           resource.MustParse("0"),
		OtherResources: make(map[string]resource.Quantity),
	}

	// Add the resources of the instance being created
	accumulateEnvResources(instanceTemplate.Spec.EnvironmentList, &totalResources)

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

		tmpl, exists := wsTemplates[workspaceInstances.Items[i].Spec.Template.Name]
		if !exists {
			warnings = append(warnings, fmt.Sprintf("template %s not found in workspace namespace for instance %s; skipping resource calculation for this instance", workspaceInstances.Items[i].Spec.Template.Name, workspaceInstances.Items[i].Name))
			continue
		}

		accumulateEnvResources(tmpl.Spec.EnvironmentList, &totalResources)
	}

	// Check against the workspace quota
	if wsQuota.Instances > 0 && totalInstances > wsQuota.Instances {
		return warnings, fmt.Errorf("quota exceeded: Instances (%d > %d)", totalInstances, wsQuota.Instances)
	}

	// Compare numeric CPU values directly using the uint32 fields
	if wsQuota.CPU > 0 && totalResources.CPU > wsQuota.CPU {
		return warnings, fmt.Errorf("quota exceeded: CPU (%d > %d)", totalResources.CPU, wsQuota.CPU)
	}

	if !wsQuota.Memory.IsZero() && totalResources.Memory.Cmp(wsQuota.Memory) > 0 {
		return warnings, fmt.Errorf("quota exceeded: Memory (%s > %s)", totalResources.Memory.String(), wsQuota.Memory.String())
	}

	if !wsQuota.Disk.IsZero() && totalResources.Disk.Cmp(wsQuota.Disk) > 0 {
		return warnings, fmt.Errorf("quota exceeded: Disk (%s > %s)", totalResources.Disk.String(), wsQuota.Disk.String())
	}

	for resourceName, usedQty := range totalResources.OtherResources {
		var quotaQty resource.Quantity // Defaults to zero quantity
		if wsQuota.OtherResources != nil {
			if q, exists := wsQuota.OtherResources[resourceName]; exists {
				quotaQty = q
			}
		}
		if usedQty.Cmp(quotaQty) > 0 {
			return warnings, fmt.Errorf("quota exceeded: %s (%s > %s)", resourceName, usedQty.String(), quotaQty.String())
		}
	}

	return warnings, nil
}

// hasLocalVMEnvironment reports whether the template boots at least one environment from a volume
// of this cluster, which is the only case a cross-namespace clone can arise from.
func hasLocalVMEnvironment(template *clv1alpha2.Template) bool {
	for i := range template.Spec.EnvironmentList {
		if template.Spec.EnvironmentList[i].EnvironmentType == clv1alpha2.ClassLocalVM {
			return true
		}
	}

	return false
}

// validateVolumeSources verifies that the actor is entitled to boot every LocalVM volume referenced
// by the template of the instance. Without this check any tenant could boot a copy of.
func (iv *InstanceValidator) validateVolumeSources(ctx context.Context, instance *clv1alpha2.Instance) error {
	// The template is resolved through its own reference, so templates living outside the tenant
	// namespace are covered as well.
	template := &clv1alpha2.Template{}
	if err := iv.Client.Get(ctx, forge.NamespacedNameFromGenericRef(instance.Spec.Template), template); err != nil {
		return fmt.Errorf("failed to get instance template: %w", err)
	}

	// Nothing to authorize unless the template boots from a volume.
	if !hasLocalVMEnvironment(template) {
		return nil
	}

	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get admission request from context: %w", err)
	}

	if utils.MatchOneInStringSlices(iv.BypassGroups, req.UserInfo.Groups) {
		return nil
	}

	tenant := &clv1alpha2.Tenant{}
	if err := iv.Client.Get(ctx, types.NamespacedName{Name: req.UserInfo.Username}, tenant); err != nil {
		return fmt.Errorf("failed to get tenant %s: %w", req.UserInfo.Username, err)
	}

	for i := range template.Spec.EnvironmentList {
		env := &template.Spec.EnvironmentList[i]
		if env.EnvironmentType != clv1alpha2.ClassLocalVM {
			continue
		}

		source, err := forge.ParseLocalVMImage(env.Image)
		if err != nil {
			return err
		}

		if !forge.TenantCanReadNamespace(tenant, source.Namespace, iv.PublicSnapshotNamespace) {
			return fmt.Errorf("environment %q cannot use volume %q from namespace %q: the source must belong to "+
				"your own tenant, to a workspace you are subscribed to, or to the public snapshot catalog",
				env.Name, source.Name, source.Namespace)
		}

		// Reaching a namespace does not imply being entitled to every volume inside it.
		if source.Namespace != iv.PublicSnapshotNamespace {
			if err := iv.checkSnapshotArtifact(ctx, source, env.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

// checkSnapshotArtifact verifies that the referenced volume was published by the snapshot
// controller, and is therefore meant to be consumed.
func (iv *InstanceValidator) checkSnapshotArtifact(
	ctx context.Context,
	source types.NamespacedName,
	envName string,
) error {
	pvc := &corev1.PersistentVolumeClaim{}
	if err := iv.APIReader.Get(ctx, source, pvc); err != nil {
		if kerrors.IsNotFound(err) {
			return fmt.Errorf("environment %q refers to volume %s, which does not exist", envName, source)
		}
		return fmt.Errorf("environment %q: cannot verify volume %s: %w", envName, source, err)
	}

	if pvc.Labels[forge.LabelSnapshotArtifactKey] != forge.LabelSnapshotArtifactValue {
		return fmt.Errorf("environment %q refers to volume %s, which is not a published snapshot: only snapshot "+
			"artifacts can be booted outside the public catalog", envName, source)
	}

	return nil
}

// ValidateCreate validates a new instance creation request.
func (iv *InstanceValidator) ValidateCreate(
	ctx context.Context,
	obj runtime.Object,
) (admission.Warnings, error) {
	var warnings admission.Warnings

	// Get the instance being created
	instance, ok := obj.(*clv1alpha2.Instance)
	if !ok {
		return warnings, fmt.Errorf("expected Instance resource but got %T", obj)
	}

	if err := iv.validateVolumeSources(ctx, instance); err != nil {
		return warnings, err
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
	oldInstance, ok := oldObj.(*clv1alpha2.Instance)
	if !ok {
		return warnings, fmt.Errorf("expected Instance resource but got %T", oldObj)
	}

	newInstance, ok := newObj.(*clv1alpha2.Instance)
	if !ok {
		return warnings, fmt.Errorf("expected Instance resource but got %T", newObj)
	}

	// If the instance is not being started, no further checks are needed
	if oldInstance.Spec.Running || !newInstance.Spec.Running {
		return warnings, nil
	}

	quotaWarnings, err := validateQuota(ctx, newInstance, iv.Client)
	if err != nil {
		return quotaWarnings, err
	}
	warnings = append(warnings, quotaWarnings...)

	if err := iv.validateVolumeSources(ctx, newInstance); err != nil {
		return warnings, err
	}

	return warnings, nil
}
