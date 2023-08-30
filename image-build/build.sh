#!/usr/bin/env bash
set -eEuo pipefail

builder_image=k8s-cluster-node-image-builder
docker build . -t $builder_image

temporary_builder_container=$(docker create $builder_image)
docker cp $temporary_builder_container:/virt-customize/output .

# Cleanup
docker rm $temporary_builder_container
