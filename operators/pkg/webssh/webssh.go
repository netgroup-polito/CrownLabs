// Copyright 2025-2030 Politecnico di Torino
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
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/crypto/ssh"
)

type config struct {
	// the user to use for SSH connections
	SSHUser string
	// the path to the private key file for SSH authentication
	PrivateKeyPath string
	// TimeoutDuration is the duration in seconds after which an SSH connection is considered idle and closed
	TimeoutDuration int
	// MaxConnectionCount is the maximum number of concurrent SSH connections allowed
	MaxConnectionCount int
	// WebsocketPort is the port on which the WebSocket server listens
	WebsocketPort string
	// VMSSHPort is the default SSH port for VMs
	VMSSHPort string
}

var (
	webSshConnections = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bastion_web_ssh_connections",
			Help: "SSH connections established through the WebSocket bridge",
		},
		[]string{"destination_ip", "destination_port"},
	)
)

func loadConfig() *config {
	vmPort := os.Getenv("WEBSSH_VM_PORT")
	if vmPort == "" {
		vmPort = "22"
		log.Println("WEBSSH_VM_PORT environment variable is not set, using default value: ", vmPort)
	}

	maxConnStr := os.Getenv("WEBSSH_MAX_CONN_COUNT")
	if maxConnStr == "" {
		maxConnStr = "1000" // Default value if not set
		log.Println("WEBSSH_MAX_CONN_COUNT environment variable is not set, using default value: ", maxConnStr)
	}

	maxConn, err := strconv.Atoi(maxConnStr)
	if err != nil {
		maxConn = 1000
		log.Println("WEBSSH_MAX_CONN_COUNT is not a valid integer, using default value: ", maxConn)
	}

	timeoutStr := os.Getenv("WEBSSH_TIMEOUT_DURATION")
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		timeout = 30
		log.Println("WEBSSH_TIMEOUT_DURATION is not a valid integer, using default value: ", timeout)
	}

	SSHUser := os.Getenv("WEBSSH_USER")
	if SSHUser == "" {
		SSHUser = "crownlabs"
		log.Println("WEBSSH_USER environment variable is not set, using default value: ", SSHUser)
	}

	privateKeyPath := os.Getenv("WEBSSH_PRIVATE_KEY_PATH")
	if privateKeyPath == "" {
		privateKeyPath = "/web-keys/ssh_web_bastion_master"
		log.Println("WEBSSH_PRIVATE_KEY_PATH environment variable is not set, using default value: ", privateKeyPath)
	}

	websocketPort := os.Getenv("WEBSSH_WEBSOCKET_PORT")
	if websocketPort == "" {
		websocketPort = "8085"
		log.Println("WEBSSH_WEBSOCKET_PORT environment variable is not set, using default value: ", websocketPort)
	}

	return &config{
		SSHUser:            SSHUser,
		PrivateKeyPath:     privateKeyPath,
		TimeoutDuration:    timeout * 60, // Convert minutes to seconds
		MaxConnectionCount: maxConn,
		WebsocketPort:      websocketPort,
		VMSSHPort:          vmPort,
	}
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			_ = r // Ignore the request, we don't need to check the origin in this case
			return true
		},
	}

	activeConnections = make(map[string]*sshConnInfo)
	connMutex         sync.Mutex // Mutex to protect access to activeConnections
)

func loadPrivateKey(path string) (ssh.Signer, error) {
	cleanPath := filepath.Clean(path)
	keyPriv, err := os.ReadFile(cleanPath)
	if err != nil {
		log.Printf("Error reading private key file at %s: %v", cleanPath, err)
		return nil, err
	}

	return ssh.ParsePrivateKey(keyPriv)
}

func returnError(ws *websocket.Conn, errMsg string) {
	if err := ws.WriteMessage(websocket.TextMessage, []byte(errMsg)); err != nil {
		log.Println("WebSocket write error:", err)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request, config *config) {
	// upgrade to the WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}

	defer func() {
		if err := ws.Close(); err != nil {
			log.Printf("failed to close ws connection: %v", err)
		}
	}()

	// wait for the first message to get the token
	_, firstMsg, err := ws.ReadMessage()
	if err != nil {
		log.Println("ReadMessage error:", err)
		returnError(ws, "Error reading initial message")
		return
	}

	// Validate the req
	connString, err := validateRequest(firstMsg, *config)
	if err != nil {
		log.Println("Request validation failed:", err)
		returnError(ws, "Invalid request")
		return
	}

	// log the connection
	ip := connString[:strings.LastIndex(connString, ":")]
	port := connString[strings.LastIndex(connString, ":")+1:]
	webSshConnections.WithLabelValues(ip, port).Inc()

	// Load the private key for SSH authentication
	signer, err := loadPrivateKey(config.PrivateKeyPath)
	if err != nil {
		log.Println("Failed to load private key:", err)
		returnError(ws, "Interal server error")
		return
	}

	log.Printf("Connecting to SSH server at %s with user %s", connString, config.SSHUser)

	sshConfig := &ssh.ClientConfig{
		User: config.SSHUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	sshConn, err := getOrCreateSSHConnection(connString, sshConfig, config.MaxConnectionCount)
	if err != nil {
		log.Println("Failed to get or create SSH connection:", err)
		returnError(ws, "Failed to connect to the SSH server")
		return
	}

	session, err := sshConn.NewSession()
	if err != nil {
		log.Println("Failed to create SSH session:", err)
		returnError(ws, "Internal server error")
		return
	}

	defer func() {
		if err := session.Close(); err != nil {
			log.Printf("failed to close SSH session: %v", err)
		}
	}()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm", 40, 80, modes); err != nil {
		log.Println("Request for pseudo terminal failed:", err)
		returnError(ws, "Internal server error")
		return
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		log.Println("Unable to setup stdin for session:", err)
		returnError(ws, "Internal server error")
		return
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		log.Println("Unable to setup stdout for session:", err)
		returnError(ws, "Internal server error")
		return
	}

	if err := session.Shell(); err != nil {
		log.Println("Failed to start shell:", err)
		returnError(ws, "Internal server error")
		return
	}

	// Start a goroutine to read from SSH stdout and write to WebSocket
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
			updateSSHConnectionLastUsed(connString) // reset timer on shell output
			if err := ws.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
				log.Println("WebSocket write error:", err)
				break
			}
		}
	}()

	// Read messages from WebSocket and write to SSH stdin
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Closing SSH session, WebSocket read error:", err)
			closeSSHConnection(connString)
			break
		}
		updateSSHConnectionLastUsed(connString) // reset timer on user input
		if _, err := stdin.Write(msg); err != nil {
			log.Println("SSH stdin write error:", err)
			break
		}
	}
}

func wsHandlerWrapper(config *config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsHandler(w, r, config)
	}
}

// StartWebSSH initializes the WebSocket SSH bridge server.
// It loads the configuration, sets up the HTTP server, and starts listening for WebSocket connections.
func StartWebSSH() {
	// Load configuration from environment variables
	config := loadConfig()

	// automatic Cleanup
	startConnectionCleanup(2*time.Minute, time.Duration(config.TimeoutDuration)*time.Second)

	// Set up the HTTP server with the WebSocket handler
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsHandlerWrapper(config))

	log.Println("WebSocket SSH bridge running on :" + config.WebsocketPort)

	server := &http.Server{
		Addr:         ":" + config.WebsocketPort,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
