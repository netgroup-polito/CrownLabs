package webssh

import (
	"encoding/json"
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

type ClientInitMessage struct {
	Token  string `json:"token"`
	VmIp   string `json:"vmIp"`
	VmPort string `json:"vmPort"`
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

/*
// --- JWT extraction ---
func extractUsernameFromToken(tokenString string) (string, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Adatta il claim secondo la configurazione Keycloak (es: "preferred_username" o "sub")
		if username, ok := claims["preferred_username"].(string); ok {
			return username, nil
		}
		if sub, ok := claims["sub"].(string); ok {
			return sub, nil
		}
	}
	return "", fmt.Errorf("username not found in token")
}

// --- RBAC check ---
func canUserAccessResource(token, namespace, instanceName string) (bool, error) {
	token = strings.TrimPrefix(token, "Bearer ")
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT")
	if host == "" || port == "" {
		return false, errors.New("KUBERNETES_SERVICE_HOST or PORT not set")
	}
	apiServer := "https://" + host + ":" + port

	config := &rest.Config{
		Host:        apiServer,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			CAFile: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
		},
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return false, err
	}

	ssar := &authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Namespace: namespace,
				Verb:      "get",
				Group:     "crownlabs.polito.it",
				Resource:  "instances",
				Name:      instanceName,
			},
		},
	}

	result, err := clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(context.Background(), ssar, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}
	return result.Status.Allowed, nil
}

// --- Instance owner check ---
func getInstanceOwner(token, namespace, instanceName string) (string, error) {
	token = strings.TrimPrefix(token, "Bearer ")
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT")
	if host == "" || port == "" {
		return "", errors.New("KUBERNETES_SERVICE_HOST or PORT not set")
	}
	apiServer := "https://" + host + ":" + port

	config := &rest.Config{
		Host:        apiServer,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			CAFile: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
		},
	}
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return "", err
	}

	gvr := schema.GroupVersionResource{
		Group:    "crownlabs.polito.it",
		Version:  "v1alpha2",
		Resource: "instances",
	}
	inst, err := dynClient.Resource(gvr).Namespace(namespace).Get(context.Background(), instanceName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	labels := inst.GetLabels()
	owner, ok := labels["crownlabs.polito.it/owner"]
	if !ok {
		return "", fmt.Errorf("owner label not found")
	}
	return owner, nil
}

// --- Combined check ---
func verifyTokenAndOwnership(token, namespace, instanceName string) bool {
	// 1. RBAC check
	allowed, err := canUserAccessResource(token, namespace, instanceName)
	if err != nil {
		log.Println("K8s permission check error:", err)
		return false
	}
	if !allowed {
		log.Println("RBAC denied access")
		return false
	}

	// 2. Owner check
	username, err := extractUsernameFromToken(token)
	if err != nil {
		log.Println("Token decode error:", err)
		return false
	}
	owner, err := getInstanceOwner(token, namespace, instanceName)
	if err != nil {
		log.Println("Error retrieving instance owner:", err)
		return false
	}
	if username != owner {
		log.Printf("User %s is not the owner (%s) of instance %s\n", username, owner, instanceName)
		return false
	}
	return true
}

/*
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
	}

	// connect to the k8s API server to verify the token

	// TODO: Implement actual token verification logic
	return true // Replace with actual token verification logic
}
*/

func validateRequest(firstMsg []byte) (string, error) {
	var initMsg ClientInitMessage
	if err := json.Unmarshal(firstMsg, &initMsg); err != nil {
		log.Println("Invalid JSON from client:", err)
		return "", errors.New("invalid JSON format")
	}

	if initMsg.VmIp == "" || initMsg.VmPort == "" || initMsg.Token == "" {
		log.Println("Missing required fields in the initialization message")
		return "", errors.New("missing required fields in the initialization message")
	}

	if !verifyToken(initMsg.Token) {
		log.Println("Invalid token provided")
		return "", errors.New("invalid token")
	}

	vmIp := initMsg.VmIp + ":" + initMsg.VmPort
	log.Printf("Valid request received for VM IP: %s, token: %s", vmIp, initMsg.Token)

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
