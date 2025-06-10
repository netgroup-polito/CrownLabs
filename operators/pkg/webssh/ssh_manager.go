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
// A piece of the webssh architecture that is used to perform and manage SSH connections
// to VMs, including connection reuse and cleanup of idle connections.
package webssh

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

var (
	activeConnections = make(map[connectionKey]*sshConnInfo)
	connMutex         sync.Mutex // Mutex to protect access to activeConnections
)

// activeConnections holds the currently active SSH connections.
type sshConnInfo struct {
	conn     *ssh.Client
	lastUsed time.Time
}

// key for the activeConnections map.
// Different users can have connections to the same VM.
type connectionKey struct {
	vmIP     string
	username string
}

func updateSSHConnectionLastUsed(connKey connectionKey) {
	connMutex.Lock()
	defer connMutex.Unlock()
	if info, exists := activeConnections[connKey]; exists {
		info.lastUsed = time.Now()
	}
}

func closeSSHConnection(connKey connectionKey) error {
	var connToClose *ssh.Client

	// Remove the connection from activeConnections map
	connMutex.Lock()
	if info, exists := activeConnections[connKey]; exists {
		connToClose = info.conn
		delete(activeConnections, connKey) // Remove from active connections
	}
	connMutex.Unlock()

	// Close the SSH connection outside the lock to to slow down contention
	if connToClose != nil {
		if err := connToClose.Close(); err != nil {
			return errors.New("failed to close SSH connection: " + err.Error())
		}
	} else {
		return errors.New("no active SSH connection to close")
	}

	return nil // no error, connection closed successfully
}

func getOrCreateSSHConnection(connKey connectionKey, port string, sshConfig *ssh.ClientConfig, maxConnCount int) (*ssh.Client, error) {
	connMutex.Lock()
	defer connMutex.Unlock()

	// Check if the connection already exists
	if conn, exists := activeConnections[connKey]; exists {
		conn.lastUsed = time.Now()
		log.Println("Reusing existing SSH connection to " + connKey.vmIP + " for user " + connKey.username)
		return conn.conn, nil
	}

	// check if the number of active connections exceeds the limit
	if len(activeConnections) >= maxConnCount {
		err := errors.New("maximum number of SSH connections reached (" + fmt.Sprintf("%d", maxConnCount) + ")")
		return nil, err
	}

	// If not, create a new SSH connection
	connString := connKey.vmIP + ":" + port

	conn, err := ssh.Dial("tcp", connString, sshConfig)
	if err != nil {
		return nil, err
	}

	// Save the new connection in the map
	activeConnections[connKey] = &sshConnInfo{
		conn:     conn,
		lastUsed: time.Now(),
	}

	log.Println("Established new SSH connection to " + connKey.vmIP + " for user " + connKey.username)
	return conn, nil
}

func startConnectionCleanup(interval, timeout time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			connMutex.Lock()
			log.Println("Starting connection cleanup...")
			now := time.Now()
			for key, conn := range activeConnections {
				if now.Sub(conn.lastUsed) > timeout {
					// remove the connection from the map and close it
					delete(activeConnections, key)
					if err := conn.conn.Close(); err != nil {
						log.Println("Error closing SSH connection to " + key.vmIP + ": " + err.Error())
					} else {
						log.Println("Closed idle SSH connection to " + key.vmIP)
					}
				}
			}
			connMutex.Unlock()
		}
	}()
}
