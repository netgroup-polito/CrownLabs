/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v7"
	"github.com/go-resty/resty/v2"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	tenantv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	controllers "github.com/netgroup-polito/CrownLabs/operators/pkg/tenant-controller"
	// +kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = tenantv1alpha1.AddToScheme(scheme)
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
	klog.InitFlags(nil)
	flag.Parse()

	if targetLabel == "" ||
		kcURL == "" || kcTnOpUser == "" || kcTnOpPsw == "" ||
		kcLoginRealm == "" || kcTargetRealm == "" || kcTargetClient == "" ||
		ncURL == "" || ncTnOpUser == "" || ncTnOpPsw == "" {
		klog.Fatal("Some flag parameters are not defined!")
	}

	targetLabelKeyValue := strings.Split(targetLabel, "=")
	if len(targetLabelKeyValue) != 2 {
		klog.Fatal("Error with target label format")
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
		klog.Fatal("Unable to start manager", err)
	}

	kcA, err := newKcActor(kcURL, kcTnOpUser, kcTnOpPsw, kcTargetRealm, kcTargetClient, kcLoginRealm)
	if err != nil {
		klog.Fatal("Error when setting up keycloak", err)
	}

	go checkAndRenewTokenPeriodically(context.Background(), kcA.Client, kcA.Token, kcTnOpUser, kcTnOpPsw, kcLoginRealm, 2*time.Minute, 5*time.Minute)

	httpClient := resty.New().SetCookieJar(nil)
	NcA := controllers.NcActor{TnOpUser: ncTnOpUser, TnOpPsw: ncTnOpPsw, Client: httpClient, BaseURL: ncURL}
	if err = (&controllers.TenantReconciler{
		Client:           mgr.GetClient(),
		Scheme:           mgr.GetScheme(),
		KcA:              kcA,
		NcA:              &NcA,
		TargetLabelKey:   targetLabelKey,
		TargetLabelValue: targetLabelValue,
	}).SetupWithManager(mgr); err != nil {
		klog.Fatal("Unable to create controller for Tenant", err)
	}
	if err = (&controllers.WorkspaceReconciler{
		Client:           mgr.GetClient(),
		Scheme:           mgr.GetScheme(),
		KcA:              kcA,
		TargetLabelKey:   targetLabelKey,
		TargetLabelValue: targetLabelValue,
	}).SetupWithManager(mgr); err != nil {
		klog.Fatal("Unable to create controller for Workspace", err)
	}
	// +kubebuilder:scaffold:builder
	// Add readiness probe
	err = mgr.AddReadyzCheck("ready-ping", healthz.Ping)
	if err != nil {
		klog.Fatal("unable add a readiness check", err)
	}

	// Add liveness probe
	err = mgr.AddHealthzCheck("health-ping", healthz.Ping)
	if err != nil {
		klog.Fatal("unable add a health check", err)
	}
	klog.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		klog.Fatal("Problem running manager", err)
	}
}

// newKcActor sets up a keycloak client with the specified parameters and performs the first login.
func newKcActor(kcURL, kcUser, kcPsw, targetRealmName, targetClient, loginRealm string) (*controllers.KcActor, error) {
	kcClient := gocloak.NewClient(kcURL)
	token, err := kcClient.LoginAdmin(context.Background(), kcUser, kcPsw, loginRealm)
	if err != nil {
		klog.Error("Unable to login as admin on keycloak", err)
		return nil, err
	}
	kcTargetClientID, err := getClientID(context.Background(), kcClient, token.AccessToken, targetRealmName, targetClient)
	if err != nil {
		klog.Errorf("Error when getting client id for %s", targetClient)
		return nil, err
	}
	return &controllers.KcActor{
		Client:                kcClient,
		Token:                 token,
		TargetClientID:        kcTargetClientID,
		TargetRealm:           targetRealmName,
		UserRequiredActions:   []string{"UPDATE_PASSWORD", "VERIFY_EMAIL"},
		EmailActionsLifeSpanS: 60 * 60 * 24 * 30, // 30 Days
	}, nil
}

// checkAndRenewTokenPeriodically checks every intervalCheck if the token is about to expire in less than expireLimit seconds or is already expired, if so it renews it.
func checkAndRenewTokenPeriodically(ctx context.Context, kcClient gocloak.GoCloak, token *gocloak.JWT, kcAdminUser, kcAdminPsw, loginRealm string, intervalCheck, expireLimit time.Duration) {
	kcRenewTokenTicker := time.NewTicker(intervalCheck)
	for {
		// wait intervalCheck
		<-kcRenewTokenTicker.C
		// take expiration date of token from tokenJWT claims
		_, claims, err := kcClient.DecodeAccessToken(ctx, token.AccessToken, loginRealm, "")
		if err != nil {
			klog.Fatal("Error when decoding token", err)
		}
		// convert expiration time in usable time
		// tokenExpiresIn :=  time.Unix(int64((*claims)["exp"].(float64)), 0).Until()
		tokenExpiresIn := time.Until(time.Unix(int64((*claims)["exp"].(float64)), 0))

		// if token is about to expire, renew it
		if tokenExpiresIn < expireLimit {
			newToken, err := kcClient.LoginAdmin(ctx, kcAdminUser, kcAdminPsw, loginRealm)
			if err != nil {
				klog.Fatal("Error when renewing token", err)
			}
			*token = *newToken
			klog.Info("Keycloak token renewed")
		}
	}
}

// getClientID returns the ID of the target client given the human id, to be used with the gocloak library.
func getClientID(ctx context.Context, kcClient gocloak.GoCloak, token, realmName, targetClient string) (string, error) {
	clients, err := kcClient.GetClients(ctx, token, realmName, gocloak.GetClientsParams{ClientID: &targetClient})
	if err != nil {
		klog.Errorf("Error when getting clientID for client %s", targetClient)
		klog.Error(err)
		return "", err
	}

	switch len(clients) {
	case 0:
		klog.Error(nil, "Error, no clientID for client %s", targetClient)
		return "", fmt.Errorf("no client ID for client %s", targetClient)
	case 1:
		targetClientID := *clients[0].ID
		return targetClientID, nil
	default:
		klog.Error(nil, "Error, got too many clientIDs for client %s", targetClient)
		return "", fmt.Errorf("too many clientIDs for client %s", targetClient)
	}
}
