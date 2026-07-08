#!/bin/bash

# TODO: refactor to include as download command in cvms.sh (optional folder or use current folder to read vms.yml)

cd "$(dirname "$(dirname "$(readlink -f "$0")")")"
root_dir=$(pwd)

# TODO: extract version
curl https://dl-cdn.alpinelinux.org/alpine/v3.24/releases/cloud/generic_alpine-3.24.1-x86_64-bios-cloudinit-r0.qcow2 -o "$root_dir/imgs/alpine.img"
curl https://cloud-images.ubuntu.com/noble/current/noble-server-cloudimg-amd64.img -o "$root_dir/imgs/ubuntu.img"
