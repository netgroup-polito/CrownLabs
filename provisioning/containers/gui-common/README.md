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
- [Websockify](websockify/): container that encapsulates the protocol used by the VNC server into a websocket and serves the noVNC client files. When a new web client connects to the remote environment, websockify delivers the Javascript package containing the noVNC code to the web client (which shows the remote desktop on the web browser and encapsulates the VNC protocol in a websocket). Once the user's browser loaded the noVNC client, a connection towards the same websockify instance is issued and the remote desktop connection takes place.

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

In CrownLabs, a custom implementation is used, derived from the go version available in the [Websockify-other repository](https://github.com/novnc/websockify-other) in order to provide a more efficient (in terms of space, memory and cpu usage) and customized component.

Customizations include connections logging and Prometheus metrics regaring the number of connections and the latency measured within such connections.
