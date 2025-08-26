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

// Package instctrl groups the functionalities related to the Instance controller.
package instctrl

import (
	"context"

	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// enforcePublicExposureNetworkPolicyPresence enforces the presence of a NetworkPolicy to allow traffic
// to the instance when public exposure is enabled and ready.
func (r *InstanceReconciler) enforcePublicExposureNetworkPolicyPresence(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)

	// If the instance is not running or public exposure is not requested, do nothing.
	if !instance.Spec.Running || instance.Status.PublicExposure == nil {
		return nil
	}

	netPol := netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      forge.PublicExposureNetworkPolicyName(instance),
			Namespace: instance.Namespace,
		},
	}

	// The instance is running and public exposure is ready, so create or update the policy.
	res, err := controllerutil.CreateOrUpdate(ctx, r.Client, &netPol, func() error {
		forge.PublicExposureNetworkPolicy(instance, &netPol)
		return ctrl.SetControllerReference(instance, &netPol, r.Scheme)
	})

	if err != nil {
		log.Error(err, "failed to enforce public exposure network policy presence", "networkpolicy", netPol.Name)
		return err
	}

	log.V(utils.FromResult(res)).Info("object enforced", "networkpolicy", netPol.Name, "result", res)
	return nil
}

// enforcePublicExposureNetworkPolicyAbsence enforces the absence of a NetworkPolicy for public exposure.
func (r *InstanceReconciler) enforcePublicExposureNetworkPolicyAbsence(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)

	netPol := netv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      forge.PublicExposureNetworkPolicyName(instance),
			Namespace: instance.Namespace,
		},
	}

	if err := utils.EnforceObjectAbsence(ctx, r.Client, &netPol, "public exposure network policy"); err != nil {
		log.Error(err, "failed to enforce public exposure network policy absence", "networkpolicy", netPol.Name)
		return err
	}

	log.V(utils.LogDebugLevel).Info("public exposure network policy absent", "networkpolicy", netPol.Name)
	return nil
}
