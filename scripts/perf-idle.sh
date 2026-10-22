#!/bin/sh
# Desktop idle checklist, or strict sampling of an existing web perf endpoint.
# Usage: scripts/perf-idle.sh [--example DIR] [--web] [--seconds N] [--url URL]
set -eu
root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$root"
example=examples/perf-benchmarks/normal
web=0
seconds=60
url=http://127.0.0.1:8128/debug/state
while [ "$#" -gt 0 ]; do
  case "$1" in
    --example|--seconds|--url)
      [ "$#" -ge 2 ] || { echo "missing value for $1" >&2; exit 1; }
      case "$1" in
        --example) example=$2 ;;
        --seconds) seconds=$2 ;;
        --url) url=$2 ;;
      esac
      shift 2 ;;
    --web) web=1; shift ;;
    -h|--help) sed -n '2,3p' "$0"; exit 0 ;;
    *) echo "unknown flag $1" >&2; exit 1 ;;
  esac
done
case "$seconds" in ''|*[!0-9]*) echo "seconds must be a positive integer" >&2; exit 1 ;; esac
[ "$seconds" -gt 0 ] || exit 1
[ -d "$example" ] || { echo "missing example: $example" >&2; exit 1; }
if [ "$web" -eq 1 ]; then
  minutes=$(python3 -c 'import sys; print(int(sys.argv[1])/60)' "$seconds")
  exec python3 scripts/perf-sample.py --minutes "$minutes" --interval 5 --url "$url" --out temp/perf-idle.csv
fi
printf 'Manual desktop checklist. This command does not launch or measure a process.
'
printf 'Build the named app once and run that binary, so the PID belongs to the window:
'
printf '  go build -o temp/idle-window ./%s
' "$example"
printf '  temp/idle-window &
  window_pid=$!
'
printf 'Leave the window untouched for %s seconds after warmup.
' "$seconds"
printf 'Sample the same process before and after: ps -o pid,pcpu,rss,etime -p "$window_pid"
'
printf 'Record CPU, RSS, elapsed time, machine, and source commit. No idle verdict is inferred here.
'
