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

package tenant

import (
	"context"
	"regexp"
	"strings"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"k8s.io/klog/v2"
)

// // randomDuration returns a duration between duration value between min and max.
// func randomDuration(min, max time.Duration) time.Duration {
// 	if min >= max {
// 		return min
// 	}

// 	return (min + time.Duration(rand.Float64()*float64(max-min))).Truncate(time.Millisecond) //nolint:gosec // don't need crypto/rand
// }

// // Helper functions to check and remove string from a slice of strings.
// func containsString(slice []string, s string) bool {
// 	for _, item := range slice {
// 		if item == s {
// 			return true
// 		}
// 	}
// 	return false
// }

// func labelSelectorPredicate(targetLabelKey, targetLabelValue string) predicate.Predicate {
// 	labelSelector := metav1.LabelSelector{
// 		MatchLabels: map[string]string{
// 			targetLabelKey: targetLabelValue,
// 		},
// 	}

// 	lsPred, err := predicate.LabelSelectorPredicate(labelSelector)
// 	if err != nil {
// 		klog.Fatalf("Failed to construct the label selector predicate: %v", err)
// 	}

// 	return lsPred
// }

func (r *TenantReconciler) updatePreservingStatus(
	ctx context.Context,
	tn *v1alpha2.Tenant,
) error {
	// the update function will overwrite the status, so we need to save it
	// and restore it after the update
	prevStatus := tn.Status
	if err := r.Update(ctx, tn); err != nil {
		klog.Errorf("Error when updating tenant %s -> %s", tn.Name, err)
		return err
	}
	tn.Status = prevStatus

	return nil
}

func cleanName(name string) string {
	okRegex := regexp.MustCompile("^[a-zA-Z0-9_]+$")
	name = strings.ReplaceAll(name, " ", "_")

	if !okRegex.MatchString(name) {
		problemChars := make([]string, 0)
		for _, c := range name {
			if !okRegex.MatchString(string(c)) {
				problemChars = append(problemChars, string(c))
			}
		}
		for _, v := range problemChars {
			name = strings.Replace(name, v, "", 1)
		}
	}

	return strings.Trim(name, "_")
}
