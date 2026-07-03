#!/usr/bin/env bash
set -euo pipefail

# this script targets ./dist for the binary(ies) to be created

version_file="VERSION"

if ! command -v go >/dev/null 2>&1; then
  echo "error: go is required" >&2
  exit 1
fi

if [[ ! -f "${version_file}" ]]; then
  echo "error: ${version_file} not found" >&2
  exit 1
fi

version="$(tr -d '[:space:]' < "${version_file}")"
tag="v${version}"

if [[ ! "${version}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "error: version must be semver without prefixes, got ${version}" >&2
  exit 1
fi

timestamp="$(date -u +%Y%m%d-%H%M%SZ)"

set -x
mkdir -p dist
go test ./...
GOOS=windows GOARCH=amd64 go build -o dist/ch.exe \
    -ldflags "-X github.com/hoza-labs/clip-history/internal/version.buildVersion=${version}-rc-${timestamp}" \
    ./cmd/ch
set +x

echo "success: local build completed successfully"
