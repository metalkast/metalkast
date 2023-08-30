#!/usr/bin/env bash
set -eEuo pipefail

NAME=eth$1
IP=$2

function check_exit_code() {
    "$@"
    local exit_code=$?
    if [ $exit_code -ne 2 ] && [ $exit_code -ne 0 ]; then
        exit $exit_code
    fi
}

check_exit_code ip link add "${NAME}" type dummy
check_exit_code ip addr add "${IP}/32" brd + dev "${NAME}" label "${NAME}:0"
ip link set dev "${NAME}" up

sleep inf
