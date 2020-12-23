package instance_creation

import (
	"context"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetPublicKeys(c client.Client, ctx context.Context, name string, namespace string, publicKeys *[]string) error {
	tenant := crownlabsv1alpha1.Tenant{}
	nsdName := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	if err := c.Get(ctx, nsdName, &tenant); err == nil {

		*publicKeys = append(*publicKeys, tenant.Spec.PublicKeys...)

		return nil
	} else {
		return err
	}
}
