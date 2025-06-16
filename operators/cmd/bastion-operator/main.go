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

// Package main contains the entrypoint for the bastion operator.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/gorilla/websocket"
	crownlabsv1alpha1 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha1"
	crownlabsv1alpha2 "github.com/netgroup-polito/CrownLabs/operators/api/v1alpha2"
	bastion_controller "github.com/netgroup-polito/CrownLabs/operators/pkg/bastion-controller"
	"golang.org/x/crypto/ssh"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = crownlabsv1alpha1.AddToScheme(scheme)
	_ = crownlabsv1alpha2.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

const (
	sshUser            = "crownlabs"
	privateKeyPath     = "/web-keys/ssh_web_bastion_master"
	timeoutDuration    = 30 * time.Minute // Duration after which an SSH connection is considered idle and closed
	maxIdleConnections = 1000             // Maximum number of idle SSH connections to keep open
)

type sshConnInfo struct {
	conn     *ssh.Client
	lastUsed time.Time
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	activeConnections = make(map[string]*sshConnInfo)
	connMutex         sync.Mutex // Mutex to protect access to activeConnections
)

// --- JWT extraction ---
func extractUsernameFromToken(tokenString string) (string, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Adatta il claim secondo la configurazione Keycloak (es: "preferred_username" o "sub")
		if username, ok := claims["preferred_username"].(string); ok {
			return username, nil
		}
		if sub, ok := claims["sub"].(string); ok {
			return sub, nil
		}
	}
	return "", fmt.Errorf("username not found in token")
}

// --- RBAC check ---
func canUserAccessResource(token, namespace, instanceName string) (bool, error) {
	token = strings.TrimPrefix(token, "Bearer ")
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT")
	if host == "" || port == "" {
		return false, errors.New("KUBERNETES_SERVICE_HOST or PORT not set")
	}
	apiServer := "https://" + host + ":" + port

	config := &rest.Config{
		Host:        apiServer,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			CAFile: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
		},
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return false, err
	}

	ssar := &authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Namespace: namespace,
				Verb:      "get",
				Group:     "crownlabs.polito.it",
				Resource:  "instances",
				Name:      instanceName,
			},
		},
	}

	result, err := clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(context.Background(), ssar, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}
	return result.Status.Allowed, nil
}

// --- Instance owner check ---
func getInstanceOwner(token, namespace, instanceName string) (string, error) {
	token = strings.TrimPrefix(token, "Bearer ")
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT")
	if host == "" || port == "" {
		return "", errors.New("KUBERNETES_SERVICE_HOST or PORT not set")
	}
	apiServer := "https://" + host + ":" + port

	config := &rest.Config{
		Host:        apiServer,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			CAFile: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
		},
	}
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return "", err
	}

	gvr := schema.GroupVersionResource{
		Group:    "crownlabs.polito.it",
		Version:  "v1alpha2",
		Resource: "instances",
	}
	inst, err := dynClient.Resource(gvr).Namespace(namespace).Get(context.Background(), instanceName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	labels := inst.GetLabels()
	owner, ok := labels["crownlabs.polito.it/owner"]
	if !ok {
		return "", fmt.Errorf("owner label not found")
	}
	return owner, nil
}

// --- Combined check ---
func verifyTokenAndOwnership(token, namespace, instanceName string) bool {
	// 1. RBAC check
	allowed, err := canUserAccessResource(token, namespace, instanceName)
	if err != nil {
		log.Println("K8s permission check error:", err)
		return false
	}
	if !allowed {
		log.Println("RBAC denied access")
		return false
	}

	// 2. Owner check
	username, err := extractUsernameFromToken(token)
	if err != nil {
		log.Println("Token decode error:", err)
		return false
	}
	owner, err := getInstanceOwner(token, namespace, instanceName)
	if err != nil {
		log.Println("Error retrieving instance owner:", err)
		return false
	}
	if username != owner {
		log.Printf("User %s is not the owner (%s) of instance %s\n", username, owner, instanceName)
		return false
	}
	return true
}

/*
func verifyToken(token string) bool {
	if token == "" {
		log.Println("No token provided")
		return false
	}

	/*if err := ih.Client.Get(r.Context(), forge.NamespacedName(inst), inst); err != nil {
			req.toket = token
		if errors.IsNotFound(err) {
			log.Error(err, "instance not found")
			WriteError(w, r, log, http.StatusNotFound, "The requested Instance does not exist.")
			return
		}
		log.Error(err, "error retrieving instance")
		WriteError(w, r, log, http.StatusInternalServerError, "Cannot retrieve the requested instance.")
		return
	}

	// connect to the k8s API server to verify the token

	// TODO: Implement actual token verification logic
	return true // Replace with actual token verification logic
}
*/

func loadPrivateKey() (ssh.Signer, error) {
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(key)
}

func getOrCreateSSHConnection(vmIp string, sshConfig *ssh.ClientConfig) (*ssh.Client, error) {
	connMutex.Lock()
	defer connMutex.Unlock()

	// Check if the connection already exists
	if info, exists := activeConnections[vmIp]; exists {
		info.lastUsed = time.Now()
		return info.conn, nil
	}

	// check if the number of active connections exceeds the limit
	if len(activeConnections) >= maxIdleConnections {
		log.Printf("Maximum number of idle SSH connections (%d) reached. Closing oldest connection.", maxIdleConnections)
		return nil, errors.New("maximum number of idle SSH connections reached")
	}

	// If not, create a new SSH connection
	conn, err := ssh.Dial("tcp", vmIp, sshConfig)
	if err != nil {
		return nil, err
	}

	// Save the new connection in the map
	activeConnections[vmIp] = &sshConnInfo{
		conn:     conn,
		lastUsed: time.Now(),
	}

	return conn, nil
}

func startConnectionCleanup(interval time.Duration, timeout time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			connMutex.Lock()
			now := time.Now()
			for ip, info := range activeConnections {
				if now.Sub(info.lastUsed) > timeout {
					log.Printf("Closing idle SSH connection to %s", ip)
					info.conn.Close()
					delete(activeConnections, ip)
				}
			}
			connMutex.Unlock()
		}
	}()
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received WebSocket connection")

	token := r.Header.Get("Authorization") // Extract token from Authorization header
	namespace := r.URL.Query().Get("namespace")
	instanceName := r.URL.Query().Get("instance")
	if namespace == "" || instanceName == "" {
		http.Error(w, "Missing namespace or instance parameter", http.StatusBadRequest)
		log.Println("Missing namespace or instance in request")
		return
	}

	if !verifyTokenAndOwnership(token, namespace, instanceName) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Println("Unauthorized access attempt")
		return
	}

	vmIp := r.URL.Query().Get("vmIp") // Extract SSH address from query parameter
	if vmIp == "" {
		http.Error(w, "Missing vmIp address", http.StatusBadRequest)
		log.Println("Missing SSH address in request")
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}

	defer func() {
		ws.Close()
	}()

	signer, err := loadPrivateKey()
	if err != nil {
		log.Println("Failed to load private key:", err)
		return
	}

	sshConfig := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // should be the case to replace it?
		Timeout:         10 * time.Second,
	}

	sshConn, err := getOrCreateSSHConnection(vmIp, sshConfig)
	if err != nil {
		log.Println("Failed to get or create SSH connection:", err)
		http.Error(w, "Failed to connect to VM, too many connections", http.StatusInternalServerError)
		return
	}

	session, err := sshConn.NewSession()
	if err != nil {
		log.Println("Failed to create SSH session:", err)
		return
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm", 40, 80, modes); err != nil {
		log.Println("Request for pseudo terminal failed:", err)
		return
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		log.Println("Unable to setup stdin for session:", err)
		return
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		log.Println("Unable to setup stdout for session:", err)
		return
	}

	if err := session.Shell(); err != nil {
		log.Println("Failed to start shell:", err)
		return
	}

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Println("SSH stdout read error:", err)
				}
				break
			}
			ws.WriteMessage(websocket.TextMessage, buf[:n])
		}
	}()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			for len(msg) > 0 {
				n, err := stdin.Write(msg)
				if err != nil {
					log.Println("SSH stdin write error:", err)
					break
				}
				msg = msg[n:] // Remove the written bytes from the message
			}
			log.Println("SSH stdin write error:", err)
			break
		}
	}
}

func setupWebSSH() {
	// automatic Cleanup
	startConnectionCleanup(1*time.Minute, timeoutDuration)

	http.HandleFunc("/ws", wsHandler)
	log.Println("WebSocket SSH bridge running on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	klog.InitFlags(nil)
	flag.Parse()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                server.Options{BindAddress: metricsAddr},
		WebhookServer:          webhook.NewServer(webhook.Options{Port: 9443}),
		LeaderElection:         enableLeaderElection,
		HealthProbeBindAddress: ":8081",
		LivenessEndpointName:   "/healthz",
		ReadinessEndpointName:  "/ready",
	})
	if err != nil {
		klog.Fatal("Unable to start manager", err)
	}

	authorizedKeysPath, isEnvSet := os.LookupEnv("AUTHORIZED_KEYS_PATH")
	if !isEnvSet {
		klog.Info("AUTHORIZED_KEYS_PATH env var is not set. Using default path \"/auth-keys-vol/authorized_keys\"")
		authorizedKeysPath = "/auth-keys-vol/authorized_keys"
	} else {
		klog.Infof("AUTHORIZED_KEYS_PATH env var found. Using path %v", authorizedKeysPath)
	}

	if err = (&bastion_controller.BastionReconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		AuthorizedKeysPath: authorizedKeysPath,
	}).SetupWithManager(mgr); err != nil {
		klog.Fatal("unable to create controller", "controller", "Bastion", err)
	}

	// +kubebuilder:scaffold:builder
	// Add readiness probe
	err = mgr.AddReadyzCheck("ready-ping", healthz.Ping)
	if err != nil {
		klog.Fatal("Unable to add a readiness check", err)
	}

	// Add liveness probe
	err = mgr.AddHealthzCheck("health-ping", healthz.Ping)
	if err != nil {
		klog.Fatal("Unable add an health check", err)
	}

	// Start the WebSocket SSH bridge in a separate goroutine
	go func() {
		klog.Info("Starting WebSocket SSH bridge")
		setupWebSSH()
	}()

	klog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		klog.Fatal("problem running manager", err)
	}
}
