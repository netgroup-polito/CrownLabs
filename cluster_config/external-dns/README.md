# external-dns

ExternalDNS synchronizes exposed Kubernetes Services and Ingresses with DNS providers (e.g. bind9).

## Configure bind9

1. Ensure that `/etc/bind/named.conf` containes the following:
    ```
    ...
    include "/etc/bind/named.conf.keys";
    include "/etc/bind/named.conf.local";
    ...
    ```
2. Create a new TSIG key to authenticate external-dns updates and transfers:
    ```sh
    # tsig-keygen -a hmac-sha512 k8s-ladispe-externaldns | tee --append /etc/bind/named.conf.keys
    ```
3. Edit `/etc/bind/named.conf.options`, and authorize the `k8s-ladispe-externaldns` key to perform zone transfers (AXFR queries):
    ```
    allow-transfer {
        ...
        key "k8s-ladispe-externaldns";
    };

    ```
3. Edit `/etc/bind/named.conf.local`, and authorize the `k8s-ladispe-externaldns` key to perform updates for the zone of interest (e.g. `ipv6.polito.it`):
    ```
    zone "ipv6.polito.it" {
        ...
        update-policy {
            grant k8s-ladispe-externaldns zonesub ANY;
        };
    };
    ```
4. Reload the bind9 configuration:
    ```sh
    # rndc reload
    ```


## Deploy external-dns

1. Edit [external-dns.yaml](external-dns.yaml) and replace `<TSIG-key>` (`--rfc2136-tsig-secret=<TSIG-key>`) with the TSIG key previously generated.
2. Deploy external-dns:
    ```sh
    $ kubectl create -f external-dns.yaml
    ```


## References
1. [external-dns](https://github.com/kubernetes-sigs/external-dns)
2. [external-dns - bind9](https://github.com/kubernetes-sigs/external-dns/blob/master/docs/tutorials/rfc2136.md)