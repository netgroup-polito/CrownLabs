// Copyright 2020-2026 Politecnico di Torino
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

// Package utils collects all the logic shared between different controllers
package utils

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
)

// ParseDockerDirectory returns a valid Docker image directory.
func ParseDockerDirectory(name string) string {
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	return strings.ToLower(reg.ReplaceAllString(name, ""))
}

// CheckLabels verifies whether a namespace is characterized by a set of required labels.
func CheckLabels(ns *corev1.Namespace, matchLabels map[string]string) bool {
	for key, value := range matchLabels {
		if v1, ok := ns.Labels[key]; !ok || v1 != value {
			return false
		}
	}
	return true
}

// CheckSelectorLabel checks if the given namespace belongs to the whitelisted namespaces where to perform reconciliation.
// TODO: Delete this and CheckLabels when the instctrl and instautoctrl will be converted to a KVLabel selector.
func CheckSelectorLabel(ctx context.Context, k8sClient client.Client, namespaceName string, matchLabels map[string]string) (bool, error) {
	namespaceLookupKey := types.NamespacedName{Name: namespaceName}
	ns := corev1.Namespace{}

	// It performs reconciliation only if the resource belongs to whitelisted namespaces
	// by checking the existence of keys in the namespace of the resource.
	if err := k8sClient.Get(ctx, namespaceLookupKey, &ns); err == nil {
		if !CheckLabels(&ns, matchLabels) {
			ctrl.LoggerFrom(ctx).V(LogDebugLevel).Info("selector labels not met")
			return false, nil
		}
	} else {
		return false, fmt.Errorf("error when retrieving namespace %q: %w", namespaceName, err)
	}

	ctrl.LoggerFrom(ctx).V(LogDebugLevel).Info("selector labels met")
	return true, nil
}

// CheckNamespaceTargetLabel checks if the given namespace has the specified target label where to perform reconciliation.
func CheckNamespaceTargetLabel(ctx context.Context, k8sClient client.Client, namespaceName string, targetLabel common.KVLabel) (bool, error) {
	namespaceLookupKey := types.NamespacedName{Name: namespaceName}
	ns := corev1.Namespace{}

	if err := k8sClient.Get(ctx, namespaceLookupKey, &ns); err != nil {
		return false, fmt.Errorf("error when retrieving namespace %q: %w", namespaceName, err)
	}

	if targetLabel.IsIncluded(ns.GetLabels()) {
		return true, nil
	}
	return false, nil
}

// MatchOneInStringSlices checks if there's at least one common string between two string slices.
func MatchOneInStringSlices(a, b []string) bool {
	for _, valA := range a {
		if slices.Contains(b, valA) {
			return true
		}
	}
	return false
}

// CheckSingleLabel checks if the instance has the label and value.
func CheckSingleLabel(obj client.Object, label, value string) bool {
	labels := obj.GetLabels()
	return labels != nil && labels[label] == value
}

// AutoEnrollEnabled checks if the specified WorkspaceAutoenroll enables any feature.
func AutoEnrollEnabled(autoEnroll clv1alpha1.WorkspaceAutoenroll) bool {
	return autoEnroll == clv1alpha1.AutoenrollImmediate ||
		autoEnroll == clv1alpha1.AutoenrollWithApproval
}
