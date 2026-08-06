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
	"bytes"
	_ "embed"
	"fmt"
	"strconv"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const (
	// NativeVNCHookConfigMapKey -> the key of the ConfigMap entry containing the hook-sidecar script.
	NativeVNCHookConfigMapKey = "onDefineDomain.sh"
	nativeVNCHookPath         = "/usr/bin/onDefineDomain"
	hookSidecarsAnnotation    = "hooks.kubevirt.io/hookSidecars"
)

//go:embed nativevnc-hook.sh
var nativeVNCHookScriptData []byte

// NativeVNCHookScript forges the onDefineDomain hook-sidecar script that injects QEMU's native
// VNC-over-websocket listener into the libvirt domain XML, without exposing it to the guest.
func NativeVNCHookScript() []byte {
	return bytes.ReplaceAll(nativeVNCHookScriptData, []byte("__NATIVEVNCPORT__"), []byte(strconv.Itoa(NativeVNCPortNumber)))
}

// NativeVNCHookConfigMapName returns the name of the ConfigMap holding the native VNC hook-sidecar script.
func NativeVNCHookConfigMapName(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) string {
	return ObjectMetaWithSuffix(instance, environment.Name+"-vnc-hook").Name
}

// VirtualMachineAnnotations forges the annotations for the VirtualMachineInstance template,
// attaching the native VNC hook-sidecar when enabled.
func VirtualMachineAnnotations(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment, annotations map[string]string) map[string]string {
	if !environment.GuiEnabled || !environment.NativeVNC {
		return annotations
	}
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[hookSidecarsAnnotation] = fmt.Sprintf(
		`[{"args": ["--version", "v1alpha3"], "configMap": {"name": %q, "key": %q, "hookPath": %q}}]`,
		NativeVNCHookConfigMapName(instance, environment), NativeVNCHookConfigMapKey, nativeVNCHookPath)
	return annotations
}
