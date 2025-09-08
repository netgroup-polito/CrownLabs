## Physical node selection

When starting a new VM/container instance, CrownLabs selects automatically the physical node where the instance has to be executed.

However, this behaviour can be changed, hence allowing users to select exactly the node (or the set of nodes) where the instance has to be executed.
This can be achieved by defining (1) the proper label selectors on the nodes, and (2) a specific _spec_ on the VM/container template, as follows.

### Add the proper label `crownlabs.polito.it/mylabel` to the worker node

This is a possible example:

    admin@k8s-master:~$ kubectl get node worker-2 -o yaml |more
    apiVersion: v1
    kind: Node
    metadata:
      annotations:
        projectcalico.org/IPv4Address: 192.168.24.25/26
    creationTimestamp: "2021-04-29T16:33:48Z"
    labels:
      beta.kubernetes.io/arch: amd64
      cpumanager: "false"
      crownlabs.polito.it/gpu-available: "true"
      crownlabs.polito.it/node-size: big
      crownlabs.polito.it/node-name: worker-2
    ...

In the above example, the node has three labels, one boolean (if a GPU is available ot not), the other which tells the size of the node (their values, in this case 'big', are arbitrary), and the third one that keeps the name of the worker.

To add a new label to the node, you can simply use the following command:

    kubectl label nodes worker-3 "crownlabs.polito.it/node-name=worker-3"

This adds label `crownlabs.polito.it/node-name: worker-3` to the `worker-3` node.

To remove the above label, you can use the following example:

    kubectl label nodes worker-3 "crownlabs.polito.it/node-name"-


### Add the proper specification in the VM/container template

This can be done as follows (this refers to a template called `vscode-rust` in namespace `workspace-sid`):

    admin@k8s-master:~$ kubectl get templates -n workspace-sid vscode-rust -o yaml
    apiVersion: crownlabs.polito.it/v1alpha2
    kind: Template
    metadata:
      annotations:
        kubectl.kubernetes.io/last-applied-configuration: ...
      creationTimestamp: "2025-04-07T14:51:49Z"
      generation: 13
      name: vscode-rust
      namespace: workspace-sid
      resourceVersion: "1846893079"
      uid: ba307d8c-fb40-4d3c-b6cd-7721dba8c388
    spec:
      allowPublicExposure: false
      deleteAfter: 30d
      environmentList:
      - containerStartupOptions:
          contentPath: /vscode/workspace
          enforceWorkdir: false
        disableControls: false
        environmentType: Standalone
        guiEnabled: true
        image: crownlabs/vscode-rust:v0.2.0
        mode: Standard
        mountMyDriveVolume: false
        name: vscode-environment
        nodeSelector: {}
        persistent: true
        ...

The key information here is the line `nodeSelector: {}`, which says that this template would support the node selector feature.
However, since the nodeSelector is empty, it means that the user can choose among any label available within the cluster that allows the selection of the node.

In this example, when the user will start this instance, it can select whether the VM/container can be started on nodes where the `gpu-available` is true, where `node-size = big`, where `node-name = worker-2`, or any other label existing in the cluster.

In case a template must be started on nodes that have a given characteristic (e.g., include a GPU), you can customize the `nodeSelector` in the template and list only the labels that are allowed for the associated instances.


### Algorithm for node selection

The algorithm for node selection is depicted in the following image.

![Node selection algorithm](./node-selection-img.svg)
