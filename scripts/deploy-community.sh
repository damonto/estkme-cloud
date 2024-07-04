#!/bin/bash

# One line command to deploy estkme-cloud-community
# curl -Lso- https://raw.githubusercontent.com/damonto/estkme-cloud/main/scripts/deploy-community.sh | bash

set -eux

# Check if docker is installed
if ! [ -x "$(command -v docker)" ]; then
  echo 'Error: docker is not installed.' >&2
  exit 1
fi

# Check if docker is running
if ! docker info > /dev/null 2>&1; then
  echo 'Error: docker is not running.' >&2
  exit 1
fi

# Check if estkme-cloud-community is already running
if [ "$(docker ps -aq -f status=running -f name=estkme-cloud-community)" ]; then
  docker stop estkme-cloud-community
fi

if [ "$(docker ps -aq -f status=exited -f name=estkme-cloud-community)" ]; then
  docker rm estkme-cloud-community
fi

curl -o Dockerfile https://raw.githubusercontent.com/damonto/estkme-cloud/main/Dockerfile.server
docker buildx build --file Dockerfile -t estkme-cloud-community:latest .
docker run -d --restart=always --name estkme-cloud-community -p 1022:22 -p 1888:1888 estkme-cloud-community:latest
rm -f Dockerfile
