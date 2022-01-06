# Jupyterlab

This is a **jupyterlab** container, supporting **CrownLabs standalone applications**.
It uses the official **python 3.8** image, with **jupyterlab** installed through **pip**.

**This application doesn't require URL rewrite**
# Pre-installed python packages 
During the build phase those python packages are installed:
- numpy 
- pandas
- jupyterlab
- matplotlib

Is possible to install additional packages using **pip** from the **jupyter terminal**.

## How to run

`docker run -e CROWNLABS_LISTEN_PORT=8001 -it --rm -p 8001:8001  harbor.crownlabs.polito.it/crownlabs-standalone/jupyterlab`

## How to build

`docker build -t harbor.crownlabs.polito.it/crownlabs-standalone/jupyterlab`