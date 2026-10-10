# Local Keycloak + OIDC — end-to-end setup

This section presents the steps needed to bring up Keycloak locally: the realm, the clients, the scopes, and the base users.
It also configures the Kubernetes API server to validate the tokens Keycloak issues.

## 0. What is keycloak

_This section explains some basic concepts and terms about keycloak.
If you are already familiar with keycloak, feel free to skip it._

Keycloak is an open source identity and access management solution.
In basic terms, it's what allows you to login once, and then authorizes you to other services &ndash; like browsing the web pages, creating VMs, ...

Here is a list of useful terms when working with keycloak:

- **group**: a group of users.
- **client**: an entity (application or service) that "does something". This can be either requesting a user's access token for other purposes, or modifying user rights within keycloak itself.
- **client scope**: describing how a client should interact with keycloak. (For example, should authentication be mandatory, or optional?)
- **realm**: a set of users, credentials, roles, and groups.
  When you have non-admin credentials, they belong to a specific realm, and allow you to login only on that realm.
  The concept of realms allows you to have the same keycloak instance manage multiple separate applications.

## 1. Setting up keycloak

In order to deploy keycloak, we need to define:

- a `Service` to accept incoming HTTP connections (remember that HTTPS is terminated earlier by envoy, so we don't have to worry about certificates here)
- an `HTTPRoute`, that allows route `keycloak.crownlabs.local` traffic towards the service
- a discovery `Service`
- a `StatefulSet` to deploy all the containers required by keycloak
- a `Deployment` and a `Service` for a deployment, ephemeral PostgreSQL database.

All these components are prepared in the `manifest/keycloak.yaml` configuration file.

However, if we just run that, we would have an empty keycloak.
That means, it does not have the crownlabs realm, nor it has any user pre-inserted.

For convenience, a basic realm was already prepared and exported.
It already has the correct realm, with a test user and all required settings for sending verification emails.
All its data is held as JSON inside the `manifest/crownlabs-real-configmap.yaml`.
This file should therefore be applied before the previous.

Both files can be applied together using the following command:

```bash
kubectl apply -f dev-local/keycloak/manifests
```

You can check that keycloak has been deployed correctly with the following command:

```bash
kubectl rollout status statefulset/keycloak
```

## 2. Using keycloak

The admin dashboard is now reachable at `https://keycloak.crownlabs.local:8443`.
Credentials to access it are `admin`/`admin`.
In the dashboard you can see everything that was imported from the basic CrownLabs-preset realm, with the ability to change it.
Once you will have a running frontend, you may also try to create a new Tenant, and it will appear among the users in this dashboard.

_Note: this is likely the first time you access any HTTPS website served using the self-signed certificate.
Due to its self-signed nature, the browser will not recognize it.
You will therefore not reach the page, and have a `ERR_CERT_AUTHORITY_INVALID` error.
You can safely continue on the website ("Advanced" &rarr; "Proceed").
This should be remembered by your browser, so when you open other websites signed with the same certificate, you will not get the error anymore._

On top of the admin account, the imported realm includes three other credential sets:

- a user account (`john.doe`/`johndoe123`, with already verified email, in the `crownlabs` realm).
  At [this link ](https://keycloak.crownlabs.local:8443/realms/crownlabs/account/) you can check that the credentials are accepted.
  However, after the login you will get an error, since the server where you should be redirected to is not present.
  Once the frontend is running, you will be able to use these credentials to login as a normal user in the website.
- credentials for a service account (`operator-local`/`operator-local-dev-secret`).
  Those are used by the main CrownLabs operator to do all the required operations on the keycloak service
- a `k8s` client (no secret), where the user authenticates himself.

_Please note that these credentials are known values, deliberately documented for local development.
For security reasons, do not reuse them in real environments._

## 3. Configure the K3s API server for OIDC

While using the development environment, the system may have to run some commands on K3s.
For example, if you start a VM from the frontend, it needs to execute the kubectl command to actually start it inside K3s.
This operation is internally done via the command `kubectl proxy`.

By default, this command authenticates you by means of your kubeconfig file.
Since your kubeconfig is also the admin of the K3s cluster, all requests will be executed as admin.
However, this is not the behavior we have on the real CrownLabs server: for example, a user can only start VMs within the workspaces they are in.

To have a more realistic environment, we need to instruct the K3s cluster to use authentication from keycloak instead of the default.
This is the purpose of this step of the guide.

Fully explaining the rationale of every action is relatively complex.
For this reason, the guide only lists the required steps, along with some very basic motivation.
The full rationale, verification steps and troubleshooting table can be found in the file [`apiserver-oidc-integration.md`](apiserver-oidc-integration.md): please refer to it if something in this guide does not behave as expected.

### 3.0. Checking that user authorization is currently not setup

_Note: in this step, nothing fundamental will be achieved.
We will setup a kubeconfig to access the cluster as user, and then we will see that it is not authorized to execute any command.
Feel free to skip this step if you don't want to set it up, or do not need to check the before and after._

As for the real CrownLabs instance, our local setup can be accessed also via kubectl commands.
To test this, we can create a new kubeconfig, to login as John Doe.
Please refer to [the base k3s guide, section 2.3](../base-k3s/README.md#23-instructing-kubectl-to-use-the-file-single-kubeconfig) if you need help to have multiple kubeconfigs available at the same time.

With the official "[Accessing a CrownLabs Sandbox Namespace](https://crownlabs.polito.it/resources/sandbox/)" guide, we can set up everything needed to connect.
Since it is a local instance, some links in the yaml file may have to be changed.
For convenience, the following is a fully ready kubeconfig file (you may still need to follow the initial installation steps in the official guide):

<!-- prettier-ignore-start -->

```yaml
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://apiserver.crownlabs.local:8443
  name: k3s_JohnDoe
contexts:
- context:
    cluster: k3s_JohnDoe
    namespace: tenant-john-doe
    user: k3s_JohnDoe-oidc
  name: k3s_JohnDoe
current-context: k3s_JohnDoe
users:
- name: k3s_JohnDoe-oidc
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      args:
      - oidc-login
      - get-token
      - --oidc-issuer-url=https://keycloak.crownlabs.local:8443/realms/crownlabs
      - --oidc-client-id=k8s
      - --skip-open-browser
      - --username=john.doe
      - --password=johndoe123
      command: kubectl
```

<!-- prettier-ignore-end -->

_Note: while working with this new kubeconfig context, remember that the main k3s context (likely called `default`) must be active when launching `WSL_bridge.sh`._

As user (aka with this new kubeconfig), you can use kubectl to get info about your own tenant, with the following command:

```bash
kubectl get tenant.crownlabs.polito.it john.doe

# error: You must be logged in to the server (the server has asked for the client to provide credentials)
```

If you run it now, you will get an error, as reported above.
After executing the next steps, this command will instead work successfully, authenticating John Doe and reporting his tenant's data.

### 3.1. Considering the root certificate as trusted

In the Envoy guide, we set up a local certification authority for the certificates we need.
Since we will use the same for the K3s cluster, we need to have our system trust it.

The first thing to do is to extract the root certificate from the CA, and add it to the trusted certificates of our system:

```bash
kubectl get secret crownlabs-tls -n default -o jsonpath='{.data.ca\.crt}' | base64 -d | sudo tee /usr/local/share/ca-certificates/crownlabs-ca.crt
sudo update-ca-certificates
```

> **SECURITY NOTE**
>
> With this command, we added the root certificate as trusted on our system.
> Should anyone get access to this certificate, they can produce "invalid" certificates that will be trusted by our device.
> An example could be a replica of our bank website, that seems completely legit, as it is signed by a "trustworthy" authority.
>
> While this eventuality is relatively unlikely &ndash; an attacker would need to first get access to our machine &ndash; it must be taken into account.
>
> A possible way to reduce the threat is to keep the certificate in the folder only while working on the system.
> When it is not needed anymore, the certificate can be removed locally from the machine with the following commands:
>
> ```bash
> sudo rm /usr/local/share/ca-certificates/crownlabs-ca.crt
> sudo update-ca-certificates
> ```
>
> When it is needed again, it can be re-installed using the command described above the security note.

### 3.2. Instructing K3s to trust this certificate

After installing the root certificate, we need to instruct K3s to trust OIDC requests signed by the root certificate itself:

```bash
sudo mkdir -p /etc/rancher/k3s/config.yaml.d
sudo tee /etc/rancher/k3s/config.yaml.d/30-keycloak-oidc.yaml > /dev/null <<'EOF'
kube-apiserver-arg:
  - "oidc-issuer-url=https://keycloak.crownlabs.local:8443/realms/crownlabs"
  - "oidc-client-id=k8s"
  - "oidc-username-claim=preferred_username"
  - "oidc-username-prefix=-"
  - "oidc-groups-claim=groups"
  - "oidc-groups-prefix=kubernetes:"
EOF
```

After requesting the changes, we have to restart the K3s cluster to have them apply.
This can be done with the following commands.
Please note that any process with an active connection to the K3s cluster will either stop, or need to be restarted (for example, `WSL_bridge`).

```bash
sudo systemctl restart k3s
journalctl -u k3s -f   # Confirm there is no "invalid authentication configuration" error, then press Ctrl+C.
```

> Note: it should not be necessary, but in case of errors the following line can be added to the end of the file:
>
> ```yaml
> - "oidc-ca-file=/usr/local/share/ca-certificates/crownlabs-ca.crt"
> ```

## Final checks

Now, everything should be setup.
Running the same command as before should now report data about the tenant John Doe:

```bash
# ensure you are using the john.doe kubeconfig

kubectl get tenant.crownlabs.polito.it john.doe

# NAME       FIRST NAME   LAST NAME   READY   AGE
# john.doe   John         Doe         true    60d
```

## Exporting the realm

As previously stated, this setup already has an import of a functional keycloak realm, fully setup for the CrownLabs development.
Should you modify it (for example, adding a new user), you may want to "freeze" and "save" this new state.
This can be done following the [`regenerating-the-realm-export.md`](regenerating-the-realm-export.md) guide.

Note that the realm is imported automatically from the ConfigMap _only_ if the same realm does not yet exist on the database.
This means that the first import works fine, but if you modify the ConfigMap alone, that would not be re-imported (as the realm is already there).
To re-import the realm from scratch, you have to delete it, with one of the following methods:

- delete the `crownlabs` realm from the admin console
- wipe the entire Postgres database
- `DELETE /admin/realms/crownlabs`
