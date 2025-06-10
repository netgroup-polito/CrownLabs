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

package forge

import (
	"fmt"

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
	// XVncPortNumber -> the port in the container in which the X server is accessible through VNC.
	XVncPortNumber = 5900
	// MetricsPortNumber -> the port in the container in which the metrics server is accessible.
	MetricsPortNumber = 9090

	ClusterPortNumber = "6443"

	//ClusterPortName -> the name of the port cluster service is exposed to.
	ClusterPortName = "kube-apiserver"
	// SSHPortName -> the name of the port the SSH daemon is exposed to.
	SSHPortName = "ssh"
	// GUIPortName -> the name of the port the NoVNC service is exposed to.
	GUIPortName = "gui"
	// MyDrivePortName -> the name of the port the "MyDrive" service is exposed to.
	MyDrivePortName = "mydrive"
	// XVncPortName -> the name of the port through which the X server is accessible through VNC.
	XVncPortName = "xvnc"
	// MetricsPortName -> the name of the port through which the metrics are exposed.
	MetricsPortName = "metrics"
)

// ServiceSpec forges the specification of a Kubernetes Service resource providing
// access to a CrownLabs environment.
func ServiceSpec(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) corev1.ServiceSpec {
	ports := make([]corev1.ServicePort, 0)

	// Do not add the ssh port on container-based instances, since no deamon is present.
	if environment.EnvironmentType == clv1alpha2.ClassVM || environment.EnvironmentType == clv1alpha2.ClassCloudVM {
		ports = append(ports, serviceSpecTCPPort(SSHPortName, SSHPortNumber))
	}

	// Add the GUI port only if enabled.
	if environment.GuiEnabled {
		ports = append(ports, serviceSpecTCPPort(GUIPortName, GUIPortNumber))
	}

	// Add the "MyDrive" port only if the environment is a container or standalone.
	if (environment.EnvironmentType == clv1alpha2.ClassStandalone || environment.EnvironmentType == clv1alpha2.ClassContainer) && environment.Mode == clv1alpha2.ModeStandard {
		ports = append(ports, serviceSpecTCPPort(MyDrivePortName, MyDrivePortNumber))
	}

	// Add the Metrics port only for container
	if environment.EnvironmentType == clv1alpha2.ClassContainer {
		ports = append(ports, serviceSpecTCPPort(MetricsPortName, MetricsPortNumber))
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

func ConfigMapData(instance *clv1alpha2.Instance, serviceName string, environment *clv1alpha2.Environment) map[string]string {
	return map[string]string{
		environment.Cluster.ClusterNet.NginxTargetPort: fmt.Sprintf("%s/%s:%s", instance.Namespace, serviceName, ClusterPortNumber),
	}
}
