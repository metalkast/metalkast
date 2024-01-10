#!/usr/bin/env bash
set -eEuo pipefail

builder_image=k8s-cluster-node-image-builder
docker build . -t $builder_image

mkdir -p output
docker run --rm --privileged -v $(pwd)/output:/virt-customize/output $builder_image
