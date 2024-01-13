#!/usr/bin/env bash
set -eEuo pipefail

# https://docs.docker.com/engine/install/ubuntu/

# Uninstall old versions
# https://docs.docker.com/engine/install/ubuntu/#uninstall-old-versions
# Replace with dpkg to ignore missing packages: https://superuser.com/a/518871
dpkg --remove docker docker-engine docker.io containerd runc
dpkg --purge docker docker-engine docker.io containerd runc

# Set up the repository
# https://docs.docker.com/engine/install/ubuntu/#set-up-the-repository

# Step 1
apt-get update
apt-get install -y ca-certificates curl gnupg

# Step 2
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
chmod a+r /etc/apt/keyrings/docker.gpg

# Step 3
echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
  tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install docker engine and containerd
# https://docs.docker.com/engine/install/ubuntu/#install-docker-engine

# Step 1
apt-get update

# Step 2
apt-get install -y containerd.io


# Configure containerd
# https://kubernetes.io/docs/setup/production-environment/container-runtimes/#containerd
# https://github.com/containerd/containerd/blob/main/docs/cri/config.md#basic-configuration
cat <<EOF > /etc/containerd/config.toml
version = 2
# Required field: https://github.com/containerd/containerd/issues/6964
[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
  runtime_type =  "io.containerd.runc.v2"
[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
  SystemdCgroup = true
EOF

# https://github.com/containerd/containerd/blob/main/docs/getting-started.md#interacting-with-containerd-via-cli
# TODO: install cri-tools to enable debugging

apt-get clean -yq
