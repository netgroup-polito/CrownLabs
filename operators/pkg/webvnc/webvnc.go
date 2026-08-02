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
// It allows users to connect to their VM desktop, exposed by KubeVirt's native
// vnc subresource, using a web-based noVNC client.
package webvnc

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/rest"
)

// ServerContext holds the context for the WebVNC server.
type ServerContext struct {
	MaxConnectionCount int32        // MaxConnectionCount is the maximum number of concurrent VNC connections allowed
	WebsocketPort      string       // WebsocketPort is the port on which the WebSocket server listens
	BaseConfig         *rest.Config // config base with all the standard Kubernetes API settings
	activeConnCount    int32        // active connection count
	BaseLogger         logr.Logger  // logger for the base context
}

// LocalContext holds the context for a local WebVNC connection.
type LocalContext struct {
	log         logr.Logger // logger for the local context
	username    string      // username for the connection
	namespace   string      // namespace of the instance
	environment string      // environment of the instance
}

func (localCtx *LocalContext) logger() logr.Logger {
	return localCtx.log.WithValues(
		"username", localCtx.username,
		"namespace", localCtx.namespace,
		"environment", localCtx.environment,
	)
}

// clientInitMessage is the first message a client sends over the WebSocket,
// before any RFB protocol byte. Browsers cannot set custom headers on a
// WebSocket upgrade request, so the JWT travels as ordinary application
// data instead.
type clientInitMessage struct {
	Token        string `json:"token"`
	InstanceName string `json:"instanceName"`
	Namespace    string `json:"namespace"`
	Environment  string `json:"environment"`
}

var webVNCConnections = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "webvnc_connections",
		Help: "VNC connections established through the WebSocket bridge",
	},
	[]string{"namespace", "environment"},
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

func (webCtx *ServerContext) wsHandler(w http.ResponseWriter, r *http.Request) {
	browserConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		webCtx.BaseLogger.Error(err, "WebSocket upgrade failed")
		return
	}
	defer browserConn.Close()

	n := atomic.AddInt32(&webCtx.activeConnCount, 1)
	defer atomic.AddInt32(&webCtx.activeConnCount, -1)
	if n > webCtx.MaxConnectionCount {
		webCtx.BaseLogger.Info("Max connection limit reached")
		return
	}

	localCtx := &LocalContext{log: webCtx.BaseLogger}

	_, firstMsg, err := browserConn.ReadMessage()
	if err != nil {
		localCtx.logger().Error(err, "Failed to read initialization message")
		return
	}

	var initMsg clientInitMessage
	if err := json.Unmarshal(firstMsg, &initMsg); err != nil {
		localCtx.logger().Error(err, "Invalid initialization message format")
		return
	}
	if initMsg.Token == "" || initMsg.InstanceName == "" || initMsg.Namespace == "" || initMsg.Environment == "" {
		localCtx.logger().Info("Missing required fields in the initialization message")
		return
	}
	localCtx.namespace = initMsg.Namespace
	localCtx.environment = initMsg.Environment

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	vmi, err := webCtx.validateRequest(ctx, initMsg.InstanceName, initMsg.Token, localCtx)
	if err != nil {
		localCtx.logger().Error(err, "Request validation failed")
		return
	}

	vncConn, err := openVNCStream(ctx, webCtx.BaseConfig, vmi)
	if err != nil {
		localCtx.logger().Error(err, "Failed to open the vnc subresource")
		return
	}
	defer vncConn.Close()

	webVNCConnections.WithLabelValues(localCtx.namespace, localCtx.environment).Inc()
	localCtx.logger().Info("VNC session started, conn number: " + strconv.Itoa(int(n)))

	done := make(chan struct{})
	go func() {
		defer close(done)
		relay(vncConn, browserConn)
	}()
	relay(browserConn, vncConn)
	<-done
}

// relay copies WebSocket messages from src to dst until either side closes
// or errors. Bytes are forwarded as-is: the RFB protocol they carry is never
// interpreted here.
func relay(dst, src *websocket.Conn) {
	for {
		msgType, data, err := src.ReadMessage()
		if err != nil {
			return
		}
		if err := dst.WriteMessage(msgType, data); err != nil {
			return
		}
	}
}

func probeHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// StartWebVNC initializes the WebSocket VNC bridge server.
func (webCtx *ServerContext) StartWebVNC() {
	mux := http.NewServeMux()
	mux.HandleFunc("/webvnc", webCtx.wsHandler)

	mux.HandleFunc("/healthz", probeHandler)
	mux.HandleFunc("/ready", probeHandler)

	prometheus.MustRegister(webVNCConnections)
	mux.Handle("/metrics", promhttp.Handler())

	webCtx.BaseLogger.Info("WebVNC server started on port: " + webCtx.WebsocketPort)

	server := &http.Server{
		Addr:         ":" + webCtx.WebsocketPort,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 0, // VNC connections can be long-lived
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		webCtx.BaseLogger.Error(err, "HTTP server failed")
	}
}
