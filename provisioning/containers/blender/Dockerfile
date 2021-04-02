# Start from https://github.com/nytimes/rd-blender-docker
FROM nytimes/blender:2.92-cpu-ubuntu18.04

# Define user and user id default arguments
ARG USER=crownlabs
ARG UID=1010

# Define basic default enviroment variables
ENV DISPLAY=:0 \
  USER=${USER} \
  HOME=/home/$USER

# Create new user and set a set a valid shell for them
RUN useradd -ms /bin/bash -u ${UID} $USER

# Set permissions on user home
RUN chown -R $USER:$USER $HOME

# Set the user to use
USER $USER

WORKDIR $HOME

CMD blender
