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

// Package bastion_controller groups the functionalities related to the Bastion controller.
package bastion_controller

import (
	"context"
	"os"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crownlabsalpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// BastionReconciler reconciles a Bastion object.
type BastionReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	AuthorizedKeysPath string

	// This function, if configured, is deferred at the beginning of the Reconcile.
	// Specifically, it is meant to be set to GinkgoRecover during the tests,
	// in order to lead to a controlled failure in case the Reconcile panics.
	ReconcileDeferHook func()
}

// Reconcile reconciles the SSH keys of a Tenant resource.
func (r *BastionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if r.ReconcileDeferHook != nil {
		defer r.ReconcileDeferHook()
	}

	klog.Infof("reconciling bastion [tenant=%s]", req.NamespacedName.Name)

	tenant := &crownlabsalpha2.Tenant{}
	deleted := false

	if err := r.Get(ctx, req.NamespacedName, tenant); apierrors.IsNotFound(err) {
		deleted = true
	} else if err != nil {
		return ctrl.Result{}, err
	}

	var keys []string

	if _, err := os.Stat(r.AuthorizedKeysPath); err == nil {
		// if the file exists, read the whole file in a []byte
		data, err := os.ReadFile(r.AuthorizedKeysPath)
		if err != nil {
			klog.Errorf("unable to read the file authorized_keys: %v", err)
			return ctrl.Result{}, err
		}

		if len(data) > 0 {
			keys = decomposeAndPurgeEntries(strings.Split(string(data), string("\n")), req.NamespacedName.Name)
		}
	}

	if !deleted {
		// if the event was NOT a deletion, add the tenant's keys. Otherwise nothing to do.
		keys = composeAndMarkEntries(keys, tenant.Spec.PublicKeys, req.NamespacedName.Name)
	}

	f, err := os.Create(r.AuthorizedKeysPath)
	if err != nil {
		klog.Errorf("unable to create the file authorized_keys: %v", err)
		return ctrl.Result{}, nil
	}

	defer closeFile(f)

	if len(keys) > 0 {
		_, err = f.WriteString(strings.Join(keys, string("\n")))
		if err != nil {
			klog.Errorf("unable to write to authorized_keys: %v", err)
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager registers a new controller for Tenant resources.
func (r *BastionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crownlabsalpha2.Tenant{}).
		Complete(r)
}
