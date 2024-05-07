#!/bin/bash

set -e

if [ "$(id -u)" != "0" ]; then
    echo "Please run as root"
    exit 1
fi

# Only support debian-based system
if [ ! -f /etc/debian_version ]; then
    echo "Unsupported system"
    exit 1
fi

# Install dependencies
apt-get update -y && apt-get install -y unzip cmake pkg-config libcurl4-openssl-dev libpcsclite-dev zip curl

# Get the latest release version
get_latest_release() {
    curl --silent "https://api.github.com/repos/$1/releases/latest" | # Get latest release from GitHub api
      grep '"tag_name":' |                                            # Get tag line
      sed -E 's/.*"([^"]+)".*/\1/' |                                    # Pluck JSON value
      sed 's/v//'                                                      # Remove the "v" from the version number
}

DST_DIR="/opt/estkme-cloud"

LPAC_VERSION=$(get_latest_release "estkme-group/lpac")
if [ -z "$LPAC_VERSION" ]; then
    echo "Invalid LPAC version"
    exit 1
fi

function download_binary {
    LPAC_BINARY_URL="https://github.com/estkme-group/lpac/releases/download/v$LPAC_VERSION/lpac-linux-x86_64.zip"
    echo "Downloading LPAC binary from $LPAC_BINARY_URL"
    curl -L -o $DST_DIR/lpac.zip $LPAC_BINARY_URL
    unzip -o $DST_DIR/lpac.zip -d $DST_DIR/data
    chmod +x $DST_DIR/data/lpac
    rm -f $DST_DIR/lpac.zip
}

function build_from_source {
    BUILD_DIR=$(mktemp -d)
    LPAC_SOURCE_CODE="https://github.com/estkme-group/lpac/archive/refs/tags/v$LPAC_VERSION.zip"
    echo "Downloading LPAC source code from $LPAC_SOURCE_CODE"
    # Download the source code
    mkdir -p $BUILD_DIR
    curl -L -o $BUILD_DIR/lpac.zip $LPAC_SOURCE_CODE
    unzip -o $BUILD_DIR/lpac.zip -d $BUILD_DIR
    cd $BUILD_DIR/lpac-$LPAC_VERSION
    mkdir -p build && cd build
    # Build the source code
    cmake .. && make -j$(nproc)
    cp -rf $BUILD_DIR/lpac-$LPAC_VERSION/build/output/lpac $DST_DIR/data
    chmod +x $DST_DIR/data/lpac
    rm -rf $BUILD_DIR
}

echo "Downloading LPAC version $LPAC_VERSION"
mkdir -p $DST_DIR/data
if [ "$(uname -m)" == "x86_64" ]; then
    download_binary
else
    build_from_source
fi

ESTKME_CLOUD_VERSION=$(get_latest_release "damonto/estkme-cloud")
if [ -z "$ESTKME_CLOUD_VERSION" ]; then
    echo "Invalid eSTK.me Cloud Enhance Server version"
    exit 1
fi

if [ "$(uname -m)" == "x86_64" ]; then
    ESTKME_CLOUD_BINARY="estkme-cloud-linux-amd64"
elif [ "$(uname -m)" == "aarch64" ]; then
    ESTKME_CLOUD_BINARY="estkme-cloud-linux-arm64"
else
    echo "Unsupported architecture"
    exit 1
fi
ESTKME_CLOUD_BINARY_URL="https://github.com/damonto/estkme-cloud/releases/download/v$ESTKME_CLOUD_VERSION/$ESTKME_CLOUD_BINARY"

SYSTEMED_UNIT="estkme-cloud.service"
SYSTEMED_UNIT_PATH="/etc/systemd/system/$SYSTEMED_UNIT"
START_CMD="/opt/estkme-cloud/estkme-cloud --data-dir=/opt/estkme-cloud/data --dont-download"
if [ -n "$1" ]; then
    START_CMD="$START_CMD --advertising='$1'"
fi
SYSTEMED_FILE="
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
"

if [ -x "$(command -v systemctl)" ]; then
  if [ "$(systemctl is-active $SYSTEMED_UNIT)" == "active" ]; then
    systemctl stop $SYSTEMED_UNIT
  fi
else
  if [ -n "$(pgrep -f "$START_CMD")" ]; then
    pkill -f "$START_CMD"
  fi
fi

# Download eSTK.me Cloud Enhance Server
echo "Downloading eSTK.me Cloud Enhance Server version $ESTKME_CLOUD_VERSION"
curl -L -o $DST_DIR/estkme-cloud $ESTKME_CLOUD_BINARY_URL
chmod +x $DST_DIR/estkme-cloud

if [ -x "$(command -v systemctl)" ]; then
    echo "Deploying eSTK.me Cloud Enhance Server to systemd"
    echo "$SYSTEMED_FILE" > $SYSTEMED_UNIT_PATH
    systemctl daemon-reload
    systemctl enable $SYSTEMED_UNIT
    systemctl start $SYSTEMED_UNIT
else
    echo "Deploying eSTK.me Cloud Enhance Server to background"
    nohup $START_CMD > /dev/null 2>&1 &
fi

echo "eSTK.me Cloud Server deployed successfully!"
