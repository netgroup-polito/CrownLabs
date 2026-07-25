# Gateway API Security Policy: Experiment (Historical Note)

> [!NOTE]
> This document serves as a historical record of an architectural experiment that was implemented and tested but ultimately **discarded**. It is preserved here to explain the rationale behind certain code structures that were temporarily introduced and to prevent future developers from duplicating the same effort.

## 1. The Context: Security Policy Definition
A `SecurityPolicy` is a Kubernetes Custom Resource Definition (CRD) provided by Envoy Gateway, designed to apply authentication policies. 

The design highlighted a key architectural advantage: **granularity**. 
- A policy can be applied **globally** directly to a Gateway (acting like a security guard outside a metro station), protecting all routes referencing that Gateway.
- Alternatively, it can be applied **granularly** to individual `HTTPRoute` resources or via labels (acting like turnstiles for specific metro lines), ensuring that only authenticated users can access specific VM GUIs.

## 2. The Historical Experiment
The goal of this historical experiment was to **integrate Envoy's native `SecurityPolicy` CRD** to manage authentication within the Gateway API ecosystem and to create a SecurityPolicy per HTTPRoute actually. The initial approach aimed to replicate the legacy Ingress flow: configuring the `SecurityPolicy` to use External Authentication (`ext_authz`) to forward authentication requests to `oauth2-proxy`, which in turn would interface with Keycloak (the authentication provider).

## 3. Code Flow & Component Responsibilities
To understand exactly *who does what, where, and why*, the experimental logic was distributed across several key components of the `instance-operator`.

### Phase 1: Operator Initialization (`main.go`)
- **What**: The operator reads the `enable_authentication` flag and retrieves the base `SecurityPolicy` template (`deploy/crownlabs/operators/templates/security-policy.yaml`).
- **Where**: During the operator's startup sequence.
- **Why (GetAPIReader)**: The template is fetched dynamically using `GetAPIReader()` rather than the standard client. This is because at this stage of execution, the operator's `Manager` (and its internal cache) has not yet been started. A direct API query to the cluster was the only way to retrieve the resource.
- **Why (The "Dummy Secret")**: The Kubernetes CRD strictly required a `clientSecret` field to pass validation. However, the authentication provider was configured in a way that didn't require one for this flow. To satisfy Kubernetes without breaking the provider, a placeholder "dummy secret" was injected into the template. 

### Phase 2: The Reconciler's Builder Pattern (`main.go`)
- **What**: The core reconciliation loop determines which resources to generate for a specific Tenant/Instance.
- **Where**: Inside the main reconciliation flow of the `instance-operator`.
- **Why**: Originally, the controller relied heavily on "early returns" (e.g., generate a resource, return it immediately). The introduction of `SecurityPolicy` required evaluating nested states: *Is Gateway API Mode enabled? If yes, is Enable Authentication also enabled?* To handle this cleanly, the logic was refactored into a **Builder pattern**. The controller sequentially evaluates the active flags and "builds" an array of required resources (`HTTPRoute`, `Service`, `SecurityPolicy`) step-by-step, returning the final accumulated list at the end.

### Phase 3: Route & Policy Exposition (`exposition.go`)
- **What**: The `EnforceSecurityPolicyPresence` function manages the routing parameters for the given Instance.
- **Where**: Inside the exposition logic that connects the VM to the outside world.
- **Why**: It calculates the expected `HTTPRoute` name and the necessary labels. Critically, it dynamically populates the `redirect URL` using the `path prefix` associated with the Instance's route. This ensures that when the authentication proxy flow finishes, the user is correctly bounced back to the specific URL path of their VM's GUI. It then delegates the actual object generation to the next layer.

### Phase 4: Policy Forging & Memory Safety (`securitypolicy.go`)
- **What**: The `ForgeSecurityPolicySpec` function merges the calculated parameters into the base template.
- **Where**: In a dedicated, modular file handling only `SecurityPolicy` resources.
- **Why (Dynamic TargetRef)**: The base template intentionally left the `targetRefs` block entirely blank. This function injects the calculated `HTTPRoute` name and sets the `kind` to `HTTPRoute` at runtime, creating a granular binding between the policy and the specific route.
- **Why (DeepCopy)**: Because the base template is loaded once globally into memory during Phase 1, `ForgeSecurityPolicySpec` *must* execute a `DeepCopy()` on the `SecurityPolicySpec` before modifying it. In Go, maps and slices (like `targetRefs`) are reference types. Modifying them directly without a deep copy would permanently alter the shared pointers in the base template, causing cross-contamination (corrupted targetRefs and race conditions) for all subsequent VM instances processed by the operator.

### Phase 5: Validation (`securitypolicy_test.go`)
- **What**: `SecurityPolicesTest` verifies the integrity of the generated policies.
- **Where**: In the unit testing suite.
- **Why**: The tests simulate the generation process to assert two critical things: 1) The `targetRef` is correctly attached to the expected route. 2) The original template parameters (like OIDC fields or proxy configs) are preserved unmodified after the DeepCopy and forging process.



## Conclusion
Although the implementation was modular, testable, and successfully bypassed `oauth2-proxy`, the approach was eventually discarded. This document remains to clarify the context behind the `GetAPIReader` template logic, the `DeepCopy` requirements, and the controller's Builder pattern evolution.
