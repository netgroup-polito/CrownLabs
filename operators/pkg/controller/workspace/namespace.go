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

// Package workspace implements the workspace controller functionality.
package workspace

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

func (r *Reconciler) enforceNamespace(
	ctx context.Context,
	ws *v1alpha1.Workspace,
) error {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: forge.GetWorkspaceNamespaceName(ws),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, ns, func() error {
		// Configure the namespace
		labels := forge.UpdateWorkspaceResourceCommonLabels(nil, r.TargetLabel)
		forge.ConfigureWorkspaceNamespace(ns, labels)

		return controllerutil.SetControllerReference(ws, ns, r.Scheme)
	}); err != nil {
		return fmt.Errorf("error while creating/updating namespace %s for workspace %s: %w",
			ns.Name, ws.Name, err)
	}

	ws.Status.Namespace = v1alpha2.NameCreated{
		Name:    ns.Name,
		Created: true,
	}

	return nil
}

func (r *Reconciler) enforceNamespaceAbsence(
	ctx context.Context,
	ws *v1alpha1.Workspace,
) error {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: forge.GetWorkspaceNamespaceName(ws),
		},
	}

	if err := utils.EnforceObjectAbsence(ctx, r.Client, ns, "Namespace"); err != nil {
		return fmt.Errorf("error while deleting namespace %s for workspace %s: %w",
			ns.Name, ws.Name, err)
	}

	ws.Status.Namespace = v1alpha2.NameCreated{
		Name:    ns.Name,
		Created: false,
	}

	return nil
}
