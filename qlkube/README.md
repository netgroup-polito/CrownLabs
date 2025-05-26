# qlkube

Our qlkube is a custom adaptation of [this repo](https://github.com/qlkube/qlkube/)

It allows us to query the kubernetes apiserver using [GraphQL](https://graphql.org/).

## Playground authentication

Qlkube exposes a Playground instance, to authenticate and be able to make certain queries you need to add an header in the playground settings by adding the Authorization header and specifying inserting the Bearer token (including `Bearer` at the beginnning).

To obtain the token you can log into the dashboard and, using the browser dev tools, check the `Authorization` headers of some requests and copy the value from there.

## Development

```bash
# setup a local cluster using [KinD](https://kind.sigs.k8s.io/)
kind create cluster
# check if cluster is ok
kubectl cluster-info
# install the CRDs
kubectl apply -f ../operators/deploy/crds
# add some test CRs
kubectl apply -f ../operators/samples
# proxy the apiserver to localhost:8001
kubectl proxy
# open a new terminal
# install the packages needed
npm install
# start qlkube in development (using nodemon for auto-restart-on-save)
npm run dev
# go to localhost:8080 on the browser to use the playground
```

## Schema

The generated graphql schema is served at /schema
