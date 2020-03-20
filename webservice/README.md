## Crown Lab web service

Component to provide a full web service which embeds in the resulting website many API for users to directly interact with a Kubernetes cluster.

This component is the result of many frameworks integrated:
    
* UserInterface: [ReactJS](https://reactjs.org/)
* Web-server and API exporting : [WebPack](https://webpack.js.org/)
* Kubernetes Javascript library: [kubernetes-client-javascript](https://github.com/scality/kubernetes-client-javascript/tree/browser), a patched version of the [official one](https://github.com/kubernetes-client/javascript)

## Variable exporting

Please **READ** this section before installing.

Both if you are going to deploy our webservice locally or in a Docker container, you need to provide 4 environment variables:

* OIDC_PROVIDER_URL
* OIDC_CLIENT_ID
* OIDC_REDIRECT_URI
* APISERVER_URL

*OIDC_PROVIDER_URL*, *OIDC_CLIENT_ID*  and *OIDC_REDIRECT_URI* are parameters used by our library to setup the **OpenID Connect** protocol.
More in detail, the provider URL is the IdentityProvider you are going to contact (in our case our keycloak).
The client id is the client id that your IdentityProvider is going to accept. The redirect URI is the URL you are going to be 
redirected after you complete the login. Please **NOTE** that if you are going to run the webservice locally, this will be
something like http://localhost:8000 , while in production this will correspond to our website URL https://crownlabs.polito.it .

*API_SERVER_URL* is the url to the API server to whom our Kubernetes library interact with. In our case, this is our Kubernetes address.

For further information about these protocols and infrastructure, please refer to [How it works](#how-it-works) section.

## Local installation

Requirements:

* Node.js

Unfortunately, the patched library we are currently using cannot be directly accessed by NodeJS, since still has not been
merged in their main project's branch. As a result we need to compile it on our own.  

From the current directory [CrownLabs/webservice](.) launch `npm install` to install the patched Kubernetes Javascript library.
Once completed this step, we move the attention to the [website](./website) directory which is where our website is defined.

From withing [website](./website) type `npm install` to install the actual webservice.
This command install all the dependencies and builds the service which now can be run from this directory. 

Export now the 3 variables as described [before](#variable-exporting):

```bash
export OIDC_PROVIDER_URL=https://2.2.2.2:4444
export OIDC_CLIENT_ID=xxxxxxx
export OIDC_REDIRECT_URI=http://localhost:8000
export APISERVER_URL=https://1.1.1.1:3333
```

To run the service, type `npm start`.

Visit http://localhost:8000 or wherever you decided to host your website (also according to the OIDC_REDIRECT_URI you have set).

## Docker installation

Requirements:
    
    * Docker
    
We offer a [Dockerfile](./Dockerfile) to build an image containing our complete web service.

When building the image, you need to set an ENV variable inside the Dockerfile whether
you are going to run your docker container locally or not.

If you are going to run it locally, please set OIDC_REDIRECT_URI=http://localhost:8000 , while if you want to deploy this docker
in our domain set it to OIDC_REDIRECT_URI=https://crownlabs.polito.it . 

To build, type `docker build -t <tag> .` from this directory.

Run the image. If running locally, map the port 8000 of your host to the port 80 of your container. Otherwise just enjoy.

`docker run -p 8000:80 --rm <image_tag>`

Visit http://localhost:8000 or https://crownlabs.polito.it .

## How it works

Now a little bit of talks.

The architecture is not that easy to understand as you may have noticed. But we do care to explain it in detail, in order to be
as comprehensible as possible.

Let's start from the user interface. 

To rapidly develop an efficient UI we have chosen ReactJS, a very wide used Javascript library. ReactJS allows you to declare
components and functions which automatically will be translated into HTML code, making you save painful days in coding your website.
In fact, all you have to do is to design all the components you are going to need, providing some ReactJS classic method for rendering.

The tricky part begins when you have to build and run the website.

React offers many scripts in NodeJS to automatically
build and configure a web server, parsing all the files you have created and offering to the user the result of that parse: the HTML page.
We noticed that under the hood these React script uses WebPack to export component client-side. So instead of using them, we wrote our own
configuration to manually run WebPack, specifying our additional libraries we want to export (the Kubernetes one).

In fact, WebPack is a static module bundler for Javascript application, which builds a dependency graph where all the modules are
mapped, and according to it they are exported as bundle. Using many plugins and modules (ex. for loading images, fonts, etc.)
we were able to deploy the website as if it was deployed using native ReactJS scripts. In addition, we included the Kubernetes Javascript Api to the build,
making them accessible directly from the client.

Talking about that, Kubernetes Javascript API are officially supported and developed by Kubernetes team and this seems to be a good news.
Unfortunately, their API are accessible only server-side using NodeJS, due to many module dependencies which can be solved only server-side.
This said, we cannot pack them using WebPack, since the client (pure Javascript) would not have those dependencies.

Digging around, we found a [fork](https://github.com/scality/kubernetes-client-javascript/tree/browser) to the official 
repository which actually performs this module resolution, allowing these API to be called also client-side. 
Basically this patched version is the same as before, but with a slightly different build file. In fact in this project
it is performed a binding of those server-side-only modules with user-accessible ones.

Unfortunately, this branch is not merged in the main one yet, so we got it locally and we integrated it with our own React web server.
 
The result of all these steps is a fully working website able to parse ReactJS components and integrate Kubernetes API inside them.

The website leverages [OpenID Connect (OIDC)](https://openid.net/connect/) combined with [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
to handle Authentication and Authorization. In particular, we tested using [Keycloak](https://www.keycloak.org/) as 
Identity and Access Management Module. However, all services and components implementing the OIDC protocol should work fine 
 
## Developing

#### UserInterface

All the ReactJS files are under [this](./website/src) directory.

#### Kubernetes API

For the Kubernetes Api, the only file you have to refer to is this [one](./website/src/k8sApi/index.js). 

#### Docker

If you want to extend or improve the provided docker, you should refer to the [Dockerfile](./Dockerfile).

#### Webpack

The configuration file used by WebPack is [here](website/webpack.config.js);

#### Change image in login page

If you want to change the image displayed in login page, you have to:

* Upload your image inside [this](./website/src/assets) directory
* Update the image name in the Login.js file (you can find it [here](./website/src/Login.js)) as follows:

    ```
    const logo = require('./assets/<image_name>');
    ```
