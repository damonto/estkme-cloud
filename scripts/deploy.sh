#!/bin/bash

set -eux

DST_DIR="/opt/estkme-cloud"
declare -A ESTKME_CLOUD_BINARIES=(
    ["x86_64"]="estkme-cloud-linux-amd64"
    ["aarch64"]="estkme-cloud-linux-arm64"
    ["mips64"]="estkme-cloud-linux-mips64"
    ["riscv64"]="estkme-cloud-linux-riscv64"
)

# Check if the architecture is supported
if [ -z "${ESTKME_CLOUD_BINARIES[$(uname -m)]}" ]; then
    echo "Unsupported architecture"
    exit 1
fi

# Delete it after the next release.
if [ -f /etc/supervisor/supervisord.conf ]; then
  if grep -q "estkme-cloud" /etc/supervisor/supervisord.conf; then
    cat > /etc/supervisor/supervisord.conf <<'_EOF'
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

[include]
files = /etc/supervisor/conf.d/*.conf
_EOF
  fi
fi

# Install dependencies.
apt-get update -y && apt-get install -y unzip cmake pkg-config libcurl4-openssl-dev zip curl

# Download the latest version of lpac and compile it.
LPAC_VERSION=$(curl -Ls https://api.github.com/repos/estkme-group/lpac/releases/latest | grep tag_name | cut -d '"' -f 4)
curl -L -o lpac.zip https://github.com/estkme-group/lpac/archive/refs/tags/"$LPAC_VERSION".zip
unzip lpac.zip && rm -f lpac.zip && cd lpac-*
cmake . -DLPAC_WITH_APDU_PCSC=off -DLPAC_WITH_APDU_AT=off && make -j $(nproc)
cp output/lpac "$DST_DIR" && cd .. && rm -rf lpac-*

# if estkme-cloud is already running stop it.
if supervisorctl status estkme-cloud | grep -q RUNNING; then
  supervisorctl stop estkme-cloud
fi

# Download and Install estkme-cloud.
ESTKME_CLOUD_VERSION=$(curl -Ls https://api.github.com/repos/damonto/estkme-cloud/releases/latest | grep tag_name | cut -d '"' -f 4)
curl -L -o "$DST_DIR"/estkme-cloud https://github.com/damonto/estkme-cloud/releases/download/"$ESTKME_CLOUD_VERSION"/${ESTKME_CLOUD_BINARIES[$(uname -m)]}
chmod +x "$DST_DIR"/estkme-cloud

START_CMD="/opt/estkme-cloud/estkme-cloud --dir=/opt/estkme-cloud --dont-download"
if [ $# -ge 1 ] && [ -n "$1" ]; then
    START_CMD="$START_CMD --advertising='$1'"
fi

tee /etc/supervisor/conf.d/estkme-cloud.conf << EOF
[program:estkme-cloud]
command=$START_CMD
autostart=true
autorestart=true
stdout_logfile=/dev/stdout
stdout_logfile_maxbytes=0
stderr_logfile=/dev/stderr
stderr_logfile_maxbytes=0
EOF

supervisorctl update
