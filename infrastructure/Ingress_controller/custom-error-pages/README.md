# nginx custom error pages

This folder contains the material necessary for the creation of the service used by the `nginx ingress controller` to serve custom error pages.

## Main components
The core components required to serve custom error pages are:

* [main.go](main.go): the source code of the server in charge of returning the desired error page, together with the companion static resources. The error page is automatically customized depending on the headers configured by the ingress controller during the request (i.e. the error code and the format requested);
* [rootfs](rootfs): the resources to be embedded into the docker image and used by the server. In detail:

    * [www/templates](rootfs/www/templates) contains the template pages which are filled in by the server with the error information before being returned;
    * [www/static](rootfs/www/static) contains the static companion resources to be served (e.g. css and images).

## Customization
The behavior of the server can be customized by setting different environment variables:

* `ERROR_SERVER_PORT`: the port the `http` server is listening to (default: `8080`);
* `ERROR_TEMPLATES_PATH`: the directory within the Docker image containing the error templates (default: `/www/templates`);
* `ERROR_STATIC_PATH`: the directory within the Docker image containing the static companion resources  (default: `/www/static`);
* `ERROR_STATIC_SERVE_PATH`: the path where the static resources are served by the handler (default: `/error-page`);
* `ERROR_STATIC_URI`: the relative or absolute pointing to the location where the static resources are made accessible (default: `/error-page`). Use an absolute URL in case multiple domains are served by the ingress controller.

## How to build
The creation of the Docker image is automatized through the [Makefile](Makefile). Please, customize the repository configuration according to your requirements.

**TL;DR:**
```bash
$ make
$ make push
```


## Additional information

* [[1]](https://github.com/kubernetes/ingress-nginx/tree/master/docs/examples/customization/custom-errors) ingress-nginx - How to setup custom errors
* [[2]](https://github.com/kubernetes/ingress-nginx/tree/master/images/custom-error-pages) ingress-nginx - The example code to serve custom errors