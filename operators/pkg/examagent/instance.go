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

// Package examagent contains the main logic and helpers
// for the crownlabs exam agent component.
package examagent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
)

// InstanceAdapter represents an Instance within the examagent.
type InstanceAdapter struct {
	ID                string                               `json:"id"`
	Template          string                               `json:"template"`
	Running           *bool                                `json:"running,omitempty"`
	CustomizationUrls clv1alpha2.InstanceCustomizationUrls `json:"customizationUrls"`
	Phase             string                               `json:"phase"`
	URL               string                               `json:"url,omitempty"`
	Labels            map[string]string                    `json:"labels"`
}

// InstanceHandler is the handler for the InstanceAdapter.
type InstanceHandler struct {
	Log             logr.Logger
	Client          client.Client
	AdapterEndpoint string
}

const (
	// XForwardedFor -> "X-Forwarded-For" header.
	XForwardedFor = "X-Forwarded-For"
)

// ServeHTTP is the Instance handler for the examagent.
func (ih *InstanceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log := ih.Log.WithValues("remote-ip", r.Header.Get(XForwardedFor), "method", r.Method, "path", r.URL.Path)

	log.Info("processing request", "query", r.URL.RawQuery)

	switch r.Method {
	case http.MethodGet:
		if ih.GetInstanceIDFromRequest(r) == "" {
			ih.HandleGetAll(w, r, log)
		} else {
			ih.HandleGet(w, r, log)
		}
	case http.MethodPut:
		ih.HandlePut(w, r, log)
	case http.MethodDelete:
		ih.HandleDelete(w, r, log)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method not allowed")
		return
	}
}

// HandleGet handles the GET request for the examagent.
func (ih *InstanceHandler) HandleGet(w http.ResponseWriter, r *http.Request, log logr.Logger) {
	inst := ih.EmptyInstanceFromRequest(r)
	log = log.WithValues("instance", inst.Name)

	if err := ih.Client.Get(r.Context(), forge.NamespacedName(inst), inst); err != nil {
		if errors.IsNotFound(err) {
			log.Error(err, "instance not found")
			WriteError(w, r, log, http.StatusNotFound, "The requested Instance does not exist.")
			return
		}
		log.Error(err, "error retrieving instance")
		WriteError(w, r, log, http.StatusInternalServerError, "Cannot retrieve the requested instance.")
		return
	}

	log = log.WithValues("phase", inst.Status.Phase)

	if !AcceptsHTML(r) {
		if err := WriteJSON(w, AdapterFromInstance(inst)); err != nil {
			log.Error(err, "cannot encode instance")
			WriteError(w, r, log, http.StatusInternalServerError, "Cannot encode the requested instance.")
		}
		log.Info("sent JSON instance")
		return
	}

	switch inst.Status.Phase {
	case clv1alpha2.EnvironmentPhaseReady:
		log.Info("redirecting", "url", inst.Status.URL)
		http.Redirect(w, r, inst.Status.URL, http.StatusFound)
	case clv1alpha2.EnvironmentPhaseOff:
		log.Error(fmt.Errorf("instance off"), "invalid phase")
		WriteError(w, r, log, http.StatusGone, "The requested Instance is not running.")
	case clv1alpha2.EnvironmentPhaseFailed, clv1alpha2.EnvironmentPhaseCreationLoopBackoff:
		log.Error(fmt.Errorf("instance failed"), "invalid phase")
		WriteError(w, r, log, http.StatusInternalServerError, "Something went wrong.")
	default:
		log.Info("sending starting-up page")

		w.Header().Add("refresh", "5")
		w.WriteHeader(http.StatusCreated)
		if err := WriteStartupPage(w); err != nil {
			log.Error(err, "error rendering startup page")
		}
	}
}

// HandlePut handles the PUT request for a InstanceAdapter api call.
func (ih *InstanceHandler) HandlePut(w http.ResponseWriter, r *http.Request, log logr.Logger) {
	if err := Options.CheckAllowedIP(r.Header.Get(XForwardedFor)); err != nil {
		log.Error(err, "unauthorized")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Forbidden")
		return
	}

	// get Instance from the request
	adapter, err := InstanceAdapterFromRequest(r, log)
	if err != nil {
		log.Error(err, "cannot parse request")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Bad request")
		return
	}

	instance := ih.EmptyInstanceFromRequest(r)
	log = log.WithValues("instance", instance.Name)

	op, err := ctrl.CreateOrUpdate(r.Context(), ih.Client, instance, func() error {
		instance.Spec = InstanceSpecFromAdapter(&adapter)
		instance.SetLabels(labels.Merge(instance.GetLabels(), adapter.Labels))
		return nil
	})

	log = log.WithValues("operation", op)

	if err != nil {
		log.Error(err, "failed performing operation")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Instance %s cannot be %s", instance.Name, op)
		return
	}

	switch op {
	case ctrlutil.OperationResultCreated:
		w.WriteHeader(http.StatusCreated)
	case ctrlutil.OperationResultUpdated:
		w.WriteHeader(http.StatusOK)
	case ctrlutil.OperationResultNone:
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err := WriteJSON(w, AdapterFromInstance(instance)); err != nil {
		fmt.Fprint(w, op)
		log.Error(err, "operation complete but cannot encode instance")
		return
	}
	log.Info("success")
}

// HandleGetAll handles the GET request for all the instances.
func (ih *InstanceHandler) HandleGetAll(w http.ResponseWriter, r *http.Request, log logr.Logger) {
	if err := Options.CheckAllowedIP(r.Header.Get(XForwardedFor)); err != nil {
		log.Error(err, "unauthorized")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Forbidden")
		return
	}

	clientOptions := []client.ListOption{client.InNamespace(Options.Namespace)}
	if r.URL.Query().Encode() != "" {
		clientOptions = append(clientOptions, client.MatchingLabels(ValuesToMap(r.URL.Query())))
	}

	var instances clv1alpha2.InstanceList
	if err := ih.Client.List(r.Context(), &instances, clientOptions...); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error retrieving instances")
		log.Error(err, "error retrieving instances")
		return
	}

	adapters := make([]InstanceAdapter, len(instances.Items))
	for i := range instances.Items {
		adapters[i] = *AdapterFromInstance(&instances.Items[i])
	}

	if err := WriteJSON(w, adapters); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		log.Error(err, "cannot encode instances")
		return
	}
}

// HandleDelete handles the DELETE request for a InstanceAdapter api call.
func (ih *InstanceHandler) HandleDelete(w http.ResponseWriter, r *http.Request, log logr.Logger) {
	if err := Options.CheckAllowedIP(r.Header.Get(XForwardedFor)); err != nil {
		log.Error(err, "unauthorized")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Forbidden")
		return
	}

	inst := ih.EmptyInstanceFromRequest(r)
	log = log.WithValues("instance", inst.Name, "operation", "delete")
	if err := ih.Client.Delete(r.Context(), inst); err != nil {
		log.Error(err, "failed performing operation")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error deleting instance")
		return
	}

	log.Info("success")
}

// GetInstanceIDFromRequest returns the instance id from the request.
func (ih *InstanceHandler) GetInstanceIDFromRequest(r *http.Request) string {
	InstanceEP := path.Join(Options.BasePath, ih.AdapterEndpoint) + "/"
	instID := strings.Replace(r.URL.Path, InstanceEP, "", 1)
	return instID
}

// EmptyInstanceFromRequest returns an Instance from a given request with just the ObjectMeta field set.
func (ih *InstanceHandler) EmptyInstanceFromRequest(r *http.Request) *clv1alpha2.Instance {
	return &clv1alpha2.Instance{
		ObjectMeta: metav1.ObjectMeta{Name: ih.GetInstanceIDFromRequest(r), Namespace: Options.Namespace},
	}
}

// InstanceAdapterFromRequest parses a InstanceAdapter from a request.
func InstanceAdapterFromRequest(r *http.Request, log logr.Logger) (InstanceAdapter, error) {
	inst := InstanceAdapter{}

	if Options.PrintRequestBody {
		body, _ := io.ReadAll(r.Body)
		log.Info("logging raw request", "body", string(body))
		r.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	err := json.NewDecoder(r.Body).Decode(&inst)
	return inst, err
}

// InstanceSpecFromAdapter creates an InstanceSpec from a given InstanceAdapter.
func InstanceSpecFromAdapter(instReq *InstanceAdapter) clv1alpha2.InstanceSpec {
	running := ptr.Deref(instReq.Running, true)
	return clv1alpha2.InstanceSpec{
		Template: clv1alpha2.GenericRef{
			Name:      instReq.Template,
			Namespace: Options.Namespace,
		},
		Running: running,
		Tenant: clv1alpha2.GenericRef{
			Name: clv1alpha2.SVCTenantName,
		},
		PrettyName:        fmt.Sprintf("Exam %s", instReq.ID),
		CustomizationUrls: &instReq.CustomizationUrls,
	}
}

// AdapterFromInstance creates an InstanceAdapter from a given Instance.
func AdapterFromInstance(inst *clv1alpha2.Instance) *InstanceAdapter {
	adapter := &InstanceAdapter{
		ID:       inst.Name,
		Template: inst.Spec.Template.Name,
		Running:  ptr.To(inst.Spec.Running),
		URL:      inst.Status.URL,
		Phase:    string(inst.Status.Phase),
		Labels:   inst.GetLabels(),
	}

	if inst.Spec.CustomizationUrls != nil {
		adapter.CustomizationUrls = *inst.Spec.CustomizationUrls
	}

	return adapter
}

// ValuesToMap converts a url.Values to a map[string]string. In case of duplicate keys, the first value is used; in case of no values or empty first value, "true" is set automatically as a value.
func ValuesToMap(values url.Values) map[string]string {
	result := make(map[string]string)
	for key, value := range values {
		if len(value) > 0 && value[0] != "" {
			result[key] = value[0]
		} else {
			result[key] = "true"
		}
	}
	return result
}
