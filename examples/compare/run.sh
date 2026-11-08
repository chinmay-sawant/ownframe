#!/bin/sh
# Apples-to-apples data gathering for examples/compare.
#
# Runs the headless benches and the synthetic mouse down/up dummies of both
# toolkits and stores the outputs side by side. No display needed.
#
# Usage: sh examples/compare/run.sh [out-dir]
# Default out-dir: temp/compare-dump/run-<utc-timestamp>
set -eu

ROOT=$(dirname "$0")/../..
STARTDIR=$(pwd)
cd "$ROOT"

STAMP=$(date -u +%Y%m%dT%H%M%SZ)
OUT=${1:-temp/compare-dump/run-$STAMP}
case "$OUT" in
/*) OUTABS="$OUT" ;;
*) OUTABS="$STARTDIR/$OUT" ;;
esac
mkdir -p "$OUT"

echo "out: $OUT"

echo "== ownframe -bench =="
go run ./examples/compare/ownframe -bench -bench-file "$OUT/ownframe-bench.txt"

echo "== ownframe -auto-click =="
go run ./examples/compare/ownframe -auto-click -auto-clicks 100 > "$OUT/ownframe-auto-click.txt"
cat "$OUT/ownframe-auto-click.txt"

echo "== ownframe -animate-frames =="
go run ./examples/compare/ownframe -animate-frames 120 > "$OUT/ownframe-animate.txt"
cat "$OUT/ownframe-animate.txt"

echo "== mygo -bench =="
if (cd examples/compare/mygo && GOWORK=off GOTOOLCHAIN=auto go run . -bench -bench-file "$OUTABS/mygo-bench.txt"); then
  :
else
  echo "WARN: mygo bench failed (needs Go >= 1.27.3); skipping" >&2
fi

echo "== mygo -auto-click =="
if (cd examples/compare/mygo && GOWORK=off GOTOOLCHAIN=auto go run . -auto-click -auto-clicks 100 > "$OUTABS/mygo-auto-click.txt"); then
  cat "$OUTABS/mygo-auto-click.txt"
else
  echo "WARN: mygo auto-click failed; skipping" >&2
fi

echo "== mygo -animate-frames =="
if (cd examples/compare/mygo && GOWORK=off GOTOOLCHAIN=auto go run . -animate-frames 120 > "$OUTABS/mygo-animate.txt"); then
  cat "$OUTABS/mygo-animate.txt"
else
  echo "WARN: mygo animate-frames failed; skipping" >&2
fi

{
  echo "# compare $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo ""
  echo "## ownframe bench"
  cat "$OUT/ownframe-bench.txt"
  echo ""
  echo "## mygo bench"
  if [ -f "$OUT/mygo-bench.txt" ]; then
    cat "$OUT/mygo-bench.txt"
  else
    echo "n/a (mygo toolchain missing)"
  fi
  echo ""
  echo "## ownframe auto-click"
  cat "$OUT/ownframe-auto-click.txt"
  echo ""
  echo "## ownframe animate-frames"
  cat "$OUT/ownframe-animate.txt"
  echo ""
  echo "## mygo auto-click"
  if [ -f "$OUT/mygo-auto-click.txt" ]; then
    cat "$OUT/mygo-auto-click.txt"
  else
    echo "n/a (mygo toolchain missing)"
  fi
  echo ""
  echo "## mygo animate-frames"
  if [ -f "$OUT/mygo-animate.txt" ]; then
    cat "$OUT/mygo-animate.txt"
  else
    echo "n/a (mygo toolchain missing)"
  fi
} > "$OUT/comparison.txt"

echo "wrote $OUT/comparison.txt"
