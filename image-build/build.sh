#!/usr/bin/env bash
set -eEuo pipefail

mkdir -p output
rm -rf output/*

export METALKAST_VERSION=$(git describe --always --dirty)
export NETBOOT_BASE_URL=https://dl.metalkast.io/node-images/

for v in 1.28.4 1.29.3; do
    export KUBERNETES_VERSION=$v
    function compose() {
        docker compose -p image-build-$(echo $v | tr '.' '_') "$@"
    }
    compose down
    compose build
    compose up --exit-code-from image-build
    compose cp image-build:/virt-customize/output .
    find output -name shasum.txt -delete
done
