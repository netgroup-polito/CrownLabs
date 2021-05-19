# qlkube

Our qlkube is a custom adaptation of [this repo](https://github.com/qlkube/qlkube/)

It allows us to query the kubernetes apiserver using [GraphQL](https://graphql.org/).

## Playground authentication

Qlkube exposes a Playground instance, to authenticate and be able to make certain queries you need to add an header in the playground specifying the token, e.g. if the token is `abcde` you need to add the following the playground `HTTP HEADERS` tab:

```json
{
  "Authorization": "Bearer abcde"
}
```

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
yarn install
# start qlkube in development (using nodemon for auto-restart-on-save)
yarn dev
# go to localhost:8080 on the browser to use the playground
```

## Schema

The generated graphql schema is served at /schema

## Playground adjust

An issue could be seen in the playground when you test the subscription due to some graphql-playground modules. This bug is about the scrolling of the responses. At the moment, the only way to overcome this issue is to use a browser extension that injects custom css into the page. A possible extension for fixing that is [Stylus](https://chrome.google.com/webstore/detail/stylus/clngdbkpkpeebahjckkjfobafhncgmne).

After you have installed the extension and opened the option panel, you can add custom css clicking on the respective button under the menu **Actions** and adding the code below:

```css
.graphiql-wrapper > div:first-child > div:first-child > div:nth-child(2) {
    height: 100%;
}
.graphiql-wrapper > div:first-child > div:first-child > div:nth-child(2) > div:nth-child(2) {
    height: 100%;
    overflow-anchor: auto;
}
```

Finally, add the correct url of **CrownLabs Playground** and save.

[setting](../documentation/settingExtension.mp4)


## Subscriptions

[example](../documentation/subscriptionExample.mp4)
