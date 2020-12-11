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
	"time"

	"github.com/Nerzal/gocloak/v7"
	tenantv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	controllers "github.com/netgroup-polito/CrownLabs/operators/pkg/tenant-controller"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
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
	var kcURL string
	var kcTenantOperatorUser string
	var kcTenantOperatorPsw string
	var kcLoginRealm string
	var kcTargetRealm string
	var kcTargetClient string

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&kcURL, "kc-URL", "", "The URL of the keycloak client.")
	flag.StringVar(&kcTenantOperatorUser, "kc-tenant-operator-user", "", "The username of the admin account for keycloak.")
	flag.StringVar(&kcTenantOperatorPsw, "kc-tenant-operator-psw", "", "The password of the admin account for keycloak.")
	flag.StringVar(&kcLoginRealm, "kc-login-realm", "", "The realm where to login the keycloak account.")
	flag.StringVar(&kcTargetRealm, "kc-target-realm", "", "The target realm for keycloak clients and roles.")
	flag.StringVar(&kcTargetClient, "kc-target-client", "", "The target client for keycloak users and roles.")
	flag.Parse()

	if kcURL == "" || kcTenantOperatorUser == "" || kcTenantOperatorPsw == "" ||
		kcLoginRealm == "" || kcTargetRealm == "" || kcTargetClient == "" {
		klog.Fatal("Some keycloak parameters are not defined")
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "f547a6ba.crownlabs.polito.it",
	})
	if err != nil {
		klog.Fatal("Unable to start manager", err)
	}

	kcA, err := newKcActor(kcURL, kcTenantOperatorUser, kcTenantOperatorPsw, kcTargetRealm, kcTargetClient, kcLoginRealm)
	if err != nil {
		klog.Fatal("Error when setting up keycloak", err)
	}

	go checkAndRenewTokenPeriodically(context.Background(), kcA.Client, kcA.Token, kcTenantOperatorUser, kcTenantOperatorPsw, kcLoginRealm, 2*time.Minute, 5*time.Minute)

	if err = (&controllers.TenantReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		klog.Fatal("Unable to create controller for Tenant", err)
	}
	if err = (&controllers.WorkspaceReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		KcA:    kcA,
	}).SetupWithManager(mgr); err != nil {
		klog.Fatal("Unable to create controller for Workspace", err)
	}
	// +kubebuilder:scaffold:builder

	klog.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		klog.Fatal("Problem running manager", err)
	}

}

// newKcActor sets up a keycloak client with the specififed parameters and performs the first login
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
		Client:         kcClient,
		Token:          token,
		TargetClientID: kcTargetClientID,
		TargetRealm:    targetRealmName,
	}, nil
}

// checkAndRenewTokenPeriodically checks every intervalCheck if the token is about in less than expireLimit or is already expired, if so it renews it
func checkAndRenewTokenPeriodically(ctx context.Context, kcClient gocloak.GoCloak, token *gocloak.JWT, kcAdminUser string, kcAdminPsw string, loginRealm string, intervalCheck time.Duration, expireLimit time.Duration) {

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

// getClientID returns the ID of the target client given the human id, to be used with the gocloak library
func getClientID(ctx context.Context, kcClient gocloak.GoCloak, token string, realmName string, targetClient string) (string, error) {

	clients, err := kcClient.GetClients(ctx, token, realmName, gocloak.GetClientsParams{ClientID: &targetClient})
	if err != nil {
		klog.Errorf("Error when getting clientID for client %s", targetClient)
		klog.Error(err)
		return "", err
	} else if len(clients) > 1 {
		klog.Error(nil, "Error, got too many clientIDs for client %s", targetClient)
		return "", err
	} else if len(clients) < 0 {
		klog.Error(nil, "Error, no clientID for client %s", targetClient)
		return "", err

	} else {
		targetClientID := *clients[0].ID
		return targetClientID, nil
	}

}
