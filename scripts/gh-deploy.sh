#!/bin/bash

# Check Environment Variables
if [ -z "$SSH_PRIVATE_KEY" ]; then
  echo "SSH_PRIVATE_KEY is not set"
  exit 1
fi

if [ -z "$SSH_SERVERS" ]; then
  echo "SSH_SERVERS is not set"
  exit 1
fi

# Deploy to Servers
for server in "$SSH_SERVERS[@]"; do
  echo "Deploying to $server"
  echo $SSH_PRIVATE_KEY | ssh -o StrictHostKeyChecking=no -i /dev/stdin $server "mkdir -p /opt/estkme-cloud"
  echo $SSH_PRIVATE_KEY | scp -i /dev/stdin -r ./deploy.sh $server:/opt/estkme-cloud
  echo $SSH_PRIVATE_KEY | ssh -o StrictHostKeyChecking=no -i /dev/stdin $server "/opt/estkme-cloud/deploy.sh"
  echo $SSH_PRIVATE_KEY | ssh -o StrictHostKeyChecking=no -i /dev/stdin $server "rm -f /opt/estkme-cloud/deploy.sh"
  echo "Deployed to $server"
done
