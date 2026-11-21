#!/bin/sh
# Full CSS property performance dump: 2-pass paired-control 688-property matrix,
# Tier-2 combinations, benchscore output, summary.md, and top-offender pprof.
# Usage: scripts/perf-props-dump.sh [--out DIR] [--iters N] [--refine-top R] [--top-pairs M] [--fuzz K] [--pprof-top P] [--filter REGEX] [--limit N] [--reference DIR]
# --reference DIR reuses the cases and refinement choices from a version 2 dump.
# Use the same sampling and fixture flags in the reference and candidate.
set -eu

ROOT=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
STAMP=$(date +%Y%m%d-%H%M%S)
OUT="$ROOT/temp/perf-dumps/props-$STAMP"
ITERS=5
REFINE_TOP=40
TOP_PAIRS=10
FUZZ=25
PPROF_TOP=9
FILTER=""
LIMIT=0
REFERENCE=""

while [ $# -gt 0 ]; do
  case "$1" in
    --out) OUT="$2"; shift 2;;
    --iters) ITERS="$2"; shift 2;;
    --refine-top) REFINE_TOP="$2"; shift 2;;
    --top-pairs) TOP_PAIRS="$2"; shift 2;;
    --fuzz) FUZZ="$2"; shift 2;;
    --pprof-top) PPROF_TOP="$2"; shift 2;;
    --filter) FILTER="$2"; shift 2;;
    --limit) LIMIT="$2"; shift 2;;
    --reference) REFERENCE="$2"; shift 2;;
    -h|--help) sed -n '2,6p' "$0"; exit 0;;
    *) echo "unknown flag $1" >&2; exit 1;;
  esac
done

case "$OUT" in
  /*) ;;
  *) OUT="$(pwd)/$OUT" ;;
esac
case "$REFERENCE" in
  ""|/*) ;;
  *) REFERENCE="$(pwd)/$REFERENCE" ;;
esac
mkdir -p "$OUT"
echo "Running CSS property benchmark dump into $OUT ..."

OWNFRAME_PROP_DUMP_DIR="$OUT" \
OWNFRAME_PROP_ITERS="$ITERS" \
OWNFRAME_PROP_REFINE_TOP="$REFINE_TOP" \
OWNFRAME_PROP_TOP_PAIRS="$TOP_PAIRS" \
OWNFRAME_PROP_FUZZ="$FUZZ" \
OWNFRAME_PROP_PPROF_TOP="$PPROF_TOP" \
OWNFRAME_PROP_FILTER="$FILTER" \
OWNFRAME_PROP_LIMIT="$LIMIT" \
OWNFRAME_PROP_REFERENCE_DIR="$REFERENCE" \
  go test -v -count=1 -timeout=10m ./examples/blinkless-all -run '^TestPropPerfDump$'

echo ""
echo "Dump complete:"
echo "  Summary:      $OUT/summary.md"
echo "  Individual:   $OUT/props-individual.jsonl"
echo "  Combinations: $OUT/props-combos.jsonl"
echo "  Benchscore:   $OUT/props-bench.txt"
echo "  Profiles:     $OUT/profiles/"
