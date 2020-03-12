## Guidelines to install & integrate Kubernetes-client/javascript into a WebServer

Date: 11/03/2020

Version: v1.0

## Cluster config

As cluster I have used `minikube (v1.8.1)`.

The only parameters you need to set in your cluster environment is the `cors-allowed-origins` to allow call between different domains in the same page via Javascript.

Let's start our cluster: `minikube start --extra-config "apiserver.cors-allowed-origins=["http://\*"]"`

Hopefully in the production release we are going to change the policy into `http://polito.it/*`.

## Steps

### Step1 - Initial setup

Lets clone the repo and pull the remote `browser` branch:

```bash
git clone git@github.com:scality/kubernetes-client-javascript.git
cd kubernetes-client-javascript
git checkout -b browser
git branch --set-upstream-to=origin/browser browser
git pull
```

### Step2 - Dependencies

You should now install the npm environment both from the project's root directory and from the `example/browser` one.

```bash
npm install
npm audit fix

cd examples/browser
npm install
npm audit fix
```

Last thing to point out, in order to be correctly executed the service needs some environment variables to be exported. You can follow these commands:

```bash
export APISERVER_URL=$(kubectl config view -o jsonpath='{range .clusters[*]}{.cluster.server}{"\n"}{end}')
export OIDC_PROVIDER_URL=https://auth.example.org
export OIDC_CLIENT_ID=example
```

`OIDC_CLIENT_ID` and `OIDC_PROVIDER_URL` are not important by now, so ignore their values.

You should now be able to run `npm start` and visit `http://localhost:8000` to check if the web server is up and running.

### Step3 - Tuning with our file

As you may notice, you are not allowed to perform anything since the `example/browser/index.js` script include an `oidc` check, thanks to we will later connect to the cluster, which we do not have yet.

In order to make it work, please overwrite `index.js` with the one I created.

### Step4- Tokenize

As previously told, to access the cluster we need a token provided by the `oidc` part we just overlooked. 

For the sake of simplicity, we now create a ClusterRoleBinding for the `default` user, giving him all the privileges. Use the `crb.yaml` file I provide.

`kubectl apply -f crb.yaml`

Visit again our website and check the displayed status. It should state that the client is working correctly, otherwise contact me.

### Step5 - Populate

This last step aims to populate your cluster with all the CRD defined in our CrownTeam repository.

Download the repo and install both `labtemplates` and `labinstance` CRDs.

```bash
git clone git@github.com:netgroup-polito/CrownLabs.git
cd CrownLabs/operators/labInstance-operator/
make install-lab-template
kubectl apply -f labTemplate/samples/template_v1_labtemplate.yaml

make install
kubectl apply -f config/samples/instance_v1_labinstance.yaml
```

Visit now the web page and check if those resources are correctly displayed.

Your installation is ended, enjoy Kubernetes.

## Pointers

https://medium.com/better-programming/k8s-tips-using-a-serviceaccount-801c433d0023
https://github.com/ramitsurana/awesome-kubernetes/tree/a3785d6c1a0d76c581aa130d08fbc61195036362#web-applications
https://www.digitalocean.com/community/tutorials/how-to-set-up-an-nginx-ingress-with-cert-manager-on-digitalocean-kubernetes