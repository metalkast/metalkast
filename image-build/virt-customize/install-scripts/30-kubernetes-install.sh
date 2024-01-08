#!/usr/bin/env bash
set -eEuo pipefail

KUBERNETES_VERSION_MINOR=${CONFIG__KUBERNETES_VERSION%.*}

# https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/#installing-kubeadm-kubelet-and-kubectl

# step 1
apt-get install -y apt-transport-https ca-certificates curl gpg

# step 2
curl -fsSL https://pkgs.k8s.io/core:/stable:/v${KUBERNETES_VERSION_MINOR}/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

# step 3
echo "deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v${KUBERNETES_VERSION_MINOR}/deb/ /" | sudo tee /etc/apt/sources.list.d/kubernetes.list

# step 4
# ${CONFIG__KUBERNETES_VERSION} should be set by install.sh script
apt-get update
apt-get -yq install \
    kubelet="${CONFIG__KUBERNETES_VERSION}-1.1" \
    kubeadm="${CONFIG__KUBERNETES_VERSION}-1.1" \
    kubectl="${CONFIG__KUBERNETES_VERSION}-1.1"
apt-mark hold kubelet kubeadm kubectl

apt-get clean -yq
