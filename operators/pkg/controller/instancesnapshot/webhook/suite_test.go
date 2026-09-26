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

package webhook_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var (
	scheme *runtime.Scheme
)

const (
	testTenant          = "tester"
	testWorkspace       = "test-workspace"
	testInstance        = "test-instance"
	testSnapshot        = "test-snapshot"
	testEnvironment     = "env1"
	testPublicNamespace = "public-snapshots"
	testBypassGroup     = "system:masters"
	testServiceAccount  = "system:serviceaccount:crownlabs:instance-operator"

	testTenantNamespace      = "tenant-" + testTenant
	testOtherTenantNamespace = "tenant-other-tester"
	testWorkspaceNamespace   = "workspace-" + testWorkspace
)

func TestInstanceSnapshotValidator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "InstanceSnapshotValidator Suite")
}

var _ = BeforeSuite(func() {
	scheme = runtime.NewScheme()
	Expect(clv1alpha2.AddToScheme(scheme)).To(Succeed())
})
