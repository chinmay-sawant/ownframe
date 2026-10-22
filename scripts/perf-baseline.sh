#!/bin/sh
# Baseline benchmarks: page stages, replay scan, platform click.
# Usage: scripts/perf-baseline.sh [--out FILE]
# Display-free: go test only, never opens a window.
set -eu
OUT=/tmp/opencode/perf-baseline-$(date +%F).txt
while [ $# -gt 0 ]; do
  case "$1" in
    --out) OUT="$2"; shift 2;;
    -h|--help) sed -n '2,4p' "$0"; exit 0;;
    *) echo "unknown flag $1" >&2; exit 1;;
  esac
done
mkdir -p "$(dirname "$OUT")"
# Avoid a tee pipeline swallowing a failing benchmark command.
if (
echo "# ownframe perf baseline $(date -u +%FT%TZ)"
echo "# $(go version)"
go test -p 1 ./internal/page -run XXX -bench 'BenchmarkRedrawStages|BenchmarkRedrawWarm|BenchmarkRedrawCold|BenchmarkRedrawHeavyWarm' -benchmem -count 1 || exit $?
go test -p 1 ./internal/replay -run XXX -bench BenchmarkDrawRectScan -benchmem -count 1 || exit $?
go test -p 1 ./examples/platform/... -run XXX -bench BenchmarkClickCount -benchmem -count 1
) >"$OUT" 2>&1; then
  cat "$OUT"
  echo "Wrote $OUT"
else
  status=$?
  cat "$OUT" >&2
  exit "$status"
fi
