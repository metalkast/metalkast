#!/usr/bin/env bash
set -eEuo pipefail

# https://kubernetes.io/docs/setup/production-environment/container-runtimes/#install-and-configure-prerequisites

cat <<EOF | tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF

# modprobe overlay
# modprobe br_netfilter

# sysctl params required by setup, params persist across reboots
cat <<EOF | tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF

# # Apply sysctl params without reboot
# sysctl --system

# lsmod | grep br_netfilter
# lsmod | grep overlay

# sysctl net.bridge.bridge-nf-call-iptables net.bridge.bridge-nf-call-ip6tables net.ipv4.ip_forward
