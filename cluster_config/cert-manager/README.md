# cert-manager

[cert-manager](https://github.com/jetstack/cert-manager) is a Kubernetes add-on to automate the management and issuance of TLS certificates from various issuing sources (e.g. [Let's Encrypt](https://letsencrypt.org/)).

In order to get a certificate from Let's Encrypt, it is necessary to prove the control of the domain names in that certificate through a *challenge*, as defined by the ACME standard. Two different types of challenges are available [[1]](https://letsencrypt.org/docs/challenge-types/),[[2]](https://cert-manager.io/docs/configuration/acme/):

* HTTP-01 challenge: requires to put a specific value in a file on the web server at a specific path;
* DNS-01 challenge: requires to put a specific value in a TXT record under that domain name.

In the following, the DNS-01 challenge is adopted. Although being a bit more complex to configure, it is more flexible (does not require the web server to be accessible at port 80) and allows to issue wildcard certificates.

## Configure bind9

1. Ensure that `/etc/bind/named.conf` contains the following:
    ```
    ...
    include "/etc/bind/named.conf.keys";
    include "/etc/bind/named.conf.local";
    ...
    ```
2. Create a new TSIG key to authenticate external-dns updates and transfers:
    ```sh
    # tsig-keygen -a hmac-sha512 k8s-ladispe-cert-manager | tee --append /etc/bind/named.conf.keys
    ```
3. Edit `/etc/bind/named.conf.local`, and authorize the `k8s-ladispe-cert-manager` key to insert TXT records for the zone of interest (e.g. `crown-labs.ipv6.polito.it`):
    ```
    zone "ipv6.polito.it" {
        ...
        update-policy {
            grant k8s-ladispe-cert-manager wildcard *.crown-labs.ipv6.polito.it. txt;
        };
    };
    ```
4. Reload the `bind9` configuration:
    ```sh
    # rndc reload
    ```


## Deploy cert-manager

1. Install the `CustomResourceDefinitions` and `cert-manager` itself:
    ```sh
    $ kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v0.14.0/cert-manager.yaml
    ```
2. Create the `ClusterIssuer` resources, which operate as an interface to request the issuance of digital certificates. In the following, two different resources are created, respectively for development (`letsencrypt-staging`) and production (`letsencrypt-production`). Indeed, the latter is associated with stricter rate limits [[3]](https://letsencrypt.org/docs/rate-limits/):
    ```sh
    $ kubectl create -f lets-encrypt-issuer-staging.yaml
    $ kubectl create -f lets-encrypt-issuer-production.yaml
    ```
    *Note:* in case a different DNS server is used, it is necessary to edit the `yaml` files with the correct configuration. Additionally, it is also possible to configure the email address that will be associated with the digital certificates issued by Let's Encrypt.
3. Create a new `Secret` resource storing the TSIG key previously generated to interact with the DNS server:
    ```sh
    $ kubectl -n cert-manager create secret generic cert-manager-tsig-secret --from-literal=cert-manager-tsig-secret-key=$(base64 <TSIG-key>)
    ```
4. Verify that the `ClusterIssuer` resources are in Ready state:
    ```sh
    $ kubectl describe ClusterIssuer -n cert-manager
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

## Additional References
1. [cert-manager documentation](https://cert-manager.io/docs/)
2. [cert-manager configuration with DNS01](https://cert-manager.io/docs/configuration/acme/dns01/)
3. [cert-manager usage](https://cert-manager.io/docs/usage/)