package instance_creation

import (
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type writeFile struct {
	Content     string `yaml:"content"`
	Path        string `yaml:"path"`
	Permissions string `yaml:"permissions"`
}

type cloudInitConfig struct {
	Network struct {
		Version int `yaml:"version"`
		ID0     struct {
			Dhcp4 bool `yaml:"dhcp4"`
		} `yaml:"id0"`
	} `yaml:"network"`
	Mounts            [][]string  `yaml:"mounts"`
	WriteFiles        []writeFile `yaml:"write_files"`
	SSHAuthorizedKeys []string    `yaml:"ssh_authorized_keys,omitempty"`
}

func createUserdata(nextUsername, nextPassword, nextCloudBaseURL string, publicKeys []string) map[string]string {
	var Userdata cloudInitConfig

	Userdata.Network.Version = 2
	Userdata.Network.ID0.Dhcp4 = true
	Userdata.Mounts = [][]string{{
		nextCloudBaseURL + "/remote.php/dav/files/" + nextUsername,
		"/media/MyDrive",
		"davfs",
		"_netdev,auto,user,rw,uid=1000,gid=1000",
		"0",
		"0"},
	// New mounts should be added here as []string
	}
	Userdata.WriteFiles = []writeFile{{
		Content:     "/media/MyDrive " + nextUsername + " " + nextPassword,
		Path:        "/etc/davfs2/secrets",
		Permissions: "0600"},
	// New write_files should be added here as []writeFile
	}
	Userdata.SSHAuthorizedKeys = publicKeys

	out, _ := yaml.Marshal(Userdata)

	headerComment := "#cloud-config\n"

	return map[string]string{"userdata": headerComment + string(out)}
}

// CreateCloudInitSecret creates and returns a Kubernetes Secret object which
// contains the cloud-init configuration required to correctly start the VMs.
func CreateCloudInitSecret(name, namespace, nextUsername, nextPassword, nextCloudBaseURL string, publicKeys []string) v1.Secret {
	secret := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "v1",
			APIVersion: "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: nil,
		StringData: createUserdata(
			nextUsername,
			nextPassword,
			nextCloudBaseURL,
			publicKeys),
		Type: v1.SecretTypeOpaque,
	}

	return secret
}
