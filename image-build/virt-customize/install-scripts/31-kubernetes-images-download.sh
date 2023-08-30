#!/usr/bin/env bash
set -eEuo pipefail

# Temporarily set RFC-1123 valid hostname to fix issues with kubeadm run
old_hostname=$(hostname)
hostname k8s-node
trap "hostname $old_hostname" EXIT

# start containerd
containerd &
trap 'kill $(jobs -p)' EXIT

kubeadm -v5 config images pull --kubernetes-version $CONFIG__KUBERNETES_VERSION
