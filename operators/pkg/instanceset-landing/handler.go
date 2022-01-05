// Copyright 2020-2022 Politecnico di Torino
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

// Package instanceset_landing contains the main logic and helpers
// for the instanceset-landing component
package instanceset_landing

import (
	_ "embed"
	"fmt"
	"net/http"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
)

const identifierQueryParam = "username"

//go:embed redirecting.html
var httpPageStartingUp string

// LandingHandler is the main handler of the instanceset-landing: starting from
// an identifier it (creates, then) redirects to the corresponding instance, if it exists.
func LandingHandler(w http.ResponseWriter, r *http.Request) {
	klog.V(3).Infof("Request received: %+v", r.URL.Query())

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method not allowed")
		return
	}

	// Obtain user identifier from query string and check validity
	identifierParam, ok := r.URL.Query()[identifierQueryParam]
	if !ok || len(identifierParam) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid request")
		klog.Errorf("Invalid request: invalid query: %v", r.URL.Query())
		return
	}

	// extract information from user identifier and check validity
	userID, courseID, err := decodeIdentifier(identifierParam[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Invalid request")
		klog.Errorf("Invalid request: invalid identifier [%v]", identifierParam)
		return
	}

	instanceName := generateInstanceName(userID, courseID)

	// try to get associated instance
	inst := &clv1alpha2.Instance{}

	err = Client.Get(r.Context(), types.NamespacedName{
		Namespace: Options.Namespace,
		Name:      instanceName,
	}, inst)

	// if instance retrieval fails
	if err != nil {
		// in case instance is not found
		if kerrors.IsNotFound(err) {
			// if dynamic startup is not enabled: report an error
			if !Options.DynamicStartup {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, "Not found")
				klog.Errorf("Tried to access non-existing instance %s", instanceName)
				return
			}
			// if dynamic startup is enabled: try to create the instance
			err = enforceOnDemandInstance(r.Context(), instanceName)
		}
	}

	// report an error if the instance exists but is failing
	if err != nil || inst.Status.Phase == clv1alpha2.EnvironmentPhaseFailed || inst.Status.Phase == clv1alpha2.EnvironmentPhaseCreationLoopBackoff {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Something went wrong. Please retry later")
		klog.Errorf("Instance %s (phase:%v) enforcement error: %w", instanceName, inst.Status.Phase, err)
		return
	}

	if inst.Status.Phase == clv1alpha2.EnvironmentPhaseReady {
		klog.Infof("Redirecting %v to %s", identifierParam[0], inst.Status.URL)
		http.Redirect(w, r, inst.Status.URL, http.StatusFound)
		return
	}

	klog.V(2).Infof("Instance %s (phase: %v): sending starting-up page", instanceName, inst.Status.Phase)

	w.Header().Add("refresh", "5")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, httpPageStartingUp)
}
