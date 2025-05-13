#!/bin/bash

docker run --rm \
    -p 5000:8000 \
    clipboard-sync-server:latest \
    