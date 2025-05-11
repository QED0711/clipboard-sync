#!/bin/bash

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/linux/x86/clipboard-sync main.go;
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o dist/linux/arm64/clipboard-sync main.go;