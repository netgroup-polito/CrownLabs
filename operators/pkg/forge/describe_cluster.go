package forge

import (
	"context"
	"fmt"
	"sort"
	"strings"

	visualizerv1 "github.com/Jont828/cluster-api-visualizer/api/v1"
	"github.com/gobuffalo/flect"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/tree"
	"sigs.k8s.io/cluster-api/controllers/external"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	maxGroupSize = 10

	addonsAmountToCollapse = 4
)

// ClusterResourceNode represents a node in the Cluster API resource tree and is used to configure the frontend with additional
// options like collapsibility and provider.
type ClusterResourceNode struct {
	Name            string                 `json:"name"`
	Namespace       string                 `json:"namespace"`
	DisplayName     string                 `json:"displayName"`
	Kind            string                 `json:"kind"`
	Group           string                 `json:"group"`
	Version         string                 `json:"version"`
	Provider        string                 `json:"provider"`
	UID             string                 `json:"uid"`
	CollapseWithTab bool                   `json:"collapseWithTab"`
	CollapseOnClick bool                   `json:"collapseOnClick"`
	Collapsible     bool                   `json:"collapsible"`
	Collapsed       bool                   `json:"collapsed"`
	Ready           bool                   `json:"ready"`
	Severity        string                 `json:"severity"`
	HasReady        bool                   `json:"hasReady"`
	Children        []*ClusterResourceNode `json:"children"`
}

type ClusterResourceTreeOptions struct {
	GroupMachines                bool
	AddControlPlaneVirtualNode   bool
	KindsToCollapse              map[string]struct{}
	VNodesToInheritChildProvider map[string]struct{}
	providerTypeOverrideMap      map[string]string
}

// ConstructClusterResourceTree returns a tree with nodes representing the Cluster API resources in the Cluster.
// Note: ObjectReferenceObjects do not have the virtual annotation so we can assume that all virtual objects are collapsible
func ConstructClusterResourceTree(ctx context.Context, defaultClient client.Client, runtimeClient ctrlclient.Client, dcOptions client.DescribeClusterOptions) (*ClusterResourceNode, *HTTPError) {
	objTree, err := defaultClient.DescribeCluster(ctx, dcOptions)
	if err != nil {
		if strings.HasSuffix(err.Error(), "not found") {
			return nil, &HTTPError{Status: 404, Message: err.Error()}
		}

		return nil, NewInternalError(err)
	}

	treeOptions := ClusterResourceTreeOptions{
		GroupMachines: true,
		KindsToCollapse: map[string]struct{}{
			"TemplateGroup":           {},
			"ClusterResourceSetGroup": {},
			"Machine":                 {},
		},
		VNodesToInheritChildProvider: map[string]struct{}{
			"ClusterResourceSetGroup": {},
			// "WorkerGroup":             {},
		},
	}

	overrides, err := injectCustomResourcesToObjectTree(ctx, runtimeClient, dcOptions, objTree)
	if err != nil {
		return nil, NewInternalError(err)
	}
	treeOptions.providerTypeOverrideMap = overrides

	resourceTree := objectTreeToResourceTree(ctx, objTree, objTree.GetRoot(), treeOptions)

	return resourceTree, nil
}

// objectTreeToResourceTree converts an clusterctl ObjectTree to a ClusterResourceNode tree.
func objectTreeToResourceTree(ctx context.Context, objTree *tree.ObjectTree, object ctrlclient.Object, treeOptions ClusterResourceTreeOptions) *ClusterResourceNode {
	log := ctrl.LoggerFrom(ctx)

	if object == nil {
		return nil
	}

	group := object.GetObjectKind().GroupVersionKind().Group
	kind := object.GetObjectKind().GroupVersionKind().Kind
	version := object.GetObjectKind().GroupVersionKind().Version

	_, collapsed := treeOptions.KindsToCollapse[kind]
	node := &ClusterResourceNode{
		Name:        object.GetName(),
		DisplayName: getDisplayName(object),
		Kind:        kind,
		Group:       group,
		Version:     version,
		Collapsed:   collapsed,
		Children:    []*ClusterResourceNode{},
		UID:         string(object.GetUID()),
	}
	if node.Namespace = object.GetNamespace(); node.Namespace == "" {
		node.Namespace = "default"
	}

	children := objTree.GetObjectsByParent(object.GetUID())
	provider, err := getProvider(ctx, object, children, treeOptions)
	if err != nil {
		log.Error(err, "failed to get provider for object", "kind", kind, "name", object.GetName())
	}
	node.Provider = provider

	setReadyFields(object, node)

	childTrees := []*ClusterResourceNode{}
	for _, child := range children {
		childTrees = append(childTrees, objectTreeToResourceTree(ctx, objTree, child, treeOptions))
	}

	log.V(4).Info("Node is", "node", node.Kind+"/"+node.Name)
	if treeOptions.GroupMachines {
		node.Children = createKindGroupNode(ctx, object.GetNamespace(), "Machine", "cluster", childTrees, maxGroupSize)
	} else {
		node.Children = childTrees
	}

	if kind == "Cluster" {
		node.Children = addAddonsGroupNode(ctx, node.Children)
	}

	// If the resource represents a real CRD we want to collapse, and it has children, we can collapse it with tab.
	node.CollapseOnClick = tree.IsVirtualObject(object)
	node.CollapseWithTab = len(node.Children) > 0 && !node.CollapseOnClick
	node.Collapsible = node.CollapseWithTab || node.CollapseOnClick

	sort.Slice(node.Children, func(i, j int) bool {
		// TODO: make sure this is deterministic!
		if getSortKeys(node.Children[i])[0] == getSortKeys(node.Children[j])[0] {
			return getSortKeys(node.Children[i])[1] < getSortKeys(node.Children[j])[1]
		}
		return getSortKeys(node.Children[i])[0] < getSortKeys(node.Children[j])[0]
	})

	return node
}

// addAddonsGroupNode finds all objects in children with `provider=addons` and create a parent node for them.
func addAddonsGroupNode(_ context.Context, children []*ClusterResourceNode) []*ClusterResourceNode {
	resultChildren := []*ClusterResourceNode{}

	addonsParent := &ClusterResourceNode{
		Name:            "",
		DisplayName:     "Add-ons",
		Kind:            "AddonsGroup",
		Provider:        "addons",
		CollapseWithTab: false,
		CollapseOnClick: true,
		Collapsible:     true,
		Collapsed:       false,
		HasReady:        false,
		Ready:           true,
		Severity:        "",
		UID:             "addons",
	}

	for _, child := range children {
		if child.Provider == "addons" {
			addonsParent.Children = append(addonsParent.Children, child)
		} else {
			resultChildren = append(resultChildren, child)
		}
	}

	if len(addonsParent.Children) == 1 && addonsParent.Children[0].Kind == "ClusterResourceSetGroup" && addonsParent.Children[0].Name == "ClusterResourceSets" {
		// If the only add-ons are the CRS group node, just remove it and make the add-ons the new parent.
		addonsParent.Children = addonsParent.Children[0].Children
	}

	if len(addonsParent.Children) >= addonsAmountToCollapse {
		addonsParent.Collapsed = true
	}

	if len(addonsParent.Children) > 0 {
		resultChildren = append(resultChildren, addonsParent)
	}

	return resultChildren
}

// createKindGroupNode finds all objects in children with `kind` and create a parent node for them.
func createKindGroupNode(ctx context.Context, namespace string, kind string, provider string, children []*ClusterResourceNode, maxGroupSize int) []*ClusterResourceNode {
	log := ctrl.LoggerFrom(ctx)

	log.V(4).Info("Starting children are ", "children", nodeArrayNames(children))

	sort.Slice(children, func(i, j int) bool {
		// TODO: make sure this is deterministic!
		if getSortKeys(children[i])[0] == getSortKeys(children[j])[0] {
			return getSortKeys(children[i])[1] < getSortKeys(children[j])[1]
		}
		return getSortKeys(children[i])[0] < getSortKeys(children[j])[0]
	})

	resultChildren := []*ClusterResourceNode{}

	// TODO: maybe in the future, we can group based on severity/error, but we'd still need a way to make sure the groups aren't too large.
	// Init a parent node, if the child groups need to be broken up. For example, if we have 100 machines, it would be
	// [MachineSet] -> [30 Machines] -> [10 Machines, 10 Machines, 10 Machines]
	groupParent := &ClusterResourceNode{
		Name:            "",
		Namespace:       namespace,
		DisplayName:     "",
		Kind:            kind,
		Provider:        provider,
		CollapseWithTab: false,
		CollapseOnClick: true,
		Collapsible:     true,
		Collapsed:       true,
		HasReady:        false,
		Ready:           true,
		Severity:        "",
		UID:             kind + ": ",
	}

	groupNodes := []*ClusterResourceNode{}
	groupNode := &ClusterResourceNode{}
	kindCount := 0
	totalKindCount := 0
	for _, child := range children {
		if kindCount == 0 {
			groupNode = &ClusterResourceNode{
				Name:            "",
				Namespace:       namespace,
				DisplayName:     "",
				Kind:            kind,
				Provider:        provider, // TODO: don't hardcode this
				CollapseWithTab: false,
				CollapseOnClick: true,
				Collapsible:     true,
				Collapsed:       true,
				Children:        []*ClusterResourceNode{},
				HasReady:        false,
				Ready:           true,
				Severity:        "",
				UID:             kind + ": ",
			}
		}
		if child.Kind == kind {
			kindCount++
			totalKindCount++
			groupNode.Group = child.Group
			groupNode.Version = child.Version
			groupNode.Children = append(groupNode.Children, child)
			groupNode.UID += child.UID + " "

			groupParent.Group = child.Group
			groupParent.Version = child.Version
			groupParent.UID += child.UID + " "
			if child.HasReady {
				groupNode.HasReady = true
				groupNode.Ready = child.Ready && groupNode.Ready
				groupNode.Severity = updateSeverityIfMoreSevere(groupNode.Severity, child.Severity)
				// Set severity based on most severe child, i.e. Error > Warning > Info > Success
				groupParent.HasReady = true
				groupParent.Ready = child.Ready && groupParent.Ready
				groupParent.Severity = updateSeverityIfMoreSevere(groupParent.Severity, child.Severity)
			}
		} else {
			resultChildren = append(resultChildren, child)
		}

		if kindCount >= maxGroupSize {
			groupNode.DisplayName = fmt.Sprintf("%d %s", kindCount, flect.Pluralize(kind))
			groupNodes = append(groupNodes, groupNode)
			kindCount = 0
		}
	}

	if totalKindCount == 1 {
		// Don't group if there is only one, and there are no other groups.
		groupNodes = append(groupNodes, groupNode.Children...)
	} else if kindCount == 1 {
		// If there is only one, and there are other groups, group it without pluralizing.
		groupNode.DisplayName = fmt.Sprintf("1 %s", kind)
		groupNodes = append(groupNodes, groupNode)
	} else if kindCount > 1 {
		// Otherwise, create a group for the remaining children, to account for remainders, i.e. 8 left overe in groups of 10.
		groupNode.DisplayName = fmt.Sprintf("%d %s", kindCount, flect.Pluralize(kind))
		groupNodes = append(groupNodes, groupNode)
	}

	if len(groupNodes) > 1 { // If we have multiple groups, add a parent node for them.
		groupParent.DisplayName = fmt.Sprintf("%d %s", totalKindCount, flect.Pluralize(kind))
		groupParent.Children = groupNodes
		resultChildren = append(resultChildren, groupParent)
	} else { // Otherwise, just add the single group node without a parent.
		resultChildren = append(resultChildren, groupNodes...)
	}

	log.V(4).Info("Result children are", "children", nodeArrayNames(resultChildren))

	return resultChildren
}

// injectCustomResourcesToObjectTree amends the clusterctl ObjectTree with custom CRDs that are not included in the clusterctl resource discovery.
// It queries all CRD types and their instances containing the visualizer label and the cluster name label.
func injectCustomResourcesToObjectTree(ctx context.Context, c ctrlclient.Client, dcOptions client.DescribeClusterOptions, objTree *tree.ObjectTree) (map[string]string, error) {
	log := ctrl.LoggerFrom(ctx)

	log.V(4).Info("Adding user specified custom resources to object tree", "namespace", dcOptions.Namespace, "clusterName", dcOptions.ClusterName)

	crds, err := getCRDList(ctx, c, ctrlclient.MatchingLabels{visualizerv1.VisualizeResourceLabel: ""})
	if err != nil {
		return nil, err
	}

	namespace := dcOptions.Namespace
	clusterName := dcOptions.ClusterName

	clusterObjSelector := []ctrlclient.ListOption{
		ctrlclient.InNamespace(namespace),
		ctrlclient.MatchingLabels{clusterv1.ClusterNameLabel: clusterName},
	}

	providerTypeOverrideMap := make(map[string]string)
	clusterObjects := []ctrlclient.Object{}
	for _, crd := range crds {
		crdLabels := crd.GetLabels()
		if crdLabels != nil {
			if provider, ok := crdLabels[visualizerv1.ProviderTypeLabel]; ok {
				switch provider {
				case "cluster":
					fallthrough
				case "bootstrap":
					fallthrough
				case "controlplane":
					fallthrough
				case "infrastructure":
					fallthrough
				case "addons":
					fallthrough
				case "virtual":
					providerTypeOverrideMap[crd.Spec.Names.Kind] = provider
				default:
					return nil, errors.Errorf("Invalid provider type %s for CRD type %s \n", provider, crd.GetName())
				}
			}
		}

		for _, version := range crd.Spec.Versions {
			typeMeta := metav1.TypeMeta{
				Kind: crd.Spec.Names.Kind,
				APIVersion: metav1.GroupVersion{
					Group:   crd.Spec.Group,
					Version: version.Name,
				}.String(),
			}

			clusterObjList, err := getObjList(ctx, c, typeMeta, clusterObjSelector)
			if err != nil {
				return nil, err
			}

			for i := range clusterObjList.Items {
				clusterObj := &clusterObjList.Items[i]
				clusterObjects = append(clusterObjects, clusterObj)

				// Add the CRD to the object tree
			}

		}
	}

	for i := range clusterObjects {
		object := clusterObjects[i]
		// Make sure not to implicitly reference loop variable!
		if err := ensureObjConnectedTotree(ctx, c, objTree, object); err != nil {
			return nil, err
		}
	}

	return providerTypeOverrideMap, nil
}

// ensureObjConnectedTotree ensures that the object is connected to the tree by adding it and its parents until a parent is owned by the Cluster (root node).
// If a parent has no owner, it is set as a child of the Cluster.
// Note: At the moment, this only supports a use case where an object has only one owner which is also set the controller.
func ensureObjConnectedTotree(ctx context.Context, c ctrlclient.Client, objTree *tree.ObjectTree, object ctrlclient.Object) error {
	log := ctrl.LoggerFrom(ctx)

	if objTree.GetObject(object.GetUID()) != nil || objTree.GetRoot().GetUID() == object.GetUID() {
		log.V(4).Info("Object already in tree", "kind", object.GetObjectKind().GroupVersionKind().Kind, "name", object.GetName(), "namespace", object.GetNamespace())
		return nil
	}

	log.V(4).Info("Adding object to tree", "kind", object.GetObjectKind().GroupVersionKind().Kind, "name", object.GetName(), "namespace", object.GetNamespace())
	var parent ctrlclient.Object
	// TODO: handle case where there is no controllerRef or how to resolve multiple owners.
	ref := pickOwner(ctx, c, object)
	if ref != nil {
		if p, err := external.Get(ctx, c, ref, object.GetNamespace()); err != nil {
			return err
		} else {
			parent = p
		}
	} else {
		// If no ownerRef, set to root.
		parent = objTree.GetRoot()
		// TODO: look into creating an add-ons virtual node.
	}

	ensureObjConnectedTotree(ctx, c, objTree, parent)

	added, _ := objTree.Add(parent, object)
	if !added {
		return fmt.Errorf("failed to add object %s to tree", object.GetName())
	}

	return nil
}
