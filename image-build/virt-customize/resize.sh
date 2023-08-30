#!/bin/sh

set -e

# https://askubuntu.com/a/937351
growpart /dev/sda 1
resize2fs /dev/sda1
