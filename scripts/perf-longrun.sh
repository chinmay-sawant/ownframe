#!/bin/sh
# Sample a running debug endpoint. For automated page interactions use make soak-check.
set -eu
root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
exec python3 "$root/scripts/perf-sample.py" "$@"
