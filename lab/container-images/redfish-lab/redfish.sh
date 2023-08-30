#!/usr/bin/env bash
set -eEuo pipefail

export NODE_INDEX="$1"
export IP=192.168.122.10$NODE_INDEX
export HOSTNAME=k8s-node-$NODE_INDEX
export NODE_UUID="$(virsh domuuid k8s-node-$NODE_INDEX)"
if [ -z "$NODE_UUID" ]; then
    exit 1
fi

CONFIG_DIR=/etc/sushy
mkdir -p $CONFIG_DIR
GENERATED_CONFIG_FILE="$CONFIG_DIR/sushy-k8s-node-${NODE_INDEX}.conf"

export CERTIFICATE_FILE="$CONFIG_DIR/sushy-k8s-node-${NODE_INDEX}.crt"
export KEY_FILE="$CONFIG_DIR/sushy-k8s-node-${NODE_INDEX}.key"
if [ ! -f "$CERTIFICATE_FILE" ] || [ ! -f "$KEY_FILE" ]; then
    openssl req \
        -new \
        -newkey ec \
        -pkeyopt ec_paramgen_curve:prime256v1 \
        -days 365 \
        -nodes \
        -x509 \
        -subj "/C=US/ST=None/L=None/O=None/CN=$IP" \
        -keyout $KEY_FILE \
        -out $CERTIFICATE_FILE
fi

envsubst < /opt/config/sushy.conf.tmpl > "$GENERATED_CONFIG_FILE"
exec /usr/local/bin/sushy-emulator --config "$GENERATED_CONFIG_FILE"
