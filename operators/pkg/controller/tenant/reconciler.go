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

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/utils"

	"time"
)

const (
	// NoWorkspacesLabel -> label to be set (to true) when no workspaces are associated to the tenant.
	NoWorkspacesLabel = "crownlabs.polito.it/no-workspaces"
	// NFSSecretName -> NFS secret name.
	NFSSecretName = "mydrive-info"
	// NFSSecretServerNameKey -> NFS Server key in NFS secret.
	NFSSecretServerNameKey = "server-name"
	// NFSSecretPathKey -> NFS path key in NFS secret.
	NFSSecretPathKey = "path"
	// ProvisionJobBaseImage -> Base container image for Personal Drive provision job.
	ProvisionJobBaseImage = "busybox"
	// ProvisionJobMaxRetries -> Maximum number of retries for Provision jobs.
	ProvisionJobMaxRetries = 3
	// ProvisionJobTTLSeconds -> Seconds for Provision jobs before deletion (either failure or success).
	ProvisionJobTTLSeconds = 3600 * 24 * 7
)

type TenantReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	TargetLabel utils.Label
	//KeepAliveTime    time.Duration
	TenantNSKeepAlive           time.Duration
	TargetLabelKey              string
	TargetLabelValue            string
	MyDrivePVCsSize             resource.Quantity
	MyDrivePVCsStorageClassName string
	MyDrivePVCsNamespace        string
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
		log.Info("Tenant %s deleted", req.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !r.TargetLabel.IsIncluded(tn.Labels) {
		// the actual Tenant is not responsibility of this controller
		log.Info("Tenant is not responsibility of this controller, skipping reconcile")
		return ctrl.Result{}, nil
	}

	verified, err := r.CheckKeycloakUserVerified(ctx, &tn)
	if err != nil {
		klog.Errorf("Error checking Keycloak status for tenant %s: %v", tn.Name, err)
		return ctrl.Result{}, err
	}

	if verified {
		// if the Tenant has already been verified, we can proceed with the reconciliation
		// and create related resources
		log.Info("create resources")
	} else {
		// if the Tenant has not been verified, we can skip the reconciliation
		// and wait for the next reconcile loop
		log.Info("Tenant not verified, skipping reconciliation")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager registers a new controller for Tenant resources.
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {

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
