#!/usr/bin/env bash
set -eEuo pipefail

mkdir -p output
rm -rf output/*

export METALKAST_VERSION=$(git describe --always --dirty)
export NETBOOT_BASE_URL=https://dl.metalkast.io/node-images/

export KUBERNETES_VERSION=1.36.1
function compose() {
    docker compose -p image-build-$(echo $v | tr '.' '_') "$@"
}
compose down
compose build
compose up --exit-code-from image-build
compose cp image-build:/virt-customize/output .
find output -name shasum.txt -delete
