# Identity Provider - Keycloak

Keycloak is an Open Source Identity and Access Management solution for modern Applications and Services.
It enables to concentrate all the tasks related to identity and access management into the same place; once authenticated, a user session can be associated to a token that can be used to validate the access of all the resources available in the cluster.

This brief guide presents how to install Keycloak in HA in a K8S cluster with a PostgreSQL Database backend (also in HA).

More info at [Keycloak's website](https://www.keycloak.org).

**If you want to connect to the CrownLabs cluster, jump to the [dedicated documentation page](https://crownlabs.polito.it/resources/sandbox/).**

## Pre-requisites
Here we assume that the following operators are installed and configured in the K8s cluster:
* [ROOK](https://rook.io/)
* [NGINX Ingress Controller](https://github.com/kubernetes/ingress-nginx)
* [cert-manager](https://cert-manager.io/)
* A namespace in K8S cluster called **keycloak-ha**

You will need the following tools installed in your workstation:
* [Helm](https://helm.sh/)

## PostgreSQL-Operator
The PostgreSQL-Operator is required to deploy the database for Keycloak. Refer to [postgres-operator](../database-operators/postgres-operator) for installation and maintenance information.

### Create a Postgres cluster

If the operator pod is running, it listens to new events regarding PostgreSQL resources. Now, it's time to submit your first Postgres cluster manifest that you can find in [manifests](manifests/) folder of this repo.
If you need to add some more features, refer to the official docs.

```bash
# create a Postgres cluster
kubectl create -f keycloak-postgres-cluster-manifest.yaml
```

After the cluster manifest is submitted and passed the validation, the operator will create *Service* and *Endpoint* resources and a *StatefulSet* which spins up new pod(s) given the number of instances specified in the manifest.
All resources are named like the cluster. The database pods can be identified by their number suffix, starting from -0. They run the Spilo container image by Zalando.
As for the services and endpoints, there will be one for the master pod and another one for all the replicas (-repl suffix).
We suggest to check if all components are coming up. Use the label `application=spilo` to filter, and check the label `spilo-role`
to see who is currently the master.

```bash
# check the deployed cluster
kubectl get postgresql

# check created database pods
kubectl get pods -l application=spilo -L spilo-role

# check created service resources
kubectl get svc -l application=spilo -L spilo-role
```

## Keycloak Server deployment
Keycloak helm repository is available at [Codecentric's Github](https://github.com/codecentric/helm-charts/tree/master/charts/keycloak).


The following commands will add the repository and deploy keycloak.
Helm values are directly commented, further documentation is available at the link above.

```bash
#add the codecentric helm repository
helm repo add codecentric https://codecentric.github.io/helm-charts
helm install keycloak-server codecentric/keycloak --namespace keycloak-ha --create-namespace --values=conf-files/keycloak-configuration.yaml
```

### Customize the email templates
In order to customize the different email templates, proceed as follows:

1. Edit the relevant files in [templates/crownlabs](templates/crownlabs);
2. Create the config maps:
   ```sh
   $ kubectl create configmap keycloak-theme-email -n keycloak-ha --from-file=templates/crownlabs/email/
   $ kubectl create configmap keycloak-theme-email-html -n keycloak-ha --from-file=templates/crownlabs/email/html
   $ kubectl create configmap keycloak-theme-email-text -n keycloak-ha --from-file=templates/crownlabs/email/text
   $ kubectl create configmap keycloak-theme-email-messages -n keycloak-ha --from-file=templates/crownlabs/email/messages
   ```
3. Restart the `keycloak-server` pods to reload the configuration.

## Configure K8S api-server to be used with Keycloak
Please follow the [official documentation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/) to allow the K8s Api-server to exploit the running Keycloak instance as identity provider.

## User Instances Authentication

In CrownLabs, the access to the graphical desktop of the user instances must be protected so that only authenticated users can connect to them.
With the transition to the Gateway API, this authentication is managed centrally and natively by **Envoy Gateway** using a global `SecurityPolicy`.

### Native OIDC Integration

Envoy Gateway natively implements the OpenID Connect (OIDC) protocol to interface directly with Keycloak (the OIDC provider). 
When a user attempts to access a protected dynamic route (e.g., a virtual machine's GUI), the Envoy Gateway intercepts the unauthenticated traffic and automatically redirects the user to Keycloak for login. Once authenticated, the gateway handles the token exchange and issues a session cookie natively.

### Enabling Authentication

The `SecurityPolicy` is deployed and managed centrally via the CrownLabs infrastructure Helm charts.
By default, the global policy protects all routes managed by the Gateway. The Instance Operator simply generates standard `HTTPRoute` resources for the user instances, and they are automatically protected without requiring any instance-specific authentication configuration or annotations.

Static routes that must bypass authentication (such as the frontend application) are labeled with `crownlabs.polito.it/public-route: "true"`. This label is targeted by a secondary, empty `SecurityPolicy` which explicitly overrides the global Gateway policy, thereby disabling OIDC enforcement for those specific routes.
