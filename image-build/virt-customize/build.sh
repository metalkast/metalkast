#!/usr/bin/env bash
set -eEuo pipefail

VERSION=k8s-v$KUBERNETES_VERSION-ubuntu-$UBUNTU_VERSION-$UBUNTU_RELEASE-amd64-$METALKAST_VERSION
OUTPUT_DIR=output/$VERSION

printenv > printenv.txt
BUILD_ENVIRONMENT_VERSION_FILE=shasum.txt
find -type f -not \( -path "./output/*" -o -name $BUILD_ENVIRONMENT_VERSION_FILE \) | sort | xargs -L1 shasum -a 256 | tee $BUILD_ENVIRONMENT_VERSION_FILE
if cmp -s "$BUILD_ENVIRONMENT_VERSION_FILE" "$OUTPUT_DIR/$BUILD_ENVIRONMENT_VERSION_FILE"; then
    echo "Build completed (cached)"
    exit 0
fi
rm -rf output/*
mkdir -p $OUTPUT_DIR

echo KUBERNETES_VERSION=$(printenv KUBERNETES_VERSION) >> install-scripts/.env

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
  name: metalkast.io/cluster-version
  annotations:
    config.kubernetes.io/local-config: "true"
data:
  k8s_version: v${KUBERNETES_VERSION}
  node_image_url: https://dl.metalkast.io/node-images/${VERSION}/cluster-node.img
  node_image_checksum: ${checksum}
EOF

./bootstrap-build.sh

# Mark build as finished and enable caching
cp $BUILD_ENVIRONMENT_VERSION_FILE $OUTPUT_DIR/$BUILD_ENVIRONMENT_VERSION_FILE
