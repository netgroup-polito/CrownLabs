# General architecture

This folder contains the components required to create the *common ground* for graphical containers, which are the following:

- **X server**: it is the foundation that allows for graphic elements to be drawn on the display. X Windows builds the primitive framework that allows moving of windows, interactions with keyboard and mouse, and draws windows. This is required for any graphical desktop.
- **Window Manager**: it is the component that controls the placement and appearance of windows, such as sub-windows or menu opened by the application. It requires X Windows but not a desktop environment.
- **VNC server**: it is a graphical desktop-sharing system that allow to remotely control another computer. It transmits the keyboard and mouse events from one computer to another, relaying the graphical-screen updates back in the other direction, over a network.
- **Web server for noVNC**: an HTML5 implmentation for a VNC client is available for modern browsers; the web server is required to transform the VNC server into a web-based protocol, which enables users using a simple web browser to connect and show the remote display.

For the sake of completeness, graphical containers do not require a Desktop Environment component (e.g., KDE, GNOME, etc), which represent a far more integrated and complex system than a Window Manager. In fact, a Desktop Environment usually includes a suite of applications that are tightly integrated with the system (control panel, networking controls, general purpose applications), as well as graphical widgets such as the *start menu*, status bars, system trays (for notifications and different widgets), and more. For more information, look at this [page](https://www.ghacks.net/2008/12/09/get-to-know-linux-desktop-environment-vs-window-manager/).


## Current implementation

The current implementation leverages three containers, each one contained in the following sub-folders:

- [X VNC server](tigervnc/): container that creates X VNC plus Fluxbox (window manager), providing the basic graphical environment.
- [Websockify](websockify/): container that encapsulates the protocol used by the VNC server into a websocket.
- [Webserver](novnc/): container that creates an `nginx` web server. This is used when a new web client connects to the remote environment to deliver the Javascript package containing the noVNC code to the web client (which shows the remote desktop on the web browser and encapsulates the VNC protocol in a websocket). In addition, this web server acts as _reverse-proxy_ for incoming websocket connections (which contain the VNC protocol) in order to redirect them to _websockify_, i.e., the component that extracts the VNC protocol from the websocket and delivers the resulting messages to the VNC server.


### X VNC Server

The X VNC server is based on a standard X server, but it has a "virtual" screen rather than a physical one. X applications display themselves on it as if it were a normal X display, but they can only be accessed via a VNC viewer.
Technically, this consists in two different servers (the actual X server and a VNC server) that are bound together and provide a virtual desktop both to the user (through the VNC protocol) and to the final application (through the X core protocol).

There are several implementations of such type of server. For CrownLabs, [TigerVNC](https://github.com/TigerVNC/tigervnc) has been chosen, since it provides automatic desktop rescaling features.
Note that TigerVNC server cannot be considered a regular X server since it cannot display the desktop on a physical screen.

This container also includes [Fluxbox](http://fluxbox.org/), a lightweight window manager which provides an effectively desktop-like user experience and, above all, it enables menus and sub-windows to be rendered correctly.

Finally note that:
- The `DISPLAY` environment variable must have the same value in the `X` server instance and in the application container. This requires the component that coordinates the deployment of the two containers to set the `DISPLAY` value accordingly when starting the two Dockers.
- The X server will create a socket file in folder `/tmp/.X11-unix/`, which may need to be shared with the container running the application. This requires the component that coordinates the deployment of the two containers to possibly mount the same shared volume under `/tmp` when starting the two Dockers.


### Websockify

Websockify is a small-footprint component that translates WebSockets traffic to normal socket traffic.
Websockify accepts the WebSockets handshake, parses it, and then begins forwarding traffic between the client and the target in both directions, hence allowing a browser (with the correct Javascript client) to communicate with any remote server.

This component is required because the client browser needs a bi-directional interation with the remote desktop, hence using a websocket connection.
However, a component is required to translate the websocket connection into standard socket data, to be consumed by the VNC server on the remote server.
See [the original repository](https://github.com/novnc/websockify) for further information.

This container has been designed to be executed as a sidecar inside the same pod of the VNC server without any further configuration.
However, it could be executed in other contexts as well by setting the necessary environment variables, if needed.

In CrownLabs, we use the [C version of Websockify](https://github.com/novnc/websockify-other) in order to provide a more efficient (in terms of space, memory and cpu usage) implementation.

Compared to the python version, the C version does not include a static web server, which is used to host the web application that connects to Websockify in order to reach the actual server running in the backend.

Since this functionality is still required, the next component (Webserver) is used to provide such web application (the actual noVNC client) to the user's browser.


### Webserver

The C version of Websockify decribed above does not include any web server.
Although not strictly needed at run-time (the Websockify software translates websockets connections into normal sockets without the necessity of any other component), a web server is still required when the client begins its connection to the remote desktop.
In fact, when the client connect to TCP port where the remote desktop server is listening to, it has to download all the required (HTML5) code that implements the VNC client running in the browser, such as the [noVNC](https://novnc.com/) project.
Without noVNC, the browser has no way to show the remote display, nor to send/receive the proper VNC commands through the established websocket.

This container consists of an `nginx` server hosting the noVNC release files, in which the necessary configuration settings are dynamically injected.
This approach allows to directly use the release archive provided by the noVNC team, without the need to maintain a forked repository.

In the context of CrownLabs, this container can be used in two ways:
- as a sidecar inside the final application pod running the VNC server
- as a disjoint deployment, decoupled from the final application pod running the VNC server

In the first case, this container is tightly bounded to the VNC server, hence creating a pod that directly accepts (multiple) websocket connections, resulting in a simpler management of the Kubernetes ingress controller.

In the second case, this container can be instantated _once_ in the entire datacenter, serving all the VNC-based containers. This results in a more efficient use of resources, but it requires the ingress controller to be able to multiplex the websocket connection towards the correct service.
