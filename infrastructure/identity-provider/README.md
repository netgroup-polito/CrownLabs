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
The following steps will install the postgresql-operator in the namespace called **keycloak-ha**.
The Postgres Operator can be installed simply by applying `yaml` manifests, after properly changing the namespace in file `operator-service-account-rbac.yaml` for the `service account` and `cluster rolebinding`.

### Manual deployment setup
For more details, please visit the [official documentation website](https://github.com/zalando/postgres-operator#documentation).

 ```bash
# First, clone the repository and change to the directory postgres-operator
git clone https://github.com/zalando/postgres-operator.git
cd postgres-operator
# apply the manifests in the following order
kubectl create -f manifests/configmap.yaml -n keycloak-ha  # configuration
kubectl create -f manifests/operator-service-account-rbac.yaml -n keycloak-ha # identity and permissions
kubectl create -f manifests/postgres-operator.yaml -n keycloak-ha # deployment
```

### Check if Postgres Operator is running
Starting the operator may take a few seconds. Check if the operator pod is running before applying a Postgres cluster manifest.

```bash
kubectl get pod -l name=postgres-operator -n keycloak-ha
```

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

In CrownLabs, the access to the graphical desktop of the user instances should be protected, so that only authenticated users can connect to them.
For this purpose, we leverage [oauth2-proxy](https://github.com/oauth2-proxy/oauth2-proxy), a solution which in this configuration stands in between the reverse-proxy (Nginx in our case) and the OIDC provider (Keycloak).

Once enabled on a per-ingress basis through the proper annotations (see below), all user requests are authenticated against oauth2-proxy, which in turn initially redirects the user to the OIDC provider for the log-in process.
Once authenticated, oauth2-proxy returns a cookie to the user, which will be validated during the following checks, without further interacting with the OIDC provider.

### Deploying oauth2-proxy

We leverage Helm to install a centralized deployment (i.e. used for all user instances) of oauth2-proxy, configuring it with multiple replicas for failure tolerance, and leveraging Redis (with Sentinel) as session storage backend.
The full configuration is described by the corresponding [values file](manifests/oauth2-proxy-values.yaml), with only the secrets redacted.
The installation/update can be performed with the following:

```bash
helm repo add oauth2-proxy https://oauth2-proxy.github.io/manifests
helm upgrade --install crownlabs-instances-auth oauth2-proxy/oauth2-proxy \
  --namespace crownlabs-instances-auth --create-namespace \
  --values manivests/oauth2-proxy-values.yaml
```

### Enabling the authentication

Once installed. user authentication can be enabled on a per-ingress basis thorough the following annotations (automatically configured by the instance-operator upon instance creation), pointing to the URLs where the oauth2-proxy deployment is exposed:

```yaml
nginx.ingress.kubernetes.io/auth-url: https://crownlabs.polito.it/app/instances/oauth2/auth
nginx.ingress.kubernetes.io/auth-signin: https://crownlabs.polito.it/app/instances/oauth2/start?rd=$escaped_request_uri
```

Currently, we perform user authentication only, hence ensuring no external users can access the graphical desktop of the user instances. Still, more complex authorization policies (e.g., group-based), could be applied both globally (i.e., inside the oauth2-proxy configuration) and [specifically for each ingress resource](https://github.com/oauth2-proxy/oauth2-proxy/pull/849).
