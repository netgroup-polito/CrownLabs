package publicexposure

import (
	"context"
	"fmt"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	metallbPoolName = "my-ip-pool" // Unifica con il controller
	sharedIPValue   = "true"       // Unifica con il controller
	basePort        = 30000
)

// Manager gestisce l'esposizione pubblica delle istanze
type Manager struct {
	client.Client
	Scheme *runtime.Scheme
}

// NewManager crea un nuovo manager per l'esposizione pubblica
func NewManager(client client.Client, scheme *runtime.Scheme) *Manager {
	return &Manager{
		Client: client,
		Scheme: scheme,
	}
}

// ReconcileExposure riconcilia l'esposizione pubblica per un'istanza
func (m *Manager) ReconcileExposure(ctx context.Context, instance *clv1alpha2.Instance) error {
	log := ctrl.LoggerFrom(ctx)
	log.Info("Reconciling instance exposure", "instance", instance.Name)

	svcName := fmt.Sprintf("instance-lb-%s", instance.Name)
	existingSvc := &v1.Service{}
	err := m.Get(ctx, types.NamespacedName{Name: svcName, Namespace: instance.Namespace}, existingSvc)
	svcExists := err == nil

	// Verifica se è richiesta l'esposizione
	if instance.Spec.PublicExposure == nil || len(instance.Spec.PublicExposure.Services) == 0 {
		if svcExists {
			log.Info("Removing LoadBalancer service as exposure is no longer required", "service", svcName)
			return m.Delete(ctx, existingSvc)
		}
		return nil
	}

	// SE IL SERVIZIO ESISTE GIÀ, VERIFICA SE È GIÀ CORRETTO PRIMA DI FARE QUALSIASI ALTRA OPERAZIONE
	if svcExists {
		// Verifica se la configurazione attuale è già quella desiderata
		if m.serviceMatchesDesiredConfig(existingSvc, instance) {
			log.Info("Service already has correct configuration, skipping update", "service", svcName)
			return nil
		}
	}

	// Ottieni le porte utilizzate (escludendo il servizio corrente)
	usedPortsByIP, err := m.updateUsedPortsByIP(ctx, instance.Namespace, svcName)
	if err != nil {
		return err
	}

	// Trova il miglior IP e assegna le porte
	targetIP, assignedPorts, err := m.findBestIPAndAssignPorts(ctx, instance, usedPortsByIP)
	if err != nil {
		return err
	}

	// Crea o aggiorna il servizio
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: instance.Namespace,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, m.Client, svc, func() error {
		// Imposta owner reference
		if err := controllerutil.SetControllerReference(instance, svc, m.Scheme); err != nil {
			return err
		}

		// Configura il servizio
		svc.Spec.Type = v1.ServiceTypeLoadBalancer
		svc.Spec.Selector = map[string]string{
			"crownlabs.polito.it/instance": instance.Name,
		}

		// Imposta le annotazioni per MetalLB
		if svc.Annotations == nil {
			svc.Annotations = make(map[string]string)
		}
		svc.Annotations["metallb.universe.tf/address-pool"] = metallbPoolName
		svc.Annotations["metallb.universe.tf/allow-shared-ip"] = sharedIPValue
		svc.Annotations["metallb.universe.tf/loadBalancerIPs"] = targetIP

		// Configura le porte del servizio
		svc.Spec.Ports = []v1.ServicePort{}
		for _, p := range assignedPorts {
			svc.Spec.Ports = append(svc.Spec.Ports, v1.ServicePort{
				Name:       p.Name,
				Port:       p.AssignedPort,
				TargetPort: intstr.FromInt(int(p.TargetPort)),
				Protocol:   v1.ProtocolTCP,
			})
		}

		return nil
	})

	if err != nil {
		return err
	}

	log.Info("Service exposure reconciled", "service", svcName, "ip", targetIP)
	return nil
}

// serviceMatchesDesiredConfig verifica se il servizio esistente corrisponde già alla configurazione desiderata
func (m *Manager) serviceMatchesDesiredConfig(svc *v1.Service, instance *clv1alpha2.Instance) bool {
	// Verifica che il servizio abbia le porte corrette
	expectedPorts := make(map[string]clv1alpha2.ServicePortMapping)
	for _, p := range instance.Spec.PublicExposure.Services {
		expectedPorts[p.Name] = p
	}

	// Se il numero di porte non corrisponde, il servizio deve essere aggiornato
	if len(svc.Spec.Ports) != len(expectedPorts) {
		return false
	}

	// Verifica ogni porta
	for _, port := range svc.Spec.Ports {
		expectedPort, exists := expectedPorts[port.Name]
		if !exists {
			return false
		}

		// Per le porte specificate (non 0), verifica che corrispondano
		if expectedPort.Port != 0 && port.Port != expectedPort.Port {
			return false
		}

		// Verifica che la target port corrisponda
		if port.TargetPort.IntVal != expectedPort.TargetPort {
			return false
		}
	}

	return true
}

// CleanupExposure rimuove il servizio LoadBalancer se non più necessario
func (m *Manager) CleanupExposure(ctx context.Context, instance *clv1alpha2.Instance) error {
	log := ctrl.LoggerFrom(ctx)
	svcName := fmt.Sprintf("instance-lb-%s", instance.Name)

	svc := &v1.Service{}
	err := m.Get(ctx, types.NamespacedName{Name: svcName, Namespace: instance.Namespace}, svc)
	if err == nil {
		log.Info("Removing LoadBalancer service", "service", svcName)
		return m.Delete(ctx, svc)
	}
	return client.IgnoreNotFound(err)
}
