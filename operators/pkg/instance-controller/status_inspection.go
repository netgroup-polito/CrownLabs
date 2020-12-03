package instance_controller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	virtv1 "kubevirt.io/client-go/api/v1"
	"net"
	"time"
)

func getVmiStatus(r *LabInstanceReconciler, ctx context.Context, log logr.Logger,
	guiEnabled bool, service v1.Service, ingress networkingv1.Ingress,
	labInstance *crownlabsv1alpha2.Instance, vmi *virtv1.VirtualMachineInstance, startTimeVM time.Time) {

	var vmStatus virtv1.VirtualMachineInstancePhase

	var ip string
	url := ingress.GetAnnotations()["crownlabs.polito.it/probe-url"]

	// iterate until the vm is running
	for {
		err := r.Client.Get(ctx, types.NamespacedName{
			Namespace: vmi.Namespace,
			Name:      vmi.Name,
		}, vmi)
		if err == nil {
			if vmStatus != vmi.Status.Phase {
				vmStatus = vmi.Status.Phase
				if len(vmi.Status.Interfaces) > 0 {
					ip = vmi.Status.Interfaces[0].IP
				}

				msg := "VirtualMachineInstance " + vmi.Name + " in namespace " + vmi.Namespace + " status update to " + string(vmStatus)
				if vmStatus == virtv1.Failed {
					setLabInstanceStatus(r, ctx, log, msg, "Warning", "Vmi"+string(vmStatus), labInstance, "", "")
					return
				}

				setLabInstanceStatus(r, ctx, log, msg, "Normal", "Vmi"+string(vmStatus), labInstance, ip, url)
				if vmStatus == virtv1.Running {
					break
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	// when the vm status is Running, it is still not available for some seconds
	// hence, wait until it starts responding
	host := service.Name + "." + service.Namespace
	port := "6080" // VNC
	if !guiEnabled {
		port = "22" // SSH
	}

	err := waitForConnection(log, host, port)
	if err != nil {
		log.Error(err, fmt.Sprintf("Unable to check whether %v:%v is reachable", host, port))
	} else {
		msg := "VirtualMachineInstance " + vmi.Name + " in namespace " + vmi.Namespace + " status update to VmiReady."
		setLabInstanceStatus(r, ctx, log, msg, "Normal", "VmiReady", labInstance, ip, url)
		readyTime := time.Now()
		bootTime := readyTime.Sub(startTimeVM)
		bootTimes.Observe(bootTime.Seconds())
	}
}

func waitForConnection(log logr.Logger, host, port string) error {
	for retries := 0; retries < 120; retries++ {
		timeout := time.Second
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
		if err != nil {
			log.Info(fmt.Sprintf("Unable to check whether %v:%v is reachable: %v", host, port, err))
			time.Sleep(time.Second)
		} else {
			// The connection succeeded, hence the VM is ready
			defer conn.Close()
			return nil
		}
	}

	return fmt.Errorf("Timeout while checking whether %v:%v is reachable", host, port)
}
