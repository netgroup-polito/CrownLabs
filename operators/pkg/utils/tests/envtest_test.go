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

package tests_test

import (
	"path/filepath"
	"testing"

	"github.com/onsi/gomega"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/tests"
)

func TestForgeEnvtestEnvUsesConfiguredAssets(t *testing.T) {
	for _, tc := range []struct {
		name           string
		includeTestRes bool
		crdDirectories []string
	}{
		{name: "operator CRDs", crdDirectories: []string{"deploy"}},
		{name: "operator and test CRDs", includeTestRes: true, crdDirectories: []string{"deploy", "tests"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			assets := t.TempDir()
			t.Setenv("KUBEBUILDER_ASSETS", assets)

			env := tests.ForgeEnvtestEnv(tc.includeTestRes)

			g.Expect(env.BinaryAssetsDirectory).To(gomega.Equal(assets))
			g.Expect(env.ErrorIfCRDPathMissing).To(gomega.BeTrue())
			g.Expect(env.CRDDirectoryPaths).To(gomega.HaveLen(len(tc.crdDirectories)))
			for i, dir := range tc.crdDirectories {
				g.Expect(filepath.IsAbs(env.CRDDirectoryPaths[i])).To(gomega.BeTrue())
				g.Expect(env.CRDDirectoryPaths[i]).To(gomega.HaveSuffix(filepath.Join(dir, "crds")))
			}
		})
	}
}
