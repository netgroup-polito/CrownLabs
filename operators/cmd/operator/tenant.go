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

// Package main contains the entrypoint for the Crownlabs unified operator.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/tenant"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/utils"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	tenantWebhookPort int
	tenantNSKeepAlive time.Duration
)

const (
	tenantWebhookPath = "/tenant-webhook"
)

func init() {
	flag.IntVar(&tenantWebhookPort, "tenant-webhook-port", 8082, "Port for the tenant webhook server for Keycloak events")
	flag.DurationVar(&tenantNSKeepAlive, "tenant-ns-keep-alive", 24*time.Hour,
		"Time elapsed after last login of tenant during which the tenant namespace should be kept alive")

	// mydrivePVCsSize := args.NewQuantity("1Gi")
	// var mydrivePVCsStorageClassName string
	// var myDrivePVCsNamespace string
	// flag.Var(&mydrivePVCsSize, "mydrive-pvcs-size", "The dimension of the user's personal space")
	// flag.StringVar(&mydrivePVCsStorageClassName, "mydrive-pvcs-storage-class-name", "rook-nfs", "The name for the user's storage class")
	// flag.StringVar(&myDrivePVCsNamespace, "mydrive-pvcs-namespace", "mydrive-pvcs", "The namespace where the PVCs are created")
}

func setup_tenant(
	mgr manager.Manager,
	targetLabel utils.Label,
) error {
	// TODO manage webhook

	tn := &tenant.TenantReconciler{
		Client:                  mgr.GetClient(),
		Scheme:                  mgr.GetScheme(),
		TargetLabel:             targetLabel,
		TenantNSKeepAlive:       tenantNSKeepAlive,
		TriggerReconcileChannel: make(chan event.GenericEvent, 10),
	}

	if err := tn.SetupWithManager(mgr); err != nil {
		return err
	}

	go startHTTPServer(tn)

	return nil
}

func startHTTPServer(tn *tenant.TenantReconciler) {
	mux := http.NewServeMux()

	// registering the handler for the tenant webhook path
	mux.HandleFunc(tenantWebhookPath, func(w http.ResponseWriter, r *http.Request) {
		tn.KeycloakEventHandler(w, r)
	})

	addr := fmt.Sprintf(":%d", tenantWebhookPort)
	log.Printf("HTTP server for Keycloak events listening on port %d", tenantWebhookPort)

	err := http.ListenAndServe(addr, mux)
	if err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
