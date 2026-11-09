#!/usr/bin/env bash
# Watches a debug-enabled ownframe window for 30s: captures heap profiles
# before/after, a 30s CPU profile, and one /debug/state snapshot per second
# as JSON lines. Run the app with -tags ownframe_debug and -debug-addr first,
# e.g. go run -tags ownframe_debug ./examples/builder/inventory -dev -perf -debug-addr 127.0.0.1:8129
#
# Usage: scripts/watch-pprof.sh [base-url] [output-dir]
#   defaults: http://127.0.0.1:8129  /tmp/opencode
set -euo pipefail
base=${1:-http://127.0.0.1:8129}
out=${2:-/tmp/opencode}
mkdir -p "$out"
curl -s "$base/debug/pprof/heap" -o "$out/heap-before-30s.pb.gz"
curl -s "$base/debug/pprof/profile?seconds=30" -o "$out/cpu-30s.pb.gz" &
cpu=$!
: > "$out/watch.jsonl"
for i in $(seq 1 31); do
  curl -s "$base/debug/state" -o "$out/state-watch-$i.json"
  python3 - "$out/state-watch-$i.json" "$i" >> "$out/watch.jsonl" <<'PY'
import json, sys
d=json.load(open(sys.argv[1]))
r=d["runtime"]
s=d["stats"]
f=d["frame"]
print(json.dumps({
    "t": int(sys.argv[2]),
    "gen": d["generation"],
    "redraws": s["Redraws"],
    "parses": s["Parses"],
    "layouts": s["Layouts"],
    "repaints": s["Repaints"],
    "alloc_frame": s["AllocFrame"],
    "last_redraw": s["LastRedraw"],
    "heap_alloc": r["heap_alloc"],
    "heap_sys": r["heap_sys"],
    "rss": r["rss_bytes"],
    "total_alloc": r["total_alloc"],
    "num_gc": r["num_gc"],
    "alloc_per_sec": r["alloc_per_sec"],
    "fps": f["present_fps"],
}))
PY
  sleep 1
done
wait "$cpu"
curl -s "$base/debug/pprof/heap" -o "$out/heap-after-30s.pb.gz"
echo "watch complete"
