package instanceset_landing

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const serviceTenantName = "service-tenant"

func generateInstanceName(userID, courseID string) string {
	return "ex-" + courseID + "-" + userID + "-crownlabs"
}

func enforceOnDemandInstance(ctx context.Context, instanceName string) error {
	klog.Infof("Creating instance %s", instanceName)

	// TODO: handle dataSource if Options.StartupOpts.InitialContentSourceURL

	instance := &clv1alpha2.Instance{}
	instance.SetName(instanceName)
	instance.SetNamespace(Options.Namespace)

	op, err := ctrl.CreateOrUpdate(ctx, Client, instance, func() error {
		if instance.CreationTimestamp.IsZero() {
			instance.Spec = clv1alpha2.InstanceSpec{
				Running:  true,
				Template: Options.Template,
				Tenant: clv1alpha2.GenericRef{
					Name: serviceTenantName,
				},
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("creation error: Instance %s can't be enforced [%w]", instanceName, err)
	}

	klog.Infof("Instance %s successfully %s", instanceName, op)

	return nil
}
