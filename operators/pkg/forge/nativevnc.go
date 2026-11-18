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

package forge

import (
	"maps"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// EnvironmentGuiMode returns how the graphical desktop of the environment is reached, or ""
// when it has none. When guiMode is unset, the deprecated guiEnabled flag decides: true means
// Web, false means no graphical desktop.
func EnvironmentGuiMode(environment *clv1alpha2.Environment) clv1alpha2.GuiMode {
	switch {
	case environment.GuiMode != "":
		return environment.GuiMode
	case environment.GuiEnabled: //nolint:staticcheck // SA1019: guiEnabled is honored only here, for templates that predate guiMode.
		return clv1alpha2.GuiModeWeb
	default:
		return ""
	}
}

// NativeVNCEnabled returns whether the environment is a VM whose graphical desktop is served
// by QEMU's VNC server.
func NativeVNCEnabled(environment *clv1alpha2.Environment) bool {
	switch environment.EnvironmentType {
	case clv1alpha2.ClassVM, clv1alpha2.ClassCloudVM, clv1alpha2.ClassLocalVM:
		return EnvironmentGuiMode(environment) == clv1alpha2.GuiModeNative
	default:
		return false
	}
}

// VirtualMachineLabels forges the labels for the VirtualMachineInstance template, marking
// VM environments whose desktop is served by QEMU's native VNC (guiMode: Native), so that
// the native-VNC KubeVirt Plugin injects QEMU's native VNC-over-websocket listener into the
// libvirt domain XML. Such VMIs also carry the namespace selector labels of the operator that
// forged them: Plugins are cluster-wide, so each deployment's Plugin matches only its own VMIs.
func VirtualMachineLabels(environment *clv1alpha2.Environment, labels, operatorSelectorLabels map[string]string) map[string]string {
	if labels == nil {
		labels = map[string]string{}
	}
	if NativeVNCEnabled(environment) {
		labels[LabelNativeVNCKey] = "true"
		maps.Copy(labels, operatorSelectorLabels)
	}
	return labels
}
