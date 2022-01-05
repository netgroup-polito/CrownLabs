// Copyright 2020-2022 Politecnico di Torino
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

	"gopkg.in/yaml.v3"
)

// userdata is a helper structure to marshal the userdata configuration.
type userdata struct {
	Network           network     `yaml:"network"`
	Mounts            [][]string  `yaml:"mounts"`
	WriteFiles        []writefile `yaml:"write_files"`
	SSHAuthorizedKeys []string    `yaml:"ssh_authorized_keys,omitempty"`
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

// writefile is a helper structure to marshal the userdata configuration to create new files.
type writefile struct {
	Content     string `yaml:"content"`
	Path        string `yaml:"path"`
	Permissions string `yaml:"permissions"`
}

// CloudInitUserData forges the yaml manifest representing the cloud-init userdata configuration.
func CloudInitUserData(nextcloudBaseURL, webdavUsername, webdavPassword string, publicKeys []string) ([]byte, error) {
	config := userdata{
		Network: network{
			Version: 2,
			ID0:     interf{DHCP4: true},
		},
		Mounts: [][]string{{
			fmt.Sprintf("%s/remote.php/dav/files/%s", nextcloudBaseURL, webdavUsername),
			"/media/MyDrive",
			"davfs",
			"_netdev,auto,user,rw,uid=1000,gid=1000",
			"0",
			"0",
		}},
		WriteFiles: []writefile{{
			Content:     fmt.Sprintf("/media/MyDrive %s %s", webdavUsername, webdavPassword),
			Path:        "/etc/davfs2/secrets",
			Permissions: "0600",
		}},
		SSHAuthorizedKeys: publicKeys,
	}

	output, err := yaml.Marshal(config)
	if err != nil {
		return []byte{}, err
	}

	output = append([]byte("#cloud-config\n"), output...)
	return output, nil
}
