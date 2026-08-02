// Copyright 2020-2026 Politecnico di Torino
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

// Package webvnc provides a WebSocket-based VNC bridge for CrownLabs instances.
// A piece of the webvnc architecture that validates incoming requests
// and retrieves VM information from the Kubernetes API.
// It ensures that the user is authenticated and authorized to access the specified VM.
package webvnc

import (
	"context"
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/forge"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

// Retrieves the instance and the requested environment from the Kubernetes API
// using the provided token for authentication.
func (webCtx *ServerContext) getInstanceEnvironment(ctx context.Context, token, environment, namespace, instanceName string) (*clv1alpha2.Instance, *clv1alpha2.InstanceStatusEnv, error) {
	if webCtx.BaseConfig == nil {
		return nil, nil, errors.New("baseConfig is not initialized")
	}

	// Create a copy of the base config and set the BearerToken, so that the
	// request is authenticated and authorized as the requesting user, not
	// as the webvnc service itself.
	cfg := &rest.Config{
		Host:            webCtx.BaseConfig.Host,
		BearerToken:     token,
		TLSClientConfig: webCtx.BaseConfig.TLSClientConfig,
	}

	k8sClient, err := utils.NewK8sClientWithConfig(cfg)
	if err != nil {
		return nil, nil, errors.New("failed to create Kubernetes client: " + err.Error())
	}

	instance := &clv1alpha2.Instance{}
	if err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: instanceName}, instance); err != nil {
		return nil, nil, errors.New("failed to get instance: " + err.Error())
	}

	for envIdx := range instance.Status.Environments {
		env := &instance.Status.Environments[envIdx]
		if env.Name == environment {
			return instance, env, nil
		}
	}

	return nil, nil, errors.New("environment not found")
}

// Extracts the username from the JWT token, for logging purposes only.
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

// Validates the incoming request and returns the namespaced name of the
// VirtualMachineInstance to connect to. The name is derived from the
// authorized Instance object itself (same naming scheme used by the
// controller that creates it, forge.NamespacedNameWithSuffix), never taken
// directly from client input, so that a request authorized on one instance
// cannot be used to reach a different one.
func (webCtx *ServerContext) validateRequest(ctx context.Context, vmName, token string, localCtx *LocalContext) (types.NamespacedName, error) {
	username, err := extractUsernameFromToken(token)
	if err != nil {
		return types.NamespacedName{}, errors.New("invalid token format: " + err.Error())
	}
	localCtx.username = username

	instance, env, err := webCtx.getInstanceEnvironment(ctx, token, localCtx.environment, localCtx.namespace, vmName)
	if err != nil {
		return types.NamespacedName{}, err
	}

	if env.Phase != clv1alpha2.EnvironmentPhaseReady {
		return types.NamespacedName{}, errors.New("environment is not running")
	}

	return forge.NamespacedNameWithSuffix(instance, env.Name), nil
}
