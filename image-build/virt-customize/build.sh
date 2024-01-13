#!/usr/bin/env bash
set -eEuo pipefail

VERSION=k8s-v$CONFIG__KUBERNETES_VERSION-ubuntu-$CONFIG__UBUNTU_VERSION-amd64
OUTPUT_DIR=output/$VERSION
mkdir -p $OUTPUT_DIR

BUILD_ENVIRONMENT_VERSION_FILE=md5sum.txt
if cmp -s "$BUILD_ENVIRONMENT_VERSION_FILE" "$OUTPUT_DIR/$BUILD_ENVIRONMENT_VERSION_FILE"; then
    echo "Build completed (cached)"
    exit 0
fi
rm -rf $OUTPUT_DIR/*

printenv | grep -E '^CONFIG__' > install-scripts/.env

ORIGINAL_UBUNTU_IMAGE=ubuntu.img
CUSTOMIZED_UBUNTU_IMAGE=$OUTPUT_DIR/cluster-node.img

# We want to use raw because otherwise ironic is going to expand the img to the size of the disk
# which will result in lots of zero writes and thus slow startup
qemu-img convert -O raw $ORIGINAL_UBUNTU_IMAGE $CUSTOMIZED_UBUNTU_IMAGE
# Give more space to OS to enable installing stuff
qemu-img resize $CUSTOMIZED_UBUNTU_IMAGE +2.5G

# Customize the image
virt-customize -v -x --commands-from-file commands -a "${CUSTOMIZED_UBUNTU_IMAGE}"

# TODO: this didn't work
# # ironic can decompress the image
# gzip $CUSTOMIZED_UBUNTU_IMAGE
# CUSTOMIZED_UBUNTU_IMAGE_COMPRESSED=$CUSTOMIZED_UBUNTU_IMAGE.gz

# generate checksum
CUSTOMIZED_UBUNTU_IMAGE_BASENAME=$(basename $CUSTOMIZED_UBUNTU_IMAGE)
CHECKSUM_FILE=${CUSTOMIZED_UBUNTU_IMAGE_BASENAME}.sha256sum
(cd $OUTPUT_DIR; sha256sum "${CUSTOMIZED_UBUNTU_IMAGE_BASENAME}" > "${CHECKSUM_FILE}")

checksum=$(cat ${OUTPUT_DIR}/${CHECKSUM_FILE} | cut -d' ' -f1)
cat <<EOF > $OUTPUT_DIR/config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8s-cluster-version
  annotations:
    config.kubernetes.io/local-config: "true"
data:
  VERSION: v${CONFIG__KUBERNETES_VERSION}
  NODE_IMAGE_URL: https://dl.metalkast.io/node-images/${VERSION}/cluster-node.img
  NODE_IMAGE_CHECKSUM: ${checksum}
EOF

./bootstrap-build.sh

# Mark build as finished and enable caching
cp $BUILD_ENVIRONMENT_VERSION_FILE $OUTPUT_DIR/$BUILD_ENVIRONMENT_VERSION_FILE
