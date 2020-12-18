package instance_creation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateCloudInitSecret(t *testing.T) {
	var (
		name             = "name"
		namespace        = "namespace"
		nextUsername     = "usertest"
		nextPassword     = "passtest"
		nextCloudBaseUrl = "nextcloud.url"
	)
	publicKeys := []string{"key1", "key2", "key3"}
	ownerRef := []metav1.OwnerReference{{
		APIVersion: "crownlabs.polito.it/v1alpha2",
		Kind:       "Instance",
		Name:       "Test1",
	}}
	secret := CreateCloudInitSecret(name, namespace, nextUsername, nextPassword, nextCloudBaseUrl, publicKeys, ownerRef)

	var (
		expectedmount       = []string{nextCloudBaseUrl + "/remote.php/dav/files/" + nextUsername, "/media/MyDrive", "davfs", "_netdev,auto,user,rw,uid=1000,gid=1000", "0", "0"}
		expectedcontent     = "/media/MyDrive " + nextUsername + " " + nextPassword
		expectedpath        = "/etc/davfs2/secrets"
		expectedpermissions = "0600"
	)

	var config cloudInitConfig

	//convert and store cloud-init config in config
	err := yaml.Unmarshal([]byte(secret.StringData["userdata"]), &config)
	assert.Equal(t, err, nil, "Yaml parser should return nil error.")

	assert.Equal(t, secret.ObjectMeta.Name, "name-secret", "Name of secret.ObjectMeta should be "+name+"-secret")
	assert.Equal(t, secret.ObjectMeta.Namespace, namespace, "Namespace of secret.ObjectMeta should be "+namespace)

	//check config
	assert.Equal(t, config.Network.Version, 2, "Network version should be set to 2.")
	assert.Equal(t, config.Network.ID0.Dhcp4, true, "DHCPv4 should be set to true.")
	assert.Equal(t, config.Mounts[0], expectedmount)
	assert.Equal(t, config.WriteFiles[0].Content, expectedcontent)
	assert.Equal(t, config.WriteFiles[0].Path, expectedpath)
	assert.Equal(t, config.WriteFiles[0].Permissions, expectedpermissions)
	assert.Equal(t, config.SSHAuthorizedKeys[0], publicKeys[0], "Public key should be set to"+publicKeys[0]+" .")
	assert.Equal(t, config.SSHAuthorizedKeys[1], publicKeys[1], "Public key should be set to"+publicKeys[1]+" .")
	assert.Equal(t, config.SSHAuthorizedKeys[2], publicKeys[2], "Public key should be set to"+publicKeys[2]+" .")

	assert.Equal(t, secret.ObjectMeta.OwnerReferences[0].APIVersion, "crownlabs.polito.it/v1alpha2")
	assert.Equal(t, secret.ObjectMeta.OwnerReferences[0].Kind, "Instance")
	assert.Equal(t, secret.ObjectMeta.OwnerReferences[0].Name, "Test1")
}
