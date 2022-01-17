# PyCharm container

This folder contains the files required to create a PyCharm container.
The `Dockerfile` can be used as a possible example about how to create _containerized_ graphical applications.

The `Dockerfile` starts building the container from a Ubuntu 20.04 image.
From this image, several libraries are installed to make PyCharm work, as detailed in the `Dockerfile`.
[PyCharm CE](https://www.jetbrains.com/pycharm/download/other.html) is then downloaded, extracted and dynamically configured in order to be able to run seamlessly.
In fact, before launching the PyCharm executable, a set of pre-configuration files are copied in the proper locations in the resulting image.
This allows the software to work out-of-the-box, without any splash screen (e.g., to confirm the license terms), for a better user experience.

The required files are located in the following directories:

- [config](config/): it contains three files for the (1) IDE theme, (2) window size, and (3) '_show tips on startup_' settings, which are copied in the container `~/.config` directory.
- [local](local/): it contains one file for the IDE usage statistics consent, which is copied in the container `~/.local` directory

Once this preparation is done, the `pycharm.sh` script is executed with an optional `PROJECT_DIR` parameter that, if set, points to a Python project root directory to further simplify the user experience.

Finally, it is worth remembering that the `DISPLAY` environment variable and the `/tmp/.X11-unix/` folder may have to be shared between this container and the ones hosting the required ancillary services (X-VNC server, etc), as detailed in the main [README](../) file.
