#!/usr/bin/env bash

env GOOS=linux GOARCH=amd64 go build -v -tags netgo -o catnip.bin