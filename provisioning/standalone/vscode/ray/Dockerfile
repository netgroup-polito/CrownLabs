# syntax = edrevo/dockerfile-plus

INCLUDE+ ./base/Dockerfile

ENV SUDO_FORCE_REMOVE yes

ENV RAY_ADDRESS="ray://raycluster-head-svc.workspace-kuberay-gpu.svc.crowng:10001"
ENV SHARED_STORAGE_PATH="/shared/models"
ENV WORKSPACE_PVC_NAME="shvol-shvol-kuberay-gpu"

# Install Python 3.11 via deadsnakes PPA + kubectl
ARG KUBECTL_VERSION=v1.30.0
RUN apt-get update && \
    apt-get install -y software-properties-common && \
    add-apt-repository ppa:deadsnakes/ppa && \
    apt-get update && \
    apt-get install -y python3.11 python3.11-venv python3.11-dev python3-pip ca-certificates curl && \
    curl -fsSL -o /usr/local/bin/kubectl \
      "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" && \
    chmod +x /usr/local/bin/kubectl && \
    apt-get purge -y curl && \
    apt-get clean

# Python venv with Ray and scientific deps
ENV VENV_PATH=/opt/venv
RUN python3.11 -m venv ${VENV_PATH}

ENV VIRTUAL_ENV=${VENV_PATH}
ENV PATH="${VENV_PATH}/bin:${PATH}"
ENV PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1 \
    PYTHONDONTWRITEBYTECODE=1

COPY requirements.txt /tmp/requirements.txt
RUN pip install --upgrade pip setuptools wheel && \
    pip install -r /tmp/requirements.txt && \
    rm -f /tmp/requirements.txt

# Activate venv automatically in every terminal session
RUN echo 'source /opt/venv/bin/activate' >> ${VSCODE_SRV_DIR}/.bashrc

# VS Code user settings (python interpreter + venv-aware terminal)
RUN mkdir -p ${VSCODE_SRV_DIR}/data/User
COPY workspace/.vscode/settings.json ${VSCODE_SRV_DIR}/data/User/settings.json

# Install VS Code extensions
RUN code-server --extensions-dir ${VSCODE_SRV_DIR}/extensions \
      --install-extension ms-python.python --force && \
    code-server --extensions-dir ${VSCODE_SRV_DIR}/extensions \
      --install-extension ms-toolsai.jupyter --force && \
    rm -rf /root/.cache /root/.local/share/code-server/CachedData 2>/dev/null || true

# Workspace files go to /example_project; start.sh --load-example seeds
# /vscode/workspace on first launch (idempotent via .startup sentinel file)
COPY --chown=${USER}:${USER} workspace/ /example_project/

# Set permissions and drop apt + sudo (CrownLabs security convention)
RUN chown -R ${USER}:${USER} ${VSCODE_SRV_DIR} /example_project && \
    apt-get remove --autoremove --purge -y apt sudo --allow-remove-essential 2>/dev/null || true

USER ${USER}

# CROWNLABS_LISTEN_PORT is injected at runtime by the platform.
# --load-example seeds the workspace on first boot (skipped on subsequent starts).
ENTRYPOINT ["/start.sh", "--load-example"]
