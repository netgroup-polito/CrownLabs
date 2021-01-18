[![Coverage Status](https://coveralls.io/repos/github/netgroup-polito/CrownLabs/badge.svg)](https://coveralls.io/github/netgroup-polito/CrownLabs)

# CrownLabs

CrownLabs is a set of services that can deliver **remote computing labs** through a **per-user virtual machine**.

Instructors can provision a set of virtual machines, properly equipped with the software required for a given lab (e.g., compilers, simulation software, etc).

Each student can connect to its own set of (remote) private environments without requiring any additional software, just a simple Web browser. No space problems on the student hard disk, no troubles in setting up the environment required to support multiple subjects on the same machine, and more.

In addition, each student can share his remote desktop with his groupmates, enabling multiple students to complete their labs in a team.

Finally, CrownLabs supports also instructors, who can connect to the remote desktop of the student and play directly with his environment, e.g., in case some help is required.

For more information, visit the CrownLabs website ([https://crownlabs.polito.it](https://crownlabs.polito.it)) and download our [scientific paper](https://ieeexplore.ieee.org/document/9136697) published in IEEE Access.


## Architecture

CrownLabs relies on two major components:

* **Frontend**, which is responsible to access Kubernetes API, guiding the user to creation of VMs.
* **Laboratory Operator**, which reacts to LabInstances creation by creating the Kubernetes objects to launch
the laboratory.

A high-level representation of the main architectural building blocks composing CrownLabs is given by the following figure. Please notice that, for the sake of clarity, the figure depicts the elements essential for the provision of the actual service (i.e. remote computing labs), while leaving out with those more low-level or associated with the cluster operation (e.g monitoring).

![CrownLabs High-Level Architecture](documentation/architecture.svg)


## External libraries

In this project we leverage and modify two external libraries:

* In the frontend component, we use a browser adaption of [Kubernetes JS Client](https://github.com/kubernetes-client/javascript)
which is available only for server-side applications. This version is based on a fork from [Scality](https://github.com/scality/kubernetes-client-javascript/tree/browser) which added browser-side support.
* The laboratory operator leverages the [Kubevirt](https://kubevirt.io/) library to create the VirtualMachineInstances.

# Installation

## Preliminary Note

CrownLabs can be installed on any Kubernetes cluster, although with a non-negligible degree of adaptation.
This would require a non-trivial knowledge of how Kubernetes (and the wonderful world of microservices) works.
No magic install procedure is unfortunately available (yet).

In a nutshell, you have to install all the components with your own custom configuration files, which may largely depend upon your physical install.
A huge degree of customization is possible in this respect: pure data-link vs. BGP-based load balancing, the number (and capabilities) of your servers, the desired degree of high availability, integration with external authentication servers, creation of admin/user credentials, your own secrets to protect the internal communication among the components.

Do not expect to complete this task in a few hours; likely, you may need several days, or even more.
Help is available on our Slack channels.
For more information, visit the CrownLabs website: [https://crownlabs.polito.it](https://crownlabs.polito.it).

## Pre-Requirements

Crownlabs has been specifically designed for bare-metal clusters and this assumption will be adopted across the documentation. To deploy CrownLabs, we have to rely on a full-fledged Kubernetes cluster where at least a subset of nodes supports Hardware Virtualization.
In [infrastructure](infrastructure/), we present all the services which should be installed on the cluster, with an example of configuration. We strongly suggest to set up on your cluster the same components that we used, in order to avoid feature mismatch.

## Laboratory Operator

The Laboratory Operator (LabOperator) implements the backend logic necessary to spawn new laboratories starting from a predefined template. The instruction to install and configure the LabOperator are detailed [here](operators).

## Frontend

The Frontend lets respectively students and professors interact with templates, for example, spawning new instances and define new templates. The logic to configure and deploy the Frontend are presented [here](webservice).

## Course Setup

Students work is normally organized in course with several professors. CrownLabs maps this logic by creating:

* a Kubernetes namespace for each student, where he/she can create his/her laboratories
* a Kubernetes namespace for the course where the professor creates the laboratory template, which can be instantiated by the final users. Those are readable only by the students of the course.

All the permissions are handled by the Kubernetes RBAC and the mapping to user and groups is done via Keycloak.
