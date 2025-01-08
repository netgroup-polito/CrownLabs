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

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
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
