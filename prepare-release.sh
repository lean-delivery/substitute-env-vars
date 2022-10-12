#!/usr/bin/env bash

echo "version = $1"

# Get version number from version tag
DOCKER_TAG=$(echo "$1" | cut -d'v' -f2)
export DOCKER_IMAGE="docker://ghcr.io/lean-delivery/substitute-env-vars:${DOCKER_TAG}"
echo "Docker tag = ${DOCKER_TAG}"
echo "Docker image = ${DOCKER_IMAGE}"

yq -i '.runs.image = strenv(DOCKER_IMAGE)' action.yml
