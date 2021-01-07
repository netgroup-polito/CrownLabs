package instance_controller

import (
	"context"
	"fmt"
	"net"
	"time"

	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	virtv1 "kubevirt.io/client-go/api/v1"

	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

func (r *LabInstanceReconciler) getVmiStatus(ctx context.Context,
	guiEnabled bool, service *v1.Service, ingress *networkingv1.Ingress,
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
					r.setLabInstanceStatus(ctx, msg, "Warning", "Vmi"+string(vmStatus), labInstance, "", "")
					return
				}

				r.setLabInstanceStatus(ctx, msg, "Normal", "Vmi"+string(vmStatus), labInstance, ip, url)
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

	err := waitForConnection(host, port)
	if err != nil {
		klog.Error(fmt.Sprintf("Unable to check whether %v:%v is reachable", host, port))
		klog.Error(err)
	} else {
		msg := "VirtualMachineInstance " + vmi.Name + " in namespace " + vmi.Namespace + " status update to VmiReady."
		r.setLabInstanceStatus(ctx, msg, "Normal", "VmiReady", labInstance, ip, url)
		readyTime := time.Now()
		bootTime := readyTime.Sub(startTimeVM)
		bootTimes.Observe(bootTime.Seconds())
	}
}

func waitForConnection(host, port string) error {
	for retries := 0; retries < 120; retries++ {
		timeout := time.Second
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
		if err != nil {
			klog.Info(fmt.Sprintf("Unable to check whether %v:%v is reachable: %v", host, port, err))
			time.Sleep(time.Second)
		} else {
			// The connection succeeded, hence the VM is ready
			defer conn.Close()
			return nil
		}
	}

	return fmt.Errorf("timeout while checking whether %v:%v is reachable", host, port)
}
