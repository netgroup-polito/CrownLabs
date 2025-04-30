# Implementation Plan: Replacing VNC with RDP/Guacamole in CrownLabs

## Overview
This implementation plan outlines the steps needed to replace the current VNC-based remote access solution with RDP (Remote Desktop Protocol) using Apache Guacamole in CrownLabs.

## 1. Guacamole Server Deployment
- Deploy Apache Guacamole server as a container in the Kubernetes cluster
- Configure Guacamole to handle RDP connections to VMs
- Set up proper authentication integration with the existing OAuth2 system
- Create Kubernetes resources (Deployment, Service, ConfigMap) for Guacamole

## 2. VM Template Modifications
- Update the provisioning scripts in `provisioning/virtual-machines/ansible/roles/crownlabs/tasks/main.yml`
- Replace TigerVNC installation with xRDP server installation
- Create systemd service for xRDP that starts automatically
- Remove noVNC and websockify components which are no longer needed

## 3. Instance Controller Updates
- Modify `operators/pkg/instctrl/exposition.go` to route connections through Guacamole
- Update `operators/pkg/forge/ingresses.go` to create routes for Guacamole instead of direct VNC connections
- Update connection probe logic in `operators/pkg/forge/virtualmachines.go` to check RDP port instead of VNC port

## 4. Authentication Integration
- Configure Guacamole to use the existing CrownLabs authentication system
- Update ingress authentication rules for Guacamole

## 5. Frontend Updates
- No significant updates needed for the frontend if Guacamole web interface is used directly
- If customization is desired, update frontend components to integrate with Guacamole API

## 6. Testing Plan
- Test RDP connection performance vs. VNC
- Verify authentication works correctly
- Validate both VM and container environment access
- Test screen sharing functionality

## 7. Detailed Technical Changes

### 7.1 VM Provisioning Updates
```yaml
# Replace current VNC installation with xRDP in ansible playbook
- name: Set the xRDP variables
  set_fact:
    xrdp_version: "0.9.15"
    xrdp_service_name: "xrdp.service"
    xrdp_service_path: "/etc/systemd/system"

- name: Install xRDP
  apt:
    name: 
      - xrdp
      - xorgxrdp
    state: present
    update_cache: yes

- name: Configure xRDP
  copy:
    src: templates/xrdp.ini
    dest: /etc/xrdp/xrdp.ini
    owner: root
    group: root
    mode: '0644'

- name: Enable the xRDP service
  systemd:
    name: "{{ xrdp_service_name }}"
    enabled: yes
    state: started
```

### 7.2 Guacamole Deployment
```yaml
# Example Kubernetes deployment for Guacamole
apiVersion: apps/v1
kind: Deployment
metadata:
  name: guacamole
  namespace: crownlabs-system
spec:
  replicas: 2
  selector:
    matchLabels:
      app: guacamole
  template:
    metadata:
      labels:
        app: guacamole
    spec:
      containers:
      - name: guacd
        image: guacamole/guacd:1.4.0
        ports:
        - containerPort: 4822
      - name: guacamole
        image: guacamole/guacamole:1.4.0
        ports:
        - containerPort: 8080
        env:
        - name: GUACD_HOSTNAME
          value: localhost
        - name: GUACD_PORT
          value: "4822"
---
apiVersion: v1
kind: Service
metadata:
  name: guacamole
  namespace: crownlabs-system
spec:
  selector:
    app: guacamole
  ports:
  - port: 8080
    targetPort: 8080
    name: http
  - port: 4822
    targetPort: 4822
    name: guacd
```

### 7.3 Instance Controller Updates
Modifications to `operators/pkg/forge/ingresses.go`:
```go
const (
    // Replace VNC-specific constants with RDP/Guacamole
    IngressRDPGUIPathSuffix = "gui"
    
    // Guacamole endpoint
    GuacamoleEndpoint = "/guacamole"
)

// Update IngressGUIPath for RDP
func IngressGUIPath(instance *clv1alpha2.Instance, environment *clv1alpha2.Environment) string {
    switch environment.EnvironmentType {
    case clv1alpha2.ClassCloudVM, clv1alpha2.ClassVM:
        return strings.TrimRight(fmt.Sprintf("%v/%v/%v", IngressInstancePrefix, instance.UID, IngressRDPGUIPathSuffix), "/")
    // Other cases remain the same
    }
    return ""
}
```

### 7.4 Instance Reconciler Updates
Modifications to expose RDP through Guacamole in `operators/pkg/instctrl/exposition.go`.

## 8. Implementation Steps
1. Begin with setting up and testing Guacamole in a development environment
2. Develop and test xRDP VM image
3. Update operators to route connections through Guacamole
4. Test with sample VMs
5. Roll out to production after thorough testing

## 9. Benefits
- Better performance: RDP generally offers better performance than VNC
- Enhanced user experience: RDP supports features like audio, clipboard, and drive mapping
- Improved security: RDP connections can be more securely tunneled through Guacamole 