// Copyright 2020-2024 Politecnico di Torino
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

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	headerXForwardedFor = "X-Forwarded-For"
	searchStringHead    = "<head>"
	hideNovncBarStyle   = "<style>#noVNC_control_bar_anchor {display:none !important;}</style>\n"
	websockifyPath      = "/websockify"
	usagesPath          = "/usages"
)

var (
	searchStringHeadBytes = []byte(searchStringHead)
)

// ConnInfo stores useful information related to the WS connection.
type ConnInfo struct {
	// latency in ms
	Latency     int64     `json:"latency"`
	IP          string    `json:"ip,omitempty"`
	UID         string    `json:"connUid,omitempty"`
	ConnTime    time.Time `json:"connTime,omitempty"`
	DisconnTime time.Time `json:"disconnTime,omitempty"`
	Active      bool      `json:"active"`
}

// NoVncHandler is the main handler for the noVNC server.
type NoVncHandler struct {
	// <uid, ConnInfo>
	connectionsTracking *sync.Map
	BasePath            string
	PingInterval        time.Duration
	NoVncFS             http.Handler
	ShowNoVncBar        bool
	TargetSocket        string
	MetricsHandler      *InstanceMetricsHandler
}

// ServeHTTP handles the HTTP request.
func (h *NoVncHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, h.BasePath) {
		http.NotFound(w, r)
		log.Printf("not found: %s", r.URL.Path)
		return
	}

	cleanedUpPath := strings.TrimPrefix(r.URL.Path, h.BasePath)

	// serve index on root
	switch cleanedUpPath {
	case "": // enforce slash terminated path.
		http.Redirect(w, r, r.URL.Path+"/", http.StatusFound)
	case "/": // serve vnc index on root.
		h.serveNoVncHome(w, r)
	case websockifyPath:
		// connUid is required for websockify connections.
		if connUID, ok := r.URL.Query()["connUid"]; ok {
			if _, ok := h.connectionsTracking.Load(connUID[0]); !ok {
				http.Error(w, "received connUid queryParam was never assigned", http.StatusBadRequest)
			} else {
				log.Println("Connection UID: ", connUID[0])
				h.serveWs(w, r)
			}
		} else {
			http.Error(w, "connUid queryParam required", http.StatusBadRequest)
		}
	case usagesPath:
		h.MetricsHandler.serveWs(w, r)
	default:
		r.URL.Path = "novnc" + strings.Replace(r.URL.Path, h.BasePath, "/", 1)
		h.NoVncFS.ServeHTTP(w, r)
	}
}

func (h *NoVncHandler) serveNoVncHome(w http.ResponseWriter, r *http.Request) {
	data, err := novncFS.ReadFile("novnc/vnc.html")
	if err != nil {
		panic(err)
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	ip := r.Header.Get(headerXForwardedFor)
	if ip == "" {
		ip = "unknown"
	}

	log.Printf("Requested novnc gui from IP=%s", ip)

	injectStr := "<head>\n"

	if !h.ShowNoVncBar {
		injectStr += hideNovncBarStyle
	}

	vncEndpoint := strings.TrimPrefix(h.BasePath+websockifyPath, "/")
	usagesEndpoint := strings.TrimPrefix(h.BasePath+usagesPath, "/")
	uid := strings.ReplaceAll(uuid.New().String(), "-", "")
	injectStr += fmt.Sprintf(`<script>window.websockifyTargetUrl='%s?connUid=%s';
	window.metricsTargetUrl='%s?connUid=%s';</script>`, vncEndpoint, uid, usagesEndpoint, uid)

	connectionInfo := &ConnInfo{IP: ip, UID: uid, Latency: 0, ConnTime: time.Now(), DisconnTime: time.Now(), Active: false}
	h.connectionsTracking.Store(connectionInfo.UID, *connectionInfo)

	data = bytes.ReplaceAll(data, searchStringHeadBytes, []byte(injectStr))

	_, err = w.Write(data)
	if err != nil {
		log.Println("index write error:", err)
	}
}
