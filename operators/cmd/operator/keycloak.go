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

	"github.com/netgroup-polito/CrownLabs/operators/pkg/controller/common"
	"k8s.io/klog/v2"
)

var (
	keycloakURL          string
	keycloakClientID     string
	keycloakClientSecret string
	keycloakRealm        string

	keycloakRolesClientID string // The client ID of the client in which the roles are defined.
)

func init() {
	flag.StringVar(&keycloakURL, "keycloak-url", "", "Keycloak URL")
	flag.StringVar(&keycloakRealm, "keycloak-realm", "", "Keycloak Realm")
	flag.StringVar(&keycloakClientID, "keycloak-client-id", "", "Keycloak Client ID")
	flag.StringVar(&keycloakClientSecret, "keycloak-client-secret", "", "Keycloak Client Secret")
	flag.StringVar(&keycloakRolesClientID, "keycloak-roles-client-id", "", "Keycloak Roles Client ID (the client in which the roles are defined)")
}

func setupKeycloak(
	log klog.Logger,
) error {
	if keycloakURL == "" || keycloakClientID == "" || keycloakClientSecret == "" || keycloakRealm == "" {
		log.Info("Keycloak actor will not be initialized (settings not provided)")
		return nil
	}

	if keycloakRolesClientID == "" {
		keycloakRolesClientID = keycloakClientID
	}

	log.Info("Keycloak settings provided, initializing Keycloak actor")
	err := common.SetupKeycloakActor(
		keycloakURL,
		keycloakClientID,
		keycloakClientSecret,
		keycloakRealm,
		keycloakRolesClientID,
	)
	return err
}
