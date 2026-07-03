#!/usr/bin/env bash
set -euo pipefail

# this script targets ./dist for the binary(ies) to be created

set -x
mkdir -p dist
go test ./...
GOOS=windows GOARCH=amd64 go build -o dist/ch.exe ./cmd/ch
set +x

echo "success: local build completed successfully"
