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
	"bytes"
	_ "embed"

	"gopkg.in/yaml.v3"
)

// userdata is a helper structure to marshal the userdata configuration.
type userdata struct {
	Users             []user     `yaml:"users"`
	Network           network    `yaml:"network"`
	Mounts            [][]string `yaml:"mounts"`
	SSHAuthorizedKeys []string   `yaml:"ssh_authorized_keys,omitempty"`
}

// user is a helper structure to marshal the userdata configuration to configure users.
type user struct {
	Name              string   `yaml:"name"`
	LockPasswd        bool     `yaml:"lock_passwd"`
	Passwd            string   `yaml:"passwd"`
	Sudo              string   `yaml:"sudo"`
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys,omitempty"`
	Shell             string   `yaml:"shell"`
}

// network is a helper structure to marshal the userdata configuration to configure the network subsystem.
type network struct {
	Version int    `yaml:"version"`
	ID0     interf `yaml:"id0"`
}

// interf is a helper structure to marshal the userdata configuration to configure a given interface.
type interf struct {
	DHCP4 bool `yaml:"dhcp4"`
}

//go:embed cloudinit-startup.sh
var scriptdata []byte

// CloudInitUserScriptData configures and forges the cloud-init startup script.
func CloudInitUserScriptData() ([]byte, error) {
	userScriptData := bytes.ReplaceAll(scriptdata, []byte("$NFSPATH"), []byte(MyDriveVolumeMountPath))
	return userScriptData, nil
}

// CloudInitUserData forges the yaml manifest representing the cloud-init userdata configuration.
func CloudInitUserData(publicKeys []string, mountInfos []NFSVolumeMountInfo) ([]byte, error) {
	config := userdata{
		Users: []user{{
			Name:       "crownlabs",
			LockPasswd: false,
			// The hash of the password ("crownlabs").
			// You can generate this hash via: "mkpasswd --method=SHA-512 --rounds=4096".
			Passwd:            "$6$rounds=4096$tBS1sNBpnw6feehB$lS9b7VKH6WMAFOB0SrHCgjD2BKs9CegDe51EiMRWbxQeCVnoGL4u0jNaRsYhvVoBFaRlXZkNsxfFhXvCBaNeQ.",
			Sudo:              "ALL=(ALL) NOPASSWD:ALL",
			SSHAuthorizedKeys: publicKeys,
			Shell:             "/bin/bash",
		}},
		Network: network{
			Version: 2,
			ID0:     interf{DHCP4: true},
		},
		SSHAuthorizedKeys: publicKeys,
	}

	config.Mounts = [][]string{}
	for _, mountInfo := range mountInfos {
		config.Mounts = append(config.Mounts, NFSVolumeMount(mountInfo.ServerAddress, mountInfo.ExportPath, mountInfo.MountPath, mountInfo.ReadOnly))
	}
	config.Mounts = append(config.Mounts, CommentMount("If you change mount options from here, not even Santa will give you 18."))

	output, err := yaml.Marshal(config)
	if err != nil {
		return []byte{}, err
	}

	output = append([]byte("#cloud-config\n"), output...)
	return output, nil
}
