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

package webvnc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	wstransport "k8s.io/client-go/transport/websocket"
)

// plainStreamProtocol is kubevirt.io/client-go/subresources.PlainStreamProtocolName,
// kept as a local constant to avoid depending on kubevirt.io/client-go for a
// single string. It is the subprotocol KubeVirt expects from browser clients
// connecting to the vnc subresource.
const plainStreamProtocol = "plain.kubevirt.io"

// openVNCStream opens the native vnc subresource exposed by KubeVirt for the
// given VirtualMachineInstance, authenticating with the webvnc service's own
// credentials (config). It relies on k8s.io/client-go/transport/websocket,
// the same mechanism used by "kubectl exec" over WebSocket, so that any
// authentication method supported by the base rest.Config (bearer token,
// client certificate, exec plugin, ...) works without extra handling here.
func openVNCStream(ctx context.Context, config *rest.Config, vmi types.NamespacedName) (*websocket.Conn, error) {
	rt, wsHolder, err := wstransport.RoundTripperFor(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build websocket round tripper: %w", err)
	}

	url := fmt.Sprintf("%s/apis/subresources.kubevirt.io/v1/namespaces/%s/virtualmachineinstances/%s/vnc?preserveSession=false",
		config.Host, vmi.Namespace, vmi.Name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	conn, err := wstransport.Negotiate(rt, wsHolder, req, plainStreamProtocol)
	if err != nil {
		return nil, fmt.Errorf("failed to negotiate websocket with the vnc subresource: %w", err)
	}
	return conn, nil
}
