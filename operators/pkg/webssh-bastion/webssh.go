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
// It allows users to connect to their VMs via SSH using a web interface.
package webssh

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/ssh"
	"k8s.io/client-go/rest"
)

// ServerContext holds the context for the WebSSH server.
type ServerContext struct {
	SSHUser            string       // the user to use for SSH connections
	PrivateKeyPath     string       // the path to the private key file for SSH authentication
	TimeoutDuration    int32        // TimeoutDuration is the duration in seconds after which an SSH connection is considered idle and closed
	MaxConnectionCount int32        // MaxConnectionCount is the maximum number of concurrent SSH connections allowed
	WebsocketPort      string       // WebsocketPort is the port on which the WebSocket server listens
	VMSSHPort          string       // VMSSHPort is the default SSH port for VMs
	BaseConfig         *rest.Config // config base with all the standard Kubernetes API settings
	activeConnCount    int32        // active connection count
	BaseLogger         logr.Logger  // logger for the base context
}

// LocalContext holds the context for a local WebSSH connection.
type LocalContext struct {
	log          logr.Logger     // logger for the local context
	connTimedOut int32           // connection timeout status
	connStarted  bool            // connection started status
	lastUsed     atomic.Value    // last used timestamp
	username     string          // username for the connection
	ip           string          // IP address of the VM
	namespace    string          // namespace of the VM
	errorMsg     string          // error to send to the client
	mainCtx      context.Context // main context for the connection
}

var (
	webSSHConnections = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bastion_web_ssh_connections",
			Help: "SSH connections established through the WebSocket bridge",
		},
		[]string{"destination_ip", "destination_port"},
	)
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			_ = r // Ignore the request, we don't need to check the origin in this case
			return true
		},
	}
)

func (webCtx *ServerContext) loadPrivateKey() (ssh.Signer, error) {
	cleanPath := filepath.Clean(webCtx.PrivateKeyPath)
	keyPriv, err := os.ReadFile(cleanPath)
	if err != nil {
		err = errors.New("failed to read private key file: " + err.Error() + " at path: " + cleanPath)
		return nil, err
	}

	return ssh.ParsePrivateKey(keyPriv)
}

func (lc *LocalContext) logger() logr.Logger {
	return lc.log.WithValues(
		"username", lc.username,
		"ip", lc.ip,
		"namespace", lc.namespace,
	)
}

func (webCtx *ServerContext) wsHandler(w http.ResponseWriter, r *http.Request) {
	// upgrade to the WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", "error", err)
		return
	}

	localCtx := &LocalContext{
		log:          webCtx.BaseLogger.WithValues(),
		connTimedOut: 0,
		connStarted:  false,
		username:     "",
		ip:           "",
		namespace:    "",
		lastUsed:     atomic.Value{},
		mainCtx:      r.Context(),
	}

	localCtx.logger().Info("New WebSocket connection established")

	defer func() {
		if localCtx.connStarted {
			atomic.AddInt32(&webCtx.activeConnCount, -1)
		}

		if atomic.LoadInt32(&localCtx.connTimedOut) == 0 {
			if localCtx.errorMsg != "" {
				if err := ws.WriteMessage(websocket.TextMessage, []byte(localCtx.errorMsg)); err != nil {
					localCtx.logger().Error(err, "WebSocket error")
				}
			}
			if err := ws.Close(); err != nil {
				localCtx.logger().Error(err, "Failed to close WebSocket connection")
			}
		}
	}()

	// wait for the first message to get the token
	_, firstMsg, err := ws.ReadMessage()
	if err != nil {
		localCtx.logger().Error(err, "ReadMessage error")
		localCtx.errorMsg = "Failed to read initialization message"
		return
	}

	// Validate the request
	err = webCtx.validateRequest(firstMsg, localCtx)
	if err != nil {
		localCtx.logger().Error(err, "Request validation failed")
		localCtx.errorMsg = "Invalid request"
		return
	}

	// log the connection
	webSSHConnections.WithLabelValues(localCtx.ip, webCtx.VMSSHPort).Inc()

	// check the number of connection
	n := atomic.LoadInt32(&webCtx.activeConnCount)
	if n >= webCtx.MaxConnectionCount {
		localCtx.logger().Error(err, "Max connection limit reached")
		localCtx.errorMsg = "Max connection limit reached"
		return
	}
	atomic.AddInt32(&webCtx.activeConnCount, 1)
	localCtx.connStarted = true

	// Load the private key for SSH authentication
	signer, err := webCtx.loadPrivateKey()
	if err != nil {
		localCtx.logger().Error(err, "Failed to load private key")
		localCtx.errorMsg = "Internal server error"
		return
	}

	sshConfig := &ssh.ClientConfig{
		User: webCtx.SSHUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		// In production, this function should check the server's host key against a trusted source (e.g., known_hosts).
		// However, in our case we are in a controlled, ephemeral student environment,
		// where host verification would add unnecessary friction and limited security value.
		// Therefore, we deliberately skip host key verification.
		//
		// #nosec G106: Ignoring host key verification is acceptable in this context.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	connString := localCtx.ip + ":" + webCtx.VMSSHPort
	sshConn, err := ssh.Dial("tcp", connString, sshConfig)
	if err != nil {
		localCtx.logger().Error(err, "Failed to establish SSH connection")
		localCtx.errorMsg = "Internal server error"
		return
	}

	session, err := sshConn.NewSession()
	if err != nil {
		localCtx.logger().Error(err, "Failed to create SSH session")
		localCtx.errorMsg = "Internal server error"
		return
	}

	defer func() {
		if err := session.Close(); err != nil {
			localCtx.logger().Error(err, "Failed to close SSH session")
			localCtx.errorMsg = "Internal server error"
		}
	}()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm", 40, 80, modes); err != nil {
		localCtx.logger().Error(err, "Request for pseudo terminal failed")
		localCtx.errorMsg = "Internal server error"
		return
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		localCtx.logger().Error(err, "Unable to setup stdin for session")
		localCtx.errorMsg = "Internal server error"
		return
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		localCtx.logger().Error(err, "Unable to setup stdout for session")
		localCtx.errorMsg = "Internal server error"
		return
	}

	if err := session.Shell(); err != nil {
		localCtx.logger().Error(err, "Failed to start shell")
		localCtx.errorMsg = "Internal server error"
		return
	}

	localCtx.logger().Info("SSH session started")

	localCtx.lastUsed.Store(time.Now())

	go func() {
		for {
			time.Sleep(time.Duration(10))
			if time.Since(localCtx.lastUsed.Load().(time.Time)) > time.Duration(webCtx.TimeoutDuration)*time.Minute {
				atomic.StoreInt32(&localCtx.connTimedOut, 1)
				localCtx.logger().Info("Connection timed out due to inactivity")
				localCtx.errorMsg = "Connection timed out due to inactivity"
				if err := ws.Close(); err != nil {
					localCtx.logger().Error(err, "Failed to close WebSocket connection goroutine")
				}
				break
			}
		}
	}()

	// Start a goroutine to read from SSH stdout and write to WebSocket
	// from VM to user
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if atomic.LoadInt32(&localCtx.connTimedOut) == 1 {
				break
			}
			if err != nil {
				if err != io.EOF {
					localCtx.logger().Error(err, "SSH stdout read error")
					localCtx.errorMsg = "SSH session error"
				}
				break
			}
			localCtx.lastUsed.Store(time.Now())

			if err := ws.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
				localCtx.logger().Error(err, "WebSocket write error")
				break
			}
		}
	}()

	// Read messages from WebSocket and write to SSH stdin
	// from user to VM
	for {
		_, msg, err := ws.ReadMessage()

		if atomic.LoadInt32(&localCtx.connTimedOut) == 1 {
			break
		}

		localCtx.lastUsed.Store(time.Now())
		if err != nil {
			localCtx.logger().Error(err, "Closing SSH session, WebSocket read error")
			break
		}
		if _, err := stdin.Write(msg); err != nil {
			localCtx.logger().Error(err, "SSH stdin write error")
			localCtx.errorMsg = "Failed to write to SSH session"
			break
		}
	}
}

func probeHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// StartWebSSH initializes the WebSocket SSH bridge server.
// It loads the configuration, sets up the HTTP server, and starts listening for WebSocket connections.
func (webCtx *ServerContext) StartWebSSH() {
	// Set up the HTTP server with the WebSocket handler
	mux := http.NewServeMux()
	mux.HandleFunc("/webssh", webCtx.wsHandler)

	mux.HandleFunc("/healthz", probeHandler) // Liveness probe endpoint
	mux.HandleFunc("/ready", probeHandler)   // Readiness probe endpoint

	prometheus.MustRegister(webSSHConnections)
	mux.Handle("/metrics", promhttp.Handler()) // Prometheus metrics endpoint

	webCtx.BaseLogger.Info("WebSSH server started on port: " + webCtx.WebsocketPort)

	server := &http.Server{
		Addr:         ":" + webCtx.WebsocketPort,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		webCtx.BaseLogger.Error(err, "HTTP server failed")
	}
}
