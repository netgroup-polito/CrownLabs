package webssh

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

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
	}*/

	// connect to the k8s API server to verify the token

	// TODO: Implement actual token verification logic
	return true // Replace with actual token verification logic
}

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

	if !verifyToken(token) {
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

func SetupWebSSH() {
	// automatic Cleanup
	startConnectionCleanup(1*time.Minute, timeoutDuration)

	http.HandleFunc("/ws", wsHandler)
	log.Println("WebSocket SSH bridge running on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
