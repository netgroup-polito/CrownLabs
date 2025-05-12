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

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

type TenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile reconciles the state of a tenant resource.
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "tenant", req.NamespacedName.Name)
	ctx = ctrl.LoggerInto(ctx, log)

	log.Info("Reconciling tenant", "name", req.NamespacedName.Name)

	return ctrl.Result{}, nil
}

// SetupWithManager registers a new controller for Tenant resources.
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// TODO
		// For(&crownlabsv1alpha2.Tenant{}, builder.WithPredicates(labelSelectorPredicate(r.TargetLabelKey, r.TargetLabelValue))).
		For(&crownlabsv1alpha2.Tenant{}).
		Owns(&v1.Secret{}).
		Owns(&v1.PersistentVolumeClaim{}).
		Owns(&v1.Namespace{}).
		Owns(&v1.ResourceQuota{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&rbacv1.ClusterRole{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&netv1.NetworkPolicy{}).
		Owns(&batchv1.Job{}).
		// TODO
		// Watches(&crownlabsv1alpha1.Workspace{},
		// 	handler.EnqueueRequestsFromMapFunc(r.workspaceToEnrolledTenants)).
		// WithOptions(controller.Options{
		// 	MaxConcurrentReconciles: r.Concurrency,
		// }).
		// WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "Tenant")).
		Complete(r)
}
