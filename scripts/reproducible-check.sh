#!/bin/sh
# Compare complete unsigned release archives from two independent builds.
set -eu
root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$root"
staging=$(mktemp -d)
trap 'rm -rf "$staging"' EXIT
SOURCE_DATE_EPOCH=${SOURCE_DATE_EPOCH:-$(git log -1 --format=%ct)}
export SOURCE_DATE_EPOCH
for run in first second; do
    OWNFRAME_DIST_DIR="$staging/$run" CGO_ENABLED=0 COSIGN=0 sh scripts/package.sh login
    python3 scripts/check-release-bundle.py "$staging/$run"
done
cmp "$staging/first/SHA256SUMS" "$staging/second/SHA256SUMS"
echo 'reproducible: complete release archive checksums match'
