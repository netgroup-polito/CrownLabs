# Gateway API - Operator Implementation Details

This document details how the CrownLabs Go operators are implemented to reconcile and manage Kubernetes Gateway API resources. For general proxy architecture and routing behavior, see the deployment documentation.

## Operational & Configuration Flags

CrownLabs operators rely on specific configuration flags to coordinate Gateway API route generation and namespace authorization:

* **`gatewayApiRefsValues` (`--gateway-api-refs-values`)**:
  - **Purpose**: Specifies the target Gateway reference formatted as `<gateway-namespace>/<gateway-name>` (e.g., `crownlabs-production/crownlabs-main`).
  - **Usage**: Required by the Instance Operator to forge the `parentRef` inside generated `HTTPRoute` specifications. Without this parameter, `HTTPRoute` resources cannot target the correct Gateway instance.

* **`TenantCommonNSLabels` (`MapFromKVString`)**:
  - **Purpose**: Defines common key-value labels stamped onto all tenant namespaces created by the Tenant Operator.
  - **Usage**: Parses configuration key-value strings (e.g., `"crownlabs.polito.it/gw-access=crownlabs-main-production"`) at operator startup into a label map (`TenantCommonNSLabels`). During namespace reconciliation, `maps.Copy` stamps these labels onto tenant namespaces, satisfying the Gateway listener's `allowedRoutes` selector (`matchLabels`) and authorizing dynamic `HTTPRoute` attachment.

## Advanced Routing & Operational Mechanics

### 1. Gateway API Acceptance Logic
The "Acceptance" mechanism verifies if the route for an instance has been correctly processed by the Gateway.
- **Gateway Validation**: When an `HTTPRoute` is created, the Gateway API operator checks its validity. It verifies if the namespace and the gateway name are correct, and if the rules are compatible with the gateway policies. If everything is correct, the Gateway "accepts" the route by injecting a `RouteConditionAccepted = True` status condition into the `HTTPRoute`. Otherwise, it rejects it.
- **Instance Status (`expositionAccepted`)**: The CrownLabs Instance controller reads this acceptance state and exposes it in the Instance status as a boolean flag (`expositionAccepted`).
  * *Troubleshooting Note*: If a public endpoint/URL is deliberately absent (e.g. for pure Bastion connections without direct exposure), an `expositionAccepted: false` value is perfectly acceptable. However, if a public URL is expected but the flag remains `false`, it indicates a configuration problem that requires investigation.

### 2. HTTPRoute Specification Structure
To keep resource definitions clean and simplify troubleshooting, CrownLabs implements a simplified `HTTPRoute` specification in the operator:
- **Single-Element Vectors**: Although the Kubernetes Gateway API standard supports multi-parent and multi-rule configurations, CrownLabs restricts `parentRefs` and `rules` lists to contain **exactly one element**.
- **Rule Components**: The single rule block contains:
  1. `matches`: Path matching logic (e.g. prefix matching `/instance/<instance-uid>/<env-name>`).
  2. `backendRefs`: Links the route directly to the target Kubernetes Service.
  3. `timeouts`: Timeout configurations for the route.
  4. `filters`: Manipulation filters (such as URL rewriting).

### 3. Headless Services Architecture
To satisfy specific resource lifecycle requirements, CrownLabs implements **Headless Services** (ClusterIP set to `"None"`) for **newly provisioned environments**. (Note: Pre-existing environments retain their legacy ClusterIP because the Kubernetes `ClusterIP` field is immutable and cannot be updated on the fly).
For environments utilizing Headless Services:
* **DNS-Based Resolution**: Services do not receive a virtual Cluster IP. Instead, they rely on the cluster DNS to resolve directly to the destination Pod's IP address.
* **Connection Stability**: Accessing environments through DNS names (which resolve dynamically to the Pod) ensures that if a Pod/VM's IP shifts within the cluster, the DNS is updated automatically and the connection remains stable, without needing configuration rewrites.

### 4. Rollout Strategy
To prevent service disruption when updating or creating exposition resources, a **safe rollout** strategy is implemented in the operator logic:
1. **Create-Before-Destroy**: The Instance Operator creates the new `HTTPRoute` resource first and verifies that the Gateway has accepted it.
