FROM python:3.8

ENV SHELL=/bin/bash

# Define user and user id default arguments
ARG USER=crownlabs
ARG UID=1010

# Create new user and set a set a valid shell for them
RUN useradd -ms /bin/bash -u ${UID} $USER

RUN cp /root/.bashrc /home/crownlabs/.bashrc && \
    cp /root/.profile /home/crownlabs/.profile

RUN pip install numpy pandas jupyterlab matplotlib

COPY start.sh start.sh
RUN chmod 755 start.sh

COPY config.sh config.sh
RUN /bin/bash config.sh && rm ./config.sh

RUN apt-get clean && \
    apt-get remove --autoremove --purge -y apt wget curl --allow-remove-essential

USER $USER
WORKDIR /home/crownlabs
ENTRYPOINT [ "/start.sh" ]