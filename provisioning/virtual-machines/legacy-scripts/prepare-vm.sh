#!/bin/bash

# Paths
VNC_PATH="/home/${USER}/.vnc"
NOVNC_PATH="/usr/share/novnc"
SYSTEMD_PATH="/etc/systemd/system"

# Services
VNC_SERVICE="vncserver@:1.service"
NOVNC_SERVICE="novnc.service"
PNE_SERVICE="prometheus_node_exporter.service"

# Install Xfce (gnome gives errors)
if ! test -f /usr/share/xsessions/xfce.desktop; then
    echo "It looks like you don't have xfce installed. To proceed you need to install it."
    while true; do
        read -p -r "Do you want to install it now? (y/n) " yn
        case $yn in
            [Yy]* ) sudo apt-get install -y xfce4; break;;
            [Nn]* ) exit 0;;
            * ) echo "Please answer yes or no.";;
        esac
    done
fi

# Block logout button
sudo mv /usr/bin/xfce4-session-logout /usr/bin/xfce4-session-logout_bak

# Install cloud-init
# Cloud-init is needed to start the VM on the cluster
# SSH right now is needed for testing
# Numpy is needed by novnc
sudo apt-get install -y openssh-server cloud-init python-numpy

# Install tigervnc
# TigerVNC is the vncserver of choice
wget -qO- https://dl.bintray.com/tigervnc/stable/tigervnc-1.10.1.x86_64.tar.gz | sudo tar xz --strip 1 -C /
mkdir -p "$VNC_PATH"

# Set vnc password
# @featureremoved
#echo "${VNC_PWD}" | vncpasswd -f > "${VNC_PATH}/passwd"
#chmod 0600 "${VNC_PATH}/passwd"

# Set vnc xstartup file
tee "${VNC_PATH}/xstartup" > /dev/null <<EOT
#!/bin/sh
unset SESSION_MANAGER
unset DBUS_SESSION_BUS_ADDRESS
exec startxfce4
EOT

chmod +x "${VNC_PATH}/xstartup"

# Create a service to autostart the vncserver at boot
sudo tee "${SYSTEMD_PATH}/${VNC_SERVICE}" > /dev/null <<EOT
[Unit]
Description=Remote desktop service (VNC)
After=syslog.target network.target

[Service]
Type=forking
User=${USER}
Group=${USER}
WorkingDirectory=${HOME}
ExecStartPre=/bin/sh -c '/usr/bin/vncserver -kill %i > /dev/null 2>&1 || :'
ExecStart=/usr/bin/vncserver %i -SecurityTypes None -localhost
ExecStop=/usr/bin/vncserver -kill %i
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
EOT

# Install NoVNC
sudo mkdir -p "$NOVNC_PATH/utils/websockify"

wget -qO- https://github.com/netgroup-polito/noVNC/archive/v1.1.3-crown.tar.gz | sudo tar xz --strip 1 -C "$NOVNC_PATH"
wget -qO- https://github.com/novnc/websockify/archive/v0.9.0.tar.gz | sudo tar xz --strip 1 -C "$NOVNC_PATH/utils/websockify"

# Create the service for NoVNC
sudo tee "${SYSTEMD_PATH}/${NOVNC_SERVICE}" > /dev/null <<EOT
[Unit]
Description=NoVNC service
After=network.target

[Service]
Type=simple
User=${USER}
Group=${USER}
ExecStart=${NOVNC_PATH}/utils/websockify/run --web ${NOVNC_PATH} 6080 localhost:5901
RemainAfterExit=yes
Nice=-10

[Install]
WantedBy=multi-user.target
EOT

# Link to NoVNC landing page for easy url access
sudo ln -s "$NOVNC_PATH/vnc.html" "$NOVNC_PATH/index.html"

# Install prometheus node exporter
# This package allows to export the node information using the 9100 TCP port
wget -qO- https://github.com/prometheus/node_exporter/releases/download/v0.18.1/node_exporter-0.18.1.linux-amd64.tar.gz | tar xz --strip 1
sudo mv node_exporter /usr/local/bin/
rm LICENSE NOTICE
sudo useradd -rs /bin/false pne_user > /dev/null

# Create prometheus node exporter service
sudo tee "${SYSTEMD_PATH}/${PNE_SERVICE}" > /dev/null <<EOT
[Unit]
Description=Prometheus Node Exporter
After=network.target

[Service]
Type=simple
User=pne_user
Group=pne_user
ExecStart=/usr/local/bin/node_exporter

[Install]
WantedBy=multi-user.target
EOT

# Install webdav support
sudo apt-get install -y debconf
echo 'davfs2 davfs2/suid_file boolean true' | sudo debconf-set-selections
sudo apt-get install -y davfs2
sudo adduser "$USER" davfs2

# Enable services
sudo systemctl daemon-reload
sudo systemctl enable $PNE_SERVICE
sudo systemctl enable $NOVNC_SERVICE
sudo systemctl enable $VNC_SERVICE

# Since the graphical desktop is accessed through VNC,
# it is useless to start the default xfce session that is not
# "seen" by anybody, but consumes memory (about 200M)
sudo systemctl set-default multi-user.target
