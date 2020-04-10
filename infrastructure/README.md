# Infrastructure
Here we summarize all the components to be installed in the Kubernetes cluster. These services
are fully documented in their own web pages and here we provide some config files tuned for
our infrastructure.
​
In each subfolder you will find a **README** file, which explains the actions carried out in order to
install the service, along with the configuration files. Our documentation is not meant to substitute
the official ones. Instead, it provides a reference about the main components to be installed in order to
make *CrownLabs* up and running. Yet, it is clearly dependent on our setup and the constraints we had to face.

There are **multiple ways** to install and run the applications in your infrastructure, but if you
choose to follow our procedures keep in mind that probably you will need to **adapt** the configuration
files to your own setup.
​
Obviously some services rely upon some others. Hence, the latter have to be installed first. Install them
in this order and you should minimize the interdependencies:
​
## [Networking](kubernetes-networking)
We used the Calico CNI to provide the networking inside our Cluster. It is not mandatory, if you already have another
CNI configured in your cluster stick with it, but configure the services to work within your environment.
​
## [Load Balancing](load-balancing)
The choice for our load-balancer went to MetalLB, which is a load-balancer implementation for bare metal
Kubernetes clusters, using standard routing protocols. If you’re running on a supported IaaS platform then skip
this step and use one of your provider.
​
## [Ingress Controller](ingress-controller)
To expose HTTP and HTTPS routes from outside the cluster to services within the cluster we use NGINX Ingress Controller,
which is a community-drive effort.
​
## [External DNS](external-dns-synchronization)
ExternalDNS makes Kubernetes resources discoverable via public DNS servers (bind9).
​
## [Certificate Provisioning](certificate-provisioning)
All our routes from outside to the cluster services goes through HTTPS and cert-manager is the operator
that makes that possible. Indeed, it automates the management and issuance of TLS certificates from Let's Encrypt.
​
## [Storage Provisioning](storage-provisioning)
A lot of application rely on persistent volumes to operate and store their configurations. We choose
the ROOK operator to manage the storage in our cluster.
​
## [Private Registry](docker-registry)
An internal docker-registry is used to store the VM images, in order to limit the internet traffic outside
the cluster and improve the user experience during the instantiation of new VMs.
​
## [Identity Provider](identity-provider)
When the matter camse to authentication and authorization of the users we went with Keycloak identity
provider. Thanks to its advanced features and rich documentation it was a breeze to configure.
​
## [Monitoring](monitoring)
All the above mentioned services have to be monitored and kept healthy, here Prometheus and
Grafana and AlertManager come into play.
