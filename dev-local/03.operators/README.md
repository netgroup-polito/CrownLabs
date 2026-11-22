# CrownLabs operator — local setup

This step is the "core" of the guide, setting up the operator(s) that are the brain of CrownLabs.
The operators are pieces of software that reflect any changes to the Kubernetes cluster, in a process called reconciliation.
For example, they are responsible of creating all the Kubernetes resources for a user, when a new user is added in the Keycloak database (something that in the previous step was done by hand).

In this guide, we will suppose that you are familiar with the concept of operator, and how they work.
We will only see how to set them up.
If you instead need more theoretical knowledge about operators, please refer to the material for the [Cloud Programming](https://cloudprog.frisso.net/) course.

## 0. Prerequisites

### 0.1. Helm

_For more information about what Helm is, refer to the Cloud Programming course material._

In addition to executing the previous guides, for this guide we also need to install Helm.
Despite not using it to deploy the operators, we still require it to build the dependencies.

You can download Helm following the [install guide](https://helm.sh/docs/intro/install/) from the Helm website.

In short, the following command suffices (_if you want to live on the edge_, as stated on the Helm website itself):

```bash
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-4 | bash
```

### 0.2. What is RBAC

RBAC stands for Role-Based Access Control.

In a normal authorization scheme, an admin grants authorizations directly to the user.
In RBAC, instead, the admin creates some roles, which have some specific authorizations: the users are then assigned one or more role, and inherit their respective authorizations.
RBAC ensures consistency among users, while also being a single point of control where an admin can easily change permissions.

Let's consider an example, in the CrownLabs environment.
When creating a Workspace, two roles are created for it:

- a user, that has the basic privileges (creating Instances from Templates, managing their own Instances, ...)
- a manager, that can manage the Workspace (add users, define Templates, manage all the Instances, ...)

Each Workspace's roles only give permissions on resources in the Workspace itself: this way, separation of concerns is guaranteed.

### 0.3. How Kubernetes implements RBAC

The Kubernetes API server can use RBAC as an authentication method, if the correct flag is set.
The local setup should have it already enabled on its own.

Kubernetes defines 4 object types to implement RBAC:

- `Role`: a set of privileges that are granted for a specific namespace.
  They are purely additive, so there is no "deny" rule.
- `ClusterRole`: a non-namespaced `Role`, that grants privileges across the whole cluster.
- `RoleBinding`: the assignation of any type of role to a specific user, limited to a specific namespace.
  The role can be:
    - a `Role`: a "special" privilege defined for the namespace, granted only for that namespace.
    - a `ClusterRole`: a "generic" privilege, granted however only within the scope of a namespace.
      An example could be a "can start VMs" `ClusterRole`, an action that is generally not namespace-specific.
      However, we may want to limit a user to start VMs in just a specific namespace, hence applying the `ClusterRole` only to a single namespace.
- `ClusterRoleBinding`: the assignation of a cluster-wide role to a specific user.
  Since it affects the whole cluster, the rule must cover the whole cluster (it must be a `ClusterRole`).

More details can be found on the [Kubernetes RBAC docs](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).

### 0.4. CrownLabs' role setup

CrownLabs implements RBAC by means of `ClusterRole`s only.
Some examples of `ClusterRole`s used by Crownlabs are:

- `crownlabs-manage-tenants`: enables a user to manage all tenants (for admins)
- `crownlabs-manage-tenant-<tenant>`: enables a user to manage a specific tenant (usually granted to the tenant itself)
- `crownlabs-manage-instances`: enables a user to manage instances
- `crownlabs-view-workspaces`: enables a user to read the list of workspaces available

## 1. Base CrownLabs RBAC and CRDs

In the production environment, all the `ClusterRole`s and `ClusterRoleBinding`s are created by the same HELM chart that runs the operator, as well as the required CRDs.
Here we will instead run the operator directly with `go run` or `make`, therefore we have to setup them manually.

Keeping a hand-maintained copy of the settings is not ideal, as they could drift from the real ones.
Instead, the best practice is to use directly the objects from the real chart (`/deploy/crownlabs`):

```bash
# The chart's roles and bindings templates are rendered relative to this folder's location.
cd dev-local/02.keycloak

# One-time step; only needed if deploy/crownlabs/charts is empty.
helm dependency build ../../deploy/crownlabs

# Apply the built templates to the K3s cluster
helm template -s templates/clusterroles.yaml -s templates/clusterrolebindings.yaml ../../deploy/crownlabs | kubectl apply -f -
```

The roles and bindings we created are persistent, so there should not be any need to rerun this every time.
Still, re-running does not cause any harm: if the RBAC templates change, feel free to simply rerun these commands.

## 2. Run the CrownLabs operator

Normally, the HELM chart compiles and deploys the operator among everything else.
To enhance flexibility, during development you will instead run the operator directly on your machine.
The development setup already takes care of linking the various systems, you just have to start the operator process:

```bash
# run from "/operators", NOT from "/dev-local/operators"
make run-operator-local
```

The Makefile has multiple options that can be changed by using environment variables.
While they are all visible in the [Makefile](../../operators/Makefile) itself (at `.PHONY: run-operator-local`), it is worth pointing out the `MYDRIVE_STORAGE_CLASS` variable.
It indicates the storage class for K3s: by default it is set to `local-path`, override it if it is different in your local setup.

While executing `run-operator-local`, on top of the required resources, some sample items are also created.
Among others, this includes:

- 4 workspaces
- some VM templates
- a sample tenant (`John Doe`), with some instances created

The full list of example resources can be seen in the relative files in the [`operators/samples`](../../operators/samples/) folder.

You can check that the resources exist by running the appropriate kubectl commands.
For example, you can check the existence of the tenant by running:

```bash
kubectl get tenant
```

If you instead want to test the reconciling of the operator, you can create manually a resource.
For example, this kubectl apply creates a new user (`Jane Doe`), with her email verified:

```bash
kubectl apply -f - <<'EOF'
apiVersion: crownlabs.polito.it/v1alpha2
kind: Tenant
metadata:
  name: jane.doe
  labels:
    crownlabs.polito.it/operator-selector: local
spec:
  email: jane.doe@example.com
  firstName: Jane
  lastName: Doe
  workspaces: []
status:
  keycloak:
    userConfirmed: True
EOF
```

After running the command, you can check that the operator indeed applied multiple updates:

- there is a new tenant in the K3s resources (`kubectl get tenant`)
- the tenant is also present in the Keycloak console
- a namespace (`kubectl get ns`) is created for Jane
- a cluster role (`kubectl get clusterrole`) exists for this new namespace
- _and much more..._

> Note: the script `run-operator-local` recreates every time the sample resources, overriding any modification that has been made.
> This may lead to confusion due to resources disappearing.
>
> For example, you could experience the following situation:
>
> - You created a new `Workspace` for `John Doe`
> - In the process of upgrading the operator, you did some changes
> - You then correctly executed `make run-operator-local` to run the updated version of the operator
> - You got no error message, therefore assumed that everything went fine
> - However, you are now getting a `403 Forbidden` error when trying to access the workspace that was working few minutes earlier: it was actually deleted when re-importing the sample resources!
>
> A solution against this problem is to simply create new resources (tenant, workspace, ...) for your testing.
> Since these would not be among the sample resources, the script inside the Makefile would simply not touch them.
>
> If you instead prefer to go on with John Doe, you will have to re-apply the changes to the sample resources every time you start the operator again.

### Forcing a reconciliation

The operator does not watch Keycloak in real time: there is no active webhook locally.
If you change something on the Keycloak side (for example, verifying a user), and you want the operator to pick it up immediately, force a new reconciliation with a harmless annotation:

```bash
kubectl annotate tenant.crownlabs.polito.it <tenant-name> crownlabs.polito.it/force-reconcile="$(date +%s)" --overwrite
```
