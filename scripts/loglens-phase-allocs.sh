#!/usr/bin/env bash
set -euo pipefail

if (($# != 1)); then
  echo "usage: $0 CAPTURE_DIRECTORY" >&2
  exit 2
fi

capture_dir=$1
if [[ ! -d $capture_dir ]]; then
  echo "capture directory not found: $capture_dir" >&2
  exit 2
fi

found=0
for phase_dir in "$capture_dir"/run-*/{hover,resize}; do
  [[ -d $phase_dir ]] || continue
  before=$phase_dir/allocs-before.pb.gz
  after=$phase_dir/allocs-after.pb.gz
  if [[ ! -f $before || ! -f $after ]]; then
    echo "missing allocation profile pair: $phase_dir" >&2
    exit 1
  fi
  found=1
  printf '\n=== %s ===\n' "$phase_dir"
  go tool pprof -top -nodecount=25 -base "$before" "$after"
done

if ((found == 0)); then
  echo "no hover or resize allocation profiles found under $capture_dir" >&2
  exit 1
fi
