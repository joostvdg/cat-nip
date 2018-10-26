#!/usr/bin/env bash

go get -d -v -t
env GOOS=linux GOARCH=amd64 go build -v -tags netgo -o catnip.bin