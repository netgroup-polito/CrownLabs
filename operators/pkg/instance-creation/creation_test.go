package instance_creation

import (
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"k8s.io/apimachinery/pkg/api/resource"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
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

func TestCreateUserData(t *testing.T) {
	var (
		nextUsername     = "usertest"
		nextPassword     = "passtest"
		nextCloudBaseUrl = "nextcloud.url"
	)
	publicKeys := []string{"key1", "key2", "key3"}

	rawConfig := createUserdata(nextUsername, nextPassword, nextCloudBaseUrl, publicKeys)

	var config cloudInitConfig

	err := yaml.Unmarshal([]byte(rawConfig["userdata"]), &config)

	assert.Equal(t, err, nil, "Yaml parser should return nil error.")

	// check if header comment is present
	hc := strings.HasPrefix(rawConfig["userdata"], "#cloud-config\n")

	var (
		expectedmount       = []string{nextCloudBaseUrl + "/remote.php/dav/files/" + nextUsername, "/media/MyDrive", "davfs", "_netdev,auto,user,rw,uid=1000,gid=1000", "0", "0"}
		expectedcontent     = "/media/MyDrive " + nextUsername + " " + nextPassword
		expectedpath        = "/etc/davfs2/secrets"
		expectedpermissions = "0600"
	)

	assert.Equal(t, hc, true, "Cloud-init head comment should be present.")
	assert.Equal(t, config.Network.Version, 2, "Network version should be set to 2.")
	assert.Equal(t, config.Network.ID0.Dhcp4, true, "DHCPv4 should be set to true.")
	assert.Equal(t, config.Mounts[0], expectedmount, "Nextcloud mount should be set to "+strings.Join(expectedmount, ", ")+".")
	assert.Equal(t, config.WriteFiles[0].Content, expectedcontent, "Nextcloud secret should be se to "+expectedcontent+" .")
	assert.Equal(t, config.WriteFiles[0].Path, expectedpath, "Nextcloud secret path should be set to "+expectedpath+".")
	assert.Equal(t, config.WriteFiles[0].Permissions, expectedpermissions, "Nextcloud secret permissions should be set to "+expectedpermissions+" .")
	assert.Equal(t, config.SSHAuthorizedKeys[0], publicKeys[0], "Public key should be set to"+publicKeys[0]+" .")
	assert.Equal(t, config.SSHAuthorizedKeys[1], publicKeys[1], "Public key should be set to"+publicKeys[1]+" .")
	assert.Equal(t, config.SSHAuthorizedKeys[2], publicKeys[2], "Public key should be set to"+publicKeys[2]+" .")

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
