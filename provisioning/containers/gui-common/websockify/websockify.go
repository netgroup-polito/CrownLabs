// Copyright 2020-2023 Politecnico di Torino
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

package main

// Websockify - go version
// Original version at https://github.com/novnc/websockify-other/tree/master/golang

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func forwardtcp(wsconn *websocket.Conn, conn net.Conn) {
	var tcpbuffer [1024]byte
	defer wsconn.Close()
	defer conn.Close()
	for {
		n, err := conn.Read(tcpbuffer[0:])
		if err != nil {
			log.Println("tcp read error:", err)
			return
		}

		if err := wsconn.WriteMessage(websocket.BinaryMessage, tcpbuffer[0:n]); err != nil {
			log.Println("ws write error:", err)
			return
		}
	}
}

func forwardweb(wsconn *websocket.Conn, conn net.Conn) {
	defer wsconn.Close()
	defer conn.Close()
	for {
		_, buffer, err := wsconn.ReadMessage()
		if err != nil {
			log.Println("ws read error:", err)
			return
		}

		if _, err := conn.Write(buffer); err != nil {
			log.Println("tcp write error: ", err)
			return
		}
	}
}

func (h *NoVncHandler) serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	vnc, err := net.Dial("tcp", h.TargetSocket)
	if err != nil {
		log.Println("vnc dial:", err)
		_ = ws.Close()
		return
	}

	ip := r.Header.Get(headerXForwardedFor)
	if ip == "" {
		ip = "unknown"
	}
	connUID := r.URL.Query().Get("connUid")

	if ci, ok := h.connectionsTracking.Load(connUID); ok {
		connectionInfo := ci.(ConnInfo)
		metric := makeLatencyObserver(ip, connUID)
		log.Printf("Incoming websocket connection on path /websockify from IP=%s", ip)
		go forwardtcp(ws, vnc)
		go forwardweb(ws, vnc)
		go h.pingCycle(ws, metric, &connectionInfo)
	} else {
		log.Println("received connUid queryParam was never assigned")
		_ = ws.Close()
		return
	}
}

func (h *NoVncHandler) pingCycle(wsconn *websocket.Conn, metric prometheus.Observer, connInfo *ConnInfo) {
	// set pong handler
	lastPing := time.Now()
	wsconn.SetPongHandler(func(data string) error {
		connInfo.Latency = time.Since(lastPing).Milliseconds()
		connInfo.Active = true
		metric.Observe(float64(connInfo.Latency))
		h.connectionsTracking.Store(connInfo.UID, *connInfo)
		log.Printf("ping latency: %dms, connUid: %s", connInfo.Latency, connInfo.UID)
		return nil
	})

	ticker := time.NewTicker(h.PingInterval)
	for lastPing = range ticker.C {
		err := wsconn.WriteControl(websocket.PingMessage, nil, time.Now().Add(h.PingInterval))
		if err != nil {
			log.Println("ping error:", err)
			log.Printf("stopping connection tracking for <IP %s; UID %s>", connInfo.IP, connInfo.UID)

			connInfo.Latency = 0
			connInfo.Active = false
			connInfo.DisconnTime = time.Now()
			h.connectionsTracking.Store(connInfo.UID, *connInfo)
			ticker.Stop()
		}
	}
}
