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

// Package common contains shared utilities for CrownLabs operators.
package common

import (
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Rescheduler is a utility to manage the requeuing of reconciliations in a controlled manner.
type Rescheduler struct {
	RequeueAfterMin time.Duration // Duration to wait before requeuing the reconciliation.
	RequeueAfterMax time.Duration // Maximum duration to wait before requeuing the reconciliation.
}

// GetRequeueAfter returns a duration to wait before requeuing the reconciliation.
func (r *Rescheduler) GetRequeueAfter() time.Duration {
	// if min and max are set, return a random duration between them
	if r.RequeueAfterMin > 0 && r.RequeueAfterMax > 0 {
		delta := r.RequeueAfterMax - r.RequeueAfterMin
		return time.Duration(
			int64(r.RequeueAfterMin) +
				time.Now().UnixNano()%int64(delta),
		)
	}
	// if only min is set, return it
	if r.RequeueAfterMin > 0 {
		return r.RequeueAfterMin
	}
	// if only max is set, return a random duration between 0 and max
	if r.RequeueAfterMax > 0 {
		return time.Duration(time.Now().UnixNano() % int64(r.RequeueAfterMax))
	}
	// if neither is set, return 0
	return 0
}

// GetReconcileResult returns a reconcile.Result based on the requeue duration.
func (r *Rescheduler) GetReconcileResult() reconcile.Result {
	reqAfter := r.GetRequeueAfter()
	if reqAfter == 0 {
		return reconcile.Result{}
	}
	return reconcile.Result{
		RequeueAfter: r.GetRequeueAfter(),
	}
}
