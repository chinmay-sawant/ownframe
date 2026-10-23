#!/usr/bin/env bash
# Run the declared version rather than accepting an arbitrary PATH binary.
set -euo pipefail
cd -- "$(dirname -- "$0")/.."
version=${1:?expected golangci-lint version}
shift
runner=$(command -v golangci-lint || true)
if [ -n "$runner" ]; then
    actual=$("$runner" version --json | python3 -c 'import json,sys; print(json.load(sys.stdin)["version"])')
else
    actual=
fi
if [ "${actual#v}" != "${version#v}" ]; then
    mkdir -p temp/tools
    runner="$PWD/temp/tools/golangci-lint-${version}"
    if [ ! -x "$runner" ]; then
        install_dir=$(mktemp -d "$PWD/temp/tools/install.XXXXXX")
        trap 'rm -rf -- "$install_dir"' EXIT
        GOBIN="$install_dir" go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$version"
        mv -- "$install_dir/golangci-lint" "$runner"
    fi
fi
actual=$("$runner" version --json | python3 -c 'import json,sys; print(json.load(sys.stdin)["version"])')
if [ "${actual#v}" != "${version#v}" ]; then
    echo "lint: expected $version, got $actual from $runner" >&2
    exit 1
fi
"$runner" version
if [ "${1:-}" = --version ]; then
    exit 0
fi
status=0
"$runner" run ./... || status=1
"$runner" run ./examples/... || status=1
exit "$status"
