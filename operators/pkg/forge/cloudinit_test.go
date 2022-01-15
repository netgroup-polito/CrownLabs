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

package forge_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("CloudInit userdata generation", func() {
	Context("The CloudInitUserData function", func() {
		const (
			baseURL  = "https://nextcloud.example.com"
			username = "username"
			password = "password"

			expected = `
#cloud-config
users:
    - name: crownlabs
      lock_passwd: false
      passwd: $6$rounds=4096$tBS1sNBpnw6feehB$lS9b7VKH6WMAFOB0SrHCgjD2BKs9CegDe51EiMRWbxQeCVnoGL4u0jNaRsYhvVoBFaRlXZkNsxfFhXvCBaNeQ.
      sudo: ALL=(ALL) NOPASSWD:ALL
      ssh_authorized_keys:
        - tenant-key-1
        - tenant-key-2
      shell: /bin/bash
network:
    version: 2
    id0:
        dhcp4: true
mounts:
    - - https://nextcloud.example.com/remote.php/dav/files/username
      - /media/MyDrive
      - davfs
      - _netdev,auto,user,rw,uid=1000,gid=1000
      - "0"
      - "0"
write_files:
    - content: /media/MyDrive username password
      path: /etc/davfs2/secrets
      permissions: "0600"
ssh_authorized_keys:
    - tenant-key-1
    - tenant-key-2
`
		)

		var (
			publicKeys []string

			output []byte
			err    error
		)

		Transformer := func(bytes []byte) string {
			return strings.TrimSpace(strings.ReplaceAll(string(bytes), "\t", "    "))
		}

		BeforeEach(func() { publicKeys = []string{"tenant-key-1", "tenant-key-2"} })
		JustBeforeEach(func() { output, err = forge.CloudInitUserData(baseURL, username, password, publicKeys) })

		It("Should succeed", func() { Expect(err).ToNot(HaveOccurred()) })
		It("Should match the expected output", func() { Expect(output).To(WithTransform(Transformer, Equal(Transformer([]byte(expected))))) })
	})
})
