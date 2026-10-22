#!/bin/sh
# Prove that release binaries are reproducible: build the login example
# twice with fixed inputs and fail when the bytes differ.
#
# Both builds use -trimpath, an explicit -buildvcs flag, and the same
# SOURCE_DATE_EPOCH, so only the toolchain and the source affect the
# output. A mismatch means the build embeds time, paths, or randomness
# and must be fixed before tagging. See documentation/supply-chain.md.
set -eu

root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$root"

first=$(mktemp)
second=$(mktemp)
trap 'rm -f "$first" "$second"' EXIT

epoch=${SOURCE_DATE_EPOCH:-0}
export SOURCE_DATE_EPOCH="$epoch"

build() {
	out=$1
	env CGO_ENABLED=0 \
		go build -trimpath -buildvcs=true -ldflags "-s -w" -o "$out" ./examples/login
}

build "$first"
build "$second"

if cmp -s "$first" "$second"; then
	echo "reproducible: identical bytes across two builds"
else
	echo "reproducible-check: builds differ" >&2
	exit 1
fi
