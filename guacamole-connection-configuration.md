# Guacamole Connection Configuration for CrownLabs

This document describes how Apache Guacamole can be configured to automatically create RDP connections for CrownLabs VM instances.

## Automatic Connection Configuration

To configure Guacamole to automatically create connections, we'll implement a REST API extension for Guacamole that communicates with the Kubernetes API to discover VM instances and create connections on-demand.

### 1. Create a Guacamole Extension

Create a Java extension for Guacamole that will:
1. Authenticate with the Kubernetes API
2. Watch for CrownLabs VM instances
3. Create appropriate RDP connections

Here's a sample implementation approach:

```java
// Sample code for CrownLabsGuacamoleExtension.java
package org.crownlabs.guacamole;

import org.apache.guacamole.GuacamoleException;
import org.apache.guacamole.net.auth.simple.SimpleUser;
import org.apache.guacamole.net.auth.Connection;
import org.apache.guacamole.net.auth.simple.SimpleConnection;
import org.apache.guacamole.protocol.GuacamoleConfiguration;
import org.apache.guacamole.token.TokenFilter;
import io.kubernetes.client.openapi.ApiClient;
import io.kubernetes.client.openapi.Configuration;
import io.kubernetes.client.openapi.apis.CoreV1Api;
import io.kubernetes.client.openapi.models.V1Pod;
import io.kubernetes.client.openapi.models.V1PodList;
import io.kubernetes.client.util.Config;

// This extension would implement Guacamole's extension API and register connection handlers
public class CrownLabsGuacamoleExtension implements AuthenticationProvider {
    
    private final Map<String, Connection> connections = new ConcurrentHashMap<>();
    
    @Override
    public void init() throws GuacamoleException {
        // Initialize Kubernetes client
        try {
            ApiClient client = Config.defaultClient();
            Configuration.setDefaultApiClient(client);
            
            // Start watching for CrownLabs VM instances
            startInstanceWatcher();
        } catch (Exception e) {
            throw new GuacamoleException("Failed to initialize Kubernetes client", e);
        }
    }
    
    private void startInstanceWatcher() {
        // Start a thread that watches for new VM instances
        new Thread(() -> {
            try {
                CoreV1Api api = new CoreV1Api();
                
                // Logic to watch for CrownLabs instances
                // When new instance is found, create a connection
                
            } catch (Exception e) {
                // Log error
            }
        }).start();
    }
    
    private void createConnectionFromInstance(String instanceId, String vmIp) {
        // Create an RDP connection configuration
        GuacamoleConfiguration config = new GuacamoleConfiguration();
        config.setProtocol("rdp");
        config.setParameter("hostname", vmIp);
        config.setParameter("port", "3389");
        config.setParameter("username", "user"); // Default username
        config.setParameter("password", ""); // No password required, as we disabled authentication in xRDP
        config.setParameter("ignore-cert", "true");
        config.setParameter("security", "nla");
        config.setParameter("resize-method", "display-update");
        config.setParameter("enable-drive", "true");
        config.setParameter("create-drive-path", "true");
        
        // Create a connection with this configuration
        SimpleConnection connection = new SimpleConnection(
            instanceId, instanceId, config);
        
        // Store the connection
        connections.put(instanceId, connection);
    }
    
    // Other required authentication provider methods
    // ...
}
```

### 2. Use Guacamole REST API

Alternatively, you can use the Guacamole REST API to create connections from the CrownLabs instance controller:

```go
// Sample Go code for instance controller that creates Guacamole connections

// createGuacamoleConnection creates a connection in Guacamole for a VM
func (r *InstanceReconciler) createGuacamoleConnection(instance *clv1alpha2.Instance, vmIP string) error {
    // Get authentication token from Guacamole
    token, err := r.getGuacamoleAuthToken()
    if err != nil {
        return err
    }
    
    // Create connection in Guacamole
    connectionData := map[string]interface{}{
        "name":              fmt.Sprintf("VM-%s", instance.UID),
        "parentIdentifier":  "ROOT",
        "protocol":          "rdp",
        "parameters": map[string]string{
            "hostname":       vmIP,
            "port":           "3389",
            "ignore-cert":    "true",
            "security":       "nla",
            "resize-method":  "display-update",
            "enable-drive":   "true",
        },
        "attributes": map[string]string{
            "guac-failover-only": "false",
        },
    }
    
    // Call Guacamole API to create the connection
    client := &http.Client{}
    jsonData, _ := json.Marshal(connectionData)
    req, _ := http.NewRequest("POST", 
        fmt.Sprintf("%s/api/session/data/kubernetes/connections", r.ServiceUrls.GuacamoleURL),
        bytes.NewBuffer(jsonData))
    
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)
    
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    // Process response
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to create Guacamole connection: %s", resp.Status)
    }
    
    return nil
}
```

## 3. Connection Parameters

Common RDP connection parameters for Guacamole:

| Parameter | Value | Description |
|-----------|-------|-------------|
| hostname | VM IP | VM's internal IP address |
| port | 3389 | Standard RDP port |
| username | user | Username for RDP connection |
| password | | Empty for passwordless login |
| security | nla | Network Level Authentication |
| ignore-cert | true | Ignore certificate validation |
| enable-audio | true | Enable audio |
| enable-printing | true | Enable printing |
| enable-drive | true | Enable drive redirection |
| drive-path | /tmp | Path for drive redirection |
| resize-method | display-update | Resize method for RDP |
| color-depth | 16 | Color depth (8, 16, 24, 32) |

## 4. Dynamic Connection Generation

For a scalable solution, connection configurations should be generated dynamically:

1. When a VM instance is created, a connection ID matching the instance UID is created in Guacamole
2. When a user accesses `/guacamole/#/client/c/<instance-uid>`, Guacamole will use this connection
3. When a VM instance is deleted, the corresponding connection should be removed from Guacamole

## 5. Authentication Integration

Guacamole can be configured to use the same OIDC provider as CrownLabs for single sign-on:

1. Configure Guacamole to use OIDC (OpenID Connect) extension
2. Point it to the same identity provider used by CrownLabs
3. Set up proper RBAC within Guacamole to restrict access to connections

## 6. Additional Features

Guacamole provides additional features that can be leveraged:

1. **Session Recording**: Record user sessions for auditing
2. **Clipboard Sharing**: Enable clipboard integration between client and VM
3. **File Transfer**: Allow file uploads/downloads
4. **Multi-user Support**: Support for screen sharing between users
5. **Connection Groups**: Organize connections by workspace/tenant 