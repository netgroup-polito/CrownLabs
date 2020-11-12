package instanceCreation

import (
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

	rawConfig := createUserdata(nextUsername, nextPassword, nextCloudBaseUrl)

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
	assert.Equal(t, config.Network.Dhcp4, true, "DHCPv4 should be set to true.")
	assert.Equal(t, config.Mounts[0], expectedmount, "Nextcloud mount should be set to "+strings.Join(expectedmount, ", ")+".")
	assert.Equal(t, config.WriteFiles[0].Content, expectedcontent, "Nextcloud secret should be se to "+expectedcontent+" .")
	assert.Equal(t, config.WriteFiles[0].Path, expectedpath, "Nextcloud secret path should be set to "+expectedpath+".")
	assert.Equal(t, config.WriteFiles[0].Permissions, expectedpermissions, "Nextcloud secret permissions should be set to "+expectedpermissions+" .")
}
