#!/usr/bin/env bash

rm -rf dist/
mkdir -p "dist"

VERSION="0.1.0"
GIT_HASH=$(git rev-parse --short=11 HEAD)

function build() {
    mkdir -p "build"
    CGO_ENABLED=0 GOOS=$1 GOARCH=$2 go build -ldflags="-s -w -X 'main.Version=${VERSION}' -X 'main.GitHash=${GIT_HASH}'" -o build/$4 github.com/webishdev/fritze-mqtt/cmd
    cd build
    cp ../README.md README.md
    shasum -a 256 $4 > checksum.txt
    zip -q -r ../dist/fritze-mqtt-$3-$2-$VERSION-$GIT_HASH.zip ./
    cd ..
    rm -rf build/
}


build darwin arm64 macos fritze-mqtt

CGO_ENABLED=0 GOOS=$1 GOARCH=$2 go build -ldflags="-s -w -X 'main.Version=${VERSION}' -X 'main.GitHash=${GIT_HASH}'" -o fritze-mqtt github.com/webishdev/fritze-mqtt/cmd