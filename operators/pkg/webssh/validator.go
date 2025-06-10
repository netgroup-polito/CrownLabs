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
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	"github.com/netgroup-polito/CrownLabs/operators/pkg/utils"
)

type clientInitMessage struct {
	Token     string `json:"token"`
	VMName    string `json:"vmName"`              // The name of the VM to connect to
	Namespace string `json:"namespace,omitempty"` // Optional namespace, can be derived from the token
}

// GetInstance retrieves an Instance CR by namespace and name.
func GetInstance(token, namespace, instanceName string) (*clv1alpha2.Instance, error) {
	// create the Kubernetes client configuration
	config := &rest.Config{
		Host:        "https://apiserver.crownlabs.polito.it",
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: false,
		},
	}

	k8sClient, err := utils.NewK8sClientWithConfig(config)
	if err != nil {
		return nil, errors.New("failed to create Kubernetes client: " + err.Error())
	}

	instance := &clv1alpha2.Instance{}
	err = k8sClient.Get(context.TODO(), client.ObjectKey{
		Namespace: namespace,
		Name:      instanceName,
	}, instance)

	if err != nil {
		return nil, errors.New("failed to get instance: " + err.Error())
	}

	return instance, nil
}

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

func validateRequest(firstMsg []byte) (ip, username string, err error) {
	var initMsg clientInitMessage
	if err := json.Unmarshal(firstMsg, &initMsg); err != nil {
		return "", "", errors.New("invalid JSON format")
	}

	if initMsg.VMName == "" || initMsg.Token == "" {
		return "", "", errors.New("missing required fields in the initialization message")
	}

	// Extract username from the token
	username, err = extractUsernameFromToken(initMsg.Token)
	if err != nil {
		return "", "", errors.New("invalid token format: " + err.Error())
	}

	// Get the namespace from the message, or derive it from the token
	namespace := initMsg.Namespace
	if namespace == "" {
		namespace = "tenant-" + username
	}

	log.Println("Validating request: ", "user:", username, " namespace:", namespace, " vmName:", initMsg.VMName)

	// get the instance by name and namespace
	instance, err := GetInstance(initMsg.Token, namespace, initMsg.VMName)
	if err != nil {
		return "", "", errors.New("failed to get instance: " + err.Error())
	}

	// Extract the IP address from the instance object
	if !instance.Spec.Running {
		return "", "", errors.New("instance is not running")
	}

	// extract the IP address from the instance status
	if instance.Status.IP == "" {
		return "", "", errors.New("instance has no IP address assigned")
	}

	return instance.Status.IP, username, nil
}
