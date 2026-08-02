⚠️ <span style="color:red"><strong>DEPRECATION NOTICE</strong></span> ⚠️

Container instances no longer provide the functionality they previously offered and will soon be discontinued. It is recommended to avoid using them, even during this transition phase.

If you're wondering what could happen, you would simply see an endless loading spinner near your Container instance when you try to start it.

<hr style="border: 0; border-top: 1px solid red; margin: 1rem 0;">

# Creating and uploading Docker-based graphical applications in CrownLabs

This folder contains the files required to create docker-based graphical applications in CrownLabs, as an alternative solution to VMs.

In order to simplify the process of creating such GUI applications, multiple containers are envisioned, running according to the *sidecar* model in Kubernetes:

- **Common components**: a first set of containers, available under `gui-common` (now removed), which previously provided the components required to expose a virtual desktop in the browser.
- **Containerized application**: a further container keeping the actual application, running in the virtual desktop above. Currently, the [pycharm](pycharm/) folder provides an example of such an application, based on the PyCharm integrated development environment for Python.

This architecture decoupled the application itself, e.g. PyCharm, from the virtual desktop logical layer previously provided by `gui-common`.


## Starting the application
The application can be started by running all the involved containers within the same pod.


## Creating a new application
A new application can be created by just *containerizing* it, i.e., writing a self-contained Dockerfile for the application layer, without introducing strict bindings with the graphical components that would make the deployment monolithic and thus more difficult to maintain.
In other words, we can deploy new applications by creating new dockerfiles modeled after the [PyCharm](pycharm/) one, provided as example, while leaving untouched the now-removed `gui-common` folder.

To the sake of completeness, two information may be considered when creating the `Dockerfile` for a new application:
- the `DISPLAY` environment variable must have the same value in the `X` server instance and in the application container. In the current implementation, this is achieved by the custom Kubernetes operator that is in charge of starting the pod, which sets the same value in both.
- the `/tmp/.X11-unix` directory that will contain the socket file used to communicate with the X server must be shared between the application container and the X VNC server. In the current implementation, this is achieved by the custom Kubernetes operator that is in charge of starting the pod, which mounts an `emptyDir` volume under `/tmp` on both containers.
