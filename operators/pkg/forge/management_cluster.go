package forge

import (
	"context"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd/api"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type MultiClusterTreeNode struct {
	Name                   string                  `json:"name"`
	Namespace              string                  `json:"namespace"`
	InfrastructureProvider string                  `json:"infrastructureProvider"`
	IsManagement           bool                    `json:"isManagement"`
	Phase                  string                  `json:"phase"`
	Ready                  bool                    `json:"ready"`
	Children               []*MultiClusterTreeNode `json:"children"`
}

// ConstructMultiClusterTree returns a tree representing the workload cluster discovered in the management cluster.
func ConstructMultiClusterTree(ctx context.Context, ctrlClient ctrlclient.Client, k8sConfigClient *api.Config) (*MultiClusterTreeNode, *HTTPError) {
	log := ctrl.LoggerFrom(ctx)

	currentContextName := k8sConfigClient.CurrentContext
	currentContext, ok := k8sConfigClient.Contexts[currentContextName]
	currentNamespace := currentContext.Namespace
	if !ok {
		return nil, &HTTPError{Status: 404, Message: "current context not found"}
	}
	name := currentContext.Cluster

	root := &MultiClusterTreeNode{
		Name:                   name,
		Namespace:              currentNamespace,
		InfrastructureProvider: "",
		Children:               []*MultiClusterTreeNode{},
		IsManagement:           true,
	}

	clusterList := &clusterv1.ClusterList{}

	// TODO: should we use ctrlClient.MatchingLabels or try to use the labelSelector itself?
	if err := ctrlClient.List(ctx, clusterList); err != nil {
		return nil, NewInternalError(err)
	}

	if clusterList == nil || len(clusterList.Items) == 0 {
		log.V(4).Info("No workload clusters found")
		return root, nil
	}
	sort.Slice(clusterList.Items, func(i, j int) bool {
		// This must be deterministic, otherwise the tree will be different between runs.
		// In this case, we can't have two clusters with the same name.
		return clusterList.Items[i].GetName() < clusterList.Items[j].GetName()
	})

	for _, cluster := range clusterList.Items {
		readyCondition := conditions.Get(&cluster, clusterv1.ReadyCondition)

		workloadCluster := MultiClusterTreeNode{
			Name:         cluster.GetName(),
			Namespace:    cluster.GetNamespace(),
			IsManagement: false,
			Phase:        cluster.Status.Phase,
			Ready:        readyCondition != nil && readyCondition.Status == corev1.ConditionTrue,
			Children:     []*MultiClusterTreeNode{},
		}

		// TODO: edge case in topology Clusters where infraRef is nil if it fails to create.
		// In that case we can try to fetch the ClusterClass and read the infraRef from there and trim the output, i.e. DockerClusterTemplate => DockerCluster.
		if cluster.Spec.InfrastructureRef != nil {
			workloadCluster.InfrastructureProvider = cluster.Spec.InfrastructureRef.Kind
		}

		root.Children = append(root.Children, &workloadCluster)
	}

	return root, nil
}
