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

package pmp_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2/textlogger"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/pmp"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/tests"
)

const (
	mirrorStorageClassName = "mirror-test"
	mirrorProvisionerName  = "pmp.crownlabs.polito.it"
)

var (
	ctx     context.Context
	pmprov  pmp.PvcMirrorProvisioner
	testEnv = envtest.Environment{}
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PVC Mirror Provisioner Suite")
}

var _ = BeforeSuite(func() {
	tests.LogsToGinkgoWriter()

	ctx = context.Background()

	testEnv.ControlPlane.GetAPIServer().Configure().Append("feature-gates", "CrossNamespaceVolumeDataSource=true")
	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	Expect(storagev1.AddToScheme(scheme.Scheme)).To(Succeed())

	logger := textlogger.NewLogger(textlogger.NewConfig())

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	pmprov = pmp.PvcMirrorProvisioner{
		Ctx:                   ctx,
		Client:                k8sClient,
		Config:                cfg,
		Logger:                logger,
		TargetLabel:           common.NewLabel("crownlabs.polito.it/operator-selector", "local"),
		MirrorStorageClass:    mirrorStorageClassName,
		MirrorProvisionerName: mirrorProvisionerName,
	}
})

var _ = AfterSuite(func() {
	Expect(testEnv.Stop()).To(Succeed())
})
