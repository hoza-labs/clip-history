#!/usr/bin/env bash
set -euo pipefail

# this script rebuilds the development CLI, stopping a locked background worker when needed,
# and then runs the new binary with all arguments passed to this script.

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"

cd "${repo_root}"

binary="${repo_root}/dist/ch.exe"
if [[ -f "${binary}" ]] && ! rm -f -- "${binary}"; then
  echo "debug: existing ch.exe is in use; stopping it before rebuilding" >&2
  "${binary}" stop || true

  if ! rm -f -- "${binary}"; then
    echo "error: could not remove ${binary}; stop any running ch.exe process and try again" >&2
    exit 1
  fi
fi

"${script_dir}/local-build.sh"

echo ""
set -x
exec "${binary}" "$@"
