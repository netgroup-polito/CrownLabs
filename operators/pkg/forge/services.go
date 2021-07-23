package forge

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const (
	// SSHPortNumber -> the port the SSH daemon is exposed to.
	SSHPortNumber = 22
	// GUIPortNumber -> the port the GUI service is exposed to.
	GUIPortNumber = 6080
	// MyDrivePortNumber -> the port the "MyDrive" service is exposed to.
	MyDrivePortNumber = 8080

	// SSHPortName -> the name of the port the SSH daemon is exposed to.
	SSHPortName = "ssh"
	// GUIPortName -> the name of the port the NoVNC service is exposed to.
	GUIPortName = "gui"
	// MyDrivePortName -> the name of the port the "MyDrive" service is exposed to.
	MyDrivePortName = "mydrive"
)

// ServiceSpec forges the specification of a Kubernetes Service resource providing
// access to a CrownLabs environment.
func ServiceSpec(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) corev1.ServiceSpec {
	ports := make([]corev1.ServicePort, 0)

	// Do not add the ssh port on container-based instances, since no deamon is present.
	if environment.EnvironmentType == clv1alpha2.ClassVM {
		ports = append(ports, serviceSpecTCPPort(SSHPortName, SSHPortNumber))
	}

	// Add the desktop port only if enabled.
	if environment.GuiEnabled {
		ports = append(ports, serviceSpecTCPPort(GUIPortName, GUIPortNumber))
	}

	// Add the "MyDrive" port only if the environment is a container.
	// TODO: add a field in the environment API to select whether to start the "MyDrive" or not.
	if environment.EnvironmentType == clv1alpha2.ClassContainer {
		ports = append(ports, serviceSpecTCPPort(MyDrivePortName, MyDrivePortNumber))
	}

	spec := corev1.ServiceSpec{
		Type:     corev1.ServiceTypeClusterIP,
		Selector: InstanceSelectorLabels(instance),
		Ports:    ports,
	}

	return spec
}

// serviceSpecTCPPort forges the service port specification, given its name and number.
// The target port is assumed to have the same number of the service one (we cannot use
// names since it is apparently not supported by kubevirt VMI definition.
func serviceSpecTCPPort(name string, number int32) corev1.ServicePort {
	return corev1.ServicePort{
		Name:       name,
		Protocol:   corev1.ProtocolTCP,
		Port:       number,
		TargetPort: intstr.FromInt(int(number)),
	}
}
