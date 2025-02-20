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

// ObjectMeta returns the namespace/name pair given an instance object.
func ObjectMeta(instance *clv1alpha2.Instance) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      canonicalName(instance.GetName()),
		Namespace: instance.GetNamespace(),
	}
}

// ObjectMetaWithSuffix returns the namespace/name pair given an instance object and a name suffix.
func ObjectMetaWithSuffix(instance *clv1alpha2.Instance, suffix string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      canonicalName(instance.GetName()) + StringSeparator + suffix,
		Namespace: instance.GetNamespace(),
	}
}

// NamespacedName returns the namespace/name pair given an instance object.
func NamespacedName(instance *clv1alpha2.Instance) types.NamespacedName {
	return types.NamespacedName{
		Name:      canonicalName(instance.GetName()),
		Namespace: instance.GetNamespace(),
	}
}

// NamespacedNameWithSuffix returns the namespace/name pair given an instance object and a name suffix.
func NamespacedNameWithSuffix(instance *clv1alpha2.Instance, suffix string) types.NamespacedName {
	return types.NamespacedName{
		Name:      canonicalName(instance.GetName()) + StringSeparator + suffix,
		Namespace: instance.GetNamespace(),
	}
}

// NamespacedNameToObjectMeta returns the ObjectMeta corresponding to a NamespacedName.
func NamespacedNameToObjectMeta(namespacedName types.NamespacedName) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      namespacedName.Name,
		Namespace: namespacedName.Namespace,
	}
}

// NamespacedNameFromSharedVolume returns the namespace/name pair of the passed SharedVolume.
func NamespacedNameFromSharedVolume(shvol *clv1alpha2.SharedVolume) types.NamespacedName {
	return types.NamespacedName{
		Name:      shvol.Name,
		Namespace: shvol.Namespace,
	}
}

// NamespacedNameFromMount returns the namespace/name pair of the SharedVolume contained in the passed SharedVolumeMountInfo.
func NamespacedNameFromMount(mountInfo clv1alpha2.SharedVolumeMountInfo) types.NamespacedName {
	return types.NamespacedName{
		Name:      mountInfo.SharedVolumeRef.Name,
		Namespace: mountInfo.SharedVolumeRef.Namespace,
	}
}

// canonicalName returns a canonical name given a resource name, to
// prevent issues with DNS style requirements.
func canonicalName(name string) string {
	return strings.ReplaceAll(name, ".", StringSeparator)
}

// CanonicalSandboxName returns a name given a tenant name.
func CanonicalSandboxName(name string) string {
	return fmt.Sprintf("sandbox-%s", canonicalName(name))
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
func CapIntegerQuantity(quantity, capQuantity uint32) uint32 {
	if quantity < capQuantity {
		return quantity
	}
	return capQuantity
}
