package forge

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/cluster-api/controllers/external"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// OwnershipGraph is a graph of objects and their ownerRefs.
type OwnershipGraph struct {
	// Objects is a map of objects indexed by their UID. They indicate nodes in the graph.
	Objects map[types.UID]ctrlclient.Object

	// OwnerRefs is a map of objects to a set of their ownerRefs. They indicate directed edges in the graph, such that an edge from
	// node I to J means that I is owned by J.
	// The set is implemented as a `map[types.UID]struct{}` (meaning a map with empty structs as values) for fast lookup.
	OwnerRefs map[types.UID]map[types.UID]struct{}
}

// NewOwnershipGraph returns a new OwnershipGraph by recursively traversing the ownerRefs of the given object.
func NewOwnershipGraph(ctx context.Context, c ctrlclient.Client, object ctrlclient.Object) *OwnershipGraph {
	ownershipGraph := &OwnershipGraph{
		Objects:   make(map[types.UID]ctrlclient.Object),
		OwnerRefs: make(map[types.UID]map[types.UID]struct{}),
	}
	constructOwnershipGraph(ctx, c, object, ownershipGraph)

	return ownershipGraph
}

// constructOwnershipGraph is a helper that recursively constructs the ownership graph by traversing the ownerRefs of the given object.
// It is implemented as follows:
// 1. Add the given object to the graph.
// 2. For each ownerRef of the given object, add the owner to the graph and add an edge from the given object to the owner.
// 3. Recursively call constructOwnershipGraph for each owner.
func constructOwnershipGraph(ctx context.Context, c ctrlclient.Client, object ctrlclient.Object, ownershipGraph *OwnershipGraph) error {
	ownershipGraph.Objects[object.GetUID()] = object
	for _, ownerRef := range object.GetOwnerReferences() {
		ref := OwnerRefToObjectRef(ownerRef, object.GetNamespace())
		owner, err := external.Get(ctx, c, ref, object.GetNamespace())
		if err != nil {
			return err
		}
		if ownershipGraph.OwnerRefs[object.GetUID()] == nil {
			ownershipGraph.OwnerRefs[object.GetUID()] = map[types.UID]struct{}{}
		}
		ownershipGraph.OwnerRefs[object.GetUID()][owner.GetUID()] = struct{}{}
		constructOwnershipGraph(ctx, c, owner, ownershipGraph)
	}

	return nil
}

// RemoveTransitiveOwners removes transitive owners from the graph, i.e. if I is owned by J and K, and J is also owned by K,
// then K is removed as an owner of I as it is implicitly owned by K through J.
func RemoveTransitiveOwners(start types.UID, ownershipGraph *OwnershipGraph) {
	removeTransitiveOwnersHelper(start, start, ownershipGraph)
}

// removeTransitiveOwnersHelper is a helper that recursively removes transitive owners from the graph.
// It is implemented as a DFS from start node i, such that:
// 1. For each owner j of i, i.e. edge (i, j) exists, start a DFS from node j.
// 2. For each node k reachable from j, remove the edge (i, k) if it exists, as k is inherently a transitive owner.
//
// In plain English: Delete any owner reachable from the owners of the start node, since if we can find it it means it's implicitly owned.
func removeTransitiveOwnersHelper(start types.UID, current types.UID, ownershipGraph *OwnershipGraph) {
	// If j exists in ownerRefs[i][j], then j is owned by i.
	for ownerUID := range ownershipGraph.OwnerRefs[current] {
		if current != start {
			delete(ownershipGraph.OwnerRefs[start], ownerUID)
		}
		removeTransitiveOwnersHelper(start, ownerUID, ownershipGraph)
	}
}
