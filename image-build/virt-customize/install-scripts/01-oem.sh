#!/bin/sh
set -e

apt-get remove --purge -y linux-virtual 'linux-image-*'
apt-get autoremove --purge -yq
apt-get clean -yq

apt-get update -y
# https://wiki.ubuntu.com/Kernel/OEMKernel
apt-get install -y linux-oem-22.04c
