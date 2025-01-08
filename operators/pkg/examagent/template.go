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

package examagent

import (
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

// TemplateHandler is the handler for the TemplateAdapter.
type TemplateHandler struct {
	Log    logr.Logger
	Client client.Client
}

// TemplateAdapter represents a Template within the examagent.
type TemplateAdapter struct {
	Name       string      `json:"name"`
	PrettyName string      `json:"prettyName"`
	Persistent bool        `json:"persistent"`
	CreatedAt  metav1.Time `json:"createdAt"`
}

// ServeHTTP is the handler for the TemplateAdapter.
func (th *TemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := th.Log.WithValues("remote-ip", r.Header.Get(XForwardedFor), "method", r.Method, "path", r.URL.Path)

	log.Info("processing request", "query", r.URL.RawQuery)

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method not allowed")
		return
	}

	templates := clv1alpha2.TemplateList{}
	if err := th.Client.List(r.Context(), &templates, client.InNamespace(Options.Namespace)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error(err, "unable to list templates")
		return
	}

	agentTemplates := make([]TemplateAdapter, len(templates.Items))

	for i := range templates.Items {
		persistent := false
		for ei := range templates.Items[i].Spec.EnvironmentList {
			if templates.Items[i].Spec.EnvironmentList[ei].Persistent {
				persistent = true
				break
			}
		}
		agentTemplates[i] = TemplateAdapter{
			Name:       templates.Items[i].Name,
			PrettyName: templates.Items[i].Spec.PrettyName,
			CreatedAt:  templates.Items[i].CreationTimestamp,
			Persistent: persistent,
		}
	}

	w.WriteHeader(http.StatusOK)
	if err := WriteJSON(w, agentTemplates); err != nil {
		th.Log.Error(err, "unable to encode templates")

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal server error")
	}
}
