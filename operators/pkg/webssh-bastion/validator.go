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

// Package webssh provides a WebSocket-based SSH bridge for CrownLabs instances.
// A piece of the webssh architecture that validates incoming requests
// and retrieves VM information from the Kubernetes API.
// It ensures that the user is authenticated and authorized to access the specified VM.
package webssh

import (
	"context"
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// Retrieves the instance information from the Kubernetes API using the provided token for authentication.
func (webCtx *ServerContext) getInstance(ctx context.Context, token, environment, namespace, instanceName string) (*clv1alpha2.Instance, error) {
	_ = environment // Currently unused, but may be used in the future for multi-environment support.

	if webCtx.BaseConfig == nil {
		return nil, errors.New("baseConfig is not initialized")
	}

	// Create a copy of the base config and set the BearerToken
	cfg := &rest.Config{
		Host:            webCtx.BaseConfig.Host,
		BearerToken:     token,
		TLSClientConfig: webCtx.BaseConfig.TLSClientConfig,
	}

	// Create a new Kubernetes client with the provided token
	k8sClient, err := utils.NewK8sClientWithConfig(cfg)
	if err != nil {
		return nil, errors.New("failed to create Kubernetes client: " + err.Error())
	}

	instance := &clv1alpha2.Instance{}
	err = k8sClient.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      instanceName,
		//Environment: environment,
	}, instance)

	if err != nil {
		return nil, errors.New("failed to get instance: " + err.Error())
	}

	return instance, nil
}

// Extracts the username from the JWT token.
func extractUsernameFromToken(tokenString string) (string, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if username, ok := claims["preferred_username"].(string); ok {
			return username, nil
		}
	}
	return "", errors.New("username not found in token claims")
}

// Validates the incoming request by checking the token, retrieving the instance.
func (webCtx *ServerContext) validateRequest(vmName, token string, localCtx *LocalContext) error {
	// Extract username from the token
	username, err := extractUsernameFromToken(token)
	if err != nil {
		return errors.New("invalid token format: " + err.Error())
	}

	localCtx.username = username

	// get the instance by name and namespace
	instance, err := webCtx.getInstance(localCtx.ctxReq, token, localCtx.environment, localCtx.namespace, vmName)
	if err != nil {
		return errors.New("failed to get instance: " + err.Error())
	}

	// check if the instance is running
	if instance.Status.Phase != clv1alpha2.EnvironmentPhaseReady {
		return errors.New("instance is not running")
	}

	// extract the IP address from the instance status
	if instance.Status.IP == "" {
		return errors.New("instance has no IP address assigned")
	}

	localCtx.ip = instance.Status.IP

	return nil
}
