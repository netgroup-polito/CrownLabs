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

	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v13/controller"

	kekrrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/go-logr/logr"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/pmp"
)

var (
	mirrorStorageClass    string
	mirrorProvisionerName string
)

func init() {
	flag.StringVar(&mirrorStorageClass, "mirror-storage-class", "nfs-mirror", "The StorageClass to be used for all PVCs which are going to be mirrors")
	flag.StringVar(&mirrorProvisionerName, "mirror-provisioner-name", "pmp.crownlabs.polito.it", "The provisioner name to be used for the mirror StorageClass")
}

func setupPmp(
	ctx context.Context,
	mgr manager.Manager,
	log logr.Logger,
	_ common.KVLabel,
) error {
	log = log.WithName("setupPmp")

	pmp := pmp.PvcMirrorProvisioner{
		Client: mgr.GetClient(),
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err, "Failed to create config")
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err, "Failed to create client")
		return err
	}

	// Check if StorageClass is available
	var outputSc storagev1.StorageClass
	if err := mgr.GetClient().Get(ctx, types.NamespacedName{Name: mirrorStorageClass}, &outputSc); err != nil {
		if kekrrors.IsNotFound(err) {
			log.Error(err, "Cannot procede without a storageclass")
		} else {
			log.Error(err, "Could not retrieve storage class")
		}
		return err
	}

	controller.NewProvisionController(ctx, clientset, mirrorProvisionerName, &pmp)

	return nil
}
