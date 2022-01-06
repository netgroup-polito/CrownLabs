# Standalone Applications

## What is a Standalone Application?

**Standalone applications** are autonomous web services that are exposed over **HTTP**. Those services have to be packaged in a **container**.
An example could be a **container** exposing a **static web page**. If the **container** includes **all the dependencies**, it can be considered a **standalone application**. Some additional examples are [listed below](#applications).

## How CrownLabs manages Standalone Applications?

**Standalone applications** are **instances** of a **template** whose **environmentType** is **Standalone**. **Instances** and **templates** are **CRD resources**, managed by the **instance operator**, which is a CrownLabs **controller** capable of creating what is needed by a **standalone application**.

## Why standalone applications?

**Standalone applications** allow users to create their **templates**, without having to manage the **cluster infrastructure** (i.e., _deployments_, _services_, _ingressess_, and _pvcs_ are created and managed by the **instance operator**). Hence, a **developer** can focus on the **application** (i.e., the container) without thinking about the **infrastructure**.

# Getting started

-   [Introduction](./docs/gettingstarted.md)    
-   [Try our example](./docs/example.md)

# Applications

-   [Visual Studio Code](./vscode/README.md)
-   [JupiterLab](./jupyterlab/README.md)
