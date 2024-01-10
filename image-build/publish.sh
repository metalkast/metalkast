#!/usr/bin/env bash
set -eEuo pipefail

rclone copy -P output/ metalkast:node-images/ --exclude "**/md5sum.txt"
