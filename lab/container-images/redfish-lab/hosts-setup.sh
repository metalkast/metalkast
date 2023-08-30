#!/usr/bin/env bash
set -eEuo pipefail

if ! virsh net-info lab; then
    virsh net-define --file /opt/config/lab-network.xml
fi

NODE_COUNT=3

# shellcheck disable=SC2086
for NODE_INDEX in $(seq 1 $NODE_COUNT); do
    NODE_NAME=k8s-node-$NODE_INDEX
    if ! virsh dumpxml "$NODE_NAME"; then
        tmpfile=$(mktemp /tmp/sushy-domain.XXXXXX)
        virt-install \
            --name "$NODE_NAME" \
            --ram 8192 \
            --disk size=15 \
            --vcpus 2 \
            --os-type linux \
            --os-variant ubuntu22.04 \
            --graphics vnc \
            --network=network=lab,bridge=virbr1,mac="52:54:00:6c:3c:0${NODE_INDEX}" \
            --boot loader.readonly=yes \
            --boot loader.type=pflash \
            --boot loader.secure=no \
            --boot loader=/usr/share/OVMF/OVMF_CODE.secboot.fd \
            --boot nvram.template=/usr/share/OVMF/OVMF_VARS.fd \
            --print-xml > "$tmpfile"
        virsh define --file "$tmpfile"
        rm "$tmpfile"
    fi
done

if ! ps -C vbmcd; then
    rm -f /root/.vbmc/master.pid
    vbmcd
fi

if [ ! -d /root/.vbmc/k8s-node-1 ]; then
    # there's some performance issue when binding to specific address, ideally we want to bind to 192.168.122.101 here
    vbmc add k8s-node-1 --address 0.0.0.0
    vbmc start k8s-node-1
fi

if [ "$(virsh net-info lab | grep Active | awk '{print $2}')" == "no" ]; then
    virsh net-start lab
fi

sleep inf
