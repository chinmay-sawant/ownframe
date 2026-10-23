#!/usr/bin/env bash
# Split the examples race suite across independent CI jobs.
set -euo pipefail
root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$root"
shard=${1:-0}
shards=${2:-1}
if ! [[ "$shard" =~ ^[0-9]+$ && "$shards" =~ ^[1-9][0-9]*$ ]] || (( shard >= shards )); then
    echo 'usage: test-examples-race.sh SHARD TOTAL_SHARDS' >&2
    exit 2
fi
listed=$(go list ./examples/...)
mapfile -t packages <<< "$listed"
selected=()
for index in "${!packages[@]}"; do
    if (( index % shards == shard )); then
        selected+=("${packages[index]}")
    fi
done
if (( ${#selected[@]} == 0 )); then
    echo 'race shard has no packages' >&2
    exit 1
fi
go test -race -p 1 "${selected[@]}"
