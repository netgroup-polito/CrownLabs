// Copyright 2020-2021 Politecnico di Torino
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

package instanceset_landing

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

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
					Name: clv1alpha2.SVCTenantName,
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
