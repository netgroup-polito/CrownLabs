FROM rayproject/ray:2.47.1.aeaf41-py311-gpu

USER root

# 1. Install Torch into the specific Ray conda path
# We use the specific CUDA 12.1 wheels to match the Ray 2.47.1 GPU base
RUN /home/ray/anaconda3/bin/pip install --no-cache-dir \
    torch==2.4.1 \
    torchvision==0.19.1 \
    torchaudio==2.4.1 \
    --index-url https://download.pytorch.org/whl/cu121

# Change ray user uid to 1010 and fix all file ownership
RUN usermod -u 1010 ray && \
    find / -user 1000 -exec chown -h 1010 {} \; 2>/dev/null || true

USER 1010