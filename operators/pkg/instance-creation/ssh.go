package instance_creation

import (
	"context"

	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetPublicKeys(c client.Client, ctx context.Context, Tenant crownlabsv1alpha2.GenericRef, Template crownlabsv1alpha2.GenericRef, publicKeys *[]string) error {
	tenant := crownlabsv1alpha1.Tenant{}
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: Tenant.Namespace,
		Name:      Tenant.Name,
	}, &tenant); err != nil {
		return err
	}

	*publicKeys = append(*publicKeys, tenant.Spec.PublicKeys...)

	template := crownlabsv1alpha2.Template{}
	if err := c.Get(ctx, types.NamespacedName{
		Namespace: Template.Namespace,
		Name:      Template.Name,
	}, &template); err != nil {
		return err
	}

	label := map[string]string{crownlabsv1alpha1.WorkspaceLabelPrefix + template.Spec.WorkspaceRef.Name: "manager"}

	var managers crownlabsv1alpha1.TenantList
	if err := c.List(context.Background(), &managers, client.MatchingLabels(label)); apierrors.IsNotFound(err) {
		// if there are no managers in this worspace there's nothing to do
		return nil
	} else if err != nil {
		return err
	}

	for _, manager := range managers.Items {
		// avoid duplicates
		if manager.Name != tenant.Name {
			*publicKeys = append(*publicKeys, manager.Spec.PublicKeys...)
		}
	}

	return nil

}
