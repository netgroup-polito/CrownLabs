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
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/tenant"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/tenant/webhook"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	tenantKeycloakWebhookAddr string
	tenantNSKeepAlive         time.Duration
	sandboxClusterRole        string
	tenantWebhookBypassGroups string
	tenantBaseWorkspaces      string
)

const (
	// TenantWebhookPath -> path on which the tenant webhook will be bound.
	TenantWebhookPath = "/tenant-webhook"

	// ValidatorWebhookPath -> path on which the validator webhook will be bound.
	ValidatorWebhookPath = "/validator-v1alpha2-tenant"
	// DefaulterWebhookPath -> path on which the defaulter webhook will be bound.
	DefaulterWebhookPath = "/defaulter-v1alpha2-tenant"
)

func init() {
	flag.StringVar(&tenantKeycloakWebhookAddr, "tenant-webhook-port", ":8082", "Port for the tenant webhook server for Keycloak events")
	flag.DurationVar(&tenantNSKeepAlive, "tenant-ns-keep-alive", 24*time.Hour,
		"Time elapsed after last login of tenant during which the tenant namespace should be kept alive")
	flag.StringVar(&sandboxClusterRole, "sandbox-cluster-role", "crownlabs-sandbox", "The cluster role defining the permissions for the sandbox namespace.")
	flag.StringVar(&tenantWebhookBypassGroups, "webhook-bypass-groups", "system:masters", "The list of groups which can skip webhooks checks, comma separated values")
	flag.StringVar(&tenantBaseWorkspaces, "base-workspaces", "", "List of comma separated workspaces to be enforced to every tenant")

	// mydrivePVCsSize := args.NewQuantity("1Gi")
	// var mydrivePVCsStorageClassName string
	// var myDrivePVCsNamespace string
	// flag.Var(&mydrivePVCsSize, "mydrive-pvcs-size", "The dimension of the user's personal space")
	// flag.StringVar(&mydrivePVCsStorageClassName, "mydrive-pvcs-storage-class-name", "rook-nfs", "The name for the user's storage class")
	// flag.StringVar(&myDrivePVCsNamespace, "mydrive-pvcs-namespace", "mydrive-pvcs", "The namespace where the PVCs are created")
}

func setup_tenant(
	mgr manager.Manager,
	targetLabel common.KVLabel,
) error {
	var baseWorkspacesList []string
	if tenantBaseWorkspaces != "" {
		baseWorkspacesList = strings.Split(tenantBaseWorkspaces, ",")
		log.Printf("Base workspaces for tenants to be enforced: %v", baseWorkspacesList)
	}

	tn := &tenant.TenantReconciler{
		Client:                  mgr.GetClient(),
		Scheme:                  mgr.GetScheme(),
		TargetLabel:             targetLabel,
		TenantNSKeepAlive:       tenantNSKeepAlive,
		TriggerReconcileChannel: make(chan event.GenericEvent, 10),
		KeycloakActor:           common.GetKeycloakActor(),
		SandboxClusterRole:      sandboxClusterRole,
		BaseWorkspaces:          baseWorkspacesList,
	}

	if err := tn.SetupWithManager(mgr); err != nil {
		return err
	}

	// Register the Keycloak event handler for tenant webhook events
	go startHTTPServer(tn)

	// Setup the webhook for tenant validation and defaulting
	if enableWebhooks {
		if err := setupTenantWebhook(mgr, targetLabel, baseWorkspacesList); err != nil {
			return err
		}
	}

	return nil
}

// starts the HTTP server for the Keycloak webhook events endpoint
func startHTTPServer(tn *tenant.TenantReconciler) {
	mux := http.NewServeMux()

	// registering the handler for the tenant webhook path
	mux.HandleFunc(TenantWebhookPath, func(w http.ResponseWriter, r *http.Request) {
		tn.KeycloakEventHandler(w, r)
	})

	log.Printf("HTTP server for Keycloak events listening on %s", tenantKeycloakWebhookAddr)

	err := http.ListenAndServe(tenantKeycloakWebhookAddr, mux)
	if err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func setupTenantWebhook(
	mgr manager.Manager,
	targetLabel common.KVLabel,
	baseWorkspaces []string,
) error {
	tnWh := webhook.TenantWebhook{
		Client:       mgr.GetClient(),
		BypassGroups: strings.Split(tenantWebhookBypassGroups, ","),
	}

	if err := ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha2.Tenant{}).
		WithValidator(&webhook.TenantValidator{
			TenantWebhook: tnWh,
		}).
		WithValidatorCustomPath(ValidatorWebhookPath).
		WithDefaulter(&webhook.TenantDefaulter{
			TenantWebhook:   tnWh,
			Decoder:         admission.NewDecoder(mgr.GetScheme()),
			OpSelectorLabel: targetLabel,
			BaseWorkspaces:  baseWorkspaces,
		}).
		WithDefaulterCustomPath(DefaulterWebhookPath).
		Complete(); err != nil {
		return err
	}

	return nil
}
