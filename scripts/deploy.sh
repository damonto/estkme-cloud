#!/bin/bash

set -xe

if [ "$(id -u)" != "0" ]; then
    echo "Please run as root"
    exit 1
fi

# Only support debian-based system
if [ ! -f /etc/debian_version ]; then
    echo "Unsupported system"
    exit 1
fi

# Check systemctl or supervisorctl
if [ ! -x "$(command -v systemctl)" ] && [ ! -x "$(command -v supervisorctl)" ]; then
    echo "Please install systemd or supervisor"
    exit 1
fi

# Check if the system is Debian 11 (bullseye)
if [ "$(lsb_release -cs)" == "bullseye" ]; then
    echo "You are using Debian 11 (bullseye), please add sid repository to install libc6 from sid repository"
    echo "If you have added sid repository, please ignore this message"
    echo "You can run the following command to add sid repository:"
    echo "echo 'deb http://deb.debian.org/debian/ sid main' > /etc/apt/sources.list.d/sid.list"
    echo "apt update && apt install -y libc6"
fi

# Install dependencies
apt-get update -y && apt-get install -y git unzip cmake pkg-config libcurl4-openssl-dev zip curl

DST_DIR="/opt/estkme-cloud"

LPAC_VERSION=$(curl -Ls https://api.github.com/repos/estkme-group/lpac/releases/latest | grep tag_name | cut -d '"' -f 4)
curl -L -o lpac.zip https://github.com/estkme-group/lpac/archive/refs/tags/$LPAC_VERSION.zip
unzip lpac.zip && rm -f lpac.zip && cd lpac-*
cmake . -DLPAC_WITH_APDU_PCSC=off -DLPAC_WITH_APDU_AT=off && make -j $(nproc)
cp output/lpac $DST_DIR && cd .. && rm -rf lpac-*

if [ -x "$(command -v systemctl)" ] && [ "$(systemctl is-active estkme-cloud.service)" == "active" ]; then
  systemctl stop estkme-cloud.service
elif [ -x "$(command -v supervisorctl)" ] &&  [ "$(supervisorctl status estkme-cloud | awk '{print $2}')" == "RUNNING" ]; then
  supervisorctl stop estkme-cloud
fi

ESTKME_CLOUD_VERSION=$(curl -Ls https://api.github.com/repos/damonto/estkme-cloud/releases/latest | grep tag_name | cut -d '"' -f 4)
if [ "$(uname -m)" == "x86_64" ]; then
    ESTKME_CLOUD_BINARY="estkme-cloud-linux-amd64"
elif [ "$(uname -m)" == "aarch64" ]; then
    ESTKME_CLOUD_BINARY="estkme-cloud-linux-arm64"
else
    echo "Unsupported architecture"
    exit 1
fi
curl -L -o $DST_DIR/estkme-cloud https://github.com/damonto/estkme-cloud/releases/download/$ESTKME_CLOUD_VERSION/$ESTKME_CLOUD_BINARY
chmod +x $DST_DIR/estkme-cloud

START_CMD="/opt/estkme-cloud/estkme-cloud --dir=/opt/estkme-cloud --dont-download"
if [ -n "$1" ]; then
    START_CMD="$START_CMD --advertising='$1'"
fi

if [ -x "$(command -v systemctl)" ]; then
    echo "Deploying eSTK.me Cloud Enhance Server to systemd"
    tee /etc/systemd/system/estkme-cloud.service << CONFIG
[Unit]
Description=eSTK.me Cloud Enhance Server
After=network.target

[Service]
Type=simple
Restart=on-failure
ExecStart=$START_CMD
RestartSec=10s
TimeoutStopSec=30s

[Install]
WantedBy=multi-user.target
CONFIG
    systemctl daemon-reload
    systemctl enable estkme-cloud.service
    systemctl start estkme-cloud.service
else
    echo "Deploying eSTK.me Cloud Enhance Server to supervisor"
    tee /etc/supervisor/supervisord.conf << CONFIG
[supervisord]
nodaemon=true
logfile=/dev/null
logfile_maxbytes=0
pidfile=/tmp/supervisord.pid

[rpcinterface:supervisor]
supervisor.rpcinterface_factory = supervisor.rpcinterface:make_main_rpcinterface

[unix_http_server]
file=/tmp/supervisor.sock

[supervisorctl]
serverurl=unix:///tmp/supervisor.sock

[program:estkme-cloud]
command=$START_CMD
autostart=true
autorestart=true
stdout_logfile=/dev/stdout
stdout_logfile_maxbytes=0
stderr_logfile=/dev/stderr
stderr_logfile_maxbytes=0
CONFIG
    supervisorctl reread
    supervisorctl update
    supervisorctl reload
fi

echo "eSTK.me Cloud Server has been deployed!"
