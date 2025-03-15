# Certificate Provisioning - cert-manager

[cert-manager](https://github.com/jetstack/cert-manager) is a Kubernetes add-on to automate the management and issuance of TLS certificates from various issuing sources (e.g. [Let's Encrypt](https://letsencrypt.org/)).

In order to get a certificate from Let's Encrypt, it is necessary to prove the control of the domain names in that certificate through a *challenge*, as defined by the ACME standard. Two different types of challenges are available [[1]](https://letsencrypt.org/docs/challenge-types/),[[2]](https://cert-manager.io/docs/configuration/acme/):

* HTTP-01 challenge: requires to put a specific value in a file on the web server at a specific path;
* DNS-01 challenge: requires to put a specific value in a TXT record under that domain name.

In the following, both types of challenges are configured and made available. Indeed, the former is easier to use but it requires a public IP and port 80 to be open and reachable. The latter, on the other hand, does not suffer from this limitation and allows to issue wildcard certificates, but it can be used only for the DNS names under our control.

## Configure bind9

1. Ensure that `/etc/bind/named.conf` contains the following:

    ```txt
    ...
    include "/etc/bind/named.conf.keys";
    include "/etc/bind/named.conf.local";
    ...
    ```

2. Create the new TSIG keys to authenticate cert-manager record updates (one for each zone):

    ```sh
    tsig-keygen -a hmac-sha512 crownlabs-certmanager | sudo tee --append /etc/bind/named.conf.keys
    tsig-keygen -a hmac-sha512 crownlabs-internal-certmanager | sudo tee --append /etc/bind/named.conf.keys
    ```

3. Edit `/etc/bind/named.conf.local`, and authorize the keys to insert TXT records for the zones of interest:

    ```txt
    zone "crownlabs.polito.it" {
        ...
        update-policy {
            grant crownlabs-certmanager zonesub txt;
        };
    };
    ...
    zone "internal.crownlabs.polito.it" {
        ...
        update-policy {
            grant crownlabs-internal-certmanager zonesub txt;
        };
    };
    ```

4. Reload the `bind9` configuration:

    ```sh
    sudo rndc reload
    ```

## Deploy cert-manager

1. Install `cert-manager` through Helm:

    ```sh
    helm repo add jetstack https://charts.jetstack.io
    helm upgrade cert-manager jetstack/cert-manager --namespace cert-manager \
        --install --create-namespace --values configurations/cert-manager-values.yaml
    ```

2. Create the `ClusterIssuer` resources, which operate as an interface to request the issuance of digital certificates. In the following, three different resources are created, respectively for self-signed certificates, development (`letsencrypt-staging`) and production (`letsencrypt-production`). Indeed, the latter is associated with stricter rate limits [[3]](https://letsencrypt.org/docs/rate-limits/):

    ```sh
    kubectl create -f configurations/self-signed.yaml
    kubectl create -f configurations/lets-encrypt-issuer-staging.yaml
    kubectl create -f configurations/lets-encrypt-issuer-production.yaml
    ```

    *Note:* in case a different DNS server is used, it is necessary to edit the `yaml` files with the correct configuration. Additionally, it is also possible to configure the email address that will be associated with the digital certificates issued by Let's Encrypt.

3. Create the new `Secret` resources storing the TSIG keys previously generated to interact with the DNS server:

    ```sh
    kubectl -n cert-manager create secret generic crownlabs-certmanager-tsig --from-literal=crownlabs-certmanager-tsig-key=<TSIG-key>
    kubectl -n cert-manager create secret generic crownlabs-internal-certmanager-tsig --from-literal=crownlabs-internal-certmanager-tsig-key=<TSIG-key>
    ```

4. Verify that the `ClusterIssuer` resources are in Ready state:

    ```sh
    kubectl describe ClusterIssuer
    ```

## Use cert-manager

### Secure Ingress Resources

A common use-case for cert-manager is requesting TLS signed certificates to secure `Ingress` resources. This operation can be performed automatically adding an ad-hoc annotation to the `Ingress` resource pointing to the `ClusterIssuer` resource of interest:

```yaml
...
annotations:
    cert-manager.io/cluster-issuer: letsencrypt-production
...
```

A valid certificate associated with the `Ingress` host is automatically generated and stored within the secret pointed by the `tls.secretName` field of the `Ingress` resource.

### Certificate Resources

A `Certificate` resource can be created to manually request the issuance of a digital certificate for a specific host name belonging to the DNS zone under control. [certificate-example.yaml](certificate-example.yaml) provides an example configuration to request a certificate; please refer to the official documentation [[4]](https://cert-manager.io/docs/usage/certificate/) for more information.

### Select the type of challenge to use (i.e. HTTP-01 or DNS-01)

By default, cert-manager has been configured to prove the control of the domain names to be certified using HTTP-01 challenges. To use DNS-01 challenges, instead, it is necessary to add the ad-hoc label to the `Ingress` or `Certificate` resource:

```yaml
labels:
    use-dns01-solver: "true"
```

## Synchronize digital certificates between namespaces
❗❗ `Kubed is no longer available and has been superseded by ConfigSyncer` 

In different scenarios, it may happen to have different `Ingress` resources in different namespaces which refer to the same domain (with different paths). Unfortunately, annotating all these ingresses with the `cert-manager.io/cluster-issuer` annotation soon leads to hitting the Let's Encrypt rate limits. Hence, it is necessary to introduce some mechanism to synchronize the secret generated between multiple namespaces. One of the projects currently providing a solution to this problem is [kubed](https://github.com/appscode/kubed).

### Install kubed

Kubed can be easily installed with helm [[5]](https://web.archive.org/web/20230605163413/https://appscode.com/products/kubed/v0.12.0/setup/install/).

```bash
helm repo add appscode https://charts.appscode.com/stable/
helm repo update

helm install kubed appscode/kubed \
  --version v0.12.0 \
  --namespace kubed \
  --create-namespace \
  --set enableAnalytics=false \
  --set config.clusterName="crownlabs"
```

### Secret synchronization

Once kubed is installed, secrets can be duplicated in multiple namespaces, and kept synchronized, by adding the ad-hoc annotation [[6]](https://cert-manager.io/v1.1-docs/faq/kubed/#syncing-arbitrary-secrets-across-namespaces-using-kubed):

```yaml
...
annotations:
    kubed.appscode.com/sync: "namespace-label=value"
...
```

Warning, the kubed annotation needs to applied to the `Secret` created by the certificate and not to the `Certificate` itself.

## Additional References

1. [cert-manager documentation](https://cert-manager.io/docs/)
2. [cert-manager configuration with HTTP01](https://cert-manager.io/docs/configuration/acme/http01/)
3. [cert-manager configuration with DNS01](https://cert-manager.io/docs/configuration/acme/dns01/)
4. [cert-manager usage](https://cert-manager.io/docs/usage/)
