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
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// VirtualMachineLabels forges the labels for the VirtualMachineInstance template, marking
// VM-family environments without a pre-installed (TigerVNC/noVNC) graphical desktop so that
// the native-VNC KubeVirt Plugin injects QEMU's native VNC-over-websocket listener into the
// libvirt domain XML.
func VirtualMachineLabels(environment *clv1alpha2.Environment, labels map[string]string) map[string]string {
	if labels == nil {
		labels = map[string]string{}
	}
	if !environment.GuiEnabled {
		labels[LabelNativeVNCKey] = "true"
	}
	return labels
}
