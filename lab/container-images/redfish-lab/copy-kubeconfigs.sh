#!/usr/bin/env bash
set -eEuo pipefail

export PATH="/root/.krew/bin:$PATH"

while true; do
    if ! kubectl ctx | grep target; then
        kubeconfig=$(find /lab -name target.kubeconfig)
        if [[ ! -z "${kubeconfig}" ]]; then
            kubectl konfig import --save $kubeconfig
            kubectl ctx target=root-admin@root
        fi
    fi

    if ! kubectl ctx | grep bootstrap; then
        kubeconfig=$(find /lab -name bootstrap.kubeconfig)
        if [[ ! -z "${kubeconfig}" ]]; then
            kubectl konfig import --save $kubeconfig
            kubectl ctx bootstrap=kubernetes-admin@kubernetes
        fi
    fi
    sleep 1
done
