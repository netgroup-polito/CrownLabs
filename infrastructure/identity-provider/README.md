# Identity Provider - Keycloak

Keycloak is an Open Source Identity and Access Management solution for modern Applications and Services.
It enables to concentrate all the tasks related to identity and access management into the same place; once authenticated, a user session can be associated to a token that can be used to validate the access of all the resources available in the cluster.

This brief guide presents how to install Keycloak in HA in a K8S cluster with a PostgreSQL Database backend (also in HA).

More info at [Keycloak's website](https://www.keycloak.org)
## Pre-requisites
Here we assume that in the K8S cluster the following operators are installed and configured:
* [ROOK](https://rook.io/)
* [NGINX Ingress Controller](https://www.nginx.com/products/nginx/kubernetes-ingress-controller/)
* [cert-manager](https://cert-manager.io/)
* A namespace in K8S cluster called **keycloak-ha**

You will need the following tools installed in your workstation:
* [Helm](https://helm.sh/)

## PostgreSQL-Operator
The following steps will install the postgresql-operator in the namespace called **keycloak-ha**.
The Postgres Operator can be installed simply by applying `yaml` manifests, after properly changing the namespace in file [operator-service-account-rbac.yaml](manifests/operator-service-account-rbac.yaml) for the `service account` and `cluster rolebinding`.

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

## TLS-Certificate
Apply the the manifest [cert-manager-keycloak-certificate-request.yaml](manifests/cert-manager-keycloak-certificate-request.yaml) to get a certificate for the keycloak domain.

```bash
kubectl create -f manifests/cert-manager-keycloak-certificate-request.yaml -n keycloak-ha
```

This command will create a `tls-secret` named `keycloak-certificate-secret`. We will need it during the keycloak server deployment.

## Keycloak Server
First of all get the helm charts from [Codecentric](https://github.com/codecentric/helm-charts/tree/master/charts/keycloak):

```bash
#add the codecentric helm repository
helm repo add codecentric https://codecentric.github.io/helm-charts
#download the codecentric/keycloak charts on your pc
helm pull codecentric/keycloak
```

After extracting the archive, replace file `keycloak/values.yaml` with `conf-files/keycloak-values.yaml`.
Note, you need to rename the later one in `values.yaml`.

The following are some changes that have bene done to user the resources deployed before in this guide:

```yaml
# add the volume and volumeMount of the tls-certificate and of the custom template for keycloak in values.yaml file
extraVolumes: |
  - name: keycloak-tls-certificate
    secret:
      defaultMode: 420
      secretName: keycloak-tls-certificate-secret
  - configMap:
      defaultMode: 420
      name: keycloak-theme-email
    name: keycloak-theme-email
  - configMap:
      defaultMode: 420
      name: keycloak-theme-email-messages
    name: keycloak-theme-email-messages
  - configMap:
      defaultMode: 420
      name: keycloak-theme-email-html
    name: keycloak-theme-email-html
  - configMap:
      defaultMode: 420
      name: keycloak-theme-email-text
    name: keycloak-theme-email-text
extraVolumeMounts: |
  - mountPath: /etc/x509/https
    name: keycloak-tls-certificate
    readOnly: true
  - mountPath: /opt/jboss/keycloak/themes/crownlabs/email
    name: keycloak-theme-email
    readOnly: true
  - mountPath: /opt/jboss/keycloak/themes/crownlabs/email/messages
    name: keycloak-theme-email-messages
    readOnly: true
  - mountPath: /opt/jboss/keycloak/themes/crownlabs/email/html
    name: keycloak-theme-email-html
    readOnly: true
  - mountPath: /opt/jboss/keycloak/themes/crownlabs/email/text
    name: keycloak-theme-email-text
    readOnly: true
```

Set the number of replicas of the server according to your preferences:
```yaml
keycloak:
  replicas: 3
```

Set the database config to use the PostgreSQL Cluster deployed before:

```yaml
  persistence:
    # If true, the Postgres chart is deployed
    deployPostgres: false
    # The database vendor. Can be either "postgres", "mysql", "mariadb", or "h2"
    dbVendor: postgres
    ## The following values only apply if "deployPostgres" is set to "false"
    dbName: keycloak
    dbHost: keycloak-db-cluster
    dbPort: 5432
    ## Database Credentials are loaded from a Secret residing in the same Namespace as keycloak.
    ## The Chart can read credentials from an existing Secret OR it can provision its own Secret.
    ## Specify existing Secret
    # If set, specifies the Name of an existing Secret to read db credentials from.
    existingSecret: "keycloak.keycloak-db-cluster.credentials"
    existingSecretPasswordKey: "password"  # read keycloak db password from existingSecret under this Key
    existingSecretUsernameKey: "username"  # read keycloak db user from existingSecret under this Key```
```

For more information visit the following [page](https://hub.docker.com/r/jboss/keycloak/).

### Install keycloak server
Type the following command:

```bash
helm install keycloak-server keycloak/ --namespace keycloak-ha
```

Now, check that the new pods are up and running. Once everything has gone smooth, apply the manifest to the service:

```bash
#apply manifests/keycloak-ingress.yaml in order to reach keycloak from outside.
kubectl create -f manifests/keycloak-ingress.yaml -n keycloak-ha
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


## Accessing K8S cluster using Keycloak as authentication server
In order to start interacting with your Kubernetes cluster, you will use a command line tool called **kubectl**. You will need to install kubectl on your local machine.

A **kubeconfig** file is a file used to configure access to Kubernetes when used in conjunction with the kubectl commandline tool.

For more details on how kubeconfig and kubectl work together, see the [Kubernetes documentation](https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/).

### Pre-requisite
You should have kubectl installed at a version compatible to your cluster.


#### Krew
First, you should install [Krew](https://krew.sigs.k8s.io/), which facilitates the use of kubectl plugins.
Here there is the commands for Linux (Bash/Zsh):

```
(
  set -x; cd "$(mktemp -d)" &&
  curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/latest/download/krew.{tar.gz,yaml}" &&
  tar zxvf krew.tar.gz &&
  KREW=./krew-"$(uname | tr '[:upper:]' '[:lower:]')_amd64" &&
  "$KREW" install --manifest=krew.yaml --archive=krew.tar.gz &&
  "$KREW" update
)
 ```

Other configurations are available on the krew official documentation.
In addition you have to enable krew by adding the following to your PATH:
```
export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"
```

To make persistent this modification, you should add permanently the previous configuration to your `bashrc/zshrc`.


## OIDC-Login
First we have to install OIDC login plugin, which enables a single sign on (SSO) to a Kubernetes cluster and other development tools:
```
kubectl krew install oidc-login
```

Now, we can proceed to use your cluster.


## Login
Once, you have created your user in your Keycloak instance, you can configure `oidc-login` by setting your credentials.
This could be done in two different ways:

1. You can use a redirect via-browser to login by putting your user/password in the Identity Provider website and store only the temporary token in your `kubeconfig`.
2. (or) You can set your username and password directly in `kubeconfig`.

```
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVrakNDQTNxZ0F3SUJBZ0lRQ2dGQlFnQUFBVk9GYzJvTGhleW5DREFOQmdrcWhraUc5dzBCQVFzRkFEQS8KTVNRd0lnWURWUVFLRXh0RWFXZHBkR0ZzSUZOcFoyNWhkSFZ5WlNCVWNuVnpkQ0JEYnk0eEZ6QVZCZ05WQkFNVApEa1JUVkNCU2IyOTBJRU5CSUZnek1CNFhEVEUyTURNeE56RTJOREEwTmxvWERUSXhNRE14TnpFMk5EQTBObG93ClNqRUxNQWtHQTFVRUJoTUNWVk14RmpBVUJnTlZCQW9URFV4bGRDZHpJRVZ1WTNKNWNIUXhJekFoQmdOVkJBTVQKR2t4bGRDZHpJRVZ1WTNKNWNIUWdRWFYwYUc5eWFYUjVJRmd6TUlJQklqQU5CZ2txaGtpRzl3MEJBUUVGQUFPQwpBUThBTUlJQkNnS0NBUUVBbk5NTThGcmxMa2UzY2wwM2c3Tm9ZekRxMXpVbUdTWGh2YjQxOFhDU0w3ZTRTMEVGCnE2bWVOUWhZN0xFcXhHaUhDNlBqZGVUbTg2ZGljYnA1Z1dBZjE1R2FuL1BRZUdkeHlHa09sWkhQL3VhWjZXQTgKU014K3lrMTNFaVNkUnh0YTY3bnNIamNBSEp5c2U2Y0Y2czVLNjcxQjVUYVl1Y3Y5YlR5V2FOOGpLa0tRRElaMApaOGgvcFpxNFVtRVVFejlsNllLSHk5djZEbGIyaG9uemhUK1hocSt3M0JydmF3MlZGbjNFSzZCbHNwa0VObldBCmE2eEs4eHVRU1hndm9wWlBLaUFsS1FUR2RNRFFNYzJQTVRpVkZycW9NN2hEOGJFZnd6Qi9vbmt4RXowdE52amoKL1BJemFyazVNY1d2eEkwTkhXUVdNNnI2aENtMjFBdkEySDNEa3dJREFRQUJvNElCZlRDQ0FYa3dFZ1lEVlIwVApBUUgvQkFnd0JnRUIvd0lCQURBT0JnTlZIUThCQWY4RUJBTUNBWVl3ZndZSUt3WUJCUVVIQVFFRWN6QnhNRElHCkNDc0dBUVVGQnpBQmhpWm9kSFJ3T2k4dmFYTnlaeTUwY25WemRHbGtMbTlqYzNBdWFXUmxiblJ5ZFhOMExtTnYKYlRBN0JnZ3JCZ0VGQlFjd0FvWXZhSFIwY0RvdkwyRndjSE11YVdSbGJuUnlkWE4wTG1OdmJTOXliMjkwY3k5awpjM1J5YjI5MFkyRjRNeTV3TjJNd0h3WURWUjBqQkJnd0ZvQVV4S2V4cEhzc2NmcmI0VXVRZGYvRUZXQ0ZpUkF3ClZBWURWUjBnQkUwd1N6QUlCZ1puZ1F3QkFnRXdQd1lMS3dZQkJBR0MzeE1CQVFFd01EQXVCZ2dyQmdFRkJRY0MKQVJZaWFIUjBjRG92TDJOd2N5NXliMjkwTFhneExteGxkSE5sYm1OeWVYQjBMbTl5WnpBOEJnTlZIUjhFTlRBegpNREdnTDZBdGhpdG9kSFJ3T2k4dlkzSnNMbWxrWlc1MGNuVnpkQzVqYjIwdlJGTlVVazlQVkVOQldETkRVa3d1ClkzSnNNQjBHQTFVZERnUVdCQlNvU21wakJIM2R1dWJST2JlbVJXWHY4Nmpzb1RBTkJna3Foa2lHOXcwQkFRc0YKQUFPQ0FRRUEzVFBYRWZOaldEamRHQlg3Q1ZXK2RsYTVjRWlsYVVjbmU4SWtDSkx4V2g5S0VpazNKSFJSSEdKbwp1TTJWY0dmbDk2UzhUaWhSelp2b3JvZWQ2dGk2V3FFQm10enczV29kYXRnK1Z5T2VwaDRFWXByLzF3WEt0eDgvCndBcEl2SlN3dG1WaTRNRlU1YU1xclNERTZlYTczTWoydGNNeW81ak1kNmptZVdVSEs4c28vam9XVW9IT1Vnd3UKWDRQbzFRWXorM2RzemtEcU1wNGZrbHhCd1hSc1cxMEtYelBNVForc09QQXZleXhpbmRtamtXOGxHeStRc1JsRwpQZlorRzZaNmg3bWplbTBZK2lXbGtZY1Y0UElXTDFpd0JpOHNhQ2JHUzVqTjJwOE0rWCtRN1VOS0VrUk9iM042CktPcWtxbTU3VEgySDNlREpBa1NuaDYvRE5GdTBRZz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K
    server: https://apiserver.crown-labs.ipv6.polito.it
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    namespace: default
    user: oidc
  name: kubernetes
current-context: kubernetes
kind: Config
preferences: {}
users:
- name: oidc
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      args:
      - oidc-login
      - get-token
      - --oidc-issuer-url=https://auth.crown-labs.ipv6.polito.it/auth/realms/crownlabs
      - --oidc-client-id=k8s
      - --oidc-client-secret=229a9d87-2bae-4e9b-8567-e8864b2bac4b
      - --skip-open-browser
      - --username=<Username>
      - --password=<Password>
      command: kubectl
      env: null
```
