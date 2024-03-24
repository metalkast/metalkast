#!/usr/bin/env bash
set -eEuo pipefail

git_root_directory="$(git rev-parse --show-toplevel)"
kustomize_version=$(grep 'sigs.k8s.io/kustomize/kustomize/v5' $git_root_directory/go.mod | awk '{print $2}')
GOBIN=${git_root_directory}/bin GO111MODULE=on go install sigs.k8s.io/kustomize/kustomize/v5@$kustomize_version

npm install > /dev/null 2>&1 && npm run generate-manifests

exit_code=0
for k in $(grep -rl --include="*.yaml" "kind: Kustomization" "$git_root_directory"); do
    dir=$(dirname $k)
    relative_path=$(realpath $dir -s --relative-to=$git_root_directory)
    echo "Running kustomize build for layer: $relative_path"

    if ! err_output=$(kustomize build --enable-helm $dir 2>&1 > /dev/null); then
        exit_code=1
        echo "$(tput setaf 1)Build failed for kustomize layer: $relative_path$(tput sgr 0)"
        echo $err_output |
            sed -E "s#'([^ ]+)'#\1#g" |
            sed "s#$git_root_directory/##g" |
            sed -E "s#'[^']+'|:[^:']+#\n- \0#g" |
            sed -E 's#^- : #- #' |
            tr -d "'" |
            sed 's#^#  #g'
    fi
done

exit $exit_code
