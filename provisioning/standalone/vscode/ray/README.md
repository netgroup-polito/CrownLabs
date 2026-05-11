# VSCode Template with Ray Support

This repository contains the Dockerfile and necessary configurations to generate a VSCode template with integrated Ray support. 

### Key Components:

- **`workspace/`**: This directory contains example scripts that can be used to test and interact with the Ray cluster.
- **`start.sh`**: The entrypoint script used to start the VSCode server. It also configures the Ray cluster connection, the shared Persistent Volume (PV), and the shared storage path within the template environment.
- **`ray-worker/`**: Contains a Dockerfile to build a custom Ray worker image with pre-installed dependencies. Using this custom image reduces startup time when leveraging `runtime_env` in your Python scripts.