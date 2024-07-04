#!/bin/bash

# curl -Lso- https://raw.githubusercontent.com/damonto/estkme-cloud/main/scripts/deploy-community.sh | bash

# Check if docker is installed
if ! [ -x "$(command -v docker)" ]; then
  echo 'Error: docker is not installed.' >&2
  exit 1
fi

curl -o Dockerfile https://raw.githubusercontent.com/damonto/estkme-cloud/main/Dockerfile.server
docker buildx build --file Dockerfile -t estkme-cloud-community:latest .
docker run -d --restart=always --name estkme-cloud-community -p 1022:22 -p 1888:1888 estkme-cloud-community:latest
