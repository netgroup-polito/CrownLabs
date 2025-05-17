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
	"fmt"
	"sync"

	"github.com/Nerzal/gocloak/v13"
	"k8s.io/klog/v2"
)

// KcActor contains the needed objects and infos to use keycloak functionalities.
type KcActor struct {
	Client                *gocloak.GoCloak
	token                 *gocloak.JWT
	tokenMutex            sync.RWMutex
	TargetRealm           string
	TargetClientID        string
	UserRequiredActions   []string
	EmailActionsLifeSpanS int
}

var actor KcActor
var initialized bool

func SetupKeycloakActor(
	url string,
	adminRealm string,
	adminUsername string,
	adminPassword string,
	targetRealm string,
	targetClientId string,
) error {
	if initialized {
		return nil
	}

	actor.Client = gocloak.NewClient(url)

	// login to keycloak
	token, err := actor.Client.LoginAdmin(context.Background(), adminUsername, adminPassword, adminRealm)
	if err != nil {
		klog.Error("Unable to login as admin on keycloak", err)
		return err
	}
	actor.SetToken(token)

	actor.TargetRealm = targetRealm

	// get the target client id
	err = getClientIdFromName(targetClientId)
	if err != nil {
		klog.Errorf("Error when getting client id for %s", targetClientId)
		return err
	}

	actor.UserRequiredActions = []string{"UPDATE_PASSWORD", "VERIFY_EMAIL"}
	actor.EmailActionsLifeSpanS = 60 * 60 * 24 * 30 // 30 Days

	initialized = true
	return nil
}

// GetKeycloakActor returns the KcActor currently used.
func GetKeycloakActor() *KcActor {
	return &actor
}

func (a *KcActor) GetToken() string {
	a.tokenMutex.RLock()
	defer a.tokenMutex.RUnlock()

	if a.token == nil {
		return ""
	}
	return a.token.AccessToken
}

func (a *KcActor) SetToken(token *gocloak.JWT) {
	a.tokenMutex.Lock()
	defer a.tokenMutex.Unlock()

	if token == nil {
		return
	}
	a.token = token
}

// getClientIdFromName returns the ID of the target client given the human id, to be used with the gocloak library.
func getClientIdFromName(clientName string) error {
	clients, err := actor.Client.GetClients(context.Background(), actor.token.AccessToken, actor.TargetRealm, gocloak.GetClientsParams{
		ClientID: &clientName,
	})
	if err != nil {
		return err
	}
	if len(clients) != 1 {
		return fmt.Errorf("client %s not found in realm %s", clientName, actor.TargetRealm)
	}

	actor.TargetClientID = *clients[0].ID
	return nil
}
