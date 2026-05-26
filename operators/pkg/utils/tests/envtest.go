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

package tests

import (
	"os"
	"path/filepath"
	"runtime"

	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func getProjectRoot() string {
	_, b, _, ok := runtime.Caller(0)
	if !ok {
		panic("Could not get current file path")
	}

	// 2. Start moving up the directory tree
	dir := filepath.Dir(b)
	for {
		// Check if go.mod exists in the current directory
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		// Move up one level
		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached filesystem root
		}
		dir = parent
	}

	panic("Could not find project root (directory containing go.mod)")
}

// ForgeEnvtestEnv creates and returns an envtest.Environment with the specified configuration (whether to include CRDs from deploy and test folders).
func ForgeEnvtestEnv(includeTestRes bool) *envtest.Environment {
	projectRoot := getProjectRoot()
	k8sBinsDir := filepath.Join(projectRoot, "bin", "k8s")
	files, err := os.ReadDir(k8sBinsDir)
	if err != nil || len(files) != 1 || !files[0].IsDir() {
		panic("Could not find envtest binaries in " + k8sBinsDir)
	}
	k8sBinsDir = filepath.Join(k8sBinsDir, files[0].Name())
	crdPaths := []string{filepath.Join(projectRoot, "deploy", "crds")}
	if includeTestRes {
		crdPaths = append(crdPaths, filepath.Join(projectRoot, "tests", "crds"))
	}

	return &envtest.Environment{
		BinaryAssetsDirectory: k8sBinsDir,
		CRDDirectoryPaths:     crdPaths,
		ErrorIfCRDPathMissing: len(crdPaths) > 0,
	}
}
