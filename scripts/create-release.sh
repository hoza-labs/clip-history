#!/usr/bin/env bash
set -euo pipefail

version_file="VERSION"

if ! command -v gh >/dev/null 2>&1; then
  echo "error: gh CLI is required" >&2
  exit 1
fi

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

if ! git diff --quiet || ! git diff --cached --quiet; then
  echo "error: working tree must be clean before creating a release" >&2
  exit 1
fi

if git rev-parse "${tag}" >/dev/null 2>&1; then
  echo "error: local tag ${tag} already exists" >&2
  exit 1
fi

if gh release view "${tag}" >/dev/null 2>&1; then
  echo "error: github release ${tag} already exists" >&2
  exit 1
fi

mkdir -p dist
go test ./...
GOOS=windows GOARCH=amd64 go build \
  -ldflags "-X github.com/hoza-labs/clip-history/internal/version.buildVersion=${version}" \
  -o "dist/ch.exe" \
  ./cmd/ch

gh release create "${tag}" "dist/ch.exe" \
  --title "${tag}" \
  --generate-notes \
  --latest

echo "success: created release ${tag}"
