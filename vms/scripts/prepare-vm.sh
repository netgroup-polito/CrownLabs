#!/bin/sh

# Vars
NOVNC_PORT=6080

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
        read -p "Do you want to install it now? (y/n) " yn
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
mkdir -p $VNC_PATH

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
ExecStartPre=/bin/sh -c '/usr/bin/vncserver -kill %i > /dev/null 2>&1 || :'
ExecStart=/usr/bin/vncserver %i -SecurityTypes None -localhost
ExecStop=/usr/bin/vncserver -kill %i
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
EOT

# Install NoVNC
sudo mkdir -p $NOVNC_PATH/utils/websockify

wget -qO- https://github.com/netgroup-polito/noVNC/archive/v1.1.1-crown.tar.gz | sudo tar xz --strip 1 -C $NOVNC_PATH
wget -qO- https://github.com/novnc/websockify/archive/v0.9.0.tar.gz | sudo tar xz --strip 1 -C $NOVNC_PATH/utils/websockify

# Create the service for NoVNC
sudo tee "${SYSTEMD_PATH}/${NOVNC_SERVICE}" > /dev/null <<EOT
[Unit]
Description=NoVNC service
After=network.target

[Service]
Type=oneshot
User=${USER}
Group=${USER}
ExecStart=${NOVNC_PATH}/utils/launch.sh --listen 6080 --vnc localhost:5901
RemainAfterExit=yes
Nice=-10

[Install]
WantedBy=multi-user.target
EOT

# Link to NoVNC landing page for easy url access
sudo ln -s $NOVNC_PATH/vnc.html $NOVNC_PATH/index.html

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
# Manually installing and configuring the davfs2 package to avoid post-install interactive configuration
sudo mkdir temp
sudo chown _apt:root temp/
cd temp/
sudo apt-get download davfs2
PACKAGE_NAME=$(find . -name *.deb)
sudo dpkg --unpack $PACKAGE_NAME

# Custom postinst script to avoid interactive config
sudo tee /var/lib/dpkg/info/davfs2.postinst > /dev/null <<EOT
#!/bin/sh -e
# postinst script for davfs2
dpkg-statoverride --update --add root root 4755 /usr/sbin/mount.davfs > /dev/null 2>&1 || true

sys_uid=\$(getent passwd davfs2 | cut -d ':' -f 3)
sys_gid=\$(getent group davfs2 | cut -d ':' -f 3)
if [ "\$sys_uid" = "" -a "\$sys_gid" = "" ]; then
    adduser --system --home "/var/cache/davfs2" --no-create-home --group davfs2 > /dev/null 2>&1 || true
elif [ "\$sys_uid" = "" ]; then
    adduser --system --home "/var/cache/davfs2" --no-create-home --ingroup davfs2 davfs2 > /dev/null 2>&1 || true
elif [ "\$sys_gid" = "" ]; then
    addgroup --system davfs2 > /dev/null 2>&1 || true
    usermod -g davfs2 davfs2 > /dev/null 2>&1 || true
fi

chown root:davfs2 /var/cache/davfs2 > /dev/null 2>&1 || true
chown root:davfs2 /var/run/mount.davfs > /dev/null 2>&1 || true
chmod 775 /var/cache/davfs2 > /dev/null 2>&1 || true
chmod 1775 /var/run/mount.davfs > /dev/null 2>&1 || true

for file in mount.davfs umount.davfs; do
    if [ ! -e /sbin/\$file ]; then
        ln -s /usr/sbin/\$file /sbin/\$file
    fi
done
EOT

sudo apt-get install -yf

cd ..
sudo rm -rf temp/

# Enable services
sudo systemctl daemon-reload
sudo systemctl enable $PNE_SERVICE
sudo systemctl enable $NOVNC_SERVICE
sudo systemctl enable $VNC_SERVICE
