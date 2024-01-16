#!/bin/sh
set -e

echo 'bootstrap ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/bootstrap
# SSH will be enabled through IPMI console
systemctl disable ssh
useradd bootstrap -m -p $(openssl passwd bootstrap) -G sudo
echo 'bootstrap ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers.d/bootstrap
