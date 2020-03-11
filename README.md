# KubeVirt

KubeVirt is a virtual machine management add-on for Kubernetes. The aim is to provide a common ground for virtualization solutions on top of Kubernetes.

As of today KubeVirt can be used to declaratively

- Create a predefined VM
- Schedule a VM on a Kubernetes cluster
- Launch a VM
- Stop a VM
- Delete a VM

# Installing KubeVirt

To install KubeVirt, the operator and the cr are going to be created with the following commands:

```sh
k8s-test.local# kubectl create -f https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-operator.yaml

k8s-test.local# kubectl create -f https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/kubevirt-cr.yaml
```

The deployment can be checked with the following command:

```sh
k8s-test.local# kubectl get pods -n kubevirt
NAME                               READY   STATUS    RESTARTS   AGE
virt-api-5546d58cc8-5sm4v          1/1     Running   0          16h
virt-api-5546d58cc8-pxkgt          1/1     Running   0          16h
virt-controller-5c749d77bf-cxxj8   1/1     Running   0          16h
virt-controller-5c749d77bf-wwkxm   1/1     Running   0          16h
virt-handler-cx7q7                 1/1     Running   0          16h
virt-operator-6b4dccb44d-bqxld     1/1     Running   0          16h
virt-operator-6b4dccb44d-m2mvf     1/1     Running   0          16h
```

Now that KubeVirt is installed is the right time to download the client tool to interact with th Virtual Machines.

```sh
k8s-test.local# wget -O virtctl https://github.com/kubevirt/kubevirt/releases/download/${KUBEVIRT_VERSION}/virtctl-${KUBEVIRT_VERSION}-linux-amd64

k8s-test.local# chmod +x virtctl

k8s-test.local# ./virtctl
Available Commands:
  console      Connect to a console of a virtual machine instance.
  expose       Expose a virtual machine instance, virtual machine, or virtual machine instance replica set as a new service.
  help         Help about any command
  image-upload Upload a VM image to a PersistentVolumeClaim.
  restart      Restart a virtual machine.
  start        Start a virtual machine.
  stop         Stop a virtual machine.
  version      Print the client and server version information.
  vnc          Open a vnc connection to a virtual machine instance.
```

This step is optional, right now anything related with the Virtual Machines can be done running the virtctl command. In case there’s a need to interact with the Virtual Machines without leaving the scope of the kubectl command, the virt plugin for Krew can be installed following the instructions below:

```sh
k8s-test.local# (
  set -x; cd "$(mktemp -d)" &&
  curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/download/v0.3.1/krew.{tar.gz,yaml}" &&
  tar zxvf krew.tar.gz &&
  ./krew-"$(uname | tr '[:upper:]' '[:lower:]')_amd64" install \
    --manifest=krew.yaml --archive=krew.tar.gz
)
...
Installed plugin: krew
WARNING: You installed a plugin from the krew-index plugin repository.
   These plugins are not audited for security by the Krew maintainers.
   Run them at your own risk.
```

The warning printed by the Krew maintainers can be ignored. To have the krew plugin available, the PATH variable has to be modified:

```sh
k8s-test.local# vim ~/.bashrc
export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"
k8s-test.local# source ~/.bashrc
```
Now, the virt plugin is going to be installed using the krew plugin manager:

```sh
k8s-test.local# kubectl krew install virt
```

# Installing the first Virtual Machine in KubeVirt

```sh
k8s-test.local# kubectl apply -f https://github.com/netgroup-polito/CrownLabs/blob/Amir/Kubevirt/vm.yaml

k8s-test.local# kubectl get vms
NAME        AGE   RUNNING   VOLUME
testvm   13s   false
```

After the Virtual Machine has been created, it has to be started, to do so, the virtctl or the kubectl can be used (depending on what method has been chosen in previous steps).

```sh
k8s-test.local# ./virtctl start testvm
VM vm-cirros was scheduled to start

k8s-test.local# kubectl get vms
NAME        AGE     RUNNING   VOLUME
testvm   7m11s   true
```

Next thing to do is to use the kubectl command for getting the IP address and the actual status of the virtual machines:

```sh
k8s-test.local# kubectl get vmis
kubectl get vmis
NAME        AGE    PHASE        IP    NODENAME
testvm    14s   Scheduling

k8s-test.local# kubectl get vmis
NAME     AGE   PHASE     IP            NODENAME
testvm   63s   Running   10.244.0.15   k8s-test
```

So, finally the Virtual Machine is running and has an IP address. To connect to that VM, the console can be used (./virtctl console testvm) or also a direct connection with SSH can be made:

```sh
k8s-test.local# ssh cirros@10.244.0.15
cirros@10.244.0.15's password: gocubsgo
$ uname -a
Linux testvm 4.4.0-28-generic #47-Ubuntu SMP Fri Jun 24 10:09:13 UTC 2016 x86_64 GNU/Linux
$ exit
```

To stop the Virtual Machine one of the following commands can be executed:


```sh
k8s-test.local# ./virtctl stop testvm
VM testvm was scheduled to stop

k8s-test.local# kubectl virt stop testvm
VM testvm was scheduled to stop
```

![](https://github.com/netgroup-polito/CrownLabs/blob/Amir/Kubevirt/pic/migration-steps.jpg)

# Running KVM VMs in Container Engine

If you’re planning to run a KVM or VMware VM in Container Engine for Kubernetes, you must first convert the disks to a compatible format
such as img, qcow2, iso. 

Before converting procedure you should apply Sysprep on the available image and then convert the to destination format, in order to implement two described steps in KVM environment, you can use available scripts as follows:


### Installation


```sh
$ ./clone-sysprep.sh VM-name
$ ./qemu-images.sh source-image destination-image
```

After the disks are converted, you can make them available to be used in Container Engine for Kubernetes. You have a few options:
  - Upload the disk into the worker nodes and running it with hostpath.
  - Create a Docker image of the raw disk and upload it into a public registry like [Oracle Cloud Infrastructure Registry][df1].
  - Clone a disk and create a persistent volume claim with it.
  
  <img src="https://github.com/netgroup-polito/CrownLabs/blob/Amir/Kubevirt/pic/import-disk.jpg" width="600" height="100" />
  

# VM defination

- Depending on original VM configuration,writing VM yaml file could be tough.
- Translation of old VM configuration to new VM yaml is done manually.

![](https://github.com/netgroup-polito/CrownLabs/blob/Amir/Kubevirt/pic/VM-yaml.jpg)


# Service defination

- All solutions of Service Discovery of Kubernetes shall work with KubeVirt VMs too

![](https://github.com/netgroup-polito/CrownLabs/blob/Amir/Kubevirt/pic/service-defination.jpg)



All of these options are explained in the [KubeVirt GitHub repo][df2] and [KubeVirt documentation][df3].


   [df1]: <https://docs.cloud.oracle.com/en-us/iaas/Content/Registry/Concepts/registryoverview.htm>
   [df2]: <https://github.com/kubevirt/kubevirt/>
   [df3]: <https://kubevirt.io/user-guide/#/creation/creating-virtual-machines>

