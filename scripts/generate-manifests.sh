#!/bin/bash
set -euo pipefail

find $(git rev-parse --show-toplevel) -name "*yaml.generate.ts" | xargs -P$(nproc) -I_f bash -c 'f=_f; o=${f%.generate.ts}; echo "Generating $o"; ts-node $f > $o'
