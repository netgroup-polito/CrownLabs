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

// Package instautoctrl contains the controller for Instance Termination and Submission automations.
package instautoctrl

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// RetrieveEnvironment retrieves the template associated to the given instance.
func RetrieveEnvironment(ctx context.Context, c client.Client, instance *clv1alpha2.Instance) (*clv1alpha2.Environment, error) {
	log := ctrl.LoggerFrom(ctx).V(utils.LogDebugLevel)

	templateName := types.NamespacedName{
		Namespace: instance.Spec.Template.Namespace,
		Name:      instance.Spec.Template.Name,
	}

	var template clv1alpha2.Template
	if err := c.Get(ctx, templateName, &template); err != nil {
		return nil, fmt.Errorf("failed retrieving the instance template")
	}

	log.Info("retrieved the instance environment", "template", templateName)

	if len(template.Spec.EnvironmentList) != 1 {
		return nil, fmt.Errorf("only one environment per template is supported")
	}

	return &template.Spec.EnvironmentList[0], nil
}

// CheckEnvironmentValidity checks whether the given environment is valid and returns it (there must be one environment that must be persistent and contestDestination within instance spec customization urls must be present).
func CheckEnvironmentValidity(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) error {
	if instance.Spec.CustomizationUrls == nil || instance.Spec.CustomizationUrls.ContentDestination == "" {
		return fmt.Errorf("missing content-destination field for instance")
	}

	if !environment.Persistent {
		return fmt.Errorf("persistent environment required for submission job")
	}

	return nil
}

var deleteAfterChanged = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		oldTemplate, oldOk := e.ObjectOld.(*clv1alpha2.Template)
		newTemplate, newOk := e.ObjectNew.(*clv1alpha2.Template)
		if !oldOk || !newOk {
			return false
		}

		oldValue := oldTemplate.Spec.DeleteAfter
		newValue := newTemplate.Spec.DeleteAfter
		fmt.Printf("template %s/%s: old deleteAfter=%s, new deleteAfter=%s\n",
			oldTemplate.Namespace, oldTemplate.Name, oldValue, newValue)

		// Requeue only if the deleteAfter field has changed
		return oldValue == "never" && newValue != "never"
	},
}

var inactivityTimeoutChanged = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		oldTemplate, oldOk := e.ObjectOld.(*clv1alpha2.Template)
		newTemplate, newOk := e.ObjectNew.(*clv1alpha2.Template)
		if !oldOk || !newOk {
			return false
		}

		oldValue := oldTemplate.Spec.InactivityTimeout
		newValue := newTemplate.Spec.InactivityTimeout
		fmt.Printf("template %s/%s: old inactivityTimeout=%s, new inactivityTimeout=%s\n",
			oldTemplate.Namespace, oldTemplate.Name, oldValue, newValue)

		// Requeue only if the deleteAfter field has changed
		return oldValue == "never" && newValue != "never"
	},
}

var inactivityIgnoreNamespace = predicate.Funcs{
	UpdateFunc: func(e event.UpdateEvent) bool {
		oldNs, oldOk := e.ObjectOld.(*corev1.Namespace)
		newNs, newOk := e.ObjectNew.(*corev1.Namespace)
		if !oldOk || !newOk {
			return false
		}

		oldValue := oldNs.Labels[forge.InstanceInactivityIgnoreNamespace]
		newValue := newNs.Labels[forge.InstanceInactivityIgnoreNamespace]
		fmt.Printf("namespace %s: old labelValue=%s, new labelValue=%s\n",
			oldNs.Namespace, oldValue, newValue)

		// Requeue only if the label on the namespace has changed
		return oldValue == forge.InstanceInactivityIgnoreNamespace && newValue == ""
	},
}
