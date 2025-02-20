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

package forge_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

var _ = Describe("CloudInit files generation", func() {
	Context("The CloudInitUserData function", func() {
		const (
			serviceName       = "rook-ceph-nfs-my-nfs-a.rook-ceph.svc.cluster.local"
			servicePath       = "/nfs/path"
			nfsShVolExpPath   = "/nfs/shvol"
			nfsShVolMountPath = "/mnt/path"
			nfsShVolReadOnly  = true

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
    - - rook-ceph-nfs-my-nfs-a.rook-ceph.svc.cluster.local:/nfs/path
      - /media/mydrive
      - nfs
      - rw,tcp,hard,intr,rsize=8192,wsize=8192,timeo=14,_netdev,user
      - "0"
      - "0"
	- - rook-ceph-nfs-my-nfs-a.rook-ceph.svc.cluster.local:/nfs/shvol
	  - /mnt/path
      - nfs
      - ro,tcp,hard,intr,rsize=8192,wsize=8192,timeo=14,_netdev,user
      - "0"
      - "0"
	- - '# If you change mount options from here, not even Santa will give you 18.'
	  - ""
	  - ""
	  - ""
	  - ""
	  - ""
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
		JustBeforeEach(func() {
			output, err = forge.CloudInitUserData(publicKeys, []forge.NFSVolumeMountInfo{
				forge.MyDriveNFSVolumeMountInfo(serviceName, servicePath),
				{
					ServerAddress: serviceName,
					ExportPath:    nfsShVolExpPath,
					MountPath:     nfsShVolMountPath,
					ReadOnly:      nfsShVolReadOnly,
				},
			})
		})

		It("Should succeed", func() { Expect(err).ToNot(HaveOccurred()) })
		It("Should match the expected output", func() { Expect(output).To(WithTransform(Transformer, Equal(Transformer([]byte(expected))))) })
	})

	Context("The CloudInitUserScriptData function", func() {
		const expected = `#!/bin/bash
mkdir -p "/media/mydrive"
chown 1000:1000 "/media/mydrive"
`

		var (
			scriptdata []byte
			err        error
		)
		JustBeforeEach(func() { scriptdata, err = forge.CloudInitUserScriptData() })

		It("Should succeed", func() { Expect(err).ToNot(HaveOccurred()) })
		It("Should match the expected output", func() { Expect(scriptdata).To(Equal([]byte(expected))) })
	})
})
