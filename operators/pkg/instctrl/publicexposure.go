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

	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
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
		instance.Spec.Running && instance.Spec.PublicExposure != nil &&
		instance.Spec.PublicExposure.Ports != nil && len(instance.Spec.PublicExposure.Ports) > 0 {
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

	instance.Status.PublicExposure.Phase = clv1alpha2.PublicExposurePhaseProvisioning
	instance.Status.PublicExposure.Message = "Provisioning public exposure: allocating IP and ports"

	// Validate the public exposure request first.
	if err := ValidatePublicExposureRequest(instance.Spec.PublicExposure.Ports); err != nil {
		log.Error(err, "invalid public exposure request")
		instance.Status.PublicExposure.Phase = clv1alpha2.PublicExposurePhaseError
		instance.Status.PublicExposure.Message = "Invalid public exposure request: " + err.Error()
		return err
	}

	service := &v1.Service{
		ObjectMeta: forge.ObjectMetaWithSuffix(instance, forge.LabelPublicExposureValue),
	}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		// Set owner reference
		if err := controllerutil.SetControllerReference(instance, service, r.Scheme); err != nil {
			return err
		}

		// Check if the current IP is still present in the current IPPool (if updating)
		currentIP := service.Annotations[r.PublicExposureOpts.LoadBalancerIPsKey]
		ipStillValid := false
		for _, ip := range r.PublicExposureOpts.IPPool {
			if ip == currentIP {
				ipStillValid = true
				break
			}
		}
		if currentIP != "" && !ipStillValid {
			// If the current IP is not valid, clear annotations/spec to force reassignment
			service.Annotations[r.PublicExposureOpts.LoadBalancerIPsKey] = ""
			service.Spec.Ports = nil
		} else if currentIP != "" {
			// If the current IP is valid, check if the requested ports match the current ones
			// If they match, do nothing, already in desired state
			// If they don't match, we need to update the service
			specPorts := instance.Spec.PublicExposure.Ports
			statusPorts := instance.Status.PublicExposure.Ports
			svcPorts := make([]clv1alpha2.PublicServicePort, len(service.Spec.Ports))
			for i, p := range service.Spec.Ports {
				svcPorts[i] = clv1alpha2.PublicServicePort{
					Name:       p.Name,
					Port:       p.Port,
					TargetPort: p.TargetPort.IntVal,
					Protocol:   p.Protocol,
				}
			}
			if !needsServiceUpdate(specPorts, statusPorts, svcPorts) {
				// If the current IP is valid and the requested ports match the current ones, do nothing, already in desired state
				log.Info("LoadBalancer service already in desired state", "service", service.GetName())
				return nil
			}
		}

		// Retrieve the map of used ports by other LoadBalancer services
		usedPortsByIP, err := UpdateUsedPortsByIP(ctx, r.Client, service.Name, instance.Namespace, &r.PublicExposureOpts)
		if err != nil {
			return fmt.Errorf("failed to get used ports by IP: %w", err)
		}

		// Find the best IP and ports to assign using the logic from ip_manager.go
		targetIP, assignedPorts, err := r.FindBestIPAndAssignPorts(ctx, instance, usedPortsByIP, currentIP)
		if err != nil {
			return fmt.Errorf("failed to assign IP and ports for public exposure: %w", err)
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

	// Update some pieces of the instance status only after LoadBalancer service is created/updated
	currentIP := service.Annotations[r.PublicExposureOpts.LoadBalancerIPsKey]
	assignedPorts := []clv1alpha2.PublicServicePort{}
	for _, p := range service.Spec.Ports {
		assignedPorts = append(assignedPorts, clv1alpha2.PublicServicePort{
			Name:       p.Name,
			Port:       p.Port,
			TargetPort: p.TargetPort.IntVal,
			Protocol:   p.Protocol,
		})
	}

	instance.Status.PublicExposure.ExternalIP = currentIP
	instance.Status.PublicExposure.Ports = assignedPorts // Use them also for the NetworkPolicy

	// Enforce the network policy before updating the instance status
	if err = r.enforcePublicExposureNetworkPolicyPresence(ctx); err != nil {
		log.Error(err, "failed to enforce public exposure network policy")
		instance.Status.PublicExposure.Phase = clv1alpha2.PublicExposurePhaseError
		instance.Status.PublicExposure.Message = "Failed to enforce the network policy, contact the administrator."
		return err
	}

	instance.Status.PublicExposure.Phase = clv1alpha2.PublicExposurePhaseReady
	instance.Status.PublicExposure.Message = "Public exposure completed successfully."
	log.Info("Public exposure successfully enforced", "instance", instance.Name, "externalIP", currentIP, "ports", assignedPorts)

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

// ValidatePublicExposureRequest checks for duplicates in ports and targetPorts.
func ValidatePublicExposureRequest(ports []clv1alpha2.PublicServicePort) error {
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

// needsServiceUpdate returns true if the requested ports in spec differ from the actual ones (status/Service),
// considering that Port==0 in spec means "any assigned port". The comparison is robust to port order.
func needsServiceUpdate(specPorts, statusPorts, svcPorts []clv1alpha2.PublicServicePort) bool {
	type portKey struct {
		Name     string
		Target   int32
		Protocol v1.Protocol
	}

	// Build set of keys from spec
	specKeys := make(map[portKey]struct{}, len(specPorts))
	for _, p := range specPorts {
		specKeys[portKey{p.Name, p.TargetPort, p.Protocol}] = struct{}{}
	}

	// Helper: check that all ports in a slice are present in specKeys, and viceversa
	matchKeys := func(ports []clv1alpha2.PublicServicePort) bool {
		if len(ports) != len(specKeys) {
			return false
		}
		seen := make(map[portKey]struct{}, len(ports))
		for _, p := range ports {
			k := portKey{p.Name, p.TargetPort, p.Protocol}
			if _, ok := specKeys[k]; !ok {
				return false
			}
			seen[k] = struct{}{}
		}
		// Check that all specKeys are present
		for k := range specKeys {
			if _, ok := seen[k]; !ok {
				return false
			}
		}
		return true
	}

	if !matchKeys(statusPorts) || !matchKeys(svcPorts) {
		return true
	}

	// For each port in spec with Port != 0, Port must match in both status and svc
	portMatch := func(ports []clv1alpha2.PublicServicePort, ref clv1alpha2.PublicServicePort) bool {
		for _, p := range ports {
			if p.Name == ref.Name && p.TargetPort == ref.TargetPort && p.Protocol == ref.Protocol && p.Port == ref.Port {
				return true
			}
		}
		return false
	}
	// Helper: for ports with Port==0 in spec, check that in status and svc the corresponding port is not also specified (i.e., it must be auto-assigned )
	autoPortMismatch := func(ports []clv1alpha2.PublicServicePort, ref clv1alpha2.PublicServicePort) bool {
		for _, q := range ports {
			if q.Name == ref.Name && q.TargetPort == ref.TargetPort && q.Protocol == ref.Protocol && q.Port < forge.BasePortForAutomaticAssignment {
				return true
			}
		}
		return false
	}

	for _, p := range specPorts {
		if p.Port != 0 {
			if !portMatch(statusPorts, p) || !portMatch(svcPorts, p) {
				return true
			}
		} else {
			if autoPortMismatch(statusPorts, p) || autoPortMismatch(svcPorts, p) {
				return true
			}
		}
	}
	return false
}
