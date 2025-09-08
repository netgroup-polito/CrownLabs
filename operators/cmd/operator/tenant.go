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
	"context"
	"flag"
	"net/http"
	"strings"
	"time"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/tenant"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/tenant/webhook"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils/args"
)

var (
	tenantKeycloakWebhookAddr     string
	tenantNSKeepAlive             time.Duration
	sandboxClusterRole            string
	tenantWebhookBypassGroups     string
	tenantBaseWorkspaces          string
	tenantMaxConcurrentReconciles int
	mydrivePVCsSize               args.Quantity
	mydrivePVCsStorageClassName   string
	myDrivePVCsNamespace          string
	waitUserVerification          bool // If true, the reconciliation will wait for the user to be verified in Keycloak before creating resources.
)

const (
	// KeycloakEventsWebhookPath -> path on which the tenant webhook will be bound.
	KeycloakEventsWebhookPath = "/tenant-webhook"

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

	mydrivePVCsSize = args.NewQuantity("1Gi")
	flag.Var(&mydrivePVCsSize, "mydrive-pvcs-size", "The dimension of the user's personal space")
	flag.StringVar(&mydrivePVCsStorageClassName, "mydrive-pvcs-storage-class-name", "rook-nfs", "The name for the user's storage class")
	flag.StringVar(&myDrivePVCsNamespace, "mydrive-pvcs-namespace", "mydrive-pvcs", "The namespace where the PVCs are created")
	flag.BoolVar(&waitUserVerification, "wait-user-verification", true, "Wait for the user to be verified in Keycloak before creating resources. If false, resources will be created immediately after the Tenant is created.")

	flag.Int64Var(&forge.CapInstance, "cap-instance", 10, "The cap number of instances that can be requested by a Tenant.")
	flag.IntVar(&forge.CapCPU, "cap-cpu", 25, "The cap amount of CPU cores that can be requested by a Tenant.")
	flag.IntVar(&forge.CapMemoryGiga, "cap-memory-giga", 50, "The cap amount of RAM memory in gigabytes that can be requested by a Tenant.")

	flag.IntVar(&tenantMaxConcurrentReconciles, "max-concurrent-reconciles", 1, "The maximum number of concurrent Reconciles which can be run")
}

func setupTenant(
	mgr manager.Manager,
	log logr.Logger,
	targetLabel common.KVLabel,
) error {
	var baseWorkspacesList []string
	if tenantBaseWorkspaces != "" {
		baseWorkspacesList = strings.Split(tenantBaseWorkspaces, ",")
		log.Info("Base workspaces for tenants to be enforced", "workspaces", baseWorkspacesList)
	}

	tn := &tenant.Reconciler{
		Client:                      mgr.GetClient(),
		Scheme:                      mgr.GetScheme(),
		TargetLabel:                 targetLabel,
		TenantNSKeepAlive:           tenantNSKeepAlive,
		TriggerReconcileChannel:     make(chan event.GenericEvent, 10),
		MyDrivePVCsSize:             mydrivePVCsSize.Quantity,
		MyDrivePVCsStorageClassName: mydrivePVCsStorageClassName,
		MyDrivePVCsNamespace:        myDrivePVCsNamespace,
		KeycloakActor:               common.GetKeycloakActor(),
		WaitUserVerification:        waitUserVerification,
		SandboxClusterRole:          sandboxClusterRole,
		BaseWorkspaces:              baseWorkspacesList,
		Concurrency:                 tenantMaxConcurrentReconciles,
		Reschedule:                  reschedule,
	}

	if err := tn.SetupWithManager(mgr, log); err != nil {
		return err
	}

	// Register the Keycloak event handler for tenant webhook events
	startKeycloakWebhookHTTPServer(tn, log, mgr)

	// Setup the webhook for tenant validation and defaulting
	if enableWebhooks {
		if err := setupTenantWebhook(mgr, targetLabel, baseWorkspacesList); err != nil {
			return err
		}
	}

	return nil
}

// starts the HTTP server for the Keycloak webhook events endpoint.
func startKeycloakWebhookHTTPServer(
	tn *tenant.Reconciler,
	log logr.Logger,
	mgr manager.Manager,
) {
	mux := http.NewServeMux()

	// registering the handler for the tenant webhook path
	mux.HandleFunc(KeycloakEventsWebhookPath, func(w http.ResponseWriter, r *http.Request) {
		log := ctrl.LoggerFrom(r.Context(), "tenant-keycloak-handler", r.RemoteAddr)
		tn.KeycloakEventHandler(log, w, r)
	})

	log.Info("HTTP server for Keycloak events listening", "address", tenantKeycloakWebhookAddr)

	srv := &http.Server{
		Addr:              tenantKeycloakWebhookAddr,
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start the server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(err, "Failed to start HTTP server")
		}
	}()

	// Register a function to be called when the manager is shutting down
	if err := mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		// This will be called when the manager is shutting down
		<-ctx.Done()

		log.Info("Shutting down HTTP server for Keycloak events")

		// Create a context with a timeout for the shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error(err, "Error during HTTP server shutdown")
			return err
		}

		log.Info("HTTP server for Keycloak events has been shut down")
		return nil
	})); err != nil {
		log.Error(err, "Failed to add HTTP server to manager")
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
