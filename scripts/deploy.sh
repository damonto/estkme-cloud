#!/bin/bash

set -eux

DST_DIR="/opt/estkme-cloud"

# Install dependencies
apt-get update -y && apt-get install -y unzip cmake pkg-config libcurl4-openssl-dev zip curl

# Download the latest version of lpac and compile it
LPAC_VERSION=$(curl -Ls https://api.github.com/repos/estkme-group/lpac/releases/latest | grep tag_name | cut -d '"' -f 4)
curl -L -o lpac.zip https://github.com/estkme-group/lpac/archive/refs/tags/"$LPAC_VERSION".zip
unzip lpac.zip && rm -f lpac.zip && cd lpac-*
cmake . -DLPAC_WITH_APDU_PCSC=off -DLPAC_WITH_APDU_AT=off && make -j $(nproc)
cp output/lpac "$DST_DIR" && cd .. && rm -rf lpac-*

# Download and Install estkme-cloud
supervisorctl stop estkme-cloud
declare -A ESTKME_CLOUD_BINARIES=(
    ["x86_64"]="estkme-cloud-linux-amd64"
    ["aarch64"]="estkme-cloud-linux-arm64"
    ["mips64"]="estkme-cloud-linux-mips64"
    ["riscv64"]="estkme-cloud-linux-riscv64"
)
if [ -z "${ESTKME_CLOUD_BINARIES[$(uname -m)]}" ]; then
    echo "Unsupported architecture"
    exit 1
fi
ESTKME_CLOUD_VERSION=$(curl -Ls https://api.github.com/repos/damonto/estkme-cloud/releases/latest | grep tag_name | cut -d '"' -f 4)
curl -L -o "$DST_DIR"/estkme-cloud https://github.com/damonto/estkme-cloud/releases/download/"$ESTKME_CLOUD_VERSION"/${ESTKME_CLOUD_BINARIES[$(uname -m)]}
chmod +x "$DST_DIR"/estkme-cloud

START_CMD="/opt/estkme-cloud/estkme-cloud --dir=/opt/estkme-cloud --dont-download"
if [ $# -ge 1 ] && [ -n "$1" ]; then
    START_CMD="$START_CMD --advertising='$1'"
fi

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
