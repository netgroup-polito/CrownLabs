# nginx custom error pages

This folder contains the material necessary for the creation of the service used by the `nginx ingress controller` to serve custom error pages.

## Main components

The core components required to serve custom error pages are:

* [main.go](server/main.go): the source code of the server in charge of returning the desired error page. The error page is automatically customized depending on the headers configured by the ingress controller during the request (i.e. the error code and the format requested);
* [templates](static/templates): the template pages which are filled in by the server with the error information before being returned. The original, non-minified version of the templates is available in the [original](static/original) directory.

## Customization

The behavior of the server can be customized through the following command line parameters:

* `--http-address`: the address the server binds to (defaults to `:8080`);
* `--templates-path`: the directory within the Docker image containing the error templates (defaults to `/templates`);

## How to build

```bash
docker build . --push -tag crownlabs/custom-error-pages:v0.x
```

## Additional information

* [[1]](https://github.com/kubernetes/ingress-nginx/tree/master/docs/examples/customization/custom-errors) ingress-nginx - How to setup custom errors
* [[2]](https://github.com/kubernetes/ingress-nginx/tree/master/images/custom-error-pages) ingress-nginx - The example code to serve custom errors
