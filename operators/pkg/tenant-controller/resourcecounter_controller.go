// Copyright 2020-2022 Politecnico di Torino
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

package tenant_controller

import (
	"context"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
)

// ResourceCounterReconciler watches for resource consumption of each instance, and updates counters in the corresponding Tenant object.
type ResourceCounterReconciler struct {
	client.Client
	ReservedToken uint32
}

// SetupWithManager registers a new controller for ResourceCounterReconciler resources.
func (r *ResourceCounterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clv1alpha2.Tenant{}).
		Complete(r)
}

// Reconcile reconciles resource counter in Tenant object.
func (r *ResourceCounterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Retrieve the tenant.
	var tenant clv1alpha2.Tenant
	if err := r.Get(ctx, req.NamespacedName, &tenant); client.IgnoreNotFound(err) != nil {
		klog.Errorf("Error when getting tenant %s before starting reconcile -> %s", req.Name, err)
		return ctrl.Result{}, err
	} else if err != nil {
		klog.Infof("Tenant %s deleted", req.Name)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var instances clv1alpha2.InstanceList

	// Retrieve the instances of the tenant.
	if err := r.List(ctx, &instances, client.InNamespace(tenant.Status.PersonalNamespace.Name)); err != nil {
		klog.Errorf("Error when retrieving instances of tenant %s", tenant.Name)
		return ctrl.Result{}, err
	}

	if tenant.Spec.ResourceToken == nil {
		tenant.Spec.ResourceToken = forgeResourceToken(r.ReservedToken)
	}

	for idx := range instances.Items {
		// Retrieve the template associated with the current instance.
		templateName := types.NamespacedName{
			Namespace: instances.Items[idx].Spec.Template.Namespace,
			Name:      instances.Items[idx].Spec.Template.Name,
		}
		var template clv1alpha2.Template
		if err := r.Get(ctx, templateName, &template); err != nil {
			klog.Errorf("failed retrieving the instance template %s", templateName)
			return ctrl.Result{}, err
		}
		ctx, _ = clctx.TemplateInto(ctx, &template)
		klog.Infof("successfully retrieved the instance template")

		increaseSecondsCounter(&instances.Items[idx], &tenant, &template)
	}
	tenant.Spec.ResourceToken.TokenCounterLastUpdated = v1.Now()

	var retrigErr error

	if err := r.Update(ctx, &tenant); err != nil {
		// if update fails, still try to reconcile later
		klog.Errorf("Unable to update tenant %s before exiting reconciler -> %s", tenant.Name, err)
		retrigErr = err
	}

	// no errors, need to normal reconcile later
	nextRequeDuration := time.Second * time.Duration(5*60) // need to use seconds value for 5 minutes duration
	return ctrl.Result{RequeueAfter: nextRequeDuration}, retrigErr
}

// increaseSecondsCounter increase the counter by considering each cpu core used.
func increaseSecondsCounter(instance *clv1alpha2.Instance, tenant *clv1alpha2.Tenant, template *clv1alpha2.Template) {
	secondsOffset := uint32(0)

	if v1.Now().Sub(instance.CreationTimestamp.Time).Seconds() < 5*60 {
		secondsOffset = uint32(v1.Now().Sub(instance.CreationTimestamp.Time).Seconds())
	} else {
		secondsOffset = uint32(5 * 60)
	}
	// Currently, only instances composed of a single environment are supported.
	newTokenUsed := secondsOffset * template.Spec.EnvironmentList[0].Resources.CPU
	tenant.Spec.ResourceToken.Used += newTokenUsed
	// update metrics counter.
	tnTokenConsumed.WithLabelValues(tenant.Name).Add(float64(newTokenUsed))
}

// forgeResourceToken creates the resourceToken field of the tenant with starting values.
func forgeResourceToken(reservedToken uint32) *clv1alpha2.TenantToken {
	return &clv1alpha2.TenantToken{
		Used:     0,
		Reserved: reservedToken,
	}
}
