# Visual Studio Code

This is a **server-based** version with a **Web Interface** of the popular editor [**VS-Code**](https://code.visualstudio.com/) by [**Microsoft®**](https://www.microsoft.com/en-us/).

_It is based on the [**code-server**](https://github.com/coder/code-server) project by [**Coder**](https://github.com/coder)_

# Getting Started

-   [Overview](#overview)
-   [How to run](#how-to-run)
-   [Basic Installation](#basic-version)
-   [Custom Installation](#custom-version)
-   [Advanced Features](#advanced-features)
-   [Others](#how-to-make-templates-for-crownlabs)

# Overview

**CrownLabs** offers a set of custom **vscode** images. Every image is created for a specific use case. All images are based on the **vscode-base** image which represents the base of the project.
To learn how to create your image, check the [**Custom Version**](#custom-version) section.

# How to run

Every **vscode image** can be run with **Docker**, passing the environment variables used by all standalone applications ([**check this**](../docs/gettingstarted.md#how-to-create-a-standalone-application))
An example is: `docker run -e CROWNLABS_LISTEN_PORT=8001 -it --rm -p 8001:8001 harbor.crownlabs.polito.it/crownlabs-standalone/vscode-c-cpp`

# Basic Version

The **base image** contained in the `./base` folder is an image that contains only the **basic features** of the editor.

**Its direct usage is not recommended** for these reasons:

-   It has not an **entrypoint** (but it contains the *start.sh* script, which will be the derived image's **entrypoint**, so it can be run anyway with `docker run vscode-base /start.sh`).
-   It runs as **root** (so it cannot be launched in **CrownLabs**).
-   It contains **sudo** and **apt**, which can be used by a user to **scale-up privileges** and to **install unwanted packages**.

# Custom Version

## How To

-   [Set Dockerfile](#set-dockerfile)
-   [Build Dockerfile](#build-dockerfile)

### Set Dockerfile

Starting from the [Base Version](#base-version) it is possible to develop a customized image.

1.   **Create** a new **Dockerfile**
2.   **Include** the base version Dockerfile using **Dockerfile+**'s `INCLUDE+` directive (see more [**here**](https://github.com/edrevo/dockerfile-plus.git))

    ```Dockerfile
    INCLUDE+ ./base/Dockerfile
    ```

3.   **Copy** the starting project files and directories to the root folder of the container.
4.   **Download**, through `apt`, all needed dependencies and packages and clean what is not necessary for your use case.
5.   **Set** the **user** and **group** permissions of the previous files.
6.   *Optionally* **install code-server extensions** through the [**CLI utility**](#install-extensions-from-cli).
7.   **Switch** to `${USER}`.
8.   **Add** `ENTRYPOINT ./start.sh ` at the end of the file.

### Build Dockerfile

To build an image it is necessary to set the **Docker context** to `/provisioning/standalone/vscode` or another parent directory. That's because we are using `INCLUDE+` (see more [**here**](https://github.com/edrevo/dockerfile-plus.git)) and the used context must contain the files used by the **base image** Dockerfile.

This is an example of how to build the **c-cpp vscode** version:

- Go inside the `/provisioning/standalone/vscode` folder
- Run: `docker build -f ./c-cpp/Dockerfile -t harbor.crownlabs.polito.it/crownlabs-standalone/vscode-c-cpp`

*Optionally*, you can install **codetogether** providing the **build-arg** `CODETOGETHER_ENABLED_ARG` (see more [here](#codetogether))

## Ready-to-use images available

This is a list of **ready-to-use images**, built by the **CrownLabs** team :
```
harbor.crownlabs.polito.it/crownlabs-standalone/vscode-c-cpp
harbor.crownlabs.polito.it/crownlabs-standalone/vscode-python
harbor.crownlabs.polito.it/crownlabs-standalone/vscode-golang
```

# Advanced Features

## Install extensions from CLI

**Extensions** can be installed from the command line using:

```Dockerfile
code-server --extensions-dir /config/extensions --install-extension ${EXTENSION_NAME}
```

The `EXTENSION_NAME` can be either a local `.vsix` file or a reference to an extension on [https://open-vsx.org/](https://open-vsx.org/) marketplace.

## Disable Marketplace

By default **vscode images** allow users to download extensions from the marketplace without limits. To avoid this, you can **disable the marketplace** by adding the `--disable-marketplace` at startup.

For example:

```
docker run -e CROWNLABS_LISTEN_PORT=8001 -it --rm -p 8001:8001 vscode-c-cpp --disable-marketplace
```

## Codetogether

**Vscode images** support [**codetogether**](https://www.codetogether.com/) extension and API, which allow **social-coding** between multiple vscode instances (similarly to **[liveshare](https://visualstudio.microsoft.com/it/services/live-share/)** by [**Microsoft®**](https://www.microsoft.com/en-us/). This [video](https://youtu.be/l4yTfduxptw) shows what it can do. To enable an image to use **codetogether** it is necessary to build it with the `CODETOGETHER_ENABLED_ARG=true` build argument.

For example:

```
docker build --build-arg=CODETOGETHER_ENABLED_ARG=true -t vscode-c-cpp-codetogether
```

## Reset persistent instances

**Persistent** standalone applications can be **paused/started**. Is possible to reset workspace directory to the initial state by deleting the `.vscode/.startup` file and **restarting** the instance.

## Internal proxy

Every **vscode image** runs an **internal proxy** which allows to access to services running inside the container on `localhost:${port_num}`, connecting to `crownlabs.polito.it/${CROWNLABS_BASE_PATH}/proxy/${PORT_NUM}`, where `$PORT_NUM` is the port exposing the service.

For example, if you run a web server inside **vscode** on **port 80**, you can access to it with `crownlabs.polito.it/${CROWNLABS_BASE_PATH}/proxy/80`

# How to make templates for Crownlabs

**Code-server** is not capable of listening on a specific **basepath**, it can only listen on `localhost:${CROWNLABS_LISTEN_PORT}`
For this reason, it is necessary to enable the **rewriteURL**, as explained [**here**](../docs/gettingstarted.md#how-to-create-a-standalone-application)

[Go back](../README.md)
