## Guidelines to install the web service

Date: 13/03/2020

Version: v1.0

## Installation

From the current directory `CrownLabs/webservice` launch `npm install` to install the Kubernetes Javascript library.

Then go into the `website` directory (`CrownLabs/webservice/website`) and type again `npm install`.

## Run

From `CrownLabs/webservice/website` type `npm start`.

Visit <http://localhost:8000>.

## Developing

### React

All the *React* files and login is under the directory `CrownLabs/webservice/website/src`. Add here your files or modify the present ones.

### K8sApi

For the Kubernetes Api, the only file you have to refert to is `CrownLabs/webservice/website/k8sApi/index.js`. 

By now those functions are only defined. The web service doesn't display errors, but we have to agree with `@simona_fiore` the return types of these APIs.

### Dockerfile

If you want to create a Docker file for the web service, you should refer to this guide's installation section and make those commands to be run inside the docker container (form the Dockerfile) after you have copied all the `CrownLabs/webservice` directory into the docker.

### Webpack config tuning

The webpack configuration file can be retrieved at `CrownLabs/webservice/website/webpack.config.js`. Feel free to modify it according to your needs (port, plugins, etc.).

By now the images are not correctly displayed client-side; it may be something related to this config file and all the module section you find inside it to tell you web server how to load and send render the objects.