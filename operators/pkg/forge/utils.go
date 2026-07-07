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

package forge

import (
	"fmt"
	"strings"

	petname "github.com/dustinkirkland/golang-petname"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const (
	// StringSeparator -> the separator used to concatenate string.
	StringSeparator = "-"
)

func init() {
	petname.NonDeterministicMode()
}

// ObjectMeta returns the namespace/name pair given a Kubernetes object.
//
// Note: Remember all k8s resource structs (e.g. Pod, Deployment, ...) do not implement the
// metav1.Object interface, but any pointer to them does.
func ObjectMeta(object metav1.Object) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      CanonicalName(object.GetName()),
		Namespace: object.GetNamespace(),
	}
}

// ObjectMetaWithSuffix returns the namespace/name pair given a Kubernetes object and a name suffix.
//
// Note: Remember all k8s resource structs (e.g. Pod, Deployment, ...) do not implement the
// metav1.Object interface, but any pointer to them does.
func ObjectMetaWithSuffix(object metav1.Object, suffix string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      CanonicalName(object.GetName()) + StringSeparator + suffix,
		Namespace: object.GetNamespace(),
	}
}

// NamespacedNameWithSuffix returns the namespace/name pair given a Kubernetes object and a name suffix.
//
// Note: Remember all k8s resource structs (e.g. Pod, Deployment, ...) do not implement the
// metav1.Object interface, but any pointer to them does.
func NamespacedNameWithSuffix(object metav1.Object, suffix string) types.NamespacedName {
	return types.NamespacedName{
		Name:      CanonicalName(object.GetName()) + StringSeparator + suffix,
		Namespace: object.GetNamespace(),
	}
}

// NamespacedNameToObjectMeta returns the ObjectMeta corresponding to a NamespacedName.
func NamespacedNameToObjectMeta(namespacedName types.NamespacedName) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      namespacedName.Name,
		Namespace: namespacedName.Namespace,
	}
}

// NamespacedNameFromObject returns the namespace/name pair of the passed Kubernetes object.
//
// Note: Remember all k8s resource structs (e.g. Pod, Deployment, ...) do not implement the
// metav1.Object interface, but any pointer to them does.
func NamespacedNameFromObject(obj metav1.Object) types.NamespacedName {
	return types.NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
}

// NamespacedNameFromGenericRef returns the namespace/name pair of the passed reference.
func NamespacedNameFromGenericRef(ref clv1alpha2.GenericRef) types.NamespacedName {
	return types.NamespacedName{
		Name:      ref.Name,
		Namespace: ref.Namespace,
	}
}

// CanonicalName returns a canonical name given a resource name, to
// prevent issues with DNS style requirements.
func CanonicalName(name string) string {
	return strings.ReplaceAll(name, ".", StringSeparator)
}

// CanonicalSandboxName returns a name given a tenant name.
func CanonicalSandboxName(name string) string {
	return fmt.Sprintf("sandbox-%s", CanonicalName(name))
}

// RandomInstancePrettyName generates a random name of 2 capitalized words.
func RandomInstancePrettyName() string {
	return cases.Title(language.AmericanEnglish).String(petname.Generate(2, " "))
}

// CapResourceQuantity compares a resource.Quantity value with a given cap and returns the lower.
func CapResourceQuantity(quantity, capQuantity resource.Quantity) resource.Quantity {
	if quantity.Cmp(capQuantity) < 0 {
		return quantity
	}
	return capQuantity
}

// CapIntegerQuantity compares an unsigned integer value with a given cap and returns the lower.
func CapIntegerQuantity(quantity, capQuantity int64) int64 {
	if quantity < capQuantity {
		return quantity
	}
	return capQuantity
}

// LastCharsOf returns a string composed by the last 'many' chars from 's'.
func LastCharsOf(s string, many int) string {
	if many < 0 {
		return ""
	}
	if len(s) > many {
		return s[len(s)-many:]
	}
	return s
}

// MapFromKVString parses a string like "key1=val1,key2=val2" into a map.
func MapFromKVString(raw string) (map[string]string, error) {
	annotations := make(map[string]string)
	if raw == "" {
		return annotations, nil
	}

	pairs := strings.Split(raw, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid annotation format: %s", pair)
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		if key == "" {
			return nil, fmt.Errorf("empty annotation key in: %s", pair)
		}
		annotations[key] = value
	}
	return annotations, nil
}
