#!/usr/bin/env bash
set -eEuo pipefail

export PATH="/root/.krew/bin:$PATH"

has_context() {
    kubectl ctx 2>/dev/null | grep -Fxq "$1"
}

find_kubeconfig() {
    local name="$1"
    [[ -d /lab ]] || return 1
    find /lab -name "${name}.kubeconfig" -print -quit
}

while true; do
    if ! has_context target; then
        kubeconfig=$(find_kubeconfig target || true)
        if [[ -n "${kubeconfig}" ]]; then
            kubectl konfig import --save "${kubeconfig}"
            kubectl ctx target=root-admin@root
        fi
    fi

    if ! has_context bootstrap; then
        kubeconfig=$(find_kubeconfig bootstrap || true)
        if [[ -n "${kubeconfig}" ]]; then
            kubectl konfig import --save "${kubeconfig}"
            kubectl ctx bootstrap=kubernetes-admin@kubernetes
        fi
    fi
    sleep 1
done
