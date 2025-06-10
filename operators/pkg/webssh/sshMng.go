package webssh

import (
	"errors"
	"log"
	"time"

	"golang.org/x/crypto/ssh"
)

type sshConnInfo struct {
	conn     *ssh.Client
	lastUsed time.Time
}

func updateSSHConnectionLastUsed(vmIP string) {
	connMutex.Lock()
	defer connMutex.Unlock()
	if info, exists := activeConnections[vmIP]; exists {
		info.lastUsed = time.Now()
	}
}

func closeSSHConnection(vmIP string) {
	var connToClose *ssh.Client

	// Remove the connection from activeConnections map
	connMutex.Lock()
	if info, exists := activeConnections[vmIP]; exists {
		connToClose = info.conn
		delete(activeConnections, vmIP) // Remove from active connections
		log.Printf("Closed SSH connection to %s", vmIP)
	} else {
		log.Printf("No active SSH connection found for %s", vmIP)
	}
	connMutex.Unlock()

	// Close the SSH connection outside the lock to to slow down contention
	if connToClose != nil {
		if err := connToClose.Close(); err != nil {
			log.Printf("failed to close SSH connection: %v", err)
		}
	}
}

func getOrCreateSSHConnection(VmIP string, sshConfig *ssh.ClientConfig, maxConnCount int) (*ssh.Client, error) {
	connMutex.Lock()
	defer connMutex.Unlock()

	// Check if the connection already exists
	if info, exists := activeConnections[VmIP]; exists {
		info.lastUsed = time.Now()
		log.Printf("Reusing existing SSH connection to %s", VmIP)
		return info.conn, nil
	}

	// check if the number of active connections exceeds the limit
	if len(activeConnections) >= maxConnCount {
		log.Printf("Maximum number of SSH connections (%d) reached. Cannot create new connection.", maxConnCount)
		return nil, errors.New("maximum number of SSH connections reached")
	}

	// If not, create a new SSH connection
	conn, err := ssh.Dial("tcp", VmIP, sshConfig)
	if err != nil {
		return nil, err
	}

	// Save the new connection in the map
	activeConnections[VmIP] = &sshConnInfo{
		conn:     conn,
		lastUsed: time.Now(),
	}

	log.Printf("New SSH connection established to %s", VmIP)

	return conn, nil
}

func startConnectionCleanup(interval, timeout time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			connMutex.Lock()
			log.Println("Starting connection cleanup...")
			now := time.Now()
			for ip, info := range activeConnections {
				if now.Sub(info.lastUsed) > timeout {
					err := info.conn.Close()
					if err != nil {
						log.Printf("Error closing SSH connection to %s: %v", ip, err)
					} else {
						log.Printf("Closed idle SSH connection to %s", ip)
					}
					delete(activeConnections, ip)
				}
			}
			connMutex.Unlock()
		}
	}()
}
