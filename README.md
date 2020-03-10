
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
  
  
  
  

All of these options are explained in the [KubeVirt GitHub repo][df2] and [KubeVirt documentation][df3].




   [df1]: <https://docs.cloud.oracle.com/en-us/iaas/Content/Registry/Concepts/registryoverview.htm>
   [df2]: <https://github.com/kubevirt/kubevirt/>
   [df3]: <https://kubevirt.io/user-guide/#/creation/creating-virtual-machines>
