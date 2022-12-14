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

// Package main contains the entrypoint for the tenant operator.
package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	clv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	controllers "github.com/netgroup-polito/CrownLabs/operators/pkg/tenant-controller"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/tenantwh"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

var (
	scheme = runtime.NewScheme()
)

const (
	// ValidatingWebhookPath -> path on which the validating webhook will be bound. Has to match the one set in the ValidatingWebhookConfiguration.
	ValidatingWebhookPath = "/validate-v1alpha2-tenant"
	// MutatingWebhookPath -> path on which the mutating webhook will be bound. Has to match the one set in the MutatingWebhookConfiguration.
	MutatingWebhookPath = "/mutate-v1alpha2-tenant"
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = clv1alpha1.AddToScheme(scheme)
	_ = clv1alpha2.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var targetLabel string
	var kcURL string
	var kcTnOpUser string
	var kcTnOpPsw string
	var kcLoginRealm string
	var kcTargetRealm string
	var kcTargetClient string
	var ncURL string
	var ncTnOpUser string
	var ncTnOpPsw string
	var maxConcurrentReconciles int
	var webhookBypassGroups string

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&targetLabel, "target-label", "", "The key=value pair label that needs to be in the resource to be reconciled. A single pair in the format key=value")
	flag.StringVar(&kcURL, "kc-url", "", "The URL of the keycloak server.")
	flag.StringVar(&kcTnOpUser, "kc-tenant-operator-user", "", "The username of the acting account for keycloak.")
	flag.StringVar(&kcTnOpPsw, "kc-tenant-operator-psw", "", "The password of the acting account for keycloak.")
	flag.StringVar(&kcLoginRealm, "kc-login-realm", "", "The realm where to login the keycloak acting account.")
	flag.StringVar(&kcTargetRealm, "kc-target-realm", "", "The target realm for keycloak clients, roles and users.")
	flag.StringVar(&kcTargetClient, "kc-target-client", "", "The target client for keycloak users and roles.")
	flag.StringVar(&ncURL, "nc-url", "", "The base URL for the nextcloud actor.")
	flag.StringVar(&ncTnOpUser, "nc-tenant-operator-user", "", "The username of the acting account for nextcloud.")
	flag.StringVar(&ncTnOpPsw, "nc-tenant-operator-psw", "", "The password of the acting account for nextcloud.")
	flag.IntVar(&maxConcurrentReconciles, "max-concurrent-reconciles", 1, "The maximum number of concurrent Reconciles which can be run")
	flag.StringVar(&webhookBypassGroups, "webhook-bypass-groups", "system:masters", "The list of groups which can skip webhooks checks, comma separated values")
	sandboxClusterRole := flag.String("sandbox-cluster-role", "crownlabs-sandbox", "The cluster role defining the permissions for the sandbox namespace.")
	enableWH := flag.Bool("enable-webhooks", true, "Enable webhooks server")

	klog.InitFlags(nil)
	flag.Parse()

	if !klog.V(5).Enabled() {
		klog.SetLogFilter(utils.LogShortenerFilter{})
	}
	ctrl.SetLogger(klogr.NewWithOptions())

	ctx := ctrl.SetupSignalHandler()
	log := ctrl.Log.WithName("setup")

	if targetLabel == "" {
		log.Error(errors.New("missing targetLabel parameter"), "Initialization failed")
		os.Exit(1)
	}
	targetLabelKeyValue := strings.Split(targetLabel, "=")
	if len(targetLabelKeyValue) != 2 {
		log.Error(errors.New("target label format error"), "Initialization failed")
		os.Exit(1)
	}
	targetLabelKey := targetLabelKeyValue[0]
	targetLabelValue := targetLabelKeyValue[1]

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "f547a6ba.crownlabs.polito.it",
		HealthProbeBindAddress: ":8081",
		LivenessEndpointName:   "/healthz",
		ReadinessEndpointName:  "/ready",
	})
	if err != nil {
		log.Error(err, "Unable to create manager")
		os.Exit(1)
	}
	if *enableWH {
		hookServer := mgr.GetWebhookServer()
		webhookBypassGroupsList := strings.Split(webhookBypassGroups, ",")
		hookServer.Register(ValidatingWebhookPath, tenantwh.MakeTenantValidator(mgr.GetClient(), webhookBypassGroupsList))
		hookServer.Register(MutatingWebhookPath, tenantwh.MakeTenantLabeler(mgr.GetClient(), webhookBypassGroupsList, targetLabelKey, targetLabelValue))
	} else {
		log.Info("Webhook set up: operation skipped")
	}

	var kcA *controllers.KcActor
	if kcURL == "" {
		log.Info("Skipping client initialization, as empty target URL", "client", "keycloak")
	} else {
		if kcTnOpUser == "" || kcTnOpPsw == "" ||
			kcLoginRealm == "" || kcTargetRealm == "" || kcTargetClient == "" {
			log.Error(errors.New("missing keycloak parameters"), "Initialization failed")
			os.Exit(1)
		}

		kcA, err = controllers.NewKcActor(kcURL, kcTnOpUser, kcTnOpPsw, kcTargetRealm, kcTargetClient, kcLoginRealm)
		if err != nil {
			log.Error(err, "Unable to setup keycloak")
			os.Exit(1)
		}

		go checkAndRenewTokenPeriodically(ctrl.LoggerInto(ctx, log), kcA, kcTnOpUser, kcTnOpPsw, kcLoginRealm, 2*time.Minute, 5*time.Minute)
	}

	var NcA controllers.NcHandler
	if ncURL == "" {
		log.Info("Skipping client initialization, as empty target URL", "client", "nextcloud")
	} else {
		if ncTnOpUser == "" || ncTnOpPsw == "" {
			log.Error(errors.New("missing nextcloud parameters"), "Initialization failed")
			os.Exit(1)
		}
		httpClient := resty.New().SetCookieJar(nil)
		NcA = &controllers.NcActor{TnOpUser: ncTnOpUser, TnOpPsw: ncTnOpPsw, Client: httpClient, BaseURL: ncURL}
	}
	if err = (&controllers.TenantReconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		KcA:                kcA,
		NcA:                NcA,
		TargetLabelKey:     targetLabelKey,
		TargetLabelValue:   targetLabelValue,
		SandboxClusterRole: *sandboxClusterRole,
		Concurrency:        maxConcurrentReconciles,
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "Unable to create controller", "controller", "tenant")
		os.Exit(1)
	}
	if err = (&controllers.WorkspaceReconciler{
		Client:           mgr.GetClient(),
		Scheme:           mgr.GetScheme(),
		KcA:              kcA,
		TargetLabelKey:   targetLabelKey,
		TargetLabelValue: targetLabelValue,
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "Unable to create controller", "controller", "workspace")
		os.Exit(1)
	}

	// Add readiness probe
	err = mgr.AddReadyzCheck("ready-ping", healthz.Ping)
	if err != nil {
		log.Error(err, "Unable to add the readiness check")
		os.Exit(1)
	}

	// Add liveness probe
	err = mgr.AddHealthzCheck("health-ping", healthz.Ping)
	if err != nil {
		log.Error(err, "Unable to add the health check")
		os.Exit(1)
	}
	klog.Info("Starting manager")
	if err := mgr.Start(ctx); err != nil {
		log.Error(err, "Failed starting manager")
		os.Exit(1)
	}
}

// checkAndRenewTokenPeriodically checks every intervalCheck if the token is about to expire in less than expireLimit seconds or is already expired, if so it renews it.
func checkAndRenewTokenPeriodically(ctx context.Context, kcA *controllers.KcActor, kcAdminUser, kcAdminPsw, loginRealm string, intervalCheck, expireLimit time.Duration) {
	log := ctrl.LoggerFrom(ctx).WithName("token-renewer")

	kcRenewTokenTicker := time.NewTicker(intervalCheck)
	for {
		// wait intervalCheck
		<-kcRenewTokenTicker.C
		// take expiration date of token from tokenJWT claims
		_, claims, err := kcA.Client.DecodeAccessToken(ctx, kcA.GetAccessToken(), loginRealm, "")
		if err != nil {
			log.Error(err, "Error when decoding token")
			os.Exit(1)
		}
		// convert expiration time in usable time
		// tokenExpiresIn :=  time.Unix(int64((*claims)["exp"].(float64)), 0).Until()
		tokenExpiresIn := time.Until(time.Unix(int64((*claims)["exp"].(float64)), 0))

		// if token is about to expire, renew it
		if tokenExpiresIn < expireLimit {
			newToken, err := kcA.Client.LoginAdmin(ctx, kcAdminUser, kcAdminPsw, loginRealm)
			if err != nil {
				log.Error(err, "Error when renewing token")
				os.Exit(1)
			}
			kcA.SetToken(newToken)
			log.Info("Keycloak token renewed")
		}
	}
}
