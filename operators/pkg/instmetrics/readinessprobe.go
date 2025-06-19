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

package instmetrics

import (
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	criapi "k8s.io/cri-api/pkg/apis/runtime/v1"

	clctx "github.com/netgroup-polito/CrownLabs/operators/pkg/context"
)

// ReadinessProbeHandler is the handler for the readiness probe requests.
type ReadinessProbeHandler struct {
	RuntimeClient *RemoteRuntimeServiceClient
	Log           logr.Logger
	Ready         bool
}

// ServeHTTP is the handler for the ReadinessProbeHandler.
func (h *ReadinessProbeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method not allowed")
		return
	}

	if !h.Ready {
		ctx := clctx.LoggerIntoContext(r.Context(), h.Log)
		if _, err := h.RuntimeClient.ListPodSandbox(ctx, &criapi.PodSandboxFilter{}); err == nil {
			h.Ready = true
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}
