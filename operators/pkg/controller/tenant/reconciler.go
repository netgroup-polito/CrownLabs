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

// Package tenant implements the tenant controller functionality.
package tenant

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// Reconciler reconciles a Tenant object.
type Reconciler struct {
	client.Client
	Scheme                      *runtime.Scheme
	TargetLabel                 common.KVLabel
	TenantNSKeepAlive           time.Duration
	TriggerReconcileChannel     chan event.GenericEvent // Channel to trigger a reconciliation of the tenant resource.
	MyDrivePVCsSize             resource.Quantity
	MyDrivePVCsStorageClassName string
	MyDrivePVCsNamespace        string
	KeycloakActor               common.KeycloakActorIface
	WaitUserVerification        bool // If true, the reconciliation will wait for the user to be verified in Keycloak before creating resources.
	SandboxClusterRole          string
	BaseWorkspaces              []string
	Concurrency                 int
	Reschedule                  common.Rescheduler
}

// Reconcile reconciles the state of a tenant resource.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx).WithValues("tenant", req.Name)
	ctx = ctrl.LoggerInto(ctx, log)

	var tn v1alpha2.Tenant
	if err := r.Get(ctx, req.NamespacedName, &tn); client.IgnoreNotFound(err) != nil {
		log.Error(err, "Error when getting tenant before starting reconcile")
		return ctrl.Result{}, err
	} else if err != nil {
		log.Info("Tenant deleted")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !r.TargetLabel.IsIncluded(tn.Labels) {
		// the actual Tenant is not responsibility of this controller
		log.Info("Tenant is not responsibility of this controller, skipping reconcile")
		return ctrl.Result{}, nil
	}
	log.Info("Reconciling tenant")

	avoidStatusUpdate := false
	defer func(original, updated *v1alpha2.Tenant) {
		if avoidStatusUpdate {
			return
		}

		// Avoid status update if not necessary.
		if !reflect.DeepEqual(original.Status, updated.Status) {
			if err := r.Status().Update(ctx, updated); err != nil {
				log.Error(err, "Error updating tenant status")
			} else {
				log.Info("Tenant status updated")
			}
		}
	}(tn.DeepCopy(), &tn)

	reschedule := r.Reschedule.GetReconcileResult()
	hasErrors := false

	// check if the Tenant is being deleted
	if !tn.DeletionTimestamp.IsZero() {
		err := r.deleteTenantDependencies(ctx, log, &tn)
		if err != nil {
			log.Error(err, "Error deleting tenant", "tenant", tn.Name)
			return reschedule, err
		}
		log.Info("Tenant deleted", "tenant", tn.Name)
		avoidStatusUpdate = true
		return ctrl.Result{}, nil
	}

	// add the finalizer if it is not already present
	if !controllerutil.ContainsFinalizer(&tn, v1alpha2.TnOperatorFinalizerName) {
		if err := utils.PatchObject(ctx, r.Client, &tn, func(t *v1alpha2.Tenant) *v1alpha2.Tenant {
			controllerutil.AddFinalizer(t, v1alpha2.TnOperatorFinalizerName)
			return t
		}); err != nil {
			log.Error(err, "Error adding finalizer to tenant", "tenant", tn.Name)
			return reschedule, err
		}
		log.Info("Finalizer added to tenant", "tenant", tn.Name)
	}

	// enforce generic labels
	if err := r.enforceTenantBaseLabels(ctx, log, &tn); err != nil {
		return reschedule, fmt.Errorf("error enforcing tenant base labels: %w", err)
	}

	// manage workspaces subscription (and related labels)
	if err := r.syncWorkspaces(ctx, log, &tn); err != nil {
		log.Error(err, "Error enforcing workspaces for tenant", "tenant", tn.Name)
		return reschedule, fmt.Errorf("error enforcing workspaces for tenant %s: %w", tn.Name, err)
	}

	// check if the tenant is already been provisioned in Keycloak
	// - if not, create the tenant in Keycloak
	// - if yes, check if the tenant is verified
	verified, err := r.CheckKeycloakUserVerified(ctx, log, &tn)
	if err != nil {
		log.Error(err, "Error checking Keycloak status for tenant", "tenant", tn.Name)
		tn.Status.Keycloak.UserSynchronized = false
		tn.Status.Ready = false
		hasErrors = true
	} else {
		tn.Status.Keycloak.UserSynchronized = true
	}

	// manage keycloak tenant authorization for workspaces
	if err := r.syncWorkspacesAuthorizationRoles(ctx, log, &tn); err != nil {
		log.Error(err, "Error updating tenant authorization roles for tenant", "tenant", tn.Name)
		tn.Status.Keycloak.UserSynchronized = false
		tn.Status.Ready = false
		hasErrors = true
	}
	log.Info("Updated tenant authorization roles for tenant", "tenant", tn.Name)

	if r.WaitUserVerification && !verified {
		// if the Tenant has not been verified, we can skip the reconciliation
		// and wait for the next reconcile loop
		log.Info("Tenant not verified, skipping resource creation")
		tn.Status.Ready = !hasErrors
		return reschedule, nil
	}

	// managing resources not related to the personal namespace
	//   if the Tenant has already been verified, we can proceed with the reconciliation
	//   and create related resources
	if err := r.enforceTenantClusterResources(ctx, log, &tn); err != nil {
		log.Error(err, "Error creating tenant cluster resources for tenant", "tenant", tn.Name)
		tnOpinternalErrors.WithLabelValues("tenant", "cluster-resources").Inc()
		return reschedule, err
	}

	// determine the Tenant resource quota based on the Spec and the existing workspaces
	if err := r.enforceServiceQuota(ctx, log, &tn); err != nil {
		log.Error(err, "Error forging service quota for tenant", "tenant", tn.Name)
		tnOpinternalErrors.WithLabelValues("tenant", "quota-forge").Inc()
		return reschedule, fmt.Errorf("error forging service quota for tenant %s: %w", tn.Name, err)
	}

	// managing resources related to the personal namespace

	// Test if namespace has been open for too long; check if it is ok to delete
	keepAlive, err := r.checkNamespaceKeepAlive(ctx, log, &tn)
	if err != nil {
		log.Error(err, "Error checking whether tenant namespace should be kept alive")
		tnOpinternalErrors.WithLabelValues("tenant", "check-keep-alive").Inc()
		return reschedule, err
	}

	if keepAlive {
		// Namespace should be kept open, so we proceed with the reconciliation
		// creating or updating the cluster resources
		if err := r.enforceResourcesRelatedToPersonalNamespace(ctx, log, &tn); err != nil {
			log.Error(err, "Error creating or updating resources related to personal namespace for tenant %s: %v", tn.Name, err)
			tnOpinternalErrors.WithLabelValues("tenant", "create-personal-namespace").Inc()
			return reschedule, fmt.Errorf("error creating or updating resources related to personal namespace for tenant %s: %w", tn.Name, err)
		}
	} else {
		// Namespace should not be kept open, so we delete all the resources related to the tenant
		if err := r.enforceResourcesRelatedToPersonalNamespaceAbsence(ctx, log, &tn); err != nil {
			log.Error(err, "Error deleting resources related to personal namespace for tenant", "tenant", tn.Name)
			tnOpinternalErrors.WithLabelValues("tenant", "delete-personal-namespace").Inc()
			return reschedule, fmt.Errorf("error deleting resources related to personal namespace for tenant %s: %w", tn.Name, err)
		}
	}

	// esporta/disesporta tutto
	if err = r.enforceSandboxResources(ctx, &tn); err != nil {
		log.Error(err, "Failed checking sandbox for tenant", "tenant", tn.Name)
		tn.Status.SandboxNamespace.Created = false
		tnOpinternalErrors.WithLabelValues("tenant", "sandbox-resources").Inc()
		return reschedule, err
	}

	// mydrive-pvcs-namespace related stuff
	// created only if the personal namespace has been created
	// (otherwise the user will not be able to access the PVC)
	if err := r.enforceMyDrivePVC(ctx, log, &tn); err != nil {
		log.Error(err, "Error creating MyDrive PVC for tenant", "tenant", tn.Name)
		return reschedule, err
	}

	tn.Status.Ready = !hasErrors

	return reschedule, nil
}

// SetupWithManager registers a new controller for Tenant resources.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, log logr.Logger) error {
	pred, err := r.TargetLabel.GetPredicate()
	if err != nil {
		log.Error(err, "Error creating predicate for tenant controller")
		return fmt.Errorf("error creating predicate for tenant controller: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha2.Tenant{}, builder.WithPredicates(pred)).
		Owns(&v1.Secret{}).
		Owns(&v1.PersistentVolumeClaim{}).
		Owns(&v1.Namespace{}).
		Owns(&v1.ResourceQuota{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&rbacv1.ClusterRole{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&netv1.NetworkPolicy{}).
		Owns(&batchv1.Job{}).
		// Watches for non-kubernetes events to trigger reconciliation.
		// This is used to trigger a reconciliation when some event is received
		// from Keycloak webhook (e.g. user email confirmation).
		WatchesRawSource(
			source.Channel(
				r.TriggerReconcileChannel,
				handler.Funcs{
					GenericFunc: func(
						_ context.Context,
						e event.TypedGenericEvent[client.Object],
						q workqueue.TypedRateLimitingInterface[ctrl.Request],
					) {
						q.Add(ctrl.Request{
							NamespacedName: client.ObjectKey{
								Name: e.Object.GetName(),
							},
						})
					},
				},
			),
		).
		Watches(&v1alpha1.Workspace{},
			handler.EnqueueRequestsFromMapFunc(r.workspaceToEnrolledTenants)).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.Concurrency,
		}).
		WithLogConstructor(utils.LogConstructor(mgr.GetLogger(), "Tenant")).
		Complete(r)
}

func (r *Reconciler) deleteTenantDependencies(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
) error {
	// delete the personal namespace
	if err := r.enforceResourcesRelatedToPersonalNamespaceAbsence(ctx, log, tn); err != nil {
		log.Error(err, "Error deleting resources related to personal namespace for tenant", "tenant", tn.Name)
		tnOpinternalErrors.WithLabelValues("tenant", "delete-personal-namespace").Inc()
		return fmt.Errorf("error deleting resources related to personal namespace for tenant %s: %w", tn.Name, err)
	}
	log.Info("Deleted resources related to personal namespace for tenant", "tenant", tn.Name)

	// delete sandbox resources if present
	if tn.Status.SandboxNamespace.Created {
		if err := r.enforceSandboxResourcesAbsence(ctx, tn); err != nil {
			log.Error(err, "Error deleting sandbox resources for tenant", "tenant", tn.Name)
			tnOpinternalErrors.WithLabelValues("tenant", "sandbox-resources").Inc()
			return fmt.Errorf("error deleting sandbox resources for tenant %s: %w", tn.Name, err)
		}
		log.Info("Deleted sandbox resources for tenant", "tenant", tn.Name)
	}

	// delete Tenant cluster-wide RBAC resources
	if err := r.enforceTenantClusterResourcesAbsence(ctx, log, tn); err != nil {
		log.Error(err, "Error deleting tenant cluster resources for tenant", "tenant", tn.Name)
		return fmt.Errorf("error deleting tenant cluster resources for tenant %s: %w", tn.Name, err)
	}
	log.Info("Deleted tenant cluster resources", "tenant", tn.Name)

	// delete MyDrivePVC
	if err := r.enforceMyDrivePVCAbsence(ctx, log, tn); err != nil {
		log.Error(err, "Error deleting MyDrive PVC for tenant", "tenant", tn.Name)
		return fmt.Errorf("error deleting MyDrive PVC for tenant %s: %w", tn.Name, err)
	}

	// remove the tenant from Keycloak
	err := r.deleteTenantInKeycloak(ctx, log, tn)
	if err != nil {
		log.Error(err, "Error deleting tenant in Keycloak", "tenant", tn.Name)
		return err
	}
	log.Info("Deleted tenant in Keycloak", "tenant", tn.Name)

	// delete the finalizer
	if controllerutil.ContainsFinalizer(tn, v1alpha2.TnOperatorFinalizerName) {
		if err := utils.PatchObject(ctx, r.Client, tn, func(t *v1alpha2.Tenant) *v1alpha2.Tenant {
			controllerutil.RemoveFinalizer(t, v1alpha2.TnOperatorFinalizerName)
			return t
		}); err != nil {
			log.Error(err, "Error removing finalizer from tenant", "tenant", tn.Name)
			return err
		}
		log.Info("Removed finalizer from tenant", "tenant", tn.Name)
	}
	return nil
}

func (r *Reconciler) workspaceToEnrolledTenants(
	ctx context.Context,
	ws client.Object,
) []ctrl.Request {
	return r.WorkspaceNameToEnrolledTenants(ctx, ws.GetName())
}

// WorkspaceNameToEnrolledTenants returns a list of requests to reconcile tenants
// that are enrolled in the specified workspace.
func (r *Reconciler) WorkspaceNameToEnrolledTenants(
	ctx context.Context,
	wsName string,
) []ctrl.Request {
	var enqueues []ctrl.Request
	var tenants v1alpha2.TenantList
	if err := r.List(ctx, &tenants, client.HasLabels{
		fmt.Sprintf("%s%s", v1alpha2.WorkspaceLabelPrefix, wsName),
	}); err != nil {
		log := ctrl.LoggerFrom(ctx)
		log.Error(err, "Error when retrieving tenants enrolled", "workspace", wsName)
		return nil
	}
	for idx := range tenants.Items {
		enqueues = append(enqueues, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name: tenants.Items[idx].GetName(),
			},
		})
	}
	return enqueues
}
