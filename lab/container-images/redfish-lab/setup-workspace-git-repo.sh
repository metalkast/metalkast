#!/bin/bash
set -euo pipefail

mkdir -p /srv/git
git init --bare /srv/git/metalkast.git
(cd /srv/git/metalkast.git && git symbolic-ref HEAD refs/heads/main)

mkdir -p /tmp/metalkast-repo
cd /tmp/metalkast-repo
cp -r /manifests .
git config --global user.email 'lab@metalkast.io'
git config --global user.name 'Metalkast Lab'
git init
git add .
git commit -m 'commit'
git branch -M main
git remote add origin file:///srv/git/metalkast.git
git push -u origin main
