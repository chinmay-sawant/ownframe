#!/bin/sh
# Repeated display-free microbenchmarks and application benchmarks.
# Usage: scripts/perf-baseline.sh [--out FILE] [--baseline REPORT.json]
set -eu
root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
exec python3 "$root/scripts/perf_baseline.py" "$@"
