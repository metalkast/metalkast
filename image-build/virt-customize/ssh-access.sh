#!/usr/bin/env bash
set -eEuo pipefail

echo 'bootstrap ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/bootstrap
apt-get -yq update
apt install -yq
# SSH should be enabled through IPMI console
systemctl disable ssh
useradd bootstrap -m -p $(openssl passwd bootstrap) -G sudo
echo 'bootstrap ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers.d/bootstrap
