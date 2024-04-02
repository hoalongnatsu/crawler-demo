#!/bin/bash
IMAGE=$1
NEW_IMAGE=$2

sudo sed -i "s|$IMAGE.*|$NEW_IMAGE|g" docker-compose-server.yaml
docker compose -f docker-compose-server.yaml up -d