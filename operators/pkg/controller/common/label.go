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
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
