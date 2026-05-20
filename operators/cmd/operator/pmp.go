// Copyright 2020-2026 Politecnico di Torino
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

package main

import (
	"context"
	"flag"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/pmp"
)

var (
	mirrorStorageClass    string
	mirrorProvisionerName string
)

func init() {
	flag.StringVar(&mirrorStorageClass, "mirror-storage-class", "pvc-mirror", "The StorageClass to be used for all PVCs which are going to be mirrors")
	flag.StringVar(&mirrorProvisionerName, "mirror-provisioner-name", "pmp.crownlabs.polito.it", "The provisioner name to be used for the mirror StorageClass")
}

func setupPmp(
	ctx context.Context,
	mgr manager.Manager,
	log logr.Logger,
	targetLabel common.KVLabel,
) error {
	log = log.WithName("pmp")

	pmprov := pmp.PvcMirrorProvisioner{
		Ctx:                   ctx,
		Client:                mgr.GetClient(),
		Config:                mgr.GetConfig(),
		Logger:                log,
		TargetLabel:           targetLabel,
		MirrorStorageClass:    mirrorStorageClass,
		MirrorProvisionerName: mirrorProvisionerName,
	}
	return mgr.Add(&pmprov)
}
