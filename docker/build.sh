#!/bin/bash

ARCH=${1:-x86}

docker build \
    -t clipboard-sync-server:latest \
    --build-arg TARGETARCH=$ARCH \
    -f docker/Dockerfile.prod .