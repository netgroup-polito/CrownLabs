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
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
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
	SSHUser            string        // the user to use for SSH connections
	PrivateKeyPath     string        // the path to the private key file for SSH authentication
	TimeoutDuration    time.Duration // TimeoutDuration is the duration in seconds after which an SSH connection is considered idle and closed
	MaxConnectionCount int32         // MaxConnectionCount is the maximum number of concurrent SSH connections allowed
	WebsocketPort      string        // WebsocketPort is the port on which the WebSocket server listens
	VMSSHPort          string        // VMSSHPort is the default SSH port for VMs
	BaseConfig         *rest.Config  // config base with all the standard Kubernetes API settings
	activeConnCount    int32         // active connection count
	BaseLogger         logr.Logger   // logger for the base context
}

// LocalContext holds the context for a local WebSSH connection.
type LocalContext struct {
	log         logr.Logger        // logger for the local context
	lastUsed    atomic.Value       // last used timestamp
	username    string             // username for the connection
	ip          string             // IP address of the VM
	namespace   string             // namespace of the VM
	environment string             // environment of the VM
	errorMsg    string             // error to send to the client
	ctxReq      context.Context    // main context for the connection
	ctxServer   context.Context    // context for the HTTP server
	cancelFun   context.CancelFunc // function to cancel the server context
	ws          *websocket.Conn    // WebSocket connection
	session     *ssh.Session       // SSH session
	wg          sync.WaitGroup     // wait group for goroutines
	stdin       io.WriteCloser     // stdin pipe for the SSH session
	stdout      io.Reader          // stdout pipe for the SSH session
}

// clientMessage represents a message from the client to the server.
type clientMessage struct {
	Type string `json:"type"`           // "resize", "input" or "ping"
	Cols int    `json:"cols,omitempty"` // used if Type is "resize"
	Rows int    `json:"rows,omitempty"` // used if Type is "resize"
	Data string `json:"data,omitempty"` // used if Type is "input"
}

// serverMessage represents a message from the server to the client.
type serverMessage struct {
	Type  string `json:"type,omitempty"`  // "error", "data" or "pong"
	Error string `json:"error,omitempty"` // non-empty if there is an error
	Data  string `json:"data,omitempty"`  // non-empty if there is data to send
}

// clientInitMessage represents the initial message sent by the client to establish the connection.
type clientInitMessage struct {
	Token         string `json:"token"`         // JWT token for authentication
	VMName        string `json:"vmName"`        // name of the VM to connect to
	Namespace     string `json:"namespace"`     // namespace of the VM
	Environment   string `json:"environment"`   // environment of the VM
	InitialWidth  int    `json:"initialWidth"`  // initial width of the terminal
	InitialHeight int    `json:"initialHeight"` // initial height of the terminal
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

func (localCtx *LocalContext) logger() logr.Logger {
	return localCtx.log.WithValues(
		"username", localCtx.username,
		"ip", localCtx.ip,
		"namespace", localCtx.namespace,
		"environment", localCtx.environment,
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
		log:         webCtx.BaseLogger.WithValues(),
		username:    "",
		ip:          "",
		namespace:   "",
		environment: "",
		lastUsed:    atomic.Value{},
		ctxReq:      r.Context(),
		ctxServer:   nil,
		ws:          ws,
		session:     nil,
		errorMsg:    "",
		wg:          sync.WaitGroup{},
		cancelFun:   nil,
		stdin:       nil,
		stdout:      nil,
	}

	// check the number of connection
	n := atomic.LoadInt32(&webCtx.activeConnCount)
	if n >= webCtx.MaxConnectionCount {
		localCtx.logger().Error(err, "Max connection limit reached")
		localCtx.errorMsg = "Max connection limit reached"
		return
	}
	atomic.AddInt32(&webCtx.activeConnCount, 1)

	defer func() {
		atomic.AddInt32(&webCtx.activeConnCount, -1)

		if localCtx.errorMsg != "" {
			// send the error message to the client
			srvMsg := serverMessage{
				Type:  "error",
				Data:  "",
				Error: localCtx.errorMsg,
			}

			msgJSON, err := json.Marshal(srvMsg)
			if err != nil {
				localCtx.logger().Error(err, "Failed to marshal server error message")
			}

			if err := ws.WriteMessage(websocket.TextMessage, msgJSON); err != nil {
				localCtx.logger().Error(err, "WebSocket error on sending msg connection closure")
			}
		}
		if err := ws.Close(); err != nil {
			localCtx.logger().Error(err, "Failed to close WebSocket connection")
		}
	}()

	// wait for the first message to get the token
	_, firstMsg, err := ws.ReadMessage()
	if err != nil {
		localCtx.logger().Error(err, "ReadMessage error")
		localCtx.errorMsg = "Failed to read initialization message"
		return
	}

	var initMsg clientInitMessage
	if err := json.Unmarshal(firstMsg, &initMsg); err != nil {
		localCtx.logger().Error(err, "Invalid JSON format")
		localCtx.errorMsg = "Invalid fist message format"
		return
	}

	if initMsg.VMName == "" ||
		initMsg.Token == "" ||
		initMsg.Namespace == "" ||
		initMsg.InitialWidth == 0 ||
		initMsg.InitialHeight == 0 ||
		initMsg.Environment == "" {
		localCtx.logger().Error(errors.New("missing required fields in the initialization message"), "invalid initialization message")
		localCtx.errorMsg = "Missing required fields in the initialization message"
		return
	}

	localCtx.namespace = initMsg.Namespace
	localCtx.environment = initMsg.Environment

	// Validate the request
	err = webCtx.validateRequest(initMsg.VMName, initMsg.Token, localCtx)
	if err != nil {
		localCtx.logger().Error(err, "Request validation failed")
		localCtx.errorMsg = "Invalid request"
		return
	}

	// log the connection
	webSSHConnections.WithLabelValues(localCtx.ip, webCtx.VMSSHPort).Inc()

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
	localCtx.session = session

	defer func() {
		if err := session.Close(); err != nil {
			localCtx.logger().Error(err, "Failed to close SSH session")
		}
	}()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm-256color", initMsg.InitialHeight, initMsg.InitialWidth, modes); err != nil {
		localCtx.logger().Error(err, "Request for pseudo terminal failed")
		localCtx.errorMsg = "Internal server error"
		return
	}

	// Setup stdin and stdout pipes
	localCtx.stdin, err = localCtx.session.StdinPipe()
	if err != nil {
		localCtx.logger().Error(err, "Unable to setup stdin for session")
		localCtx.errorMsg = "Internal server error"
		return
	}

	localCtx.stdout, err = localCtx.session.StdoutPipe()
	if err != nil {
		localCtx.logger().Error(err, "Unable to setup stdout for session")
		localCtx.errorMsg = "Internal server error"
		return
	}

	// Start the shell
	if err := session.Shell(); err != nil {
		localCtx.logger().Error(err, "Failed to start shell")
		localCtx.errorMsg = "Internal server error"
		return
	}

	localCtx.logger().Info("SSH session started")

	localCtx.lastUsed.Store(time.Now())
	localCtx.ctxServer, localCtx.cancelFun = context.WithCancel(context.Background())

	// Start a goroutine to monitor for timeouts and send pings
	localCtx.wg.Add(1)
	go func() {
		webCtx.timeoutHandler(localCtx)
	}()

	// Start a goroutine to read from SSH stdout and write to WebSocket: from VM to user
	localCtx.wg.Add(1)
	go func() {
		webCtx.serverToClient(localCtx)
	}()

	// Read messages from WebSocket and write to SSH stdin: from user to VM
	localCtx.wg.Add(1)
	webCtx.clientToServer(localCtx)

	localCtx.wg.Wait()
}

func (webCtx *ServerContext) timeoutHandler(localCtx *LocalContext) {
	defer localCtx.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-localCtx.ctxServer.Done():
			localCtx.logger().Info("server context done, closing timeoutHandler")
			return
		case <-localCtx.ctxReq.Done():
			localCtx.logger().Info("request context done, closing timeoutHandler")
			return
		case <-ticker.C:
			// Check for timeout
			if webCtx.TimeoutDuration != 0 &&
				time.Since(localCtx.lastUsed.Load().(time.Time)) > webCtx.TimeoutDuration {
				localCtx.closeContexts(
					errors.New("connection timed out"),
					"Connection timed out due to inactivity",
					"Connection timed out due to inactivity",
				)
				return
			} else {
				// Send a ping message to the client to keep the connection alive
				srvMsg := serverMessage{
					Type:  "pong",
					Data:  "",
					Error: "",
				}

				msgJSON, err := json.Marshal(srvMsg)
				if err != nil {
					localCtx.closeContexts(err, "Failed to marshal server message - timeoutHandler", "Internal server error")
					return
				}

				if err := localCtx.ws.WriteMessage(websocket.PongMessage, msgJSON); err != nil {
					localCtx.closeContexts(err, "WebSocket write error - timeoutHandler", "Internal server error")
					return
				}
			}
		}
	}
}

func (localCtx *LocalContext) closeContexts(err error, logMsg, logForUser string) {
	localCtx.logger().Error(err, logMsg)
	localCtx.errorMsg = logForUser
	localCtx.cancelFun()
}

func (webCtx *ServerContext) serverToClient(localCtx *LocalContext) {
	defer localCtx.wg.Done()

	buf := make([]byte, 1024)
	for {
		done := make(chan struct{})
		var n int
		var err error

		go func() {
			n, err = localCtx.stdout.Read(buf)
			close(done)
		}()

		// Check if the contexts are done
		select {
		case <-done:
			// read completed
		case <-localCtx.ctxServer.Done():
			localCtx.logger().Info("server context done, closing serverToClient")
			return
		case <-localCtx.ctxReq.Done():
			localCtx.logger().Info("request context done, closing serverToClient")
			return
		}

		// parse the received data
		if err != nil && !errors.Is(err, io.EOF) {
			localCtx.closeContexts(err, "SSH stdout read error", "Internal server error")
			return
		}

		if n == 0 {
			localCtx.closeContexts(io.EOF, "SSH session closed", "SSH session closed")
			return
		}

		if webCtx.TimeoutDuration != 0 {
			// Update the last used timestamp
			localCtx.lastUsed.Store(time.Now())
		}

		srvMsg := serverMessage{
			Type:  "data",
			Data:  string(buf[:n]),
			Error: "",
		}

		msgJSON, err := json.Marshal(srvMsg)
		if err != nil {
			localCtx.closeContexts(err, "Failed to marshal server message", "Internal server error")
			return
		}

		if err := localCtx.ws.WriteMessage(websocket.TextMessage, msgJSON); err != nil {
			localCtx.closeContexts(err, "WebSocket write error", "Internal server error")
			return
		}
	}
}

func (webCtx *ServerContext) clientToServer(localCtx *LocalContext) {
	defer localCtx.wg.Done()

	for {
		done := make(chan struct{})
		var err error
		var msg []byte

		go func() {
			_, msg, err = localCtx.ws.ReadMessage()
			close(done)
		}()

		// Check if the contexts are done
		select {
		case <-localCtx.ctxServer.Done():
			localCtx.logger().Info("server context done, closing clientToServer")
			return
		case <-localCtx.ctxReq.Done():
			localCtx.logger().Info("request context done, closing clientToServer")
			return
		case <-done:
			// read completed
		}

		if err != nil {
			localCtx.closeContexts(err, "Closing SSH session, WebSocket read error", "Internal server error")
			return
		}

		// parse the received data
		var clientMsg clientMessage
		if err := json.Unmarshal(msg, &clientMsg); err != nil {
			localCtx.closeContexts(err, "Invalid message format", "Invalid message format")
			return
		}

		switch clientMsg.Type {
		case "resize":
			if err := localCtx.session.WindowChange(clientMsg.Rows, clientMsg.Cols); err != nil {
				localCtx.closeContexts(err, "Failed to resize terminal", "Internal server error")
				return
			}
		case "input":
			if _, err := localCtx.stdin.Write([]byte(clientMsg.Data)); err != nil {
				localCtx.closeContexts(err, "Failed to write to SSH stdin", "Internal server error")
				return
			}
		case "ping":
			continue
		default:
			localCtx.logger().Error(err, "Unknown message type")
			continue
		}

		if webCtx.TimeoutDuration != 0 {
			// Update the last used timestamp
			localCtx.lastUsed.Store(time.Now())
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
