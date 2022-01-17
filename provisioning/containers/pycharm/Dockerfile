FROM ubuntu:20.04

# Prevent apt from prompting the user for input during install/configuration
ENV DEBIAN_FRONTEND=noninteractive

# Install required libraries to render PyCharm
RUN apt-get update && \
  apt-get install -y --no-install-recommends libxrender1 libxtst6 libxi6 libfontconfig1 && \
  apt-get install -y python3 python3-venv python3-setuptools tzdata && \
  apt-get clean -y && \
  rm -rf /var/lib/apt/lists/*.*

# Define PyCharm version default arguments
ARG PYCHARM_VERSION="2021.2"
ARG REVISION=".1"

# Install wget, download and unpack PyCharm, and clean up
RUN apt-get update && apt-get install -y wget && \
  mkdir -p /opt/pycharm && \
  wget -qO- https://download.jetbrains.com/python/pycharm-community-${PYCHARM_VERSION}${REVISION}.tar.gz | tar xz --strip 1 -C /opt/pycharm && \
  apt-get purge --autoremove wget -y && \
  apt-get clean -y && \
  rm -rf /var/lib/apt/lists/*.*

# Define user and user id default arguments
ARG USER=crownlabs
ARG UID=1010

# Define basic default enviroment variables
ENV DISPLAY=:0 \
  USER=${USER} \
  HOME=/home/$USER \
  TZ=Europe/Rome

# Create new user and set a set a valid shell for them
RUN useradd -ms /bin/bash -u ${UID} $USER

# Copy PyCharm pre-configuration (PyCharm configuration and consent options on license and usage stats)
COPY config $HOME/.config/JetBrains/PyCharmCE${PYCHARM_VERSION}
COPY local $HOME/.local/share/JetBrains/consentOptions

# Set permissions on user home
RUN chown -R $USER:$USER $HOME

# Set the user to use
USER $USER

# CMD [ "executable", "parameter", ... ] form throws an error if $PROJECT_DIR is not set
# Therefore, bash-like form is used
ENTRYPOINT ["/opt/pycharm/bin/pycharm.sh"]
