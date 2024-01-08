#!/usr/bin/env bash
set -eEuo pipefail

# Temporarily set RFC-1123 valid hostname to fix issues with kubeadm run
old_hostname=$(hostname)
hostname k8s-node
trap "hostname $old_hostname" EXIT

# start containerd
containerd &
timeout 5s bash -c 'until [ -e /var/run/containerd/containerd.sock ]; do sleep 0.1; done'
trap 'kill $(jobs -p)' EXIT

kubeadm -v5 config images pull --kubernetes-version $CONFIG__KUBERNETES_VERSION
