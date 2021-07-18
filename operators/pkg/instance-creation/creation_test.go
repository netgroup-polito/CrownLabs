package instance_creation

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	virtv1 "kubevirt.io/client-go/api/v1"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
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
	c1 := utils.CheckLabels(&ns1, labels)
	c2 := utils.CheckLabels(&ns2, labels)
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
	vm := virtv1.VirtualMachineInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "name",
			Namespace: "namespace"},
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualMachineInstance",
			APIVersion: "kubevirt.io/v1alpha3",
		},
	}
	UpdateVirtualMachineInstanceSpec(&vm, tc1)
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
	assert.Equal(t, vm.Spec.Volumes[1].VolumeSource.CloudInitNoCloud.UserDataSecretRef.Name, "name")
}

func TestCreatePersistentVirtualMachine(t *testing.T) {
	tc1 := &v1alpha2.Environment{
		Name:       "Test1Pers",
		GuiEnabled: true,
		Resources: v1alpha2.EnvironmentResources{
			CPU:                   1,
			ReservedCPUPercentage: 25,
			Memory:                resource.MustParse("1024M"),
		},
		EnvironmentType: v1alpha2.ClassVM,
		Persistent:      true,
		Image:           "test/image",
	}
	vm := virtv1.VirtualMachine{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VirtualMachine",
			APIVersion: "kubevirt.io/v1alpha3",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "name",
			Namespace: "namespace"},
	}
	instance := &v1alpha2.Instance{
		Spec: v1alpha2.InstanceSpec{
			Running: true,
		},
	}

	UpdateVirtualMachineSpec(&vm, tc1, instance.Spec.Running)
	assert.Equal(t, len(vm.Spec.Template.Spec.Volumes), 2, "The VMI has a number of volume different from expected")
	assert.Equal(t, len(vm.Spec.Template.Spec.Domain.Devices.Disks), 2, "The VMI has a number of devices different from the expected")
	assert.Equal(t, vm.Spec.Template.Spec.Domain.Devices.Disks[0].Name, "containerdisk")
	assert.Equal(t, vm.Spec.Template.Spec.Domain.Devices.Disks[1].Name, "cloudinitdisk")
	assert.Equal(t, vm.Spec.Template.Spec.Domain.CPU.Cores, uint32(1))
	assert.Equal(t, vm.Spec.Template.Spec.Domain.Resources.Limits.Memory().String(), "1524M")
	assert.Equal(t, vm.Spec.Template.Spec.Domain.Memory.Guest.String(), "1024M")
	assert.Equal(t, vm.Spec.Template.Spec.Domain.Resources.Limits.Cpu().String(), "1500m")
	assert.Equal(t, vm.Spec.Template.Spec.Domain.Resources.Requests.Cpu().String(), "250m")
	assert.Equal(t, vm.Spec.Template.Spec.Volumes[0].Name, "containerdisk")
	assert.Equal(t, vm.Spec.Template.Spec.Volumes[0].DataVolume.Name, "name")
	assert.Equal(t, vm.Spec.Template.Spec.Volumes[1].Name, "cloudinitdisk")
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
	assert.Equal(t, utils.CheckLabels(&ns, labels), true)
	assert.Equal(t, utils.CheckLabels(&ns1, labels), false)
	assert.Equal(t, utils.CheckLabels(&ns2, labels), false)
	assert.Equal(t, utils.CheckLabels(&ns3, labels), false)
}

func TestComputeCPULimits(t *testing.T) {
	var (
		CPU                   uint32  = 1
		HypervisorCoefficient float32 = 0.2
	)
	rawLimit := computeCPULimits(CPU, HypervisorCoefficient)
	limit, err := strconv.ParseFloat(rawLimit, 32) /* if bitSize is 32, ParseFloat return always a Float64 but it's convertible to Float32 without changing its value */

	assert.Equal(t, err, nil, "ParseFloat should not return an error")

	assert.Equal(t, float32(limit), float32(CPU)+HypervisorCoefficient)
}

func TestComputeCPURequests(t *testing.T) {
	var (
		CPU        uint32 = 1
		percentage uint32 = 30
	)
	rawRequests := computeCPURequests(CPU, percentage)
	requests, err := strconv.ParseFloat(rawRequests, 32) /* if bitSize is 32, ParseFloat return always a Float64 but it's convertible to Float32 without changing its value */

	assert.Equal(t, err, nil, "ParseInt should not return an error")

	assert.Equal(t, float32(requests), float32(CPU*percentage)/100)
}
