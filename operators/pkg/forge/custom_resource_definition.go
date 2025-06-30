package forge

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/cluster-api/controllers/external"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func GetCustomResource(ctx context.Context, runtimeClient ctrlclient.Client, kind string, apiVersion string, namespace string, name string) (*unstructured.Unstructured, *HTTPError) {
	objectRef := corev1.ObjectReference{
		Kind:       kind,
		Namespace:  namespace,
		Name:       name,
		APIVersion: apiVersion,
	}
	object, err := external.Get(ctx, runtimeClient, &objectRef, namespace)
	if err != nil {
		return nil, &HTTPError{404, err.Error()}
	}

	// Remove managed fields for now since they're read only and don't show in `kubectl get -o yaml`
	object.SetManagedFields(nil)

	return object, nil
}
