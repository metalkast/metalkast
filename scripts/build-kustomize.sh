#!/bin/bash
set -euo pipefail

root_directory="$(git rev-parse --show-toplevel)"
kustomize_version=$(grep 'sigs.k8s.io/kustomize/kustomize/v5' $root_directory/go.mod | awk '{print $2}')
GOBIN=${root_directory}/bin GO111MODULE=on go install sigs.k8s.io/kustomize/kustomize/v5@$kustomize_version

grep -rl --include="*.yaml" "kind: Kustomization" "$root_directory" |
xargs -L1 -I_k -P$(nproc) bash -c '
    dir=$(dirname _k)
    echo "Running kustomize build for layer: $dir"

    if ! err=$(kustomize build --enable-helm $dir 2>&1 > /dev/null); then
        echo "Build failed for kustomize layer: $dir"
        echo "$err"
    fi
'
