# this dockerfile generates a SCRATCH image containing _only_ the built binaries

FROM golang:1.26-alpine AS builder
RUN apk add --no-cache bash

WORKDIR /src
COPY . .

RUN <<-EOF
    echo "do local build"
    set -eu
    set -x
    ls -la
    echo "dist dir before:"
    ls -la ./dist/
    ./scripts/local-build.sh
    echo "dist dir after:"
    ls -la ./dist/
    set +x
    echo "local build completed"
EOF

RUN <<-EOF
    echo "I am here now"
    ls -la ./dist/
EOF

FROM scratch
COPY --from=builder /src/dist/ch.exe ch.exe

# ENTRYPOINT ["/usr/local/bin/ch"]
