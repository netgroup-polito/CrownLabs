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
	"fmt"

	"github.com/go-logr/logr"

	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
)

var (
	keycloakURL          string
	keycloakClientID     string
	keycloakClientSecret string
	keycloakRealm        string

	keycloakRolesClientID string // The client ID of the client in which the roles are defined.

	keycloakCompatibilityMode bool // If true, the Keycloak actor will use the compatibility mode for Keycloak old clients.
)

func init() {
	flag.StringVar(&keycloakURL, "keycloak-url", "", "Keycloak URL")
	flag.StringVar(&keycloakRealm, "keycloak-realm", "", "Keycloak Realm")
	flag.StringVar(&keycloakClientID, "keycloak-client-id", "", "Keycloak Client ID")
	flag.StringVar(&keycloakClientSecret, "keycloak-client-secret", "", "Keycloak Client Secret")
	flag.StringVar(&keycloakRolesClientID, "keycloak-roles-client-id", "", "Keycloak Roles Client ID (the client in which the roles are defined)")
	flag.BoolVar(&keycloakCompatibilityMode, "keycloak-compatibility-mode", false, "Enable Keycloak compatibility mode for old clients")
}

func setupKeycloak(
	ctx context.Context,
	log logr.Logger,
) error {
	if keycloakURL == "" || keycloakClientID == "" || keycloakClientSecret == "" || keycloakRealm == "" {
		err := fmt.Errorf("missing parameters for Keycloak configuration")
		log.Error(err, "Keycloak actor will not be initialized (settings not provided)")
		return err
	}

	if keycloakRolesClientID == "" {
		keycloakRolesClientID = keycloakClientID
	}

	log.Info("Initializing Keycloak actor")

	if keycloakCompatibilityMode {
		log.Info("WARNING: Keycloak compatibility mode is enabled")

		err := common.SetupKeycloakActorCompatibility(
			ctx,
			keycloakURL,
			keycloakClientID,
			keycloakClientSecret,
			keycloakRealm,
			keycloakRolesClientID,
			log,
		)
		if err != nil {
			return err
		}
	} else {
		err := common.SetupKeycloakActor(
			ctx,
			keycloakURL,
			keycloakClientID,
			keycloakClientSecret,
			keycloakRealm,
			keycloakRolesClientID,
			log,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
