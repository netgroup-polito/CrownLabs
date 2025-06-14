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
package workspace

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

type WorkspaceReconciler struct {
	TargetLabel common.KVLabel
}

// Reconcile reconciles the state of a Workspace resource.
func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// SetupWithManager registers a new controller for Workspace resources.
func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred, err := r.TargetLabel.GetPredicate()
	if err != nil {
		klog.Errorf("Error creating predicate for tenant controller: %v", err)
		return fmt.Errorf("error creating predicate for tenant controller: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&crownlabsv1alpha1.Workspace{}, builder.WithPredicates(pred)).
		Owns(&v1.Namespace{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&rbacv1.RoleBinding{}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "Workspace")).
		Complete(r)
}
