package controller

import (
	"context"
	"fmt"

	networkingv1alpha1 "github.com/crownlabs/service-operator/api/v1alpha1" // Modifica questo con il tuo modulo effettivo
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	kubevirtv1 "kubevirt.io/api/core/v1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	finalizerName   = "servicerequest.networking.example.com/finalizer"
	metallbPoolName = "my-ip-pool"
	sharedIPValue   = "true" // valore per l'annotazione allow-shared-ip
	basePort        = 30000  // porta iniziale per l'assegnazione automatica
)

type ServiceRequestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// updateUsedPortsByIP aggiorna la mappa dei port in uso per ciascun IP
func (r *ServiceRequestReconciler) updateUsedPortsByIP(ctx context.Context, namespace string) (map[string]map[int]bool, error) {
	usedPortsByIP := make(map[string]map[int]bool)
	logger := log.FromContext(ctx)

	// Ottieni tutti i servizi LoadBalancer
	svcList := &corev1.ServiceList{}
	if err := r.List(ctx, svcList, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	for _, svc := range svcList.Items {
		if svc.Spec.Type != corev1.ServiceTypeLoadBalancer {
			continue
		}

		// Ottieni l'IP esterno assegnato
		var externalIP string
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			externalIP = svc.Status.LoadBalancer.Ingress[0].IP
		} else {
			// Se il servizio non ha ancora un IP assegnato, prova a trovarlo nelle annotazioni
			if specIP, ok := svc.Annotations["metallb.universe.tf/loadBalancerIPs"]; ok {
				externalIP = specIP
			} else {
				continue // Non possiamo determinare l'IP
			}
		}

		// Inizializza la mappa per questo IP se non esiste
		if _, exists := usedPortsByIP[externalIP]; !exists {
			usedPortsByIP[externalIP] = make(map[int]bool)
		}

		// Registra le porte utilizzate
		for _, port := range svc.Spec.Ports {
			usedPortsByIP[externalIP][int(port.Port)] = true
			logger.Info("Porta registrata come in uso", "ip", externalIP, "porta", port.Port)
		}
	}

	return usedPortsByIP, nil
}

// getMetalLBIPPool ottiene il pool di IP configurato in MetalLB
func (r *ServiceRequestReconciler) getMetalLBIPPool(ctx context.Context) ([]string, error) {
	// TODO: Implementa la logica per ottenere il pool di IP da MetalLB dal cluster, interrogare risorsa ConfigMap o MetalLB CRD
	// Per adesso restituito pool statico di che usiamo tra gli esempi
	return []string{
		"172.18.0.240", "172.18.0.241", "172.18.0.242", "172.18.0.243",
		"172.18.0.244", "172.18.0.245", "172.18.0.246", "172.18.0.247",
		"172.18.0.248", "172.18.0.249", "172.18.0.250",
	}, nil
}

// findBestIP trova l'IP migliore per le porte richieste e gestisce l'assegnazione delle porte
func (r *ServiceRequestReconciler) findBestIPAndAssignPorts(ctx context.Context, sr *networkingv1alpha1.ServiceRequest, usedPortsByIP map[string]map[int]bool) (string, []networkingv1alpha1.Service, error) {
	logger := log.FromContext(ctx)

	// 1. Ottieni il pool di IP disponibili da MetalLB
	ipPool, err := r.getMetalLBIPPool(ctx)
	if err != nil {
		return "", nil, err
	}

	// Crea una copia locale delle porte usate per simulare l'assegnazione
	simulatedUsedPorts := make(map[string]map[int]bool)
	for ip, ports := range usedPortsByIP {
		simulatedUsedPorts[ip] = make(map[int]bool)
		for port := range ports {
			simulatedUsedPorts[ip][port] = true
		}
	}

	// 2. Suddividi le service ports in quelle con porte specificate e quelle con porte automatiche
	var specifiedPorts, autoPorts []networkingv1alpha1.Service
	for _, svcPort := range sr.Spec.Services {
		if svcPort.Port != 0 {
			specifiedPorts = append(specifiedPorts, svcPort)
		} else {
			autoPorts = append(autoPorts, svcPort)
		}
	}

	// Scegli l'IP migliore considerando prima le porte specifiche
	var bestIP string
	var allAssignedPorts []networkingv1alpha1.Service

	// 3. Esamina ogni IP disponibile
	for _, ip := range ipPool {
		// Inizializza la mappa delle porte se non esiste
		if simulatedUsedPorts[ip] == nil {
			simulatedUsedPorts[ip] = make(map[int]bool)
		}

		// Flag per tracciare se questo IP è compatibile con tutte le porte specificate
		isIPCompatible := true

		// 4. Prima verifica se è possibile assegnare tutte le porte specifiche
		var tempAssignedSpecific []networkingv1alpha1.Service
		for _, port := range specifiedPorts {
			// Verifica se la porta richiesta è già in uso
			if simulatedUsedPorts[ip][port.Port] {
				isIPCompatible = false
				logger.Info("Porta specifica già in uso", "ip", ip, "porta", port.Port)
				break
			}

			// Simula l'assegnazione della porta
			simulatedUsedPorts[ip][port.Port] = true
			assignedPort := networkingv1alpha1.Service{
				Name:         port.Name,
				TargetPort:   port.TargetPort,
				Port:         port.Port,
				AssignedPort: port.Port, // Per porte specifiche, AssignedPort = Port
			}
			tempAssignedSpecific = append(tempAssignedSpecific, assignedPort)
		}

		// Se non è compatibile con le porte specifiche, prova il prossimo IP
		if !isIPCompatible {
			continue
		}

		// 5. Ora assegna le porte automatiche
		var tempAssignedAuto []networkingv1alpha1.Service
		allAutoPortsAssignable := true

		for _, port := range autoPorts {
			// Cerca una porta libera nell'intervallo 30000-32767
			var assignedPort int
			for potentialPort := 30000; potentialPort <= 32767; potentialPort++ {
				// Verifica che la porta non sia già in uso o richiesta da altre ServiceRequest
				if !simulatedUsedPorts[ip][potentialPort] {
					assignedPort = potentialPort
					simulatedUsedPorts[ip][potentialPort] = true
					break
				}
			}

			if assignedPort == 0 {
				// Non è stato possibile trovare una porta libera
				allAutoPortsAssignable = false
				logger.Info("Non è possibile trovare una porta libera per l'assegnazione automatica", "ip", ip)
				break
			}

			tempAssignedAuto = append(tempAssignedAuto, networkingv1alpha1.Service{
				Name:         port.Name,
				TargetPort:   port.TargetPort,
				Port:         0, // La porta originale è 0 (automatica)
				AssignedPort: assignedPort,
			})
		}

		// 6. Se tutte le porte sono assegnabili, questo è l'IP migliore
		if allAutoPortsAssignable {
			bestIP = ip
			allAssignedPorts = append(tempAssignedSpecific, tempAssignedAuto...)
			logger.Info("Trovato IP compatibile", "ip", bestIP)
			break
		}
	}

	if bestIP == "" {
		return "", nil, fmt.Errorf("nessun IP disponibile può supportare tutte le porte richieste")
	}

	// Aggiorna la mappa reale delle porte utilizzate
	for _, port := range allAssignedPorts {
		if usedPortsByIP[bestIP] == nil {
			usedPortsByIP[bestIP] = make(map[int]bool)
		}
		usedPortsByIP[bestIP][port.AssignedPort] = true
		logger.Info("Porta registrata come in uso", "ip", bestIP, "porta", port.AssignedPort)
	}

	return bestIP, allAssignedPorts, nil
}

// Reconcile gestisce la riconciliazione degli oggetti ServiceRequest
func (r *ServiceRequestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 1. Carica la risorsa -----------------------------------------------------------------
	sr := &networkingv1alpha1.ServiceRequest{}
	if err := r.Get(ctx, req.NamespacedName, sr); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("ServiceRequest non esiste più", "name", req.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// 2. Finalizer -------------------------------------------------------------------------
	if sr.ObjectMeta.DeletionTimestamp.IsZero() {
		// non in cancellazione → assicura finalizer
		if !controllerutil.ContainsFinalizer(sr, finalizerName) {
			controllerutil.AddFinalizer(sr, finalizerName)
			if err := r.Update(ctx, sr); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// in cancellazione → cleanup
		if controllerutil.ContainsFinalizer(sr, finalizerName) {
			logger.Info("Cleanup risorse associate prima della rimozione", "name", sr.Name)

			// Elimina Service LB
			svc := &corev1.Service{}
			if err := r.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("service-%s", sr.Name), Namespace: sr.Spec.Namespace}, svc); err == nil {
				if err := r.Delete(ctx, svc); err != nil && !errors.IsNotFound(err) {
					return ctrl.Result{}, err
				}
			}

			// Rimuovi finalizer
			controllerutil.RemoveFinalizer(sr, finalizerName)
			if err := r.Update(ctx, sr); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil // stop reconciliation quando è in delete
	}

	// 3. Controllo esistenza risorse --------------------------------------------------------
	vm := &kubevirtv1.VirtualMachine{}
	vmKey := types.NamespacedName{Name: sr.Spec.VMName, Namespace: sr.Spec.Namespace}
	vmExists := r.Get(ctx, vmKey, vm) == nil

	svc := &corev1.Service{}
	svcKey := types.NamespacedName{Name: fmt.Sprintf("service-%s", sr.Name), Namespace: sr.Spec.Namespace}
	svcExists := r.Get(ctx, svcKey, svc) == nil

	// 3.a: se VM NON esiste ma Service sì → cleanup & requeue
	if !vmExists && svcExists {
		logger.Info("VM rimossa manualmente: elimino Service", "service", svc.Name)
		if err := r.Delete(ctx, svc); err != nil && !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}

		// reset stato e riesegui reconcile
		sr.Status.Status = ""
		sr.Status.AssignedPorts = nil
		if err := r.Status().Update(ctx, sr); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	// 3.b: se VM e Service esistono già e stato == Created → nulla da fare
	if sr.Status.Status == "Created" && vmExists && svcExists {
		return ctrl.Result{}, nil
	}

	// 4. Aggiorna la mappa dei porti usati -------------------------------------------------
	usedPortsByIP, err := r.updateUsedPortsByIP(ctx, sr.Spec.Namespace)
	if err != nil {
		logger.Error(err, "Errore nell'aggiornare la mappa delle porte usate")
		return ctrl.Result{}, err
	}

	// 5.   miglior IP e assegna le porte --------------------------------------------
	targetIP, assigned, err := r.findBestIPAndAssignPorts(ctx, sr, usedPortsByIP)
	if err != nil {
		logger.Error(err, "Errore nel trovare un IP adatto")
		return ctrl.Result{}, err
	}

	// 6. Crea VM se manca ------------------------------------------------------------------
	runStrategy := kubevirtv1.RunStrategyAlways
	if !vmExists {
		vm := &kubevirtv1.VirtualMachine{
			ObjectMeta: ctrl.ObjectMeta{
				//GenerateName: fmt.Sprintf("vm-%s-", serviceRequest.Name), // Usa generateName per creare un nome univoco
				Name:      sr.Spec.VMName,
				Namespace: sr.Spec.Namespace,
			},
			Spec: kubevirtv1.VirtualMachineSpec{
				RunStrategy: &runStrategy, // Modifica qui
				Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"kubevirt.io/domain": sr.Spec.VMName,
						},
					},
					Spec: kubevirtv1.VirtualMachineInstanceSpec{
						Domain: kubevirtv1.DomainSpec{
							Devices: kubevirtv1.Devices{
								Disks: []kubevirtv1.Disk{
									{
										Name: "containerdisk",
										DiskDevice: kubevirtv1.DiskDevice{
											Disk: &kubevirtv1.DiskTarget{
												Bus: "virtio",
											},
										},
									},
									{
										Name: "cloudinitdisk",
										DiskDevice: kubevirtv1.DiskDevice{
											Disk: &kubevirtv1.DiskTarget{
												Bus: "virtio",
											},
										},
									},
								},
								Interfaces: []kubevirtv1.Interface{
									{
										Name: "default",
										InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
											Masquerade: &kubevirtv1.InterfaceMasquerade{},
										},
									},
								},
							},
							Resources: kubevirtv1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("1024Mi"), // Modifica qui
								},
							},
						},
						Networks: []kubevirtv1.Network{
							{
								Name: "default",
								NetworkSource: kubevirtv1.NetworkSource{
									Pod: &kubevirtv1.PodNetwork{},
								},
							},
						},
						Volumes: []kubevirtv1.Volume{
							{
								Name: "containerdisk",
								VolumeSource: kubevirtv1.VolumeSource{
									ContainerDisk: &kubevirtv1.ContainerDiskSource{
										Image: "kubevirt/fedora-cloud-container-disk-demo",
									},
								},
							},
							{
								Name: "cloudinitdisk",
								VolumeSource: kubevirtv1.VolumeSource{
									CloudInitNoCloud: &kubevirtv1.CloudInitNoCloudSource{
										UserData: `#cloud-config
package_update: true
packages:
  - nginx
  - openssh-server
  - openssh-clients
ssh_pwauth: true
disable_root: false
users:
  - name: fedora
    groups: sudo
    shell: /bin/bash
    sudo: ["ALL=(ALL) NOPASSWD:ALL"]
    lock_passwd: false
chpasswd:
  list: |
    fedora:fedora
  expire: False
runcmd:
  - echo "Ciao mondo" > /usr/share/nginx/html/index.html
  - systemctl enable sshd
  - systemctl start sshd
  - systemctl enable nginx
  - systemctl start nginx`,
									},
								},
							},
						},
					},
				},
			},
		}
		if err := r.Create(ctx, vm); err != nil {
			return ctrl.Result{}, err
		}

		// Ricarica la VM
		if err := r.Get(ctx, vmKey, vm); err != nil {
			return ctrl.Result{}, err
		}
	}

	// 7. ServiceRequest diventa FIGLIA della VM per GC automatico -------------------------
	if err := controllerutil.SetControllerReference(vm, sr, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.Update(ctx, sr); err != nil {
		if errors.IsConflict(err) {
			// Se c'è un conflitto, riprova
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}

	// 8. Crea/aggiorna Service --------------------------------------------------------------
	if !svcExists {
		annotations := map[string]string{
			"metallb.universe.tf/address-pool":    metallbPoolName,
			"metallb.universe.tf/allow-shared-ip": sharedIPValue, // Sempre presente
		}

		// Se abbiamo un IP specifico, aggiungilo come annotazione
		if targetIP != "" {
			annotations["metallb.universe.tf/loadBalancerIPs"] = targetIP
			logger.Info("Richiesto IP specifico", "ip", targetIP)
		}

		svc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:        fmt.Sprintf("service-%s", sr.Name),
				Namespace:   sr.Spec.Namespace,
				Annotations: annotations,
			},
			Spec: corev1.ServiceSpec{
				Type:                  corev1.ServiceTypeLoadBalancer,
				Selector:              map[string]string{"kubevirt.io/domain": sr.Spec.VMName},
				ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeCluster,
			},
		}

		for _, p := range assigned {
			svc.Spec.Ports = append(svc.Spec.Ports, corev1.ServicePort{
				Name:       p.Name,
				Protocol:   corev1.ProtocolTCP,
				Port:       int32(p.AssignedPort),
				TargetPort: intstr.FromInt(p.TargetPort),
			})
		}

		// SR è la owner del Service
		if err := controllerutil.SetControllerReference(sr, svc, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, svc); err != nil {
			return ctrl.Result{}, err
		}
	}

	// 9. Aggiorna Status --------------------------------------------------------------------
	sr.Status.Status = "Created"
	sr.Status.AssignedPorts = assigned
	sr.Status.AssignedIP = targetIP
	if err := r.Status().Update(ctx, sr); err != nil {
		if errors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, err
	}

	logger.Info("ServiceRequest riconciliata con successo",
		"name", sr.Name,
		"namespace", sr.Namespace,
		"assignedPorts", len(sr.Status.AssignedPorts))

	return ctrl.Result{}, nil
}

// SetupWithManager configura il controller con il manager
func (r *ServiceRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1alpha1.ServiceRequest{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

/* disapplica tutto
kubectl delete servicerequest --all -n default
kubectl delete services --all -n ns1
kubectl delete virtualmachines --all -n ns1
*/
