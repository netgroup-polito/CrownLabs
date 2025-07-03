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

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"github.com/go-logr/logr"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
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

	"time"
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
	SandboxClusterRole          string
	BaseWorkspaces              []string
	Concurrency                 int
}

// Reconcile reconciles the state of a tenant resource.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx, "tenant", req.NamespacedName.Name)
	ctx = ctrl.LoggerInto(ctx, log)

	log.Info("Reconciling tenant", "name", req.NamespacedName.Name)

	var tn v1alpha2.Tenant
	if err := r.Get(ctx, req.NamespacedName, &tn); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting tenant %s before starting reconcile -> %s", req.Name, err)
		return ctrl.Result{}, err
	} else if err != nil {
		log.Info("Tenant deleted", "name", req.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !r.TargetLabel.IsIncluded(tn.Labels) {
		// the actual Tenant is not responsibility of this controller
		log.Info("Tenant is not responsibility of this controller, skipping reconcile")
		return ctrl.Result{}, nil
	}

	defer func() {
		// update the Tenant status
		if err := r.Status().Update(ctx, &tn); err != nil {
			klog.Errorf("Error updating status for tenant %s: %v", tn.Name, err)
		}
	}()

	// check if the Tenant is being deleted
	if !tn.DeletionTimestamp.IsZero() {
		err := r.deleteTenant(ctx, log, &tn)
		if err != nil {
			klog.Errorf("Error deleting tenant %s: %v", tn.Name, err)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// add the finalizer if it is not already present
	if !controllerutil.ContainsFinalizer(&tn, v1alpha2.TnOperatorFinalizerName) {
		controllerutil.AddFinalizer(&tn, v1alpha2.TnOperatorFinalizerName)
		if err := r.Update(ctx, &tn); err != nil {
			klog.Errorf("Error adding finalizer to tenant %s: %v", tn.Name, err)
			return ctrl.Result{}, err
		}
		klog.Infof("Finalizer %s added to tenant %s", v1alpha2.TnOperatorFinalizerName, tn.Name)
	}

	if tn.Status.Subscriptions == nil {
		tn.Status.Subscriptions = make(map[string]v1alpha2.SubscriptionStatus)
	}

	// manage generic labels
	if err := r.updateTenantBaseLabels(ctx, &log, &tn); err != nil {
		return ctrl.Result{}, fmt.Errorf("error updating tenant base labels: %w", err)
	}

	// manage workspaces subscription (and related labels)
	if err := r.manageWorkspaces(ctx, &tn); err != nil {
		klog.Errorf("Error managing workspaces for tenant %s: %v", tn.Name, err)
		return ctrl.Result{}, fmt.Errorf("error managing workspaces for tenant %s: %w", tn.Name, err)
	}

	// check if the tenant is already been provisioned in Keycloak
	// - if not, create the tenant in Keycloak
	// - if yes, check if the tenant is verified
	verified, err := r.CheckKeycloakUserVerified(ctx, log, &tn)
	if err != nil {
		klog.Errorf("Error checking Keycloak status for tenant %s: %v", tn.Name, err)
		tn.Status.Subscriptions["keycloak"] = v1alpha2.SubscrFailed
		tn.Status.Ready = false
		return ctrl.Result{}, err
	}
	tn.Status.Subscriptions["keycloak"] = v1alpha2.SubscrOk

	// manage keycloak tenant authorization for workspaces
	if err := r.updateWorkspacesAuthorizationRoles(ctx, &log, &tn); err != nil {
		klog.Errorf("Error updating tenant authorization roles for tenant %s: %v", tn.Name, err)
		tn.Status.Subscriptions["keycloak"] = v1alpha2.SubscrFailed
		tn.Status.Ready = false
		return ctrl.Result{}, fmt.Errorf("error updating tenant authorization roles for tenant %s: %w", tn.Name, err)
	}

	if !verified {
		// if the Tenant has not been verified, we can skip the reconciliation
		// and wait for the next reconcile loop
		log.Info("Tenant not verified, skipping resource creation")
		return ctrl.Result{}, nil
	}

	// managing resources not related to the personal namespace
	//   if the Tenant has already been verified, we can proceed with the reconciliation
	//   and create related resources
	if err := r.createTenantClusterResources(ctx, log, &tn); err != nil {
		klog.Errorf("Error creating tenant cluster resources for tenant %s: %v", tn.Name, err)
		tnOpinternalErrors.WithLabelValues("tenant", "cluster-resources").Inc()
		return ctrl.Result{}, err
	}

	// determine the Tenant resource quota based on the Spec and the existing workspaces
	if err := r.forgeServiceQuota(ctx, &tn); err != nil {
		klog.Errorf("Error forging service quota for tenant %s: %v", tn.Name, err)
		tnOpinternalErrors.WithLabelValues("tenant", "quota-forge").Inc()
		return ctrl.Result{}, fmt.Errorf("error forging service quota for tenant %s: %w", tn.Name, err)
	}

	// managing resources related to the personal namespace

	// Test if namespace has been open for too long; check if it is ok to delete
	keepAlive, err := r.checkNamespaceKeepAlive(ctx, &tn)
	if err != nil {
		klog.Errorf("Error checking whether tenant namespace should be kept alive: %s", err)
		tnOpinternalErrors.WithLabelValues("tenant", "check-keep-alive").Inc()
		return ctrl.Result{}, err
	}

	if keepAlive {
		// Namespace should be kept open, so we proceed with the reconciliation
		// creating or updating the cluster resources
		if err := r.createResourcesRelatedToPersonalNamespace(ctx, log, &tn); err != nil {
			klog.Errorf("Error creating or updating resources related to personal namespace for tenant %s: %v", tn.Name, err)
			tnOpinternalErrors.WithLabelValues("tenant", "create-personal-namespace").Inc()
			return ctrl.Result{}, fmt.Errorf("error creating or updating resources related to personal namespace for tenant %s: %w", tn.Name, err)
		}
	} else {
		// Namespace should not be kept open, so we delete all the resources related to the tenant
		if err := r.deleteResourcesRelatedToPersonalNamespace(ctx, log, &tn); err != nil {
			klog.Errorf("Error deleting resources related to personal namespace for tenant %s: %v", tn.Name, err)
			tnOpinternalErrors.WithLabelValues("tenant", "delete-personal-namespace").Inc()
			return ctrl.Result{}, fmt.Errorf("error deleting resources related to personal namespace for tenant %s: %w", tn.Name, err)
		}
	}

	if err = r.EnforceSandboxResources(ctx, &tn); err != nil {
		klog.Errorf("Failed checking sandbox for tenant %s -> %s", tn.Name, err)
		tn.Status.SandboxNamespace.Created = false
		tnOpinternalErrors.WithLabelValues("tenant", "sandbox-resources").Inc()
		return ctrl.Result{}, err
	}

	// mydrive-pvcs-namespace related stuff here
	if err := r.createMyDrivePVC(ctx, &tn); err != nil {
		klog.Errorf("Error creating MyDrive PVC for tenant %s: %v", tn.Name, err)
		return ctrl.Result{}, err
	}

	tn.Status.Ready = true

	return ctrl.Result{}, nil
}

// SetupWithManager registers a new controller for Tenant resources.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred, err := r.TargetLabel.GetPredicate()
	if err != nil {
		klog.Errorf("Error creating predicate for tenant controller: %v", err)
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

func (r *Reconciler) deleteTenant(
	ctx context.Context,
	log logr.Logger,
	tn *v1alpha2.Tenant,
) error {
	// delete the personal namespace
	if err := r.deleteResourcesRelatedToPersonalNamespace(ctx, log, tn); err != nil {
		klog.Errorf("Error deleting resources related to personal namespace for tenant %s: %v", tn.Name, err)
		tnOpinternalErrors.WithLabelValues("tenant", "delete-personal-namespace").Inc()
		return fmt.Errorf("error deleting resources related to personal namespace for tenant %s: %w", tn.Name, err)
	}
	log.Info("Deleted resources related to personal namespace for tenant", "name", tn.Name)

	// delete Tenant cluster-wide RBAC resources
	if err := r.deleteTenantClusterResources(ctx, log, tn); err != nil {
		klog.Errorf("Error deleting tenant cluster resources for tenant %s: %v", tn.Name, err)
		return fmt.Errorf("error deleting tenant cluster resources for tenant %s: %w", tn.Name, err)
	}
	log.Info("Deleted tenant cluster resources", "name", tn.Name)

	//delete MyDrivePVC
	if err := r.deleteMyDrivePVC(ctx, tn); err != nil {
		klog.Errorf("Error deleting MyDrive PVC for tenant %s: %v", tn.Name, err)
		return fmt.Errorf("error deleting MyDrive PVC for tenant %s: %w", tn.Name, err)
	}

	// remove the tenant from Keycloak
	err := r.deleteTenantInKeycloak(ctx, log, tn)
	if err != nil {
		klog.Errorf("Error deleting tenant %s in Keycloak: %v", tn.Name, err)
		return err
	}
	log.Info("Deleted tenant in Keycloak", "name", tn.Name)

	// delete the finalizer
	if controllerutil.ContainsFinalizer(tn, v1alpha2.TnOperatorFinalizerName) {
		controllerutil.RemoveFinalizer(tn, v1alpha2.TnOperatorFinalizerName)
		if err := r.Update(ctx, tn); err != nil {
			klog.Errorf("Error removing finalizer from tenant %s: %v", tn.Name, err)
			return err
		}
		log.Info("Removed finalizer from tenant", "name", tn.Name)
	}
	return nil
}

func (r *Reconciler) workspaceToEnrolledTenants(
	ctx context.Context,
	ws client.Object,
) []ctrl.Request {
	var enqueues []ctrl.Request
	var tenants v1alpha2.TenantList
	if err := r.List(ctx, &tenants, client.HasLabels{
		fmt.Sprintf("%s%s", v1alpha2.WorkspaceLabelPrefix, ws.GetName()),
	}); err != nil {
		klog.Errorf("Error when retrieving tenants enrolled in %s -> %s", ws.GetName(), err)
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
