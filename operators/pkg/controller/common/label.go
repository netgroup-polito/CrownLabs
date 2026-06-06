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

// Package common contains shared utilities for CrownLabs operators.
package common

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// KVLabel represents a key-value pair label.
type KVLabel struct {
	key   string
	value string
}

// NewLabel creates a new Label instance.
func NewLabel(key, value string) KVLabel {
	return KVLabel{
		key:   key,
		value: value,
	}
}

// ParseLabel parses a label string key=value into a Label instance.
func ParseLabel(label string) (KVLabel, error) {
	parts := strings.SplitN(label, "=", 2)
	if len(parts) != 2 {
		return KVLabel{}, fmt.Errorf("invalid label format: %s", label)
	}
	return NewLabel(parts[0], parts[1]), nil
}

// ParseLabelSelectorAsMap parses a Kubernetes label selector string and converts
// it into an exact-match map representation.
//
// Only exact-match compatible operators are allowed:
// - '=' and '=='
// - 'in' with a single value.
func ParseLabelSelectorAsMap(selector string) (map[string]string, error) {
	parsed, err := k8slabels.Parse(selector)
	if err != nil {
		return nil, fmt.Errorf("invalid label selector %q: %w", selector, err)
	}

	requirements, selectable := parsed.Requirements()
	if !selectable {
		return nil, fmt.Errorf("label selector %q is not selectable", selector)
	}

	labels := make(map[string]string, len(requirements))
	for i := range requirements {
		req := requirements[i]
		values := req.ValuesUnsorted()

		switch req.Operator() {
		case selection.Equals, selection.DoubleEquals:
			if len(values) != 1 {
				return nil, fmt.Errorf("selector requirement %q must have exactly one value", req.String())
			}
			labels[req.Key()] = values[0]
		case selection.In:
			if len(values) != 1 {
				return nil, fmt.Errorf("selector requirement %q must have a single value to convert into map", req.String())
			}
			labels[req.Key()] = values[0]
		default:
			return nil, fmt.Errorf("selector requirement %q uses unsupported operator %q for map conversion", req.String(), req.Operator())
		}
	}

	return labels, nil
}

// ExtractLabelFromMap extracts a key/value label from a map and returns it as KVLabel.
func ExtractLabelFromMap(labels map[string]string, key string) (KVLabel, error) {
	if key == "" {
		return KVLabel{}, fmt.Errorf("label key cannot be empty")
	}
	value, ok := labels[key]
	if !ok {
		return KVLabel{}, fmt.Errorf("required label key %q not found", key)
	}
	return NewLabel(key, value), nil
}

// GetKey returns the key of the label.
func (l KVLabel) GetKey() string {
	return l.key
}

// GetValue returns the value of the label.
func (l KVLabel) GetValue() string {
	return l.value
}

// GetPredicate returns a Predicate for builder.
func (l KVLabel) GetPredicate() (predicate.Predicate, error) {
	labelSelector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			l.key: l.value,
		},
	}

	pred, err := predicate.LabelSelectorPredicate(labelSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to construct the label selector predicate: %w", err)
	}

	return pred, nil
}

// IsIncluded checks if the label is included in the given labels.
func (l KVLabel) IsIncluded(labels map[string]string) bool {
	if labels == nil {
		return false
	}
	value, ok := labels[l.key]
	if !ok {
		return false
	}
	return value == l.value
}
