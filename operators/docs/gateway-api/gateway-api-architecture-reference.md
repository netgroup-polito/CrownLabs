# Gateway API - Envoy Gateway

CrownLabs uses [*Envoy Gateway*](https://gateway.envoyproxy.io/) implementing the Kubernetes Gateway API (`gateway.networking.k8s.io/v1`) to manage L7 traffic routing, domain unification, and access security policies for all cluster services and user environments.

## Architecture & Gateway Choice Rationale

CrownLabs originally relied on a legacy NGINX Ingress Controller setup using custom annotations for L7 ingress routing.

During the transition to the Kubernetes Gateway API, an implementation based on **Cilium** was evaluated following an initial Proof of Concept (PoC). While Cilium offers efficient eBPF-based kernel networking, testing revealed critical L7 policy limitations:
* **Lack of `SecurityPolicy` Resource Support**: When native Kubernetes `SecurityPolicy` custom resources were applied, Cilium did not support or recognize them at all, forcing access control rules to rely on fragmented annotation-based workarounds.
* **Focused Ingress vs CNI Scope**: Cilium's Gateway API controller is tightly coupled with its CNI networking engine rather than acting as a dedicated, pure application-level proxy controller.

To achieve a clean, maintainable, and standards-compliant architecture, CrownLabs selected **Envoy Gateway**, which provides the following key advantages:
* **Native `SecurityPolicy` Support**: Full native recognition and processing of `SecurityPolicy` custom resources, enabling a clean two-level security policy model (Global Gateway policy + per-route overrides).
* **CNCF Reference Standard**: As the official CNCF reference controller for the Kubernetes Gateway API project, it adheres strictly to upstream standards without vendor lock-in or proprietary CRD extensions.
* **High Performance & Low Latency**: Powered directly by Envoy Proxy (the industry-standard C++ data plane), providing high request throughput, low latency, and efficient memory usage.
* **Decoupled Control & Data Plane**: Clear architectural separation between the Gateway Controller (control plane) and the Envoy Proxy fleet (data plane), enabling independent horizontal scaling.
* **Native Policy Extensibility & Authentication**: First-class support for authentication policies (OIDC), TLS termination, and advanced L7 traffic management (URL rewriting, path prefix matching, timeouts) without custom annotation friction or service mesh operational overhead.






## Architecture

The gateway architecture separates the control plane from the data plane:
* **Gateway Controller**: Reconciles `Gateway`, `HTTPRoute`, and `SecurityPolicy` resources, programming the underlying proxy configuration.
* **Envoy Proxy**: Data plane proxies dispatching incoming HTTPS traffic to target services.

### Single Global Gateway Model

CrownLabs adopts a **single global Gateway architecture** (`Crownlabs main`, class `envoy-public`) as the sole public entrypoint for the entire platform. The rationale behind this design includes:
* **Centralized Ingress Management**: Concentrates L7 traffic routing, domain unification, TLS certificate management, and global security policies under a single unified entrypoint.
* **Safe Multi-Tenancy via Route Selectors**: Rather than deploying separate gateway instances for different workspaces or tenants, a single shared Gateway safely serves all static core services and dynamic per-user instances by restricting route attachment using namespace label selectors (`allowedRoutes` / `matchLabels`).

## Deployment & Configuration Model


Core Gateway API resources are declared in the main CrownLabs Helm chart (`deploy/crownlabs/`).

### Template Standards vs Values Customization

To ensure architectural consistency, non-changing properties are hardcoded within the resource template ([`deploy/crownlabs/templates/gateway-api.yaml`](../../../deploy/crownlabs/templates/gateway-api.yaml)):
* `apiVersion`: `gateway.networking.k8s.io/v1`
* `kind`: `Gateway`
* `protocol`: `HTTPS`
* `port`: `443`
* `tls.mode`: `Terminate`

Deployment-specific properties are configured via [`deploy/crownlabs/values.yaml`](../../../deploy/crownlabs/values.yaml):

```yaml
global:
  gateway:
    gatewayApiMode: true
    name: "crownlabs-main"
    hostname: crownlabs.polito.it

gatewayAPI:
  class: envoy-public
  listenerName: https
  tls:
    certificateRefName: crownlabs-main-gw-cert
  allowedRoutesSimple:
    key: "crownlabs.polito.it/gw-access"
    value: "crownlabs-main-production"
```

## Centralized Authentication (SecurityPolicy & OIDC)

With the transition to Envoy Gateway, CrownLabs leverages a **centralized authentication architecture** powered natively by the Gateway API `SecurityPolicy` CRD. This removes authentication logic from the individual Go operators. The Envoy Gateway natively implements the OpenID Connect (OIDC) protocol to interface directly with Keycloak.

### Static vs Dynamic Routes (The `public` label)
Authentication policies are applied globally at the Gateway level, but individual routes can override or bypass them using labels:
* **Dynamic Routes**: By default, standard routes exposing protected resources or user environments do not possess any special labels. They inherit the global `SecurityPolicy`, meaning Envoy will intercept unauthenticated traffic and redirect users to Keycloak.
* **Static / Public Routes**: Core infrastructure routes (like the frontend dashboard or WebSSH) that do not require gateway-level authentication are labeled with `crownlabs.polito.it/public-route: "true"`. A dedicated, empty `SecurityPolicy` uses a `targetSelector` with this `matchLabels` condition to explicitly override the global policy, thereby disabling OIDC enforcement for these specific routes and allowing the applications to handle authentication internally.

### The Callback URL (`instout/callback`)
During the OIDC flow, Keycloak redirects the user back to the Gateway to exchange the authorization code for a session token. 
Initially, the callback URL (`instout/callback`) fell under the frontend's static route prefix. Because the frontend route bypassed authentication, Envoy ignored the callback request, failed to exchange the token, and failed to write the session cookie (resulting in a blank screen). 
To resolve this, a dedicated, explicit `HTTPRoute` for `instout/callback` was created **without** the `public` label. This ensures the callback route is intercepted by the global authentication policy, allowing Envoy to successfully finalize the OIDC handshake and authenticate the user.

## Route Selection and Security (`allowedRoutes` & `matchLabels`)

### Rationale

In a multi-tenant Kubernetes environment, allowing any namespace to freely attach `HTTPRoute` resources to a shared public Gateway presents a significant security risk. Without restriction, unauthorized users could create arbitrary routes to expose unmonitored external services (e.g., private servers or unauthorized web applications) through CrownLabs' primary domain and public TLS endpoint.

To prevent unauthorized attachment, the Gateway enforces a strict namespace route selector using `allowedRoutes.namespaces.from: Selector` configured with a `matchLabels` condition (e.g., `crownlabs.polito.it/gw-access: crownlabs-main-production`).

### How It Works

1. **Gateway Filtering**: The Gateway listener evaluates incoming `HTTPRoute` binding requests and accepts only routes whose parent namespace contains the matching key-value label pair.
2. **Dynamic Tenant Namespaces**: When tenant namespaces are provisioned by the Tenant Operator, the operator reconciles a unified common label map (`TenantCommonNSLabels`, parsed from key-value configuration strings via `MapFromKVString`). This map includes the Gateway access label (`crownlabs.polito.it/gw-access`), which is stamped onto all tenant namespaces using map copying (`maps.Copy`). The Instance Operator then reconciles dynamic per-instance `HTTPRoute` resources inside these labeled namespaces, enabling them to bind successfully to the Gateway.

3. **Static Routes via Helm Sub-charts**: Core infrastructure services (`frontend-app`, `qlkube`, `bastion`) are declared as Helm sub-charts of the main CrownLabs chart. During `helm install` or `helm upgrade`, Helm automatically assigns all sub-chart resources to the primary release namespace (e.g., `crownlabs-production`). Because the release namespace is configured with the matching `gw-access` label, static `HTTPRoute` resources automatically bind to the Gateway.


## Domain Unification and Path Routing

CrownLabs consolidates all core web services under a single primary domain (e.g., `crownlabs.polito.it`) using path-prefix matching on `HTTPRoute` resources:

| Service | Sub-chart | Path Prefix | Description |
| :--- | :--- | :--- | :--- |
| **Frontend** | `frontend-app` | `/` | CrownLabs Dashboard UI |
| **QL Cube** | `qlkube` | `/graph` | GraphQL API |
| **Web SSH** | `bastion` | `/webssh` | Command-line Web SSH interface |

Per-user VM and container instances generated by the Instance Operator receive dynamic `HTTPRoute` resources appended under unique instance path prefixes.

## Operational & Configuration Flags

CrownLabs operators and Helm charts rely on specific configuration flags to coordinate Gateway API route generation and namespace authorization:

* **`global.gateway.gatewayApiMode` (`--gateway-api-mode`)**:
  - **Purpose**: Controls whether CrownLabs operates in Gateway API mode or legacy Ingress mode.
  - **Usage**: When enabled (`true`), the Instance Operator reconciles Gateway API `HTTPRoute` resources (`enforceInstanceExpositionHTTPRoutePresence`). When disabled (`false`), it falls back to legacy Kubernetes `Ingress` resources.

* **`gatewayApiRefsValues` (`--gateway-api-refs-values`)**:
  - **Purpose**: Specifies the target Gateway reference formatted as `<gateway-namespace>/<gateway-name>` (e.g., `crownlabs-production/crownlabs-main`).
  - **Usage**: Required by the Instance Operator to forge the `parentRef` inside generated `HTTPRoute` specifications. Without this parameter, `HTTPRoute` resources cannot target the correct Gateway instance.

* **`TenantCommonNSLabels` (`MapFromKVString`)**:
  - **Purpose**: Defines common key-value labels stamped onto all tenant namespaces created by the Tenant Operator.
  - **Usage**: Parses configuration key-value strings (e.g., `"crownlabs.polito.it/gw-access=crownlabs-main-production"`) at operator startup into a label map (`TenantCommonNSLabels`). During namespace reconciliation, `maps.Copy` stamps these labels onto tenant namespaces, satisfying the Gateway listener's `allowedRoutes` selector (`matchLabels`) and authorizing dynamic `HTTPRoute` attachment.

* **`global.gateway.securityPolicy.enabled`**:
  - **Purpose**: Controls whether Envoy Gateway OIDC SecurityPolicy resources and callback routes are rendered at the deployment level.
  - **Usage**: Allows administrators or developers in staging/testing environments to toggle authentication enforcement centrally via Helm values, disabling OIDC policies conditionally without modifying underlying HTTPRoute definitions or operator code.

* **`authentication.enabled` (`EnableAuthentication`)**:
  - **Purpose**: Toggles authentication enforcement for instance exposition routes.
  - **Usage**: Allows administrators or developers in staging/testing environments to bypass authentication requirements conditionally without removing underlying route definitions.

## Advanced Routing & Operational Mechanics

### 1. Gateway API Acceptance Logic
The "Acceptance" mechanism verifies if the route for an instance has been correctly processed by the Gateway.
- **Gateway Validation**: When an `HTTPRoute` is created, the Gateway API operator checks its validity. It verifies if the namespace and the gateway name are correct, and if the rules are compatible with the gateway policies. If everything is correct, the Gateway "accepts" the route by injecting a `RouteConditionAccepted = True` status condition into the `HTTPRoute`. Otherwise, it rejects it.
- **Instance Status (`expositionAccepted`)**: The CrownLabs Instance controller reads this acceptance state and exposes it in the Instance status as a boolean flag (`expositionAccepted`). 
  * *Legacy Ingress Compatibility*: Since legacy `Ingress` resources do not natively report an acceptance status like `HTTPRoute` does, the operator immediately hardcodes `expositionAccepted = true` upon Ingress creation.
  * *Troubleshooting Note*: If a public endpoint/URL is deliberately absent (e.g. for pure Bastion connections without direct exposure), an `expositionAccepted: false` value is perfectly acceptable. However, if a public URL is expected but the flag remains `false`, it indicates a configuration problem that requires investigation.
- **Safe Migration**: This status acts as a safety parachute during hot migrations. The system verifies the successful creation and acceptance of the new `HTTPRoute` before destroying the legacy `Ingress`, minimizing any service disruption.

### 2. HTTPRoute Specification Structure
To keep resource definitions clean and simplify troubleshooting, CrownLabs implements a simplified `HTTPRoute` specification:
- **Single-Element Vectors**: Although the Kubernetes Gateway API standard supports multi-parent and multi-rule configurations, CrownLabs restricts `parentRefs` and `rules` lists to contain **exactly one element**.
- **Rule Components**: The single rule block contains:
  1. `matches`: Path matching logic (e.g. prefix matching `/instance/<instance-uid>/<env-name>`).
  2. `backendRefs`: Links the route directly to the target Kubernetes Service.
  3. `timeouts`: Timeout configurations for the route.
  4. `filters`: Manipulation filters (such as URL rewriting).

### 3. Route Logic Optimization

#### Timeouts
Two separate timeout values are configured to prevent resource hangs:
* **Static Routes (e.g. WebSSH, QLKube, Frontend)**: Both `request` (maximum time for the gateway to process the request) and `backend request` (time the gateway waits for the backend service to respond) timeouts are set to **600 seconds** via Helm values.
* **Dynamic Routes (e.g. Interactive Virtual Machine Sessions)**: For dynamic routes (like the ones for interactive virtual machine sessions), the timeout is set to **3600 seconds** (1 hour).

##### URL Rewriting vs. Redirect
When cleaning up request paths before forwarding them to user backends, CrownLabs implements **Prefix URL Rewriting** instead of a Redirect logic. This relies on a strict routing schema (e.g., `https://<domain>/app/instance/<tenant-name>/<instance-name>/<environment-name>/`).
* **Mechanism**: The Gateway API `URLRewrite` filter intercepts traffic matching the instance-specific path prefix (which acts as the fundamental routing contract). It strips this entire prefix and replaces it with an empty string (`""`), passing only the trailing path (e.g. `/ssh`) to the backend service. Any divergence from this schema will result in an HTTP 404 error because the route won't match.
* **Rationale**: Using Rewrite instead of Redirects is faster and executes entirely in the Envoy data plane, saving client round-trips. It also abstracts away the complex path structures from the underlying user environments, which can simply serve their content from the root `/` path.
* **Path Management**: Under the legacy Ingress architecture, path stripping was managed via regex matching and the `rewrite-target` annotation. The Gateway API replaces this fragile parsing with a clean, native prefix replacement filter defined directly in the `HTTPRoute`.

#### Legacy Annotations Obsolescence
Under Gateway API mode, `HTTPRoute` resources **do not inherit or use any legacy NGINX Ingress annotations**. The previous annotation-based configurations have been entirely replaced by native mechanisms:
* **Timeouts (WebSSH/Instances)**: Migrated to the native `timeouts` block inside the `HTTPRoute` specification.
* **TLS & Cert-Manager (Frontend)**: The `cert-manager.io/cluster-issuer` annotation is no longer needed on individual routes. TLS termination and certificate management are now handled centrally on the global `Gateway` resource.
* **Buffer Sizes**: Obsolete. Envoy natively accommodates much larger headers (e.g., up to 60K for request headers) without requiring manual tuning of memory chunks, thus removing the need for `proxy-buffer-size` annotations.
* **Custom HTTP Errors**: Obsolete. Envoy manages error flows directly. While NGINX required the `custom-http-errors` annotation (e.g. for error 418), under Gateway API this custom logic is no longer implemented at the route level. (If needed in the future, it can be handled via `EnvoyPatchPolicy`).

#### WebSockets Support
In Gateway API v1.3, WebSocket integration is greatly simplified. Instead of using custom proxy annotations to upgrade connections, administrators define `App Protocol` as `"kubernetes.io/ws"` directly on the Service port definition. The Envoy proxy automatically detects this protocol and upgrades connections natively.

#### Buffer Size & Large Headers
The legacy NGINX Ingress constraint of a 4K/8K buffer size is obsolete under Envoy. NGINX traditionally uses fixed-size memory chunks to parse response headers, requiring administrators to manually tune annotations like `proxy-buffer-size` to prevent 502 Bad Gateway errors when dealing with large headers.
Envoy natively accommodates much larger headers (by default up to 60K) without requiring manual tuning of memory chunks, thus making legacy buffer tuning annotations entirely unnecessary.

### 4. Headless Services Architecture
To satisfy specific resource lifecycle requirements, CrownLabs implements **Headless Services** (ClusterIP set to `"None"`) for **newly provisioned environments**. (Note: Pre-existing environments retain their legacy ClusterIP because the Kubernetes `ClusterIP` field is immutable and cannot be updated on the fly).
For environments utilizing Headless Services:
* **DNS-Based Resolution**: Services do not receive a virtual Cluster IP. Instead, they rely on the cluster DNS to resolve directly to the destination Pod's IP address.
* **Connection Stability**: Accessing environments through DNS names (which resolve dynamically to the Pod) ensures that if a Pod/VM's IP shifts within the cluster, the DNS is updated automatically and the connection remains stable, without needing configuration rewrites.

### 5. Rollout Strategy
To prevent service disruption when updating or creating exposition resources, a **safe rollout** strategy is implemented:
1. **Create-Before-Destroy**: The Instance Operator creates the new `HTTPRoute` resource first and verifies that the Gateway has accepted it.
2. **Downtime Minimization**: Only after the route is successfully accepted and DNS is synchronized does the operator destroy any older or conflicting resources (like legacy Ingresses during transitions). This act of creating the new path before tearing down the old one acts as a safety net to ensure zero downtime.

---

## Legacy Ingress Annotations to Gateway API Mapping

The following table summarizes how legacy NGINX Ingress annotations are mapped to native Gateway API specs or Envoy mechanisms within CrownLabs:

| Legacy Ingress Annotation | Gateway API / Envoy Native Equivalent | Description |
| :--- | :--- | :--- |
| `nginx.ingress.kubernetes.io/proxy-read-timeout` / `proxy-send-timeout` | `HTTPRoute.spec.rules[].timeouts.backendRequest` | Configured to 600s for static routes (3600s for VM sessions) |
| `nginx.ingress.kubernetes.io/proxy-connect-timeout` | `HTTPRoute.spec.rules[].timeouts.request` | Configured to 600s for static routes (3600s for VM sessions) |
| `nginx.ingress.kubernetes.io/rewrite-target` | `HTTPRoute.spec.rules[].filters` (Type: `URLRewrite`) | Replaces matched prefix path with `""` |
| `nginx.ingress.kubernetes.io/websocket-services` | Service Port `appProtocol: kubernetes.io/ws` | Gateway API 1.3 native WebSocket upgrade |
| `nginx.ingress.kubernetes.io/proxy-buffer-size` | Native Envoy Stream Handling | Obsolete. Envoy natively supports headers/streams up to 60K |
| `nginx.ingress.kubernetes.io/custom-http-errors` | Native Envoy Error Flows | Obsolete. Envoy manages errors directly; custom logic can be patched via Envoy Patch Policy if needed in the future |


