#!/usr/bin/env bash
set -eEuo pipefail

# https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/#installing-kubeadm-kubelet-and-kubectl

# step 1
apt-get install -y apt-transport-https ca-certificates curl

# step 2
# curl -fsSLo /etc/apt/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
# hotfix: https://github.com/kubernetes/release/issues/2862#issuecomment-1554211504
curl -fsSLo /etc/apt/keyrings/kubernetes-archive-keyring.gpg https://dl.k8s.io/apt/doc/apt-key.gpg

# step 3
echo "deb [signed-by=/etc/apt/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | tee /etc/apt/sources.list.d/kubernetes.list

# step 4
# ${CONFIG__KUBERNETES_VERSION} should be set by install.sh script
apt-get update
apt-get -yq install \
    kubelet="${CONFIG__KUBERNETES_VERSION}-00" \
    kubeadm="${CONFIG__KUBERNETES_VERSION}-00" \
    kubectl="${CONFIG__KUBERNETES_VERSION}-00"
apt-mark hold kubelet kubeadm kubectl
