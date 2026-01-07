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

package webhook_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

var (
	scheme *runtime.Scheme
)

const (
	testTenant           = "pippo"
	testWorkspace        = "test"
	testTemplate         = "tmpl1"
	testExistingInstance = "inst1"
	testNewInstance      = "inst2"
	testEnvironment      = "env1"
	testMissingTemplate  = "missing"

	testTenantNamespace    = "tenant-" + testTenant
	testWorkspaceNamespace = "workspace-" + testWorkspace
)

var _ = BeforeSuite(func() {
	scheme = runtime.NewScheme()
	Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
	Expect(v1alpha2.AddToScheme(scheme)).To(Succeed())
})
