#!/usr/bin/env bash
# Run the declared version rather than accepting an arbitrary PATH binary.
set -euo pipefail
cd -- "$(dirname -- "$0")/.."
version=${1:?expected golangci-lint version}
shift
# v1 and v2 disagree on the machine-readable flag (`version --format json`
# against `version --json`), so read the plain line both print.
lint_version() {
    "$1" version | python3 -c '
import re
import sys
match = re.search(r"\bversion\s+(\S+)", sys.stdin.read())
if not match:
    raise SystemExit("lint: could not read golangci-lint version")
print(match.group(1))
'
}
runner=$(command -v golangci-lint || true)
if [ -n "$runner" ]; then
    actual=$(lint_version "$runner")
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
actual=$(lint_version "$runner")
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
