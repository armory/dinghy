#!/bin/bash -xe
set -e
cd "$(dirname "$0")/.."

yarn

for file in diagrams/*.mmd; do
  ./node_modules/.bin/mmdc -i $file # -o $file.png
done
