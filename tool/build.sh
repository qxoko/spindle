#!/bin/bash

build_dir="build_v2"

set -e

printf "[building]\n"

build() {
	echo "$1 $2"
	name="$1_$2"

	if [ $1 = 'darwin' ]; then
		name="macOS_"

		if [ $2 = 'arm64' ]; then
			name+="M1"
		else
			name+="Intel"
		fi
	fi

	mkdir -p "$build_dir/$name"
	env GOOS="$1" GOARCH="$2" go build -ldflags "-s -w" -trimpath -o "$build_dir/$name/spindle$3" ./source

build windows amd64 ".exe"

build linux amd64
build linux arm64

build darwin amd64
build darwin arm64