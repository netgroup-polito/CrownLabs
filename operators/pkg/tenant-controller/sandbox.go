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

// Package tenant_controller groups the functionalities related to the Tenant controller.
package tenant_controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// EnforceSandboxResources ensures the presence/absence of a sandbox.
func (r *TenantReconciler) EnforceSandboxResources(ctx context.Context, tenant *clv1alpha2.Tenant) error {
	if tenant.Spec.CreateSandbox {
		return r.enforceSandboxResourcesPresence(ctx, tenant)
	}
	return r.enforceSandboxResourcesAbsence(ctx, tenant)
}

// enforceSandboxResourcesPresence ensures the presence of a sandbox.
func (r *TenantReconciler) enforceSandboxResourcesPresence(ctx context.Context, tenant *clv1alpha2.Tenant) error {
	sandboxnsName := forge.CanonicalSandboxName(tenant.Name)
	log := ctrl.LoggerFrom(ctx, "environment", "sandbox")

	// Enforce the namespace presence
	namespace := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: sandboxnsName}}
	res, err := ctrl.CreateOrUpdate(ctx, r.Client, &namespace, func() error {
		namespace.SetLabels(forge.SandboxObjectLabels(namespace.GetLabels(), tenant.Name))
		return ctrl.SetControllerReference(tenant, &namespace, r.Scheme)
	})

	if err != nil {
		log.Error(err, "failed to enforce resource", "namespace", klog.KObj(&namespace))
		return err
	}
	log.V(utils.FromResult(res)).Info("sandbox namespace correctly enforced", "sandbox namespace", klog.KObj(&namespace), "result", res)
	tenant.Status.SandboxNamespace.Name = sandboxnsName

	// Enforce role binding precence
	roleBind := rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "sandbox-editor", Namespace: sandboxnsName}}
	res, err = ctrl.CreateOrUpdate(ctx, r.Client, &roleBind, func() error {
		roleBind.SetLabels(forge.SandboxObjectLabels(roleBind.GetLabels(), tenant.Name))
		roleBind.RoleRef = rbacv1.RoleRef{Kind: "ClusterRole", Name: r.SandboxClusterRole, APIGroup: rbacv1.GroupName}
		roleBind.Subjects = []rbacv1.Subject{{Kind: rbacv1.UserKind, Name: tenant.Name, APIGroup: rbacv1.GroupName}}
		return ctrl.SetControllerReference(tenant, &roleBind, r.Scheme)
	})
	if err != nil {
		log.Error(err, "failed to enforce resource", "role binding", klog.KObj(&roleBind))
		return err
	}
	log.V(utils.FromResult(res)).Info("sandbox role binding correctly enforced", "role binding", klog.KObj(&roleBind), "result", res)

	// Enforce resource quota presence
	resourceQuota := corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{Name: "sandbox-resource-quota", Namespace: sandboxnsName},
	}
	res, err = ctrl.CreateOrUpdate(ctx, r.Client, &resourceQuota, func() error {
		resourceQuota.SetLabels(forge.SandboxObjectLabels(resourceQuota.GetLabels(), tenant.Name))
		resourceQuota.Spec.Hard = forge.SandboxResourceQuotaSpec()
		return ctrl.SetControllerReference(tenant, &resourceQuota, r.Scheme)
	})
	if err != nil {
		log.Error(err, "failed to enforce resource", "resource quota", klog.KObj(&resourceQuota))
		return err
	}
	log.V(utils.FromResult(res)).Info("sandbox resource quota correctly enforced", "resource quota", klog.KObj(&resourceQuota), "result", res)
	tenant.Status.SandboxNamespace.Created = true

	// Enforce limit range presence
	limitRange := corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{Name: "sandbox-limit-range", Namespace: sandboxnsName},
	}
	res, err = ctrl.CreateOrUpdate(ctx, r.Client, &limitRange, func() error {
		limitRange.SetLabels(forge.SandboxObjectLabels(resourceQuota.GetLabels(), tenant.Name))
		limitRange.Spec = forge.SandboxLimitRangeSpec()
		return ctrl.SetControllerReference(tenant, &limitRange, r.Scheme)
	})
	if err != nil {
		log.Error(err, "failed to enforce resource", "limit range", klog.KObj(&limitRange))
		return err
	}
	log.V(utils.FromResult(res)).Info("sandbox limit range correctly enforced", "limit range", klog.KObj(&limitRange), "result", res)
	tenant.Status.SandboxNamespace.Created = true

	return err
}

// enforceInstanceExpositionAbsence ensures the sandbox's absence.
func (r *TenantReconciler) enforceSandboxResourcesAbsence(ctx context.Context, tenant *clv1alpha2.Tenant) error {
	sandboxnsName := forge.CanonicalSandboxName(tenant.Name)

	// Enforce namespace absence
	namespace := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: sandboxnsName}}
	if err := utils.EnforceObjectAbsence(ctx, r.Client, &namespace, "sandbox namespace"); err != nil {
		return err
	}

	tenant.Status.SandboxNamespace.Created = false
	tenant.Status.SandboxNamespace.Name = ""

	return nil
}
