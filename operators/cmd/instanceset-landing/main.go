// Copyright 2020-2021 Politecnico di Torino
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

package main

import (
	"fmt"
	"net/http"

	"k8s.io/klog/v2"

	isetlanding "github.com/netgroup-polito/CrownLabs/operators/pkg/instanceset-landing"
)

func main() {
	isetlanding.Options.Init()
	isetlanding.Options.Parse()

	if err := isetlanding.Options.Validate(); err != nil {
		klog.Fatalf("invalid configuration: %w", err)
	}

	if isetlanding.Options.DynamicStartup {
		if err := isetlanding.PrepareClient(); err != nil {
			klog.Fatal(err)
		}
	}

	http.HandleFunc("/healthz", healthzHandler)
	http.HandleFunc("/", isetlanding.LandingHandler)

	klog.Info("Instanceset landing listening on port ", isetlanding.Options.ListenerAddr)
	klog.Fatal(http.ListenAndServe(isetlanding.Options.ListenerAddr, nil))
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method not allowed")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}
