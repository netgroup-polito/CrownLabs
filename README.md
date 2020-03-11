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

