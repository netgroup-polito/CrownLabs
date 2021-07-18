package forge

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// ObjectMeta returns the namespace/name pair given an instance object.
func ObjectMeta(instance *clv1alpha2.Instance) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      canonicalName(instance.GetName()),
		Namespace: instance.GetNamespace(),
	}
}

// NamespaceName returns the namespace/name pair given an instance object.
func NamespaceName(instance *clv1alpha2.Instance) types.NamespacedName {
	return types.NamespacedName{
		Name:      canonicalName(instance.GetName()),
		Namespace: instance.GetNamespace(),
	}
}

// canonicalName returns a canonical name given a resource name, to
// prevent issues with DNS style requirements.
func canonicalName(name string) string {
	return strings.ReplaceAll(name, ".", "-")
}
