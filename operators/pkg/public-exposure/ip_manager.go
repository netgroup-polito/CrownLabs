package publicexposure

import (
	"context"
	"fmt"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// getMetalLBIPPool recupera il pool di IP configurato in MetalLB
func (m *Manager) getMetalLBIPPool(ctx context.Context) ([]string, error) {
	// TODO: Implementare logica per ottenere il pool di IP da MetalLB nel cluster
	// Per ora, restituisce un pool statico
	return []string{
		"172.18.0.240", "172.18.0.241", "172.18.0.242", "172.18.0.243",
		"172.18.0.244", "172.18.0.245", "172.18.0.246", "172.18.0.247",
		"172.18.0.248", "172.18.0.249", "172.18.0.250",
	}, nil
}

// findBestIPAndAssignPorts trova il miglior IP per le porte richieste e gestisce l'assegnazione delle porte
func (m *Manager) findBestIPAndAssignPorts(ctx context.Context, instance *clv1alpha2.Instance, usedPortsByIP map[string]map[int]bool) (string, []clv1alpha2.ServicePortMapping, error) {
	log := ctrl.LoggerFrom(ctx)

	// Controlla se esiste già un servizio per questa istanza e usa lo stesso IP se possibile
	svcName := fmt.Sprintf("instance-lb-%s", instance.Name)
	existingSvc := &v1.Service{}
	err := m.Get(ctx, types.NamespacedName{Name: svcName, Namespace: instance.Namespace}, existingSvc)

	var preferredIP string
	if err == nil {
		// Prova a usare l'ip già assegnato
		if ip, ok := existingSvc.Annotations["metallb.universe.tf/loadBalancerIPs"]; ok {
			preferredIP = ip
		} else if len(existingSvc.Status.LoadBalancer.Ingress) > 0 {
			preferredIP = existingSvc.Status.LoadBalancer.Ingress[0].IP
		}
	}

	// 1. Ottieni il pool di IP disponibili da MetalLB
	ipPool, err := m.getMetalLBIPPool(ctx)
	if err != nil {
		return "", nil, err
	}

	// Se abbiamo un IP preferito, mettilo all'inizio della lista
	if preferredIP != "" {
		// Rimuovi l'IP preferito dalla lista e aggiungilo all'inizio
		newPool := []string{preferredIP}
		for _, ip := range ipPool {
			if ip != preferredIP {
				newPool = append(newPool, ip)
			}
		}
		ipPool = newPool
		log.Info("Using preferred IP from existing service", "ip", preferredIP)
	}

	// Crea una copia locale delle porte utilizzate per la simulazione
	simulatedUsedPorts := make(map[string]map[int]bool)
	for ip, ports := range usedPortsByIP {
		simulatedUsedPorts[ip] = make(map[int]bool)
		for port := range ports {
			simulatedUsedPorts[ip][port] = true
		}
	}

	// 2. Dividi le porte del servizio in porte specificate e automatiche
	var specifiedPorts, autoPorts []clv1alpha2.ServicePortMapping
	for _, svcPort := range instance.Spec.PublicExposure.Services {
		if svcPort.Port != 0 {
			specifiedPorts = append(specifiedPorts, svcPort)
		} else {
			autoPorts = append(autoPorts, svcPort)
		}
	}

	// Scegli il miglior IP considerando prima le porte specificate
	var bestIP string
	var allAssignedPorts []clv1alpha2.ServicePortMapping

	// 3. Esamina ogni IP disponibile
	for _, ip := range ipPool {
		// Inizializza la mappa delle porte se non esiste
		if simulatedUsedPorts[ip] == nil {
			simulatedUsedPorts[ip] = make(map[int]bool)
		}

		// Flag per tracciare se questo IP è compatibile con tutte le porte specificate
		isIPCompatible := true

		// 4. Prima controlla se tutte le porte specificate possono essere assegnate
		var tempAssignedSpecific []clv1alpha2.ServicePortMapping
		for _, port := range specifiedPorts {
			// Controlla se la porta richiesta è già in uso
			if simulatedUsedPorts[ip][int(port.Port)] {
				isIPCompatible = false
				log.Info("Specified port already in use", "ip", ip, "port", port.Port)
				break
			}

			// Simula l'assegnazione della porta
			simulatedUsedPorts[ip][int(port.Port)] = true
			assignedPort := clv1alpha2.ServicePortMapping{
				Name:         port.Name,
				TargetPort:   port.TargetPort,
				Port:         port.Port,
				AssignedPort: port.Port, // Per le porte specificate, AssignedPort = Port
			}
			tempAssignedSpecific = append(tempAssignedSpecific, assignedPort)
		}

		// Se non è compatibile con le porte specificate, prova il prossimo IP
		if !isIPCompatible {
			continue
		}

		// 5. Ora assegna le porte automatiche
		var tempAssignedAuto []clv1alpha2.ServicePortMapping
		allAutoPortsAssignable := true

		for _, port := range autoPorts {
			// Trova una porta libera nell'intervallo 30000-32767
			var assignedPort int32
			for potentialPort := basePort; potentialPort <= 32767; potentialPort++ {
				// Verifica che la porta non sia già utilizzata o richiesta da altre ServiceRequest
				if !simulatedUsedPorts[ip][potentialPort] {
					assignedPort = int32(potentialPort)
					simulatedUsedPorts[ip][potentialPort] = true
					break
				}
			}

			if assignedPort == 0 {
				// Non è possibile trovare una porta libera
				allAutoPortsAssignable = false
				log.Info("Cannot find free port for automatic assignment", "ip", ip)
				break
			}

			tempAssignedAuto = append(tempAssignedAuto, clv1alpha2.ServicePortMapping{
				Name:         port.Name,
				TargetPort:   port.TargetPort,
				Port:         0, // La porta originale è 0 (automatica)
				AssignedPort: assignedPort,
			})
		}

		// 6. Se tutte le porte sono assegnabili, questo è il miglior IP
		if allAutoPortsAssignable {
			bestIP = ip
			allAssignedPorts = append(tempAssignedSpecific, tempAssignedAuto...)
			log.Info("Found compatible IP", "ip", bestIP)
			break
		}
	}

	if bestIP == "" {
		return "", nil, fmt.Errorf("no available IP can support all requested ports")
	}

	// Aggiorna la mappa reale delle porte utilizzate
	for _, port := range allAssignedPorts {
		if usedPortsByIP[bestIP] == nil {
			usedPortsByIP[bestIP] = make(map[int]bool)
		}
		usedPortsByIP[bestIP][int(port.AssignedPort)] = true
		log.Info("Port registered as in use", "ip", bestIP, "port", port.AssignedPort)
	}

	return bestIP, allAssignedPorts, nil
}

// updateUsedPortsByIP aggiorna la mappa delle porte in uso per ogni IP
func (m *Manager) updateUsedPortsByIP(ctx context.Context, namespace string, excludeServiceName string) (map[string]map[int]bool, error) {
	usedPortsByIP := make(map[string]map[int]bool)
	logger := log.FromContext(ctx)

	// Ottieni tutti i servizi LoadBalancer
	svcList := &v1.ServiceList{}
	if err := m.List(ctx, svcList, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	for _, svc := range svcList.Items {
		// SALTA IL SERVIZIO CORRENTE per evitare conflitti con se stesso
		if svc.Name == excludeServiceName {
			logger.Info("Skipping current service from used ports calculation", "service", svc.Name)
			continue
		}

		if svc.Spec.Type != v1.ServiceTypeLoadBalancer {
			continue
		}

		// Ottieni l'IP esterno assegnato
		var externalIP string
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			externalIP = svc.Status.LoadBalancer.Ingress[0].IP
		} else {
			// Se il servizio non ha ancora un IP, prova a trovarlo nelle annotazioni
			if specIP, ok := svc.Annotations["metallb.universe.tf/loadBalancerIPs"]; ok {
				externalIP = specIP
			} else {
				continue // Non è possibile determinare l'IP
			}
		}

		// Inizializza la mappa per questo IP se non esiste
		if _, exists := usedPortsByIP[externalIP]; !exists {
			usedPortsByIP[externalIP] = make(map[int]bool)
		}

		// Registra le porte utilizzate
		for _, port := range svc.Spec.Ports {
			usedPortsByIP[externalIP][int(port.Port)] = true
			logger.Info("Port registered as in use", "ip", externalIP, "port", port.Port)
		}
	}

	return usedPortsByIP, nil
}
