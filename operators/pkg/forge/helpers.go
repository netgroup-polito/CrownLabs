package forge

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/tree"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func getCRDList(ctx context.Context, c ctrlclient.Client, opts ...ctrlclient.ListOption) ([]apiextensionsv1.CustomResourceDefinition, error) {
	crdList := &apiextensionsv1.CustomResourceDefinitionList{}
	if err := c.List(ctx, crdList, opts...); err != nil {
		return nil, errors.Wrap(err, "failed to get the list of CRDs required for the move discovery phase")
	}

	return crdList.Items, nil
}

// getObjList is a helper function to list objects of a specific type for constructing the DescribeCluster resource tree.
func getObjList(ctx context.Context, c ctrlclient.Client, typeMeta metav1.TypeMeta, selectors []ctrlclient.ListOption) (*unstructured.UnstructuredList, error) {
	objList := new(unstructured.UnstructuredList)
	objList.SetAPIVersion(typeMeta.APIVersion)
	objList.SetKind(typeMeta.Kind)

	if err := c.List(ctx, objList, selectors...); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed to list %q resources", objList.GroupVersionKind())
	}

	return objList, nil
}

// updateSeverityIfMoreSevere takes an existing severity and a new severity and returns the more severe of the two based on the rule that Error > Warning > Info > None.
// This is used to determine the severity of a group node, i.e. a node representing 2 Machines in the DescribeCluster resource tree.
func updateSeverityIfMoreSevere(existingSev string, newSev string) string {
	switch {
	case existingSev == "":
		return newSev
	case existingSev == "Info":
		if newSev == "Error" || newSev == "Warning" {
			return newSev
		}
		return existingSev
	case existingSev == "Warning":
		if newSev == "Error" {
			return newSev
		}
		return existingSev
	case existingSev == "Error":
		return existingSev
	}

	return existingSev
}

// getProvider returns the provider type for an object in the Cluster resource tree. If the object is a virtual object and its kind is
// listed in treeOptions.VNodesToInheritChildProvider, the provider type of the object's children is checked. If all children have the
// same provider type, the provider type is inherited. If the object is not a virtual object, the provider type is looked up directly.
func getProvider(ctx context.Context, object ctrlclient.Object, children []ctrlclient.Object, treeOptions ClusterResourceTreeOptions) (string, error) {
	log := ctrl.LoggerFrom(ctx)

	if override, ok := treeOptions.providerTypeOverrideMap[object.GetObjectKind().GroupVersionKind().Kind]; ok {
		return override, nil
	}

	if tree.IsVirtualObject(object) {
		_, inherit := treeOptions.VNodesToInheritChildProvider[object.GetObjectKind().GroupVersionKind().Kind]
		if !inherit {
			return "virtual", nil
		}
		log.V(4).Info("Aggregating object w/ kind, name, and metaName", "kind", object.GetObjectKind().GroupVersionKind().Kind, "name", object.GetName(), "metaName", tree.GetMetaName(object))

		prev := ""
		for i, child := range children {
			provider, err := lookUpProvider(child)
			if err != nil {
				return "", err
			}
			log.V(4).Info("Child object w/ kind, name, and provider", "kind", object.GetObjectKind().GroupVersionKind().Kind, "name", object.GetName(), "metaName", tree.GetMetaName(object))

			if provider == "virtual" { // Do not inherit virtual provider
				return "virtual", nil
			}
			if i > 0 && provider != prev { // If two children have different providers, don't inherit
				return "virtual", nil
			}
			prev = provider
		}

		return prev, nil
	} else {
		return lookUpProvider(object)
	}
}

// lookUpProvider returns the provider type of an object in the Cluster resource tree based off of the group in the GVK.
// If the object is a virtual object, the provider type is "virtual".
func lookUpProvider(object ctrlclient.Object) (string, error) {
	group := object.GetObjectKind().GroupVersionKind().Group
	if capiAPIVersionIndex := strings.Index(group, "cluster.x-k8s.io"); capiAPIVersionIndex < 0 {
		return "virtual", nil
	}

	providerIndex := strings.IndexByte(group, '.')
	if tree.IsVirtualObject(object) {
		return "virtual", nil
	} else if providerIndex > -1 {
		return group[:providerIndex], nil
	} else {
		return "", errors.Errorf("No provider found for object %s of %s \n", object.GetName(), object.GetObjectKind().GroupVersionKind().String())
	}
}

// getDisplayName returns the name of an object or the metaName if the object is virtual or has no name.
func getDisplayName(object ctrlclient.Object) string {
	metaName := tree.GetMetaName(object)
	displayName := object.GetName()
	if metaName != "" {
		if object.GetName() == "" || tree.IsVirtualObject(object) {
			displayName = metaName
		}
	}

	return displayName
}

// setReadyFields sets a marker on if an object has a ready condtion, and if so, whether it is ready or not and the severity of the condition.
func setReadyFields(object ctrlclient.Object, node *ClusterResourceNode) {
	if readyCondition := tree.GetReadyCondition(object); readyCondition != nil {
		node.HasReady = true
		node.Ready = readyCondition.Status == corev1.ConditionTrue
		node.Severity = string(readyCondition.Severity)
	}
}

// nodeArrayNames is a debug function that returns the <Kind>/<Name> of all nodes in a slice of ClusterResourceNodes.
func nodeArrayNames(nodes []*ClusterResourceNode) string {
	result := ""
	for _, node := range nodes {
		result += node.Kind + "/" + node.Name + " "
	}

	return result
}

// getSortKeys returns the sort keys for a node in the DescribeCluster resource tree. The sort keys are used to sort the children of a node.
func getSortKeys(node *ClusterResourceNode) []string {
	if node.Group == "virtual.cluster.x-k8s.io" {
		return []string{node.DisplayName, ""}
	}
	return []string{node.Kind, node.DisplayName}
}

// OwnerRefToObjectRef returns an object reference to the ownerRef of an object.
func OwnerRefToObjectRef(ownerRef metav1.OwnerReference, namespace string) *corev1.ObjectReference {
	return &corev1.ObjectReference{
		APIVersion: ownerRef.APIVersion,
		Kind:       ownerRef.Kind,
		Name:       ownerRef.Name,
		Namespace:  namespace,
	}
}

// pickOwner returns an object reference to the ownerRef of an object. If the object has multiple ownerRefs, it will attempt to pick the most appropriate one.
func pickOwner(ctx context.Context, c ctrlclient.Client, object ctrlclient.Object) *corev1.ObjectReference {
	log := ctrl.LoggerFrom(ctx)

	graph := NewOwnershipGraph(ctx, c, object)

	// Remove transitive owners from the graph to simplify the case where an object is owned by multiple objects, and one of the owners is owned by another owner.
	RemoveTransitiveOwners(object.GetUID(), graph)
	// Now that the graph is simplified, find the owners of the object.
	owners := graph.OwnerRefs[object.GetUID()]

	if len(owners) == 0 {
		log.V(4).Info("No owners found for object", "object", object.GetName())
		return nil
	}

	if len(owners) == 1 {
		for ownerUID := range owners {
			owner := graph.Objects[ownerUID]
			log.V(4).Info("Owner found for object", "ownerKind", owner.GetObjectKind().GroupVersionKind().Kind, "ownerName", owner.GetName(), "object", object.GetName())
			return &corev1.ObjectReference{
				APIVersion: owner.GetObjectKind().GroupVersionKind().GroupVersion().String(),
				Kind:       owner.GetObjectKind().GroupVersionKind().Kind,
				Name:       owner.GetName(),
				Namespace:  owner.GetNamespace(),
			}
		}
	}

	if len(owners) > 1 {
		log.V(4).Info("Found multiple owners for object", "numOwners", len(owners), "object", object.GetName())
		controllerRef := metav1.GetControllerOf(object)
		if controllerRef != nil {
			return OwnerRefToObjectRef(*controllerRef, object.GetNamespace())
		}

		return nil
	}

	return nil
}
