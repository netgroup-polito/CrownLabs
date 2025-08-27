# Standalone Applications

## How to create a standalone application?

To create a container that can be used  as a **standalone application**, some simple rules have to be followed:

-   The **port** to which the service is exposed has to be set using the environment variable `$CROWNLABS_LISTEN_PORT`.
-   The **user** running the **container** cannot be **root** and should have UID 1010.
-   The relevant files have to be accessible by the **user** running the **container**.
-   The listening service has to use as **basepath** the environment variable `$CROWNLABS_BASE_PATH` or setup the **rewriteURL** inside the template (_see next paragraph_).
-   Must answer a **200** if a **GET request** is sent to `$CROWNLABS_BASE_PATH/` (*rewriteURL to false*) or `/` (*rewriteURL to true*).

### RewriteURL vs CROWNLABS_BASE_PATH

Let's make an example. We are deploying an application as a **standalone application**, so we will create a **template** with a **yaml file** that looks like the following:

```yaml
apiVersion: crownlabs.polito.it/v1alpha2
kind: Template
metadata:
    name: my-app
    namespace: workspace-sid
spec:
    prettyName: My Application
    description: Just an example template
    environmentList:
        - name: my-app-environment
          environmentType: Standalone
          mode: Standard
          image: myapplicationimage
          resources:
              cpu: 2
              memory: 2G
              reservedCPUPercentage: 25
          rewriteURL: false
          persistent: false
    workspace.crownlabs.polito.it/WorkspaceRef:
        name: sid
    deleteAfter: 30d
    inactivityTimeout: 14d
```

Then we will create an **instance**, either through a **yaml** file or using the **CrownLabs frontend**. The generated **URL** to access the **instance** would resemble `https://crownlabs.polito.it/instance/4e46/app/`. So, when we will send a request to this **URL**, the **application** will receive a **request** for `/instance/4e46/app/` and not just `/`. Hence, the **application** has to be aware of it (see the [example](example.md)).

In some cases, our **containers** do not run software written by ourselves. When is not possible to configure the application **basepath**, the **rewriteUrl** flag inside the **template** has to be set. This enables the **rewrite-URL** feature inside the generated **ingress** resource, which translates all **request's URLs** (towards our application) into paths relative to `/`. Hence, allowing the application to work correctly even if unaware of the **basepath**.

[Go back](../README.md)
