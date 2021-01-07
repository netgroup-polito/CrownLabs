package instance_creation

import (
	"strconv"
	"testing"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ns1 = v1.Namespace{
	TypeMeta: metav1.TypeMeta{},
	ObjectMeta: metav1.ObjectMeta{
		Name: "test",
		Labels: map[string]string{
			"test": "true",
		},
	},
	Spec:   v1.NamespaceSpec{},
	Status: v1.NamespaceStatus{},
}

var ns2 = v1.Namespace{
	TypeMeta: metav1.TypeMeta{},
	ObjectMeta: metav1.ObjectMeta{
		Name: "production",
		Labels: map[string]string{
			"production": "true",
		},
	},
	Spec:   v1.NamespaceSpec{},
	Status: v1.NamespaceStatus{},
}

var labels = map[string]string{
	"test": "true",
}

func TestWhitelist(t *testing.T) {
	c1 := CheckLabels(ns1, labels)
	c2 := CheckLabels(ns2, labels)
	assert.Equal(t, c1, true, "The two label set should be identical and return true.")
	assert.Equal(t, c2, false, "The two labels set should be different and return false.")
}

func TestCreateVirtualMachineInstance(t *testing.T) {
	tc1 := &v1alpha2.Environment{
		Name:       "Test1",
		GuiEnabled: true,
		Resources: v1alpha2.EnvironmentResources{
			CPU:                   1,
			ReservedCPUPercentage: 25,
			Memory:                resource.MustParse("1024M"),
		},
		EnvironmentType: v1alpha2.ClassVM,
		Persistent:      false,
		Image:           "test/image",
	}
	ownerRef := []metav1.OwnerReference{{
		APIVersion: "crownlabs.polito.it/v1alpha2",
		Kind:       "Instance",
		Name:       "Test1",
	},
	}
	vm, err := CreateVirtualMachineInstance("name", "namespace", tc1, "instance-name", "secret-name", ownerRef)
	assert.Equal(t, err, nil, "Errors while generating the VMI")
	assert.Equal(t, vm.Name, "name-vmi", "The VMI has not the expected name")
	assert.Equal(t, vm.Namespace, "namespace", "The VMI is not created in the expected namespace")
	assert.Equal(t, len(vm.Spec.Volumes), 2, "The VMI has a number of volume different from expected")
	assert.Equal(t, len(vm.Spec.Domain.Devices.Disks), 2, "The VMI has a number of devices different from the expected")
	assert.Equal(t, vm.Spec.Domain.Devices.Disks[0].Name, "containerdisk")
	assert.Equal(t, vm.Spec.Domain.Devices.Disks[1].Name, "cloudinitdisk")
	assert.Equal(t, vm.Spec.Domain.CPU.Cores, uint32(1))
	assert.Equal(t, vm.Spec.Domain.Resources.Limits.Memory().String(), "1524M")
	assert.Equal(t, vm.Spec.Domain.Memory.Guest.String(), "1024M")
	assert.Equal(t, vm.Spec.Domain.Resources.Limits.Cpu().String(), "1500m")
	assert.Equal(t, vm.Spec.Domain.Resources.Requests.Cpu().String(), "250m")
	assert.Equal(t, vm.Spec.Volumes[0].Name, "containerdisk")
	assert.Equal(t, vm.Spec.Volumes[0].ContainerDisk.ImagePullPolicy, v1.PullIfNotPresent)
	assert.Equal(t, vm.Spec.Volumes[0].ContainerDisk.ImagePullSecret, registryCred)
	assert.Equal(t, vm.Spec.Volumes[1].Name, "cloudinitdisk")
	assert.Equal(t, vm.Spec.Volumes[1].VolumeSource.CloudInitNoCloud.UserDataSecretRef.Name, "secret-name")
	assert.Equal(t, vm.Kind, "VirtualMachineInstance")
	assert.Equal(t, vm.APIVersion, "kubevirt.io/v1alpha3")
}

func TestCheckLabels(t *testing.T) {
	labels := map[string]string{
		"crownlabs.polito.it/operator-selector": "production",
	}
	ns := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"crownlabs.polito.it/operator-selector": "production",
			},
		},
	}
	ns1 := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"crownlabs.polito.it/operator-selector": "preprod",
			},
		},
	}
	ns2 := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"crownlabs.polito.it/other": "production",
			},
		},
	}
	ns3 := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{},
		},
	}
	assert.Equal(t, CheckLabels(ns, labels), true)
	assert.Equal(t, CheckLabels(ns1, labels), false)
	assert.Equal(t, CheckLabels(ns2, labels), false)
	assert.Equal(t, CheckLabels(ns3, labels), false)
}

func TestComputeCPULimits(t *testing.T) {
	var (
		CPU                   uint32  = 1
		HypervisorCoefficient float32 = 0.2
	)
	rawLimit := ComputeCPULimits(CPU, HypervisorCoefficient)
	limit, err := strconv.ParseFloat(rawLimit, 32) /* if bitSize is 32, ParseFloat return always a Float64 but it's convertible to Float32 without changing its value */

	assert.Equal(t, err, nil, "ParseFloat should not return an error")

	assert.Equal(t, float32(limit), float32(CPU)+HypervisorCoefficient)
}

func TestComputeCPURequests(t *testing.T) {
	var (
		CPU        uint32 = 1
		percentage uint32 = 30
	)
	rawRequests := ComputeCPURequests(CPU, percentage)
	requests, err := strconv.ParseFloat(rawRequests, 32) /* if bitSize is 32, ParseFloat return always a Float64 but it's convertible to Float32 without changing its value */

	assert.Equal(t, err, nil, "ParseInt should not return an error")

	assert.Equal(t, float32(requests), float32(CPU*percentage)/100)
}
