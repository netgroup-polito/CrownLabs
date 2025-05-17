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
package utils

import (
	"context"

	"github.com/Nerzal/gocloak/v13"
	"k8s.io/klog/v2"
)

// KcActor contains the needed objects and infos to use keycloak functionalities.
type KcActor struct {
	initialized bool
	Client      *gocloak.GoCloak
	Realm       string
	// token                 *gocloak.JWT
	// tokenMutex            sync.RWMutex
	// UserRequiredActions   []string
	// EmailActionsLifeSpanS int
	credentials struct {
		ClientID     string
		ClientSecret string
	}
}

var actor KcActor

func SetupKeycloakActor(
	url string,
	clientID string,
	clientSecret string,
	realm string,
) error {
	if actor.initialized {
		return nil
	}

	actor.Client = gocloak.NewClient(url)

	// login to keycloak
	_, err := actor.Client.LoginClient(context.Background(), clientID, clientSecret, realm)
	if err != nil {
		klog.Error("Unable to login as admin on keycloak", err)
		return err
	}

	// actor.UserRequiredActions = []string{"UPDATE_PASSWORD", "VERIFY_EMAIL"}
	// actor.EmailActionsLifeSpanS = 60 * 60 * 24 * 30 // 30 Days

	actor.Realm = realm
	actor.credentials.ClientID = clientID
	actor.credentials.ClientSecret = clientSecret
	actor.initialized = true
	return nil
}

// GetKeycloakActor returns the KcActor currently used.
func GetKeycloakActor() *KcActor {
	return &actor
}

// func (a *KcActor) GetAccessToken() string {
// 	a.tokenMutex.RLock()
// 	defer a.tokenMutex.RUnlock()

// 	if a.token == nil {
// 		return ""
// 	}
// 	return a.token.AccessToken
// }

// func (a *KcActor) SetToken(token *gocloak.JWT) {
// 	a.tokenMutex.Lock()
// 	defer a.tokenMutex.Unlock()

// 	if token == nil {
// 		return
// 	}
// 	a.token = token
// }
