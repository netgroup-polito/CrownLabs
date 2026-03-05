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
	"flag"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/sharedvolume"
)

var (
	sharedVolumeStorageClass     string
	maxConcurrentShVolReconciles int
)

const (
	sharedVolumeCtrl = "SharedVolume"
)

func init() {
	flag.StringVar(&sharedVolumeStorageClass, "shared-volume-storage-class", "rook-nfs", "The StorageClass to be used for all SharedVolumes' PVC (if unique can be used to enforce ResourceQuota on Workspaces, about number and size of ShVols)")
	flag.IntVar(&maxConcurrentShVolReconciles, "max-concurrent-reconciles-shvol", 1, "The maximum number of concurrent Reconciles which can be run for the Instance Shared Volume controller")
}

func setupSharedVolume(
	mgr manager.Manager,
	targetLabel common.KVLabel,
) error {
	shvol := &sharedvolume.Reconciler{
		Client:          mgr.GetClient(),
		TargetLabel:     targetLabel,
		EventsRecorder:  mgr.GetEventRecorderFor(sharedVolumeCtrl),
		PVCStorageClass: sharedVolumeStorageClass,
	}

	if err := shvol.SetupWithManager(mgr, maxConcurrentShVolReconciles); err != nil {
		return err
	}

	return nil
}
