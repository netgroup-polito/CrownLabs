# Cloud Image Registry

## Table of Contents
- [What is it?](#what-is-it)
- [Why do we need it?](#why-do-we-need-it)
- [How to use it](#how-to-use-it)

---

## What is it?

**Cloud Image Registry** is an image storage solution designed to store and serve VM images within a Kubernetes cluster. It consists of:
- A custom **Golang server** running as a Kubernetes **Deployment**.
- A **Persistent Volume Claim (PVC)** that acts as the storage backend for the images.

The server provides a set of HTTP endpoints to **upload, download, and manage images and their metadata**.

---

## Why do we need it?

Presently, the VM images are stored in container disks in the **[Harbor registry](../docker-registry/README.md)**. Storing and using VM images directly would make the process of VM instantiation more efficient. This is accomplished by introducing the **Cloud Image Registry**. 

The new component also stores metadata about every image, which is necessary for future CrownLabs features that require additional image-related information.

---

## How to use it

### **Deployment Setup**
The registry currently runs inside the Kubernetes cluster and is **accessible only via an internal network**. The setup consists of:
- A **Deployment** running the custom Golang server.
- A **ClusterIP Service** for internal access.
- A **Persistent Volume Claim (PVC)** to store image files.

**Note:** Future updates may include exposing the service externally via an **Ingress** resource.

### **Configuration Options**
Some server options can be configured through **command-line flags** in the deployment [manifest](../../operators/deploy/cloudimg-registry/templates/deployment.yaml):
- `--listener-addr` → Server listen address (default: `localhost:8080`)
- `--data-root` → Root directory of the registry inside PVC
- `--read-header-timeout-secs` → Timeout for reading headers in HTTP requests

### **API Endpoints**
All interactions happen via HTTP requests. Below is a table summarizing the available endpoints:

| Method | Endpoint | Description |
|--------|----------------------------------|-----------------------------|
| **GET** | `/reg` | Lists all existing repositories. |
| **GET** | `/reg/{repo}` | Lists all images inside a repository. |
| **GET** | `/reg/{repo}/{image}` | Lists all available tags for an image. |
| **GET** | `/reg/{repo}/{image}/{tag}` | Serves the image file for an image tag. |
| **GET** | `/reg/{repo}/{image}/{tag}/meta` | Serves the metadata JSON file for an image tag. |
| **POST** | `/reg/{repo}/{image}/{tag}` | Uploads an image with its metadata. **Request body:** `meta` (metadata JSON), `img` (image file). |
| **DELETE** | `/reg/{repo}/{image}/{tag}` | Deletes an image and its metadata. If it's the last tag/image, the image/repo directory is removed. |

**Note:** `reg` in the endpoints above is a default registry prefix which can also be configured using `--repo-prefix` command-line flag.

**Example Base URL:**  
For a server running at `192.168.0.1:8080`, with the default registry prefix (`reg`), the following request would be valid:

- **List all repositories:**  
  ```sh
  curl -X GET http://192.168.0.1:8080/reg
