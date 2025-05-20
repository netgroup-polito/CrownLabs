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
package tenant

import (
	"context"
	"fmt"
	"strings"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/crownlabs-controller/utils"

	  "time"
    "github.com/netgroup-polito/CrownLabs/operators/pkg/crownlabs-controller/tenant/namespaces"
)

type TenantReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	TargetLabel utils.Label
	namespaceManager *namespaces.NamespaceManager
	KeepAliveTime    time.Duration
}

// Reconcile reconciles the state of a tenant resource.
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "tenant", req.NamespacedName.Name)
	ctx = ctrl.LoggerInto(ctx, log)

	log.Info("Reconciling tenant", "name", req.NamespacedName.Name)

	var tn crownlabsv1alpha2.Tenant
	if err := r.Get(ctx, req.NamespacedName, &tn); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting tenant %s before starting reconcile -> %s", req.Name, err)
		return ctrl.Result{}, err
	} else if err != nil {
		klog.Infof("Tenant %s deleted", req.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}



	if !r.TargetLabel.IsIncluded(tn.Labels) {
		// the actual Tenant is not responsibility of this controller
		log.Info("Tenant is not responsible for this controller, skipping reconcile")
		return ctrl.Result{}, nil
	}

	// Add namespace reconciliation
    nsName := fmt.Sprintf("tenant-%s", strings.ReplaceAll(tn.Name, ".", "-"))
    
    keepNsOpen, err := r.namespaceManager.CheckNamespaceKeepAlive(ctx, &tn, nsName)
    if err != nil {
        log.Error(err, "Failed to check namespace keep-alive status")
        return ctrl.Result{}, err
    }

    _, err = r.namespaceManager.EnforceClusterResources(ctx, &tn, nsName, keepNsOpen)
    if err != nil {
        log.Error(err, "Failed to enforce cluster resources")
        return ctrl.Result{}, err
    }


	r.CheckKeycloakStatus(ctx, &tn)

	return ctrl.Result{}, nil
}

// SetupWithManager registers a new controller for Tenant resources.
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Set up the NamespaceManager
	 r.namespaceManager = namespaces.NewNamespaceManager(
        r.Client,
        r.Scheme,
        r.KeepAliveTime,
        r.TargetLabel.GetKey(),
        r.TargetLabel.GetValue(),
    )
	labelPredicate, err := r.TargetLabel.GetPredicate()
	if err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&crownlabsv1alpha2.Tenant{}, builder.WithPredicates(labelPredicate)).
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

// CheckKeycloakStatus checks if the Tenant has already been created in Keycloak.
// If it has not been created, it creates it.
// It returns true if the Tenant has confrmed his/her email, false otherwise.
func (r *TenantReconciler) CheckKeycloakStatus(
	ctx context.Context,
	tenant *crownlabsv1alpha2.Tenant,
) (bool, error) {
	actor := utils.GetKeycloakActor()
	if !actor.IsInitialized() {
		klog.Warningf("Keycloak actor not initialized, skipping Keycloak status check for tenant %s", tenant.Name)
		return true, nil
	}

	// Check if the tenant exists in Keycloak
	_, err := actor.GetUser(ctx, tenant.Name)
	if err != nil {
		if err.Error() == "404" {
			klog.Infof("Tenant %s not found in Keycloak, creating it", tenant.Name)

			// Create the tenant in Keycloak
			err = r.createTenantInKeycloak(ctx, tenant)
			if err != nil {
				klog.Errorf("Error creating tenant %s in Keycloak: %v", tenant.Name, err)
				return false, err
			}

			klog.Infof("Tenant %s created in Keycloak", tenant.Name)
		} else {
			klog.Errorf("Error checking Keycloak status for tenant %s: %v", tenant.Name, err)
			return false, err
		}
	}

	// TODO check if the mail is confirmed

	return false, nil
}

func (r *TenantReconciler) createTenantInKeycloak(
	ctx context.Context,
	tenant *crownlabsv1alpha2.Tenant,
) error {
	actor := utils.GetKeycloakActor()
	if !actor.IsInitialized() {
		klog.Warningf("Keycloak actor not initialized, skipping Keycloak creation for tenant %s", tenant.Name)
		return nil
	}

	// Create the tenant in Keycloak
	userId, err := actor.CreateUser(
		ctx,
		tenant.Name,
		tenant.Spec.Email,
		tenant.Spec.FirstName,
		tenant.Spec.LastName,
	)
	if err != nil {
		klog.Errorf("Error creating tenant %s in Keycloak: %v", tenant.Name, err)
		return err
	}

	tenant.Status.Keycloak = crownlabsv1alpha2.KeycloakStatus{
		UserCreated: crownlabsv1alpha2.NameCreated{
			Name:    userId,
			Created: true,
		},
		UserConfirmed: false,
	}

	return nil
}
