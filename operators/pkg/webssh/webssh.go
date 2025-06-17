package webssh

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

const (
	sshUser            = "crownlabs"
	privateKeyPath     = "/web-keys/ssh_web_bastion_master"
	timeoutDuration    = 30 * time.Minute // Duration after which an SSH connection is considered idle and closed
	maxIdleConnections = 1000             // Maximum number of idle SSH connections to keep open
	vmPort             = "22"             // Default SSH port for VMs can be overridden in the ClientInitMessage
	ServerPort         = "8080"           // Port on which the WebSocket server listens
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

type ClientInitMessage struct {
	Token  string `json:"token"`
	VmIp   string `json:"vmIp"`
	VmPort string `json:"vmPort"` // Optional, can be used to specify a different port
}

func extractUsernameFromToken(tokenString string) (string, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {

		if username, ok := claims["preferred_username"].(string); ok {
			return username, nil
		}
	}
	return "", errors.New("username not found in token claims")
}

func getInstances(token string, namespace string) (map[string]any, error) {
	// create the Kubernetes client configuration
	config := &rest.Config{
		Host:        "https://apiserver.crownlabs.polito.it",
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: false, // TLS ??
		},
	}

	// Create a dynamic client to interact with custom resources
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, errors.New("failed to create dynamic client: " + err.Error())
	}

	gvr := schema.GroupVersionResource{
		Group:    "crownlabs.polito.it",
		Version:  "v1alpha2",
		Resource: "instances",
	}

	list, err := dynamicClient.Resource(gvr).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.New("failed to list instances: " + err.Error())
	}

	instances := make(map[string]any)
	for _, item := range list.Items {
		instanceName := item.GetName()
		instances[instanceName] = item.Object // Store the entire object or specific fields as needed
	}

	return instances, nil
}

func validateRequest(firstMsg []byte) (string, error) {

	var initMsg ClientInitMessage
	if err := json.Unmarshal(firstMsg, &initMsg); err != nil {
		log.Println("Invalid JSON from client:", err)
		return "", errors.New("invalid JSON format")
	}

	if initMsg.VmIp == "" || initMsg.Token == "" {
		log.Println("Missing required fields in the initialization message")
		return "", errors.New("missing required fields in the initialization message")
	}

	// Extract username from the token
	username, err := extractUsernameFromToken(initMsg.Token)
	if err != nil {
		log.Println("Token decode error:", err)
		return "", errors.New("invalid token format")
	}

	namespace := "tenant-" + username

	log.Printf("Validating request for user: %s, namespace: %s", username, namespace)

	// Get the list of instances for the user
	instances, err := getInstances(initMsg.Token, namespace)
	if err != nil {
		log.Println("Error retrieving instances:", err)
		return "", errors.New("error retrieving instances")
	}

	// check if the requested VM IP is in the list of instances
	found := false
	for _, i := range instances {
		if instanceMap, ok := i.(map[string]any); ok {
			if status, ok := instanceMap["status"].(map[string]any); ok {
				if ip, ok := status["ip"].(string); ok {
					log.Println("IP: ", ip)
					if ip == initMsg.VmIp {
						found = true
						log.Printf("Found instance with IP: %s in namespace: %s", ip, namespace)
						break // Exit the loop if we found the instance
					}
				}
			}
		}
	}

	if !found {
		log.Printf("No instance found for VM IP: %s in namespace: %s", initMsg.VmIp, namespace)
		return "", errors.New("no instance found for the provided VM IP")
	}

	// ssh port used to connect to the VM
	port := initMsg.VmPort
	if port == "" {
		port = vmPort // Default port if not specified
	}

	vmIp := initMsg.VmIp + ":" + port // ip:port

	return vmIp, nil
}

func loadPrivateKey() (ssh.Signer, error) {
	key, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Printf("Error reading private key file at %s: %v", privateKeyPath, err)
		return nil, err
	}
	return ssh.ParsePrivateKey(key)
}

func updateSSHConnectionLastUsed(vmIp string) {
	connMutex.Lock()
	defer connMutex.Unlock()
	if info, exists := activeConnections[vmIp]; exists {
		info.lastUsed = time.Now()
	}
}

func closeSSHConnection(vmIp string) {
	var connToClose *ssh.Client

	// Remove the connection from activeConnections map
	connMutex.Lock()
	if info, exists := activeConnections[vmIp]; exists {
		connToClose = info.conn
		delete(activeConnections, vmIp) // Remove from active connections
		log.Printf("Closed SSH connection to %s", vmIp)
	} else {
		log.Printf("No active SSH connection found for %s", vmIp)
	}
	connMutex.Unlock()

	// Close the SSH connection outside the lock to to slow down contention
	if connToClose != nil {
		connToClose.Close() // Close the SSH connection outside the lock
	}
}

func getOrCreateSSHConnection(vmIp string, sshConfig *ssh.ClientConfig) (*ssh.Client, error) {
	connMutex.Lock()
	defer connMutex.Unlock()

	// Check if the connection already exists
	if info, exists := activeConnections[vmIp]; exists {
		info.lastUsed = time.Now()
		log.Printf("Reusing existing SSH connection to %s", vmIp)
		return info.conn, nil
	}

	// check if the number of active connections exceeds the limit
	if len(activeConnections) >= maxIdleConnections {
		log.Printf("Maximum number of idle SSH connections (%d) reached. Cannot create new connection.", maxIdleConnections)
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

	log.Printf("New SSH connection established to %s", vmIp)

	return conn, nil
}

func startConnectionCleanup(interval time.Duration, timeout time.Duration) {
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
func wsHandler(w http.ResponseWriter, r *http.Request) {

	// upgrade to the WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}

	defer func() {
		ws.Close()
	}()

	// Validate the req
	// wait for the first message to get the token
	_, firstMsg, err := ws.ReadMessage()
	if err != nil {
		log.Println("ReadMessage error:", err)
		return
	}

	sshStr, err := validateRequest(firstMsg)
	if err != nil {
		log.Println("Request validation failed:", err)
		ws.WriteMessage(websocket.TextMessage, []byte("Invalid request"))
		ws.Close()
		return
	}

	// Load the private key for SSH authentication
	signer, err := loadPrivateKey()
	if err != nil {
		log.Println("Failed to load private key:", err)
		ws.WriteMessage(websocket.TextMessage, []byte("Failed to load private key"))
		ws.Close()
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

	sshConn, err := getOrCreateSSHConnection(sshStr, sshConfig)
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
			updateSSHConnectionLastUsed(sshStr) // reset timer on shell output
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
			closeSSHConnection(sshStr)
			break
		}
		updateSSHConnectionLastUsed(sshStr) // reset timer on user input
		if _, err := stdin.Write(msg); err != nil {
			log.Println("SSH stdin write error:", err)
			break
		}
	}
}

func StartWebSSH() {
	// automatic Cleanup
	startConnectionCleanup(1*time.Minute, timeoutDuration)

	http.HandleFunc("/ws", wsHandler)
	log.Println("WebSocket SSH bridge running on :" + ServerPort)
	err := http.ListenAndServe(":"+ServerPort, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
