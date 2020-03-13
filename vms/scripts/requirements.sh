#!/bin/sh

VNC_PWD="ccroot"
VNC_DEPTH=24
VNC_GEOMETRY="1440x1024"

# CLOUD-INIT
sudo apt-get install -y openssh-server cloud-init

# TIGERVNC
wget -qO- https://dl.bintray.com/tigervnc/stable/tigervnc-1.10.1.x86_64.tar.gz | sudo tar xz --strip 1 -C /
mkdir -p "/home/${USER}/.vnc"
echo "${VNC_PWD}" | vncpasswd -f > .vnc/passwd
chmod 0600 .vnc/passwd
touch .vnc/config

tee "/home/${USER}/.vnc/xstartup" > /dev/null <<EOT
#!/bin/sh
unset SESSION_MANAGER
unset DBUS_SESSION_BUS_ADDRESS
exec dbus-launch gnome-session
EOT
chmod +x "/home/${USER}/.vnc/xstartup"

echo "vncserver :1 -localhost -depth ${VNC_DEPTH} -geometry ${VNC_GEOMETRY} &" >> "/home/${USER}/.profile"

# NOVNC
sudo apt-get install -y novnc websockify python-numpy
sudo openssl req -x509 -nodes -newkey rsa:2048 -keyout novnc.pem -out /etc/ssl/novnc.pem -days 365
sudo chmod 644 /etc/ssl/novnc.pem

sudo tee /etc/systemd/system/novnc_start.service > /dev/null <<EOT
[Unit]
Description=Remote desktop service (VNC)
After=network.target

[Service]
Type=oneshot
User=${USER}
Group=${USER}
ExecStart=/usr/bin/websockify -D --web=/usr/share/novnc/ --cert=/etc/ssl/novnc.pem 6080 localhost:5901
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
EOT

sudo systemctl daemon-reload
sudo systemctl enable novnc_start.service

sudo ln -s /usr/share/novnc/vnc.html /usr/share/novnc/index.html

# NODE EXPORTER
wget -qO- https://github.com/prometheus/node_exporter/releases/download/v0.18.1/node_exporter-0.18.1.linux-amd64.tar.gz | tar xz --strip 1
sudo mv node_exporter /usr/local/bin/
rm LICENSE NOTICE

# node_exporter user
sudo useradd -rs /bin/false node_exporter > /dev/null

# node_exporter service
sudo tee /etc/systemd/system/node_exporter.service > /dev/null <<EOT
[Unit]
Description=Node Exporter
After=network.target
 
[Service]
User=node_exporter
Group=node_exporter
Type=simple
ExecStart=/usr/local/bin/node_exporter
 
[Install]
WantedBy=multi-user.target
EOT

sudo systemctl daemon-reload
sudo systemctl enable node_exporter
