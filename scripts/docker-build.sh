#!/usr/bin/env bash
set -euo pipefail

# this script builds the binary(ies) in an isolated docker container,
# sending them to the `./dist` folder.
# this script does not create or update a docker image of any kind.

set -x
mkdir -p dist
if command -v dos2unix &> /dev/null; then
    echo "ensuring Dockerfile uses unix line endings..."
    dos2unix Dockerfile
    echo "...done"
fi
docker buildx build . -f Dockerfile -o ./dist --no-cache --progress=plain
