// Copyright 2020-2025 Politecnico di Torino
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

package instctrl

import (
	"context"
	"fmt"
	"reflect"

	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// EnforcePublicExposure ensures the presence or absence of the LoadBalancer service for public exposure.
func (r *InstanceReconciler) EnforcePublicExposure(ctx context.Context) error {
	instance := clctx.InstanceFrom(ctx)
	template := clctx.TemplateFrom(ctx)

	if template.Spec.AllowPublicExposure &&
		instance.Spec.Running && instance.Spec.PublicExposure != nil && len(instance.Spec.PublicExposure.Ports) > 0 {
		return r.enforcePublicExposurePresence(ctx)
	}
	return r.enforcePublicExposureAbsence(ctx)
}

// enforcePublicExposurePresence ensures the presence and correctness of the LoadBalancer service.
func (r *InstanceReconciler) enforcePublicExposurePresence(ctx context.Context) error {
	log := ctrl.LoggerFrom(ctx)
	instance := clctx.InstanceFrom(ctx)

	// Initialize PublicExposure status if nil
	if instance.Status.PublicExposure == nil {
		instance.Status.PublicExposure = &clv1alpha2.InstancePublicExposureStatus{}
	}

	// Validate the public exposure request first.
	if err := validatePublicExposureRequest(instance.Spec.PublicExposure.Ports); err != nil {
		log.Error(err, "invalid public exposure request")
		instance.Status.PublicExposure.Phase = clv1alpha2.PublicExposurePhaseError
		instance.Status.PublicExposure.Message = err.Error()
		return err
	}

	service := &v1.Service{
		ObjectMeta: forge.ObjectMetaWithSuffix(instance, forge.LabelPublicExposureValue),
	}

	// Try to get the existing service
	err := r.Get(ctx, client.ObjectKey{Name: service.Name, Namespace: instance.Namespace}, service)
	serviceExists := err == nil

	// If the service exists, check if its current spec matches the desired spec
	if serviceExists {
		desiredPorts := instance.Spec.PublicExposure.Ports
		currentPorts := []clv1alpha2.PublicServicePort{}
		for _, p := range service.Spec.Ports {
			currentPorts = append(currentPorts, clv1alpha2.PublicServicePort{
				Name:       p.Name,
				Port:       p.Port,
				TargetPort: p.TargetPort.IntVal,
				Protocol:   p.Protocol,
			})
		}
		currentIP := service.Annotations[r.PublicExposureOpts.LoadBalancerIPsKey]

		// If the current IP and ports match the desired, skip update
		if reflect.DeepEqual(desiredPorts, currentPorts) && currentIP != "" {
			log.Info("Service already matches desired state, skipping update")
			// Also update status if needed
			newStatus := &clv1alpha2.InstancePublicExposureStatus{
				ExternalIP: currentIP,
				Ports:      currentPorts,
				Phase:      clv1alpha2.PublicExposurePhaseReady,
			}
			if !reflect.DeepEqual(instance.Status.PublicExposure, newStatus) {
				instance.Status.PublicExposure = newStatus
				instance.Status.PublicExposure.Message = "" // Clear any previous error message
			}
			return nil
		}
	}

	// 1. Retrieve the map of used ports by other LoadBalancer services
	usedPortsByIP, err := UpdateUsedPortsByIP(ctx, r.Client, service.Name, instance.Namespace, &r.PublicExposureOpts)
	if err != nil {
		log.Error(err, "failed to get used ports by IP")
		instance.Status.PublicExposure.Phase = clv1alpha2.PublicExposurePhaseError
		instance.Status.PublicExposure.Message = "Failed to check free ports, contact the administrator."
		return err
	}

	instance.Status.PublicExposure.Phase = clv1alpha2.PublicExposurePhaseProvisioning
	instance.Status.PublicExposure.Message = "Provisioning in progress."

	// 2. Find the best IP and ports to assign using the logic from ip_manager.go
	targetIP, assignedPorts, err := r.FindBestIPAndAssignPorts(ctx, r.Client, instance, usedPortsByIP)
	if err != nil {
		log.Error(err, "failed to assign IP and ports for public exposure")
		instance.Status.PublicExposure.Phase = clv1alpha2.PublicExposurePhaseError
		instance.Status.PublicExposure.Message = err.Error()
		return err
	}

	// 3. Create or update the LoadBalancer Service
	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		// Set owner reference
		if err := controllerutil.SetControllerReference(instance, service, r.Scheme); err != nil {
			return err
		}

		// Set labels
		if service.Labels == nil {
			service.Labels = forge.LoadBalancerServiceLabels()
		}

		// Set annotations using the new options
		service.Annotations = forge.LoadBalancerServiceAnnotations(targetIP, &r.PublicExposureOpts)

		// Set spec
		service.Spec = forge.LoadBalancerServiceSpec(instance, assignedPorts)

		return nil
	})

	if err != nil {
		log.Error(err, "failed to create or update LoadBalancer service", "service", service.GetName())
		instance.Status.PublicExposure.Phase = clv1alpha2.PublicExposurePhaseError
		instance.Status.PublicExposure.Message = "Failed to create or update the LoadBalancer service, contact the administrator."
		return err
	}
	log.V(utils.FromResult(op)).Info("LoadBalancer service enforced", "service", service.GetName(), "result", op)

	// 4. Update the instance status
	instance.Status.PublicExposure.ExternalIP = targetIP
	instance.Status.PublicExposure.Ports = assignedPorts
	instance.Status.PublicExposure.Phase = clv1alpha2.PublicExposurePhaseReady
	instance.Status.PublicExposure.Message = "Public exposure completed successfully."

	// Enforce the network policy only after the service is ready.
	if err = r.enforcePublicExposureNetworkPolicyPresence(ctx); err != nil {
		log.Error(err, "failed to enforce public exposure network policy")
		instance.Status.PublicExposure.Phase = clv1alpha2.PublicExposurePhaseError
		instance.Status.PublicExposure.Message = "Failed to enforce the network policy, contact the administrator."
		return err
	}

	return nil
}

// enforcePublicExposureAbsence ensures the absence of the LoadBalancer service.
func (r *InstanceReconciler) enforcePublicExposureAbsence(ctx context.Context) error {
	instance := clctx.InstanceFrom(ctx)
	service := &v1.Service{
		ObjectMeta: forge.ObjectMetaWithSuffix(instance, forge.LabelPublicExposureValue),
	}

	// Remove the service if it exists
	if err := utils.EnforceObjectAbsence(ctx, r.Client, service, "service"); err != nil {
		return err
	}

	// Also remove the associated network policy.
	if err := r.enforcePublicExposureNetworkPolicyAbsence(ctx); err != nil {
		return err
	}

	// Clean up the status
	instance.Status.PublicExposure = nil
	return nil
}

// validatePublicExposureRequest checks for duplicates in ports and targetPorts.
func validatePublicExposureRequest(ports []clv1alpha2.PublicServicePort) error {
	seenPorts := make(map[int32]bool)
	seenTargetPorts := make(map[int32]bool)

	for _, p := range ports {
		// Check for duplicate TargetPorts
		if seenTargetPorts[p.TargetPort] {
			return fmt.Errorf("duplicate desired targetPort %d found in public exposure request", p.TargetPort)
		}
		seenTargetPorts[p.TargetPort] = true

		// Check for duplicate specified Ports (ignore 0, which is for auto-assignment)
		if p.Port != 0 {
			if seenPorts[p.Port] {
				return fmt.Errorf("duplicate requested port %d found in public exposure request", p.Port)
			}
			seenPorts[p.Port] = true
		}
	}

	return nil
}
