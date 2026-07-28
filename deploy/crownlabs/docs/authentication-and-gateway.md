# Authentication and Centralized Gateway

CrownLabs uses **Envoy Gateway** implementing the Kubernetes Gateway API (`gateway.networking.k8s.io/v1`) for L7 traffic routing and authentication.

## Gateway Configuration Model

* **Hardcoded Parameters** (defined directly in the resource template `templates/gateway-api.yaml`):
  - `apiVersion`: `gateway.networking.k8s.io/v1`
  - `kind`: `Gateway`
  - `protocol`: `HTTPS`
  - `port`: `443`
  - `tls.mode`: `Terminate`
* **Configurable Parameters** (specified via `values.yaml`):
  - Gateway Name (`name: crownlabs-main`)
  - Gateway Class (`class: envoy-public`)
  - Listener Name (`listenerName: https`)
  - Hostname (`hostname`)
  - TLS Certificate References (`tls.certificateRefName`)
  - Route Selector (`allowedRoutes` / `allowedRoutesSimple`)

## Route Selection and Security (`allowedRoutes` & `matchLabels`)

In a multi-tenant Kubernetes environment, allowing any namespace to freely attach `HTTPRoute` resources to a shared public Gateway presents a significant security risk. Without restriction, unauthorized users could create arbitrary routes to expose unmonitored external services (e.g., private servers or unauthorized web applications) through CrownLabs' primary domain and public TLS endpoint.

To prevent unauthorized attachment, the Gateway enforces a strict namespace route selector using `allowedRoutes.namespaces.from: Selector` configured with a `matchLabels` condition (e.g., `crownlabs.polito.it/gw-access: crownlabs-main-production`).

1. **Gateway Filtering**: The Gateway listener evaluates incoming `HTTPRoute` binding requests and accepts only routes whose parent namespace contains the matching key-value label pair.
2. **Dynamic Tenant Namespaces**: When tenant namespaces are provisioned by the Tenant Operator, the operator reconciles a unified common label map (including `crownlabs.polito.it/gw-access`). The Instance Operator then reconciles dynamic per-instance `HTTPRoute` resources inside these labeled namespaces, enabling them to bind successfully to the Gateway.
3. **Static Routes via Helm Sub-charts**: Core infrastructure services are declared as Helm sub-charts. During `helm install`, Helm automatically assigns all sub-chart resources to the primary release namespace (e.g., `crownlabs-production`). Because the release namespace is configured with the matching `gw-access` label, static `HTTPRoute` resources automatically bind to the Gateway.

## Control Flags

* `global.gateway.gatewayApiMode` (`--gateway-api-mode`): Toggle flag (enabled by default) that controls whether CrownLabs deploys Gateway API `HTTPRoute` resources instead of legacy `Ingress` objects.
* `configurations.generic.gatewayApiRefsValues` (`--gateway-api-refs-values`): Configures the target Gateway reference formatted as `<namespace>/<name>` (e.g., `crownlabs-production/crownlabs-main`), enabling the Instance Operator to forge valid `parentRef` targets for `HTTPRoute` resources.
* `global.gateway.securityPolicy.enabled`: Helm deployment flag (enabled by default) that controls whether Envoy Gateway SecurityPolicy OIDC authentication resources and callback routes are rendered.
* `authentication.enabled` (`EnableAuthentication`): Operator flag that controls whether authentication enforcement is active for instance exposition routes.

## Centralized Authentication (SecurityPolicy & OIDC)

With the transition to Envoy Gateway, CrownLabs leverages a **centralized authentication architecture** powered natively by the Gateway API `SecurityPolicy` CRD. This removes authentication logic from the individual Go operators. The Envoy Gateway natively implements the OpenID Connect (OIDC) protocol to interface directly with Keycloak (the OIDC provider).

*(Note: An early experiment to manage SecurityPolicies directly inside the Instance controller to enforce per-instance auth rules was developed in the [`monument/mrMela/secpol-in-instctrl`](https://github.com/netgroup-polito/CrownLabs/tree/monument/mrMela/secpol-in-instctrl) branch, but the architecture eventually shifted to the centralized global model described here).*

When a user attempts to access a protected dynamic route (e.g., a virtual machine's GUI), the Envoy Gateway intercepts the unauthenticated traffic and automatically redirects the user to Keycloak for login. Once authenticated, the gateway handles the token exchange and issues a session cookie natively.

### Static vs Dynamic Routes (The `public` label)

Authentication policies are applied globally at the Gateway level, but individual routes can override or bypass them using labels:
* **Dynamic Routes**: By default, standard routes exposing protected resources or user environments do not possess any special labels. They inherit the global `SecurityPolicy`, meaning Envoy will intercept unauthenticated traffic and redirect users to Keycloak.
* **Static / Public Routes**: Core infrastructure routes (like the frontend dashboard or WebSSH) that do not require gateway-level authentication are labeled with `crownlabs.polito.it/public-route: "true"`. A dedicated, empty `SecurityPolicy` uses a `targetSelector` with this `matchLabels` condition to explicitly override the global policy, thereby disabling OIDC enforcement for these specific routes and allowing the applications to handle authentication internally.

### The Callback URL (`instout/callback`)

During the OIDC flow, Keycloak redirects the user back to the Gateway to exchange the authorization code for a session token. 
Initially, the callback URL (`instout/callback`) fell under the frontend's static route prefix. Because the frontend route bypassed authentication, Envoy ignored the callback request, failed to exchange the token, and failed to write the session cookie (resulting in a blank screen). 
To resolve this, a dedicated, explicit `HTTPRoute` for `instout/callback` was created **without** the `public` label. This ensures the callback route is intercepted by the global authentication policy, allowing Envoy to successfully finalize the OIDC handshake and authenticate the user.
